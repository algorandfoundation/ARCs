/* eslint-disable no-param-reassign */
import algosdk from 'algosdk';
import fs from 'fs';
import path from 'path';
// eslint-disable-next-line import/extensions,import/no-unresolved
import ARC12 from '../src/index';
import masterABI from '../../contracts/Master.abi.json';

const token = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa';
const server = 'http://localhost';
const indexerClient = new algosdk.Indexer('', server, 8980);
const algodClient = new algosdk.Algodv2(token, server, 4001);
const kmdClient = new algosdk.Kmd(token, server, 4002);
const kmdWallet = 'unencrypted-default-wallet';
const kmdPassword = '';

jest.setTimeout(10_000);
interface TestState {
  master: number,
  vault: number,
  assets: number[],
  sender: algosdk.Account
  receiver: algosdk.Account
  arc12: ARC12
}

// Based on https://github.com/algorand-devrel/demo-abi/blob/master/js/sandbox.ts
async function getAccounts(): Promise<algosdk.Account[]> {
  const { wallets } = await kmdClient.listWallets();

  // find kmdWallet
  let walletId;
  wallets.forEach((wallet: any) => {
    if (wallet.name === kmdWallet) walletId = wallet.id;
  });
  if (walletId === undefined) throw Error(`No wallet named: ${kmdWallet}`);

  // get handle
  const handleResp = await kmdClient.initWalletHandle(walletId, kmdPassword);
  const handle = handleResp.wallet_handle_token;

  // get account keys
  const { addresses } = await kmdClient.listKeys(handle);
  const acctPromises: Promise<{private_key: Buffer}>[] = [];
  addresses.forEach((addr: string) => {
    acctPromises.push(kmdClient.exportKey(handle, kmdPassword, addr));
  });
  const keys = await Promise.all(acctPromises);

  // release handle
  await kmdClient.releaseWalletHandle(handle);

  // return all algosdk.Account objects derived from kmdWallet
  return keys.map((k) => {
    const addr = algosdk.encodeAddress(k.private_key.slice(32));
    return { sk: k.private_key, addr } as algosdk.Account;
  });
}

// https://developer.algorand.org/docs/get-details/dapps/smart-contracts/frontend/apps/#create
async function compileProgram(programSource: string) {
  const encoder = new TextEncoder();
  const programBytes = encoder.encode(programSource);
  const compileResponse = await algodClient.compile(programBytes).do();
  return new Uint8Array(Buffer.from(compileResponse.result, 'base64'));
}

async function createASA(state: TestState, amount: number = 1): Promise<number> {
  const asaTxn = algosdk.makeAssetCreateTxnWithSuggestedParams(
    state.sender.addr,
    undefined,
    amount,
    0,
    false,
    undefined,
    undefined,
    undefined,
    undefined,
    'TEST',
    `TEST${Math.random().toString()}`,
    undefined,
    undefined,
    await algodClient.getTransactionParams().do(),
  ).signTxn(state.sender.sk);

  const { txId } = await algodClient.sendRawTransaction(asaTxn).do();
  return (await algosdk.waitForConfirmation(algodClient, txId, 3))['asset-index'];
}

async function createMaster(state: TestState) {
  const masterContract = new algosdk.ABIContract(masterABI);
  const creator = state.sender as algosdk.Account;

  const txn = algosdk.makeApplicationCreateTxn(
    creator.addr,
    await algodClient.getTransactionParams().do(),
    algosdk.OnApplicationComplete.NoOpOC,
    await compileProgram(fs.readFileSync(path.join(__dirname, '../../contracts/Master.teal')).toString()),
    await compileProgram(fs.readFileSync(path.join(__dirname, 'clear.teal')).toString()),
    0,
    0,
    0,
    0,
    [algosdk.getMethodByName(masterContract.methods, 'create').getSelector()],
  ).signTxn(creator.sk);

  const { txId } = await algodClient.sendRawTransaction(txn).do();
  state.master = (await algosdk.waitForConfirmation(algodClient, txId, 3))['application-index'];

  const payTxn = algosdk.makePaymentTxnWithSuggestedParams(
    state.sender.addr,
    algosdk.getApplicationAddress(state.master),
    100_000,
    undefined,
    undefined,
    await algodClient.getTransactionParams().do(),
  ).signTxn(state.sender.sk);

  await algodClient.sendRawTransaction(payTxn).do();
}

async function newASA(state: TestState) {
  const atc = new algosdk.AtomicTransactionComposer();
  const signer = algosdk.makeBasicAccountTransactionSigner(state.sender);

  const asa = await createASA(state);

  await state.arc12.vaultOptIn(
    atc,
    state.sender.addr,
    signer,
    asa,
    state.vault,
  );

  const axfer = algosdk.makeAssetTransferTxnWithSuggestedParams(
    state.sender.addr,
    algosdk.getApplicationAddress(state.vault),
    undefined,
    undefined,
    1,
    undefined,
    asa,
    await algodClient.getTransactionParams().do(),
  );
  atc.addTransaction({ txn: axfer, signer });

  await atc.execute(algodClient, 3);

  return asa;
}

describe('ARC12 SDK', () => {
  // @ts-ignore
  const state: TestState = { assets: [] };

  beforeAll(async () => {
    const accounts = await getAccounts();
    state.sender = accounts.pop() as algosdk.Account;
    state.receiver = accounts.pop() as algosdk.Account;
    await createMaster(state);
    state.arc12 = new ARC12(indexerClient, algodClient, state.master);
  });

  it('createVault and getVault', async () => {
    const atc = new algosdk.AtomicTransactionComposer();
    const signer = algosdk.makeBasicAccountTransactionSigner(state.sender);
    await state.arc12.createVault(
      atc,
      state.sender.addr,
      signer,
      state.receiver.addr,
      state.master,
    );
    const res = await atc.execute(algodClient, 3);

    // Wait for indexer to catch up
    // eslint-disable-next-line no-promise-executor-return
    await new Promise((r) => setTimeout(r, 50));

    state.vault = Number(res.methodResults[0].returnValue as algosdk.ABIValue);
    expect(await state.arc12.getVault(state.receiver.addr)).toBe(state.vault);
  });

  it('optIn and getHolding', async () => {
    state.assets.push(await newASA(state));
    state.assets.push(await newASA(state));
    state.assets.push(await newASA(state));
    const holding = await state.arc12.getHolding(state.receiver.addr, state.assets[0]);

    expect(holding).toStrictEqual({ optedIn: false, vault: state.vault, vaultOptedIn: true });
  });

  it('claim and getAssets', async () => {
    const atc = new algosdk.AtomicTransactionComposer();
    await state.arc12.claim(
      atc,
      state.receiver.addr,
      algosdk.makeBasicAccountTransactionSigner(state.receiver),
      state.assets[0],
      state.vault,
    );

    await atc.execute(algodClient, 3);

    // Wait for indexer to catch up
    // eslint-disable-next-line no-promise-executor-return
    await new Promise((r) => setTimeout(r, 50));

    const { accountAssets } = await state.arc12.getAssets(state.receiver.addr);

    const asaBalance = accountAssets.assets.find((a: any) => a['asset-id'] === state.assets[0]).amount;

    expect(asaBalance).toBe(1);
  });

  it('reject', async () => {
    const atc = new algosdk.AtomicTransactionComposer();
    await state.arc12.reject(
      atc,
      state.receiver.addr,
      algosdk.makeBasicAccountTransactionSigner(state.receiver),
      state.assets[1],
      state.vault,
    );

    await atc.execute(algodClient, 3);

    // Wait for indexer to catch up
    // eslint-disable-next-line no-promise-executor-return
    await new Promise((r) => setTimeout(r, 50));

    const creatorBalance = (await indexerClient.lookupAccountAssets(state.sender.addr)
      .assetId(state.assets[1]).do()).assets[0].amount;

    expect(creatorBalance).toBe(1);
  });

  it('deleteVault', async () => {
    const atc = new algosdk.AtomicTransactionComposer();
    await state.arc12.claim(
      atc,
      state.receiver.addr,
      algosdk.makeBasicAccountTransactionSigner(state.receiver),
      state.assets[2],
      state.vault,
    );

    await atc.execute(algodClient, 3);

    expect(async () => {
      await algodClient.getApplicationByID(state.vault).do();
    }).rejects.toThrow('application does not exist');
  });

  it('send', async () => {
    const asa = await createASA(state, 3);
    let holding = await state.arc12.getHolding(state.receiver.addr, asa);
    expect(holding.optedIn).toBe(false);
    expect(holding.vault).toBe(undefined);

    await state.arc12.send(
      new algosdk.AtomicTransactionComposer(),
      state.sender.addr,
      algosdk.makeBasicAccountTransactionSigner(state.sender),
      asa,
      state.receiver.addr,
      1,
    );

    // Wait for indexer to catch up
    // eslint-disable-next-line no-promise-executor-return
    await new Promise((r) => setTimeout(r, 50));

    const vault = await state.arc12.getVault(state.receiver.addr);

    holding = await state.arc12.getHolding(state.receiver.addr, asa);
    expect(holding.optedIn).toBe(false);
    expect(holding.vault).toBe(vault);
    expect(holding.vaultOptedIn).toBe(true);

    const secondAsa = await createASA(state, 1);

    holding = await state.arc12.getHolding(state.receiver.addr, secondAsa);
    expect(holding.optedIn).toBe(false);
    expect(holding.vault).toBe(vault);
    expect(holding.vaultOptedIn).toBe(false);

    await state.arc12.send(
      new algosdk.AtomicTransactionComposer(),
      state.sender.addr,
      algosdk.makeBasicAccountTransactionSigner(state.sender),
      secondAsa,
      state.receiver.addr,
      1,
    );

    // Wait for indexer to catch up
    // eslint-disable-next-line no-promise-executor-return
    await new Promise((r) => setTimeout(r, 50));

    holding = await state.arc12.getHolding(state.receiver.addr, secondAsa);
    expect(holding.optedIn).toBe(false);
    expect(holding.vault).toBe(vault);
    expect(holding.vaultOptedIn).toBe(true);

    await state.arc12.send(
      new algosdk.AtomicTransactionComposer(),
      state.sender.addr,
      algosdk.makeBasicAccountTransactionSigner(state.sender),
      asa,
      state.receiver.addr,
      1,
    );

    // Wait for indexer to catch up
    // eslint-disable-next-line no-promise-executor-return
    await new Promise((r) => setTimeout(r, 50));

    holding = await state.arc12.getHolding(state.receiver.addr, asa);
    expect(holding.optedIn).toBe(false);
    expect(holding.vault).toBe(vault);
    expect(holding.vaultOptedIn).toBe(true);

    const claimAtc = new algosdk.AtomicTransactionComposer();
    await state.arc12.claim(
      claimAtc,
      state.receiver.addr,
      algosdk.makeBasicAccountTransactionSigner(state.receiver),
      asa,
      await state.arc12.getVault(state.receiver.addr) as number,
    );

    claimAtc.execute(algodClient, 3);

    // Wait for indexer to catch up
    // eslint-disable-next-line no-promise-executor-return
    await new Promise((r) => setTimeout(r, 50));

    holding = await state.arc12.getHolding(state.receiver.addr, asa);
    expect(holding.optedIn).toBe(true);
    expect(holding.vault).toBe(undefined);
    expect(holding.vaultOptedIn).toBe(undefined);

    await state.arc12.send(
      new algosdk.AtomicTransactionComposer(),
      state.sender.addr,
      algosdk.makeBasicAccountTransactionSigner(state.sender),
      asa,
      state.receiver.addr,
      1,
    );

    // Wait for indexer to catch up
    // eslint-disable-next-line no-promise-executor-return
    await new Promise((r) => setTimeout(r, 50));

    holding = await state.arc12.getHolding(state.receiver.addr, asa);
    expect(holding.optedIn).toBe(true);
    expect(holding.vault).toBe(undefined);
    expect(holding.vaultOptedIn).toBe(undefined);
  });
});
