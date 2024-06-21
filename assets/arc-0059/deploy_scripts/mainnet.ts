/* eslint-disable no-console */
import algosdk from 'algosdk';
import { Arc59Client } from '../contracts/clients/Arc59Client';
import 'dotenv/config';
import * as algokit from '@algorandfoundation/algokit-utils';

algokit.Config.configure({ populateAppCallResources: true });

async function deploy() {
  if (process.env.MAINNET_DEPLOYER_MNEMONIC === undefined) {
    throw new Error('MAINNET_DEPLOYER_MNEMONIC not set');
  }
  const deployer = algosdk.mnemonicToSecretKey(process.env.MAINNET_DEPLOYER_MNEMONIC);

  const algod = algokit.getAlgoClient(algokit.getAlgoNodeConfig('mainnet', 'algod'));

  const appClient = new Arc59Client(
    {
      sender: deployer,
      resolveBy: 'id',
      id: 0,
    },
    algod
  );

  const createResult = await appClient.create.createApplication({});

  console.debug(`App ${createResult.appId} created in transaction ${createResult.transaction.txID()}`);

  await appClient.appClient.fundAppAccount({ amount: algokit.microAlgos(100_000) });

  console.debug(`App ${createResult.appId} funded`);
}

deploy();
