import { describe, test, beforeAll, beforeEach, expect } from '@jest/globals';
import { algorandFixture } from '@algorandfoundation/algokit-utils/testing';
import * as algokit from '@algorandfoundation/algokit-utils';
import algosdk, { makeBasicAccountTransactionSigner } from 'algosdk';
import { ERR_ALLOWANCE_EXCEEDED, ERR_CANNOT_CALL_OTHER_APPS_DURING_REKEY, ERR_MALFORMED_OFFSETS, ERR_METHOD_ON_COOLDOWN, ERR_PLUGIN_DOES_NOT_EXIST, ERR_PLUGIN_EXPIRED, ERR_PLUGIN_ON_COOLDOWN } from './errors';
import { AbstractedAccountClient, AbstractedAccountFactory } from '../artifacts/abstracted_account/AbstractedAccountClient';
import { OptInPluginClient, OptInPluginFactory } from '../artifacts/plugins/optin/OptInPluginClient';
import { EscrowFactoryFactory } from '../artifacts/escrow/EscrowFactoryClient';
import { PayPluginClient, PayPluginFactory } from '../artifacts/plugins/pay/PayPluginClient';

const ZERO_ADDRESS = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAY5HFKQ';

const PluginInfoAbiType = algosdk.ABIType.from('(bool,uint8,uint64,uint64,uint64,(byte[4],uint64,uint64)[],bool,bool,uint64,uint64)')
type PluginInfoTuple = [boolean, bigint, bigint, bigint, bigint, [string, bigint, bigint][], boolean, boolean, number, number]

const EscrowInfoAbiType = algosdk.ABIType.from('uint64');

const AllowanceInfoAbiType = algosdk.ABIType.from('(uint8,uint64,uint64,uint64,uint64,uint64,uint64,bool)');
type AllowanceInfoTuple = [bigint, bigint, bigint, bigint, bigint, bigint, bigint, boolean];


algokit.Config.configure({ populateAppCallResources: true });

describe('ARC58 Plugin Permissions', () => {
  /** Alice's externally owned account (ie. a keypair account she has in Pera) */
  let aliceEOA: algosdk.Account;
  /** The client for Alice's abstracted account */
  let abstractedAccountClient: AbstractedAccountClient;
  /** The client for the dummy plugin */
  let optInPluginClient: OptInPluginClient;
  /** The client for the pay plugin */
  let payPluginClient: PayPluginClient;
  /** The account that will be calling the plugin */
  let caller: algosdk.Account;
  /** optin plugin app id */
  let plugin: bigint;
  /** pay plugin app id */
  let payPlugin: bigint;
  /** The suggested params for transactions */
  let suggestedParams: algosdk.SuggestedParams;
  /** The maximum uint64 value. Used to indicate a never-expiring plugin */
  const MAX_UINT64 = BigInt('18446744073709551615');
  /** a created asset id to use */
  let asset: bigint;
  /** the name of the escrow in use during this test */
  let escrow: string = '';

  const fixture = algorandFixture();

  async function callPayPlugin(
    caller: algosdk.Account,
    payClient: PayPluginClient,
    receiver: string,
    asset: bigint,
    amount: bigint,
    offsets: number[] = [],
    global: boolean = true,
  ) {
    const payPluginTxn = (
      await (payClient
        .createTransaction
        .pay({
          sender: caller.addr,
          signer: makeBasicAccountTransactionSigner(caller),
          args: {
            walletId: abstractedAccountClient.appId,
            rekeyBack: true,
            receiver,
            asset,
            amount
          },
          extraFee: (1_000).microAlgos()
        }))
    ).transactions[0];

    await abstractedAccountClient
      .newGroup()
      .arc58RekeyToPlugin({
        sender: caller.addr,
        signer: makeBasicAccountTransactionSigner(caller),
        args: {
          plugin: payPlugin,
          global,
          methodOffsets: offsets,
          fundsRequest: [[asset, amount]]
        },
        extraFee: (2000).microAlgos()
      })
      .addTransaction(payPluginTxn, makeBasicAccountTransactionSigner(caller))
      .arc58VerifyAuthAddr({
        sender: caller.addr,
        signer: makeBasicAccountTransactionSigner(caller),
        args: {}
      })
      .send();
  }

  async function callOptinPlugin(
    caller: algosdk.Account,
    receiver: string,
    suggestedParams: algosdk.SuggestedParams,
    pluginClient: OptInPluginClient,
    asset: bigint,
    offsets: number[] = [],
    global: boolean = true
  ) {
    const mbrPayment = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
      sender: caller.addr,
      receiver,
      amount: 100_000,
      suggestedParams,
    });

    const optInGroup = (
      await (pluginClient
        .createTransaction
        .optInToAsset({
          sender: caller.addr,
          signer: makeBasicAccountTransactionSigner(caller),
          args: {
            walletId: abstractedAccountClient.appId,
            rekeyBack: true,
            assets: [asset],
            mbrPayment
          },
          extraFee: (1_000).microAlgos()
        }))
    ).transactions;

    await abstractedAccountClient
      .newGroup()
      .arc58RekeyToPlugin({
        sender: caller.addr,
        signer: makeBasicAccountTransactionSigner(caller),
        args: {
          plugin,
          global,
          methodOffsets: offsets,
          fundsRequest: []
        },
        extraFee: (2000).microAlgos()
      })
      // Add the mbr payment
      .addTransaction(optInGroup[0], makeBasicAccountTransactionSigner(caller)) // mbrPayment
      // Add the opt-in plugin call
      .addTransaction(optInGroup[1], makeBasicAccountTransactionSigner(caller)) // optInToAsset
      .arc58VerifyAuthAddr({
        sender: caller.addr,
        signer: makeBasicAccountTransactionSigner(caller),
        args: {}
      })
      .send();
  }

  beforeEach(async () => {
    await fixture.beforeEach();

    const { algorand } = fixture.context;

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

    await abstractedAccountClient.appClient.fundAppAccount({ amount: (100_000).microAlgos() });
  });

  beforeAll(async () => {
    await fixture.beforeEach();
    const { testAccount } = fixture.context;
    const { algorand } = fixture;
    aliceEOA = testAccount;
    caller = algorand.account.random().account;
    const dispenser = await algorand.account.dispenserFromEnvironment();

    suggestedParams = await algorand.getSuggestedParams();

    await algorand.account.ensureFunded(
      aliceEOA.addr,
      dispenser,
      (100).algos(),
    );

    await algorand.account.ensureFunded(
      caller.addr,
      dispenser,
      (100).algos(),
    );

    const optinPluginMinter = new OptInPluginFactory({
      defaultSender: aliceEOA.addr,
      defaultSigner: makeBasicAccountTransactionSigner(aliceEOA),
      algorand
    });
    const optInMintResults = await optinPluginMinter.send.create.bare();

    optInPluginClient = optInMintResults.appClient;
    plugin = optInPluginClient.appId;

    // Create an asset
    const txn = await algorand.send.assetCreate({
      sender: aliceEOA.addr,
      total: BigInt(1_000_000_000_000),
      decimals: 6,
      defaultFrozen: false,
    });

    asset = BigInt(txn.confirmation!.assetIndex!);

    const payPluginMinter = new PayPluginFactory({
      defaultSender: aliceEOA.addr,
      defaultSigner: makeBasicAccountTransactionSigner(aliceEOA),
      algorand
    });

    const payMintResults = await payPluginMinter.send.create.bare();
    payPluginClient = payMintResults.appClient;
    payPlugin = payPluginClient.appId;
  });

  test('both are valid, global is used', async () => {
    const { algorand } = fixture;

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({ args: { methodCount: 0, pluginName: '', escrowName: '' } })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins * BigInt(2)}`)
    const minFundingAmount = mbr.plugins * BigInt(2) // we install plugins twice here so double it

    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, minFundingAmount.microAlgo())

    await abstractedAccountClient.send.arc58AddPlugin({
      args: {
        app: plugin,
        allowedCaller: caller.addr.toString(),
        admin: false,
        delegationType: 3,
        escrow: '',
        lastValid: MAX_UINT64,
        cooldown: 0,
        methods: [],
        useRounds: false
      }
    });

    await abstractedAccountClient.send.arc58AddPlugin({
      args: {
        app: plugin,
        allowedCaller: ZERO_ADDRESS,
        admin: false,
        delegationType: 3,
        escrow: '',
        lastValid: MAX_UINT64,
        cooldown: 1,
        methods: [],
        useRounds: false
      }
    });

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [], true);

    const globalPluginBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('p'),
          Buffer.from(algosdk.encodeUint64(plugin)),
          algosdk.decodeAddress(ZERO_ADDRESS).publicKey,
        ])
      ),
      PluginInfoAbiType
    )) as PluginInfoTuple;

    const ts = (await algorand.client.algod.status().do())
    const block = (await algorand.client.algod.block((ts.lastRound - 1n)).do());

    expect(globalPluginBox[8]).toBe(BigInt(block.block.header.timestamp));
  });

  test('global valid, global is used', async () => {
    const { algorand } = fixture;

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 0,
        pluginName: '',
        escrowName: ''
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)

    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    await abstractedAccountClient.send.arc58AddPlugin({
      sender: aliceEOA.addr,
      signer: makeBasicAccountTransactionSigner(aliceEOA),
      args: {
        app: plugin,
        allowedCaller: ZERO_ADDRESS,
        admin: false,
        delegationType: 3,
        escrow: '',
        lastValid: MAX_UINT64,
        cooldown: 1,
        methods: [],
        useRounds: false
      }
    });

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [], true);

    const globalPluginBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('p'),
          Buffer.from(algosdk.encodeUint64(plugin)),
          algosdk.decodeAddress(ZERO_ADDRESS).publicKey,
        ])
      ),
      PluginInfoAbiType
    )) as PluginInfoTuple;

    const ts = (await algorand.client.algod.status().do())
    const block = (await algorand.client.algod.block(ts.lastRound - 1n).do());

    expect(globalPluginBox[8]).toBe(BigInt(block.block.header.timestamp));
  });

  test('global does not exist, sender valid', async () => {
    const { algorand } = fixture;

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 0,
        pluginName: '',
        escrowName: ''
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)

    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    await abstractedAccountClient.send.arc58AddPlugin({
      sender: aliceEOA.addr,
      signer: makeBasicAccountTransactionSigner(aliceEOA),
      args: {
        app: plugin,
        allowedCaller: caller.addr.toString(),
        admin: false,
        delegationType: 3,
        escrow: '',
        lastValid: MAX_UINT64,
        cooldown: 1,
        methods: [],
        useRounds: false
      }
    });

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [], false);

    const callerPluginBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('p'),
          Buffer.from(algosdk.encodeUint64(plugin)),
          caller.addr.publicKey,
        ])
      ),
      PluginInfoAbiType
    )) as PluginInfoTuple;

    const ts = (await algorand.client.algod.status().do())
    const block = (await algorand.client.algod.block(ts.lastRound - 1n).do());

    expect(callerPluginBox[8]).toBe(BigInt(block.block.header.timestamp));
  });

  test('global does not exist, sender valid, method allowed', async () => {
    const { algorand } = fixture;

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 3,
        pluginName: '',
        escrowName: ''
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)

    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    const optInToAssetSelector = optInPluginClient.appClient.getABIMethod('optInToAsset').getSelector();
    await abstractedAccountClient.send.arc58AddPlugin({
      sender: aliceEOA.addr,
      signer: makeBasicAccountTransactionSigner(aliceEOA),
      args: {
        app: plugin,
        allowedCaller: caller.addr.toString(),
        admin: false,
        delegationType: 3,
        escrow: '',
        lastValid: MAX_UINT64,
        cooldown: 1,
        methods: [
          [optInToAssetSelector, 0],
          [Buffer.from('dddd'), 0],
          [Buffer.from('aaaa'), 0]
        ],
        useRounds: false
      }
    });

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    console.log('optInToAssetSelector', new Uint8Array([...optInToAssetSelector]))

    await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [0], false);

    // const capturedLogs = logs.testLogger.capturedLogs
    // console.log('capturedLogs', capturedLogs)

    const callerPluginBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('p'),
          Buffer.from(algosdk.encodeUint64(plugin)),
          caller.addr.publicKey,
        ])
      ),
      PluginInfoAbiType
    )) as PluginInfoTuple;

    const ts = (await algorand.client.algod.status().do())
    const block = (await algorand.client.algod.block(ts.lastRound - 1n).do());

    expect(callerPluginBox[8]).toBe(BigInt(block.block.header.timestamp));
  });

  test('methods on cooldown', async () => {
    const { algorand } = fixture;

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 1,
        pluginName: '',
        escrowName: ''
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)

    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    const optInToAssetSelector = optInPluginClient.appClient.getABIMethod('optInToAsset').getSelector();
    await abstractedAccountClient.send.arc58AddPlugin({
      sender: aliceEOA.addr,
      signer: makeBasicAccountTransactionSigner(aliceEOA),
      args: {
        app: plugin,
        allowedCaller: ZERO_ADDRESS,
        admin: false,
        delegationType: 3,
        escrow: '',
        lastValid: MAX_UINT64,
        cooldown: 0,
        methods: [
          [optInToAssetSelector, 100] // cooldown of 1 so we can call it at most once per round
        ],
        useRounds: false
      }
    });

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [0]);

    const callerPluginBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('p'),
          Buffer.from(algosdk.encodeUint64(plugin)),
          algosdk.decodeAddress(ZERO_ADDRESS).publicKey,
        ])
      ),
      PluginInfoAbiType
    )) as PluginInfoTuple;

    const ts = (await algorand.client.algod.status().do());
    const block = (await algorand.client.algod.block(ts.lastRound - 1n).do());

    expect(callerPluginBox[5][0][2]).toBe(BigInt(block.block.header.timestamp));

    let error = 'no error';
    try {
      await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [0]);
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_METHOD_ON_COOLDOWN)
  });

  test('methods on cooldown, single group', async () => {
    const { algorand } = fixture;

    const optInToAssetSelector = optInPluginClient.appClient.getABIMethod('optInToAsset').getSelector();

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 1,
        pluginName: '',
        escrowName: ''
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)

    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    await abstractedAccountClient.send.arc58AddPlugin({
      sender: aliceEOA.addr,
      signer: makeBasicAccountTransactionSigner(aliceEOA),
      args: {
        app: plugin,
        allowedCaller: ZERO_ADDRESS,
        admin: false,
        delegationType: 3,
        escrow: '',
        lastValid: MAX_UINT64,
        cooldown: 0,
        methods: [
          [optInToAssetSelector, 1] // cooldown of 1 so we can call it at most once per round
        ],
        useRounds: false
      }
    });

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbrPayment = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
      sender: caller.addr,
      receiver: abstractedAccountClient.appAddress,
      amount: 100_000,
      suggestedParams,
    });

    const optInGroup = (
      await (optInPluginClient
        .createTransaction
        .optInToAsset({
          sender: caller.addr,
          signer: makeBasicAccountTransactionSigner(caller),
          args: {
            walletId: abstractedAccountClient.appId,
            rekeyBack: false,
            assets: [asset],
            mbrPayment
          },
          extraFee: (1_000).microAlgos()
        }))
    ).transactions;

    const mbrPaymentTwo = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
      sender: caller.addr,
      receiver: abstractedAccountClient.appAddress,
      amount: 100_000,
      suggestedParams,
      note: new Uint8Array(Buffer.from('two'))
    });

    const optInGroupTwo = (
      await (optInPluginClient
        .createTransaction
        .optInToAsset({
          sender: caller.addr,
          signer: makeBasicAccountTransactionSigner(caller),
          args: {
            walletId: abstractedAccountClient.appId,
            rekeyBack: true,
            assets: [asset],
            mbrPayment: mbrPaymentTwo
          },
          extraFee: (1_000).microAlgos(),
          note: 'two'
        }))
    ).transactions;

    let error = 'no error';
    try {
      await abstractedAccountClient
        .newGroup()
        .arc58RekeyToPlugin({
          sender: caller.addr,
          signer: makeBasicAccountTransactionSigner(caller),
          args: {
            plugin,
            global: true,
            methodOffsets: [0, 0],
            fundsRequest: []
          },
          extraFee: (1000).microAlgos()
        })
        // Add the mbr payment
        .addTransaction(optInGroup[0], makeBasicAccountTransactionSigner(caller)) // mbrPayment
        // Add the opt-in plugin call
        .addTransaction(optInGroup[1], makeBasicAccountTransactionSigner(caller)) // optInToAsset
        .addTransaction(optInGroupTwo[0], makeBasicAccountTransactionSigner(caller)) // mbrPayment
        .addTransaction(optInGroupTwo[1], makeBasicAccountTransactionSigner(caller)) // optInToAsset
        .arc58VerifyAuthAddr({
          sender: caller.addr,
          signer: makeBasicAccountTransactionSigner(caller),
          args: {}
        })
        .send();
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_METHOD_ON_COOLDOWN);
  });

  test('plugins on cooldown', async () => {
    const { algorand } = fixture;

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 0,
        pluginName: '',
        escrowName: ''
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)

    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    await abstractedAccountClient.send.arc58AddPlugin({
      sender: aliceEOA.addr,
      signer: makeBasicAccountTransactionSigner(aliceEOA),
      args: {
        app: plugin,
        allowedCaller: caller.addr.toString(),
        admin: false,
        delegationType: 3,
        escrow: '',
        lastValid: MAX_UINT64,
        cooldown: 100,
        methods: [],
        useRounds: false
      }
    });

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [], false);

    let error = 'no error';
    try {
      await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [], false);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_PLUGIN_ON_COOLDOWN);
  });

  test('neither sender nor global plugin exists', async () => {
    let error = 'no error';
    try {
      await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_PLUGIN_DOES_NOT_EXIST);
  });

  test('expired', async () => {
    const { algorand } = fixture;

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 0,
        pluginName: '',
        escrowName: ''
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)

    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    await abstractedAccountClient.send.arc58AddPlugin({
      sender: aliceEOA.addr,
      signer: makeBasicAccountTransactionSigner(aliceEOA),
      args: {
        app: plugin,
        allowedCaller: ZERO_ADDRESS,
        admin: false,
        delegationType: 3,
        escrow: '',
        lastValid: 1,
        cooldown: 0,
        methods: [],
        useRounds: false
      }
    });

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    let error = 'no error';
    try {
      await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_PLUGIN_EXPIRED);
  });

  test('erroneous app call in sandwich', async () => {
    const { algorand } = fixture;

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 0,
        pluginName: '',
        escrowName: ''
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)

    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    await abstractedAccountClient.send.arc58AddPlugin({
      sender: aliceEOA.addr,
      signer: makeBasicAccountTransactionSigner(aliceEOA),
      args: {
        app: plugin,
        allowedCaller: ZERO_ADDRESS,
        admin: false,
        delegationType: 3,
        escrow: '',
        lastValid: MAX_UINT64,
        cooldown: 0,
        methods: [],
        useRounds: false
      }
    });

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbrPayment = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
      sender: caller.addr,
      receiver: abstractedAccountClient.appAddress,
      amount: 100_000,
      suggestedParams,
    });

    // create an extra app call that on its own would succeed
    const erroneousAppCall = (
      await abstractedAccountClient.createTransaction.arc58AddPlugin({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          app: plugin,
          allowedCaller: caller.addr.toString(),
          admin: false,
          delegationType: 3,
          escrow: '',
          lastValid: MAX_UINT64,
          cooldown: 0,
          methods: [],
          useRounds: false
        }
      })
    ).transactions[0];

    const optInGroup = (
      await (optInPluginClient
        .createTransaction
        .optInToAsset({
          sender: caller.addr,
          signer: makeBasicAccountTransactionSigner(caller),
          args: {
            walletId: abstractedAccountClient.appId,
            rekeyBack: true,
            assets: [asset],
            mbrPayment
          },
          extraFee: (1_000).microAlgos()
        }))
    ).transactions;

    let error = 'no error';
    try {
      await abstractedAccountClient
        .newGroup()
        .arc58RekeyToPlugin({
          sender: caller.addr,
          signer: makeBasicAccountTransactionSigner(caller),
          args: {
            plugin,
            global: true,
            methodOffsets: [],
            fundsRequest: []
          },
          extraFee: (1000).microAlgos()
        })
        // Add the mbr payment
        .addTransaction(optInGroup[0], makeBasicAccountTransactionSigner(caller)) // mbrPayment
        // Add the opt-in plugin call
        .addTransaction(optInGroup[1], makeBasicAccountTransactionSigner(caller)) // optInToAsset
        .addTransaction(erroneousAppCall, makeBasicAccountTransactionSigner(aliceEOA)) // erroneous app call
        .arc58VerifyAuthAddr({
          sender: caller.addr,
          signer: makeBasicAccountTransactionSigner(caller),
          args: {}
        })
        .send();
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_CANNOT_CALL_OTHER_APPS_DURING_REKEY);
  });

  test('malformed methodOffsets', async () => {
    const { algorand } = fixture;

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 1,
        pluginName: '',
        escrowName: ''
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)

    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    await abstractedAccountClient.send.arc58AddPlugin({
      sender: aliceEOA.addr,
      signer: makeBasicAccountTransactionSigner(aliceEOA),
      args: {
        app: plugin,
        allowedCaller: ZERO_ADDRESS,
        admin: false,
        delegationType: 0,
        escrow: '',
        lastValid: MAX_UINT64,
        cooldown: 0,
        methods: [
          [new Uint8Array(Buffer.from('dddd')), 0]
        ],
        useRounds: false
      }
    });

    let error = 'no error';
    try {
      await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, []);
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_MALFORMED_OFFSETS);
  });

  test('allowance - flat', async () => {
    const { algorand } = fixture;
    escrow = 'pay_plugin';

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 0,
        pluginName: '',
        escrowName: escrow
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)
    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    const randomAccount = algorand.account.random().account;

    await algorand.account.ensureFunded(
      randomAccount.addr,
      dispenser,
      (100).algos(),
    );

    const mbrPayment = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
      sender: aliceEOA.addr,
      receiver: abstractedAccountClient.appAddress,
      amount: 100_000,
      suggestedParams,
    });

    await abstractedAccountClient.send
      .arc58AddPlugin({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          app: plugin,
          allowedCaller: ZERO_ADDRESS,
          admin: false,
          delegationType: 3,
          escrow: '',
          lastValid: MAX_UINT64,
          cooldown: 1,
          methods: [],
          useRounds: false
        }
      })

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [], true);

    const escrowCreationCost = BigInt(112_100 + 100_000) // Global.minBalance
    const amount = mbr.plugins + mbr.allowances + mbr.escrows + escrowCreationCost

    console.log(`Funding arc58 account with amount: ${amount}`)
    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, amount.microAlgo())

    await abstractedAccountClient.newGroup()
      .addTransaction(
        algosdk.makeAssetTransferTxnWithSuggestedParamsFromObject({
          sender: randomAccount.addr,
          receiver: randomAccount.addr,
          amount: 0,
          assetIndex: asset,
          suggestedParams
        }),
        makeBasicAccountTransactionSigner(randomAccount)
      )
      .addTransaction(
        algosdk.makeAssetTransferTxnWithSuggestedParamsFromObject({
          sender: aliceEOA.addr,
          receiver: abstractedAccountClient.appAddress,
          amount: 100_000_000,
          assetIndex: asset,
          suggestedParams
        }),
        makeBasicAccountTransactionSigner(aliceEOA)
      )
      .arc58AddPlugin({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          app: payPlugin,
          allowedCaller: ZERO_ADDRESS,
          admin: false,
          delegationType: 3,
          escrow,
          lastValid: MAX_UINT64,
          cooldown: 1,
          methods: [],
          useRounds: false
        }
      })
      .arc58AddAllowances({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          escrow,
          allowances: [[
            asset,
            1, // type
            10_000_000, // allowed
            0, // max
            300, // interval
            false, // useRounds
          ]]
        },
      })
      .arc58PluginOptinEscrow({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          app: payPlugin,
          allowedCaller: ZERO_ADDRESS,
          assets: [asset],
          mbrPayment,
        },
        extraFee: (8000).microAlgos(),
      })
      .send()

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    // use the full amount
    await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 6_000_000n, [], true);

    const escrowAppID = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(Buffer.concat([Buffer.from('e'), Buffer.from('pay_plugin')])),
      EscrowInfoAbiType
    )) as bigint;

    console.log('escrowAppID', escrowAppID);

    let allowanceBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('a'),
          Buffer.from(algosdk.encodeUint64(escrowAppID)),
          Buffer.from(algosdk.encodeUint64(asset)),
        ])
      ),
      AllowanceInfoAbiType
    )) as AllowanceInfoTuple;

    expect(allowanceBox[3]).toBe(6_000_000n); // type 2 is window

    await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 2_000_000n, [], true);

    allowanceBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('a'),
          Buffer.from(algosdk.encodeUint64(escrowAppID)),
          Buffer.from(algosdk.encodeUint64(asset)),
        ])
      ),
      AllowanceInfoAbiType
    )) as AllowanceInfoTuple;

    expect(allowanceBox[3]).toBe(8_000_000n); // type 2 is window

    // try to use more
    let error = 'no error';
    try {
      await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 8_000_000n, [], true)
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_ALLOWANCE_EXCEEDED);
  })

  test('allowance - window', async () => {
    const { algorand } = fixture;
    escrow = 'pay_plugin_window'

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 0,
        pluginName: '',
        escrowName: escrow,
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)
    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo())

    const randomAccount = algorand.account.random().account;

    await algorand.account.ensureFunded(
      randomAccount.addr,
      dispenser,
      (100).algos(),
    );

    const mbrPayment = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
      sender: aliceEOA.addr,
      receiver: abstractedAccountClient.appAddress,
      amount: 100_000,
      suggestedParams,
    });

    await abstractedAccountClient.send
      .arc58AddPlugin({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          app: plugin,
          allowedCaller: ZERO_ADDRESS,
          admin: false,
          delegationType: 3,
          escrow: '',
          lastValid: MAX_UINT64,
          cooldown: 1,
          methods: [],
          useRounds: false
        }
      })

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [], true);

    const escrowCreationCost = BigInt(112_100 + 100_000) // Global.minBalance
    const amount = mbr.plugins + mbr.allowances + mbr.escrows + escrowCreationCost
    console.log(`Funding arc58 account with amount: ${amount}`)
    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, amount.microAlgo())

    await abstractedAccountClient.newGroup()
      .addTransaction(algosdk.makeAssetTransferTxnWithSuggestedParamsFromObject({
        sender: randomAccount.addr,
        receiver: randomAccount.addr,
        amount: 0,
        assetIndex: asset,
        suggestedParams
      }), makeBasicAccountTransactionSigner(randomAccount))
      .addTransaction(algosdk.makeAssetTransferTxnWithSuggestedParamsFromObject({
        sender: aliceEOA.addr,
        receiver: abstractedAccountClient.appAddress,
        amount: 100_000_000,
        assetIndex: asset,
        suggestedParams
      }), makeBasicAccountTransactionSigner(aliceEOA))
      .arc58AddPlugin({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          app: payPlugin,
          allowedCaller: ZERO_ADDRESS,
          admin: false,
          delegationType: 3,
          escrow,
          lastValid: MAX_UINT64,
          cooldown: 1,
          methods: [],
          useRounds: false
        }
      })
      .arc58AddAllowances({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          escrow,
          allowances: [
            [
              asset,
              2, // type
              10_000_000, // allowed
              0, // max
              300, // interval
              false, // useRounds
            ]
          ]
        },
      })
      .arc58PluginOptinEscrow({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          app: payPlugin,
          allowedCaller: ZERO_ADDRESS,
          assets: [asset],
          mbrPayment,
        },
        extraFee: (8000).microAlgos(),
      })
      .send()

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    // use the full amount
    await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 10_000_000n, [], true);

    const escrowAppID = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(Buffer.concat([Buffer.from('e'), Buffer.from('pay_plugin_window')])),
      EscrowInfoAbiType
    )) as bigint;

    let allowanceBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('a'),
          Buffer.from(algosdk.encodeUint64(escrowAppID)),
          Buffer.from(algosdk.encodeUint64(asset)),
        ])
      ),
      AllowanceInfoAbiType
    )) as AllowanceInfoTuple;

    expect(allowanceBox[3]).toBe(10_000_000n); // type 2 is window

    let globalPluginBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('p'),
          Buffer.from(algosdk.encodeUint64(payPlugin)),
          algosdk.decodeAddress(ZERO_ADDRESS).publicKey,
        ])
      ),
      PluginInfoAbiType
    )) as PluginInfoTuple;

    const spendingAddress = algosdk.getApplicationAddress(globalPluginBox[2]);

    const spendingAddressInfo = await algorand.account.getInformation(spendingAddress.toString())

    expect(spendingAddressInfo.authAddr?.toString()).toBe(abstractedAccountClient.appAddress.toString());

    // try to use more
    let error = 'no error';
    try {
      await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 1n, [], true)
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_ALLOWANCE_EXCEEDED);

    // wait for the next window
    for (let i = 0; i < 3; i++) {
      await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 0n, [], true)
    }

    // use more
    await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 1_000_000n, [], true);

    allowanceBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('a'),
          Buffer.from(algosdk.encodeUint64(escrowAppID)),
          Buffer.from(algosdk.encodeUint64(asset)),
        ])
      ),
      AllowanceInfoAbiType
    )) as AllowanceInfoTuple;

    expect(allowanceBox[3]).toBe(1_000_000n); // type 2 is window

    await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 8_000_000n, [], true);

    // try to use more
    error = 'no error';
    try {
      await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 2_000_000n, [], true)
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_ALLOWANCE_EXCEEDED);
  })

  test('allowance - drip', async () => {
    const { algorand } = fixture;
    escrow = 'pay_plugin_drip';

    const dispenser = await algorand.account.dispenserFromEnvironment();

    let accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    const mbr = (await abstractedAccountClient.send.mbr({
      args: {
        methodCount: 0,
        pluginName: '',
        escrowName: escrow,
      }
    })).return

    if (mbr === undefined) {
      throw new Error('MBR is undefined');
    }

    console.log(`Funding arc58 account with amount: ${mbr.plugins}`)
    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, mbr.plugins.microAlgo());

    const randomAccount = algorand.account.random().account;

    await algorand.account.ensureFunded(
      randomAccount.addr,
      dispenser,
      (100).algos(),
    );

    const mbrPayment = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
      sender: aliceEOA.addr,
      receiver: abstractedAccountClient.appAddress,
      amount: 100_000,
      suggestedParams,
    });

    await abstractedAccountClient.send
      .arc58AddPlugin({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          app: plugin,
          allowedCaller: ZERO_ADDRESS,
          admin: false,
          delegationType: 3,
          escrow: '',
          lastValid: MAX_UINT64,
          cooldown: 1,
          methods: [],
          useRounds: false
        }
      })

    accountInfo = await algorand.account.getInformation(abstractedAccountClient.appAddress)
    expect(accountInfo.balance.microAlgos).toEqual(accountInfo.minBalance.microAlgos)

    await callOptinPlugin(caller, abstractedAccountClient.appAddress.toString(), suggestedParams, optInPluginClient, asset, [], true);

    const escrowCreationCost = BigInt(112_100 + 100_000) // Global.minBalance
    const amount = mbr.plugins + mbr.allowances + mbr.escrows + escrowCreationCost;
    console.log(`Funding arc58 account with amount: ${amount}`)
    await algorand.account.ensureFunded(abstractedAccountClient.appAddress, dispenser, amount.microAlgo());

    await abstractedAccountClient.newGroup()
      .addTransaction(algosdk.makeAssetTransferTxnWithSuggestedParamsFromObject({
        sender: randomAccount.addr,
        receiver: randomAccount.addr,
        amount: 0,
        assetIndex: asset,
        suggestedParams
      }), makeBasicAccountTransactionSigner(randomAccount))
      .addTransaction(algosdk.makeAssetTransferTxnWithSuggestedParamsFromObject({
        sender: aliceEOA.addr,
        receiver: abstractedAccountClient.appAddress,
        amount: 100_000_000,
        assetIndex: asset,
        suggestedParams
      }), makeBasicAccountTransactionSigner(aliceEOA))
      .arc58AddPlugin({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          app: payPlugin,
          allowedCaller: ZERO_ADDRESS,
          admin: false,
          delegationType: 3,
          escrow,
          lastValid: MAX_UINT64,
          cooldown: 1,
          methods: [],
          useRounds: true
        }
      })
      .arc58AddAllowances({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          escrow,
          allowances: [
            [
              asset,
              3, // type
              1_000_000, // allowed
              50_000_000, // max
              1, // interval
              true, // useRounds
            ]
          ]
        },
      })
      .arc58PluginOptinEscrow({
        sender: aliceEOA.addr,
        signer: makeBasicAccountTransactionSigner(aliceEOA),
        args: {
          app: payPlugin,
          allowedCaller: ZERO_ADDRESS,
          assets: [asset],
          mbrPayment,
        },
        extraFee: (8000).microAlgos(),
      })
      .send()

    // use the full amount
    await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 1_000_000n, [], true);

    const escrowAppID = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(Buffer.concat([Buffer.from('e'), Buffer.from('pay_plugin_drip')])),
      EscrowInfoAbiType
    )) as bigint;

    let allowanceBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('a'),
          Buffer.from(algosdk.encodeUint64(escrowAppID)),
          Buffer.from(algosdk.encodeUint64(asset)),
        ])
      ),
      AllowanceInfoAbiType
    )) as AllowanceInfoTuple;

    expect(allowanceBox[3]).toBe(49_000_000n);

    await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 48_000_000n, [], true)

    // try to use more
    let error = 'no error';
    try {
      await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 5_000_000n, [], true)
    } catch (e: any) {
      error = e.message;
    }

    expect(error).toContain(ERR_ALLOWANCE_EXCEEDED);

    // wait for the next window
    for (let i = 0; i < 3; i++) {
      await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 0n, [], true)
    }

    await callPayPlugin(caller, payPluginClient, randomAccount.addr.toString(), asset, 0n, [], true);

    allowanceBox = (await abstractedAccountClient.appClient.getBoxValueFromABIType(
      new Uint8Array(
        Buffer.concat([
          Buffer.from('a'),
          Buffer.from(algosdk.encodeUint64(escrowAppID)),
          Buffer.from(algosdk.encodeUint64(asset)),
        ])
      ),
      AllowanceInfoAbiType
    )) as AllowanceInfoTuple;

    expect(allowanceBox[3]).toBe(6_000_000n);
  })
});
