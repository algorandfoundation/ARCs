/* eslint-disable no-console */
import algosdk from 'algosdk';
import { MutableArc59Client } from '../contracts/clients/MutableARC59Client';
import 'dotenv/config';
import * as algokit from '@algorandfoundation/algokit-utils';

algokit.Config.configure({ populateAppCallResources: true });

async function deploy() {
  if (process.env.TESTNET_DEPLOYER_MNEMONIC === undefined) {
    throw new Error('TESTNET_DEPLOYER_MNEMONIC not set');
  }
  const deployer = algosdk.mnemonicToSecretKey(process.env.TESTNET_DEPLOYER_MNEMONIC);
  if (process.env.TESTNET_APP_ID === undefined) {
    throw new Error('TESTNET_APP_ID not set');
  }

  const id = parseInt(process.env.TESTNET_APP_ID, 10);

  const algod = algokit.getAlgoClient(algokit.getAlgoNodeConfig('testnet', 'algod'));

  const appClient = new MutableArc59Client(
    {
      sender: deployer,
      resolveBy: 'id',
      id,
    },
    algod
  );

  if (id === 0) {
    const result = await appClient.create.createApplication({});

    console.debug(`App ${result.appId} created in transaction ${result.transaction.txID()}`);
  } else {
    const result = await appClient.update.updateApplication({});
    console.debug(`App ${id} updated in transaction ${result.transaction.txID()}`);
  }

  // Create Asset
  const assetCreate = algosdk.makeAssetCreateTxnWithSuggestedParamsFromObject({
    from: deployer.addr,
    total: 100,
    decimals: 0,
    defaultFrozen: false,
    suggestedParams: await algod.getTransactionParams().do(),
  });

  const atc = new algosdk.AtomicTransactionComposer();

  atc.addTransaction({ txn: assetCreate, signer: algosdk.makeBasicAccountTransactionSigner(deployer) });

  const assetCreateResult = await algokit.sendAtomicTransactionComposer({ atc }, algod);

  const assetId = Number(assetCreateResult.confirmations![0].assetIndex);

  console.debug(`Created asset ${assetId}`);

  await appClient.appClient.fundAppAccount({ amount: algokit.microAlgos(200_000) });

  await appClient.arc59OptRouterIn({ asa: assetId }, { sendParams: { fee: algokit.microAlgos(2_000) } });

  console.debug(`Opted router in to asset ${assetId}`);

  // Send asset to a new account's inbox
  const receiver = algosdk.generateAccount().addr;
  const arc59RouterAddress = (await appClient.appClient.getAppReference()).appAddress;

  const sendInfo = (await appClient.arc59GetSendAssetInfo({ asset: assetId, receiver })).return;

  const itxns = sendInfo![0];
  const mbr = sendInfo![1];

  const composer = appClient.compose();

  if (mbr) {
    const mbrPayment = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
      from: deployer.addr,
      to: arc59RouterAddress,
      amount: mbr,
      suggestedParams: await algod.getTransactionParams().do(),
    });

    composer.addTransaction({ transaction: mbrPayment, signer: deployer });
  }

  const axfer = algosdk.makeAssetTransferTxnWithSuggestedParamsFromObject({
    from: deployer.addr,
    to: arc59RouterAddress,
    assetIndex: assetId,
    amount: 1,
    suggestedParams: await algod.getTransactionParams().do(),
  });

  const result = await composer
    .arc59SendAsset({ axfer, receiver }, { sendParams: { fee: algokit.microAlgos(1000 + 1000 * Number(itxns)) } })
    .execute();

  console.debug(`Sent asset ${assetId} to ${receiver}'s inbox (${result.returns[0]})`);
}

deploy();
