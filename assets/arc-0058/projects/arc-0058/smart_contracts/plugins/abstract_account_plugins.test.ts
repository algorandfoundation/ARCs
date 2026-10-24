import { describe, test, beforeAll, beforeEach, expect } from '@jest/globals';
import { algorandFixture } from '@algorandfoundation/algokit-utils/testing';
import * as algokit from '@algorandfoundation/algokit-utils';
import algosdk, { encodeUint64, makeBasicAccountTransactionSigner } from 'algosdk';
import { AbstractedAccountClient, AbstractedAccountFactory } from '../artifacts/abstracted_account/AbstractedAccountClient';
import { SubscriptionPluginClient, SubscriptionPluginFactory } from '../artifacts/plugins/subscription/SubscriptionPluginClient';
import { OptInPluginClient, OptInPluginFactory } from '../artifacts/plugins/optin/OptInPluginClient';
import { EscrowFactoryFactory } from '../artifacts/escrow/EscrowFactoryClient';

const ZERO_ADDRESS = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAY5HFKQ';
algokit.Config.configure({ populateAppCallResources: true });
const fixture = algorandFixture();

describe('Abstracted Subscription Program', () => {
  /** Alice's externally owned account (ie. a keypair account she has in Pera) */
  let aliceEOA: algosdk.Account;
  /** The address of Alice's new abstracted account. Sends app calls from aliceEOA unless otherwise specified */
  let aliceAbstractedAccount: string;
  /** The client for Alice's abstracted account */
  let abstractedAccountClient: AbstractedAccountClient;
  /** The client for the subscription plugin */
  let subPluginClient: SubscriptionPluginClient;
  /** The ID of the subscription plugin */
  let subPluginID: bigint; 
  /** The client for the opt-in plugin */
  let optInPluginClient: OptInPluginClient;
  /** The ID of the opt-in plugin */
  let optInPluginID: bigint;
  /** The suggested params for transactions */
  let suggestedParams: algosdk.SuggestedParams;

  /** The maximum uint64 value. Used to indicate a never-expiring plugin */
  const maxUint64 = BigInt('18446744073709551615');

  beforeEach(fixture.beforeEach);

  beforeAll(async () => {
    await fixture.beforeEach();
    const { algorand, testAccount } = fixture.context;
    suggestedParams = await algorand.getSuggestedParams();
    aliceEOA = testAccount;

    const escrowFactory = new EscrowFactoryFactory({
      defaultSender: aliceEOA.addr,
      defaultSigner: makeBasicAccountTransactionSigner(aliceEOA),
      algorand
    })

    const escrowFactoryResults = await escrowFactory.send.create.bare()

    await escrowFactoryResults.appClient.appClient.fundAppAccount({ amount: (100_000).microAlgos() });

    const minter = new AbstractedAccountFactory({
      defaultSender: aliceEOA.addr,
      defaultSigner: makeBasicAccountTransactionSigner(aliceEOA),
      algorand
    });

    const results = await minter.send.create.createApplication({
      args: {
        admin: aliceEOA.addr.toString(),
        controlledAddress: ZERO_ADDRESS,
        escrowFactory: escrowFactoryResults.appClient.appId,
      },
    });
    abstractedAccountClient = results.appClient;
    aliceAbstractedAccount = abstractedAccountClient.appAddress.toString();

    // Fund the abstracted account with 0.2 ALGO so it can hold an ASA
    await abstractedAccountClient.appClient.fundAppAccount({ amount: algokit.microAlgos(200_000) });

    const subPluginMinter = new SubscriptionPluginFactory({
      defaultSender: aliceEOA.addr,
      defaultSigner: makeBasicAccountTransactionSigner(aliceEOA),
      algorand
    });
    
    const subPluginMintResults = await subPluginMinter.send.create.bare({
      deployTimeParams: {
        FREQUENCY: BigInt(1), // 1 round
        AMOUNT: BigInt(100_000) // 100_000 microAlgos
      }
    });
    subPluginClient = subPluginMintResults.appClient;
    subPluginID = subPluginClient.appId;

    const optinPluginMinter = new OptInPluginFactory({
      defaultSender: aliceEOA.addr,
      defaultSigner: makeBasicAccountTransactionSigner(aliceEOA),
      algorand
    });

    const optInMintResults = await optinPluginMinter.send.create.bare();
    optInPluginClient = optInMintResults.appClient;
    optInPluginID = optInPluginClient.appId;
  });

  describe('Unnamed Subscription Plugin', () => {
    /** Another account that the subscription payments will go to */
    const joe = '46XYR7OTRZXISI2TRSBDWPUVQT4ECBWNI7TFWPPS6EKAPJ7W5OBXSNG66M';
    // const joe = ZERO_ADDRESS;
    /** The box key for the subscription plugin */
    let pluginBox: Uint8Array;
    /** The boxes to pass to app calls */
    let boxes: Uint8Array[];

    beforeAll(() => {
      /** The box key for a plugin is `p + plugin ID + allowed caller`  */
      pluginBox = new Uint8Array(
        Buffer.concat([
          Buffer.from('p'),
          Buffer.from(encodeUint64(subPluginID)),
          algosdk.decodeAddress(ZERO_ADDRESS).publicKey,
        ])
      );
      boxes = [pluginBox];
    });

    test('Alice adds the app to the abstracted account', async () => {
      // const { algorand } = fixture.context;

      const mbr = (await abstractedAccountClient.send.mbr({
        args: {
          methodCount: 0,
          plugin: '',
          escrow: '',
          groups: 0n
        }
      })).return

      if (mbr === undefined) {
        throw new Error('MBR is undefined');
      }

      await abstractedAccountClient.appClient.fundAppAccount({
        amount: mbr.plugins.microAlgo()
      });

      await abstractedAccountClient.send.arc58AddPlugin({
        args: {
          // Add the subscription plugin
          plugin: subPluginID,
          // Set address to ZERO_ADDRESS so anyone can call it
          caller: ZERO_ADDRESS,
          // Dont create an escrow for this plugin
          escrow: '',
          // Set adminPrivileges to false so it cant change the admin account
          admin: false,
          // Delegation type of other
          delegationType: 3,
          // Set end to maxUint64 so it never expires
          lastValid: maxUint64,
          // Set cooldown to 0 so it can always be called
          cooldown: 0,
          methods: [],
          useRounds: false,
          useExecutionKey: false,
          defaultToEscrow: false
        },
      });
    });

    test('Someone calls the program to trigger payment', async () => {
      const { algod, testAccount } = fixture.context;

      boxes = [
        new Uint8Array(
          Buffer.concat([
            Buffer.from('p'),
            Buffer.from(encodeUint64(subPluginID)),
            algosdk.decodeAddress(ZERO_ADDRESS).publicKey,
          ])
        ),
      ];

      const alicePreBalance = await algod.accountInformation(aliceAbstractedAccount).do();
      const joePreBalance = await algod.accountInformation(joe).do();

      // Get the call to the subscription plugin
      const makePaymentTxn = (
        (await subPluginClient
          .createTransaction
          .makePayment({
            sender: testAccount.addr,
            // Send a payment from the abstracted account to Joe
            args: {
              wallet: abstractedAccountClient.appId,
              rekeyBack: true,
              _acctRef: joe
            },
            // Double the fee to cover the inner txn fee
            extraFee: (1_000).microAlgos(),
          })
        ).transactions
      )[0];

      // Compose the group needed to actually use the plugin
      await abstractedAccountClient
        .newGroup()
        // Step one: rekey to the plugin
        .arc58RekeyToPlugin({
          sender: testAccount.addr,
          signer: makeBasicAccountTransactionSigner(testAccount),
          args: {
            plugin: subPluginID,
            global: true,
            escrow: '',
            methodOffsets: [],
            fundsRequest: []
          },
          extraFee: (1_000).microAlgos(),
          boxReferences: boxes,
          accountReferences: [aliceAbstractedAccount, joe],
        })
        // Step two: Call the plugin
        .addTransaction(makePaymentTxn, testAccount.signer)
        // Step three: Call verify auth addr to rekey back to the abstracted account
        .arc58VerifyAuthAddress()
        .send();

      // Verify the payment was made
      const alicePostBalance = await algod.accountInformation(aliceAbstractedAccount).do();
      const joePostBalance = await algod.accountInformation(joe).do();
      expect(alicePostBalance.amount).toBe(alicePreBalance.amount - BigInt(100_000));
      expect(joePostBalance.amount).toBe(joePreBalance.amount + BigInt(100_000));
    });
  });

  describe('Named OptIn Plugin', () => {
    let bob: algosdk.Account;
    let asset: bigint;

    const nameBox = new Uint8Array(Buffer.concat([Buffer.from('n'), Buffer.from('optIn')]));

    let pluginBox: Uint8Array;

    const boxes: Uint8Array[] = [nameBox];

    beforeAll(async () => {
      bob = fixture.context.testAccount;
      const { algorand } = fixture.context;

      // Create an asset
      const txn = await algorand.send.assetCreate({
        sender: bob.addr,
        total: BigInt(1),
        decimals: 0,
        defaultFrozen: false,
      });

      asset = BigInt(txn.confirmation!.assetIndex!);

      pluginBox = new Uint8Array(
        Buffer.concat([
          Buffer.from('p'),
          Buffer.from(encodeUint64(optInPluginID)),
          algosdk.decodeAddress(ZERO_ADDRESS).publicKey,
        ])
      );

      boxes.push(pluginBox);
    });

    test('Alice adds the app to the abstracted account', async () => {
      const mbr = (await abstractedAccountClient.send.mbr({
        args: {
          methodCount: 0,
          plugin: 'optIn',
          escrow: '',
          groups: 0n
        }
      })).return

      if (mbr === undefined) {
        throw new Error('MBR is undefined');
      }

      await abstractedAccountClient.appClient.fundAppAccount({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        amount: algokit.microAlgos(Number(mbr.plugins + mbr.namedPlugins))
      });

      // Add opt-in plugin
      await abstractedAccountClient.send.arc58AddNamedPlugin({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          name: 'optIn',
          plugin: optInPluginID,
          caller: ZERO_ADDRESS,
          admin: false,
          delegationType: 3,
          escrow: '',
          lastValid: maxUint64,
          cooldown: 0,
          methods: [],
          useRounds: false,
          useExecutionKey: false,
          defaultToEscrow: false
        }
      });
    });

    test("Bob opts Alice's abstracted account into the asset", async () => {
      // Form a payment from bob to alice's abstracted account to cover the MBR
      const mbrPayment = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
        sender: bob.addr,
        receiver: aliceAbstractedAccount,
        amount: 100_000,
        suggestedParams,
      });

      // Form the group txn needed to call the opt-in plugin
      const optInGroup = (
        await (optInPluginClient
          .createTransaction
          .optInToAsset({
            sender: bob.addr,
            args: {
              wallet: abstractedAccountClient.appId,
              rekeyBack: true,
              assets: [asset],
              mbrPayment
            },
            extraFee: (1_000).microAlgo()
          }))
      ).transactions;

      // Compose the group needed to actually use the plugin
      await abstractedAccountClient
        .newGroup()
        // Rekey to the opt-in plugin
        .arc58RekeyToNamedPlugin({
          sender: bob.addr,
          signer: makeBasicAccountTransactionSigner(bob),
          args: {
            name: 'optIn',
            global: true,
            escrow: '',
            methodOffsets: [],
            fundsRequest: []
          },
          extraFee: (1_000).microAlgo(),
          boxReferences: boxes,
          assetReferences: [asset],
        })
        // Add the mbr payment
        .addTransaction(optInGroup[0], makeBasicAccountTransactionSigner(bob)) // mbrPayment
        // Add the opt-in plugin call
        .addTransaction(optInGroup[1], makeBasicAccountTransactionSigner(bob)) // optInToAsset
        // Call verify auth addr to verify the abstracted account is rekeyed back to itself
        .arc58VerifyAuthAddress()
        .send();
    });
  });
});
