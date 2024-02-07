import algosdk from 'algosdk';
import vaultABI from '../../contracts/Vault.abi.json';
import masterABI from '../../contracts/Master.abi.json';

interface Holding {
  optedIn: boolean,
  vault?: number,
  vaultOptedIn?: boolean,
}

interface Assets {
  vaultAssets?: any
  accountAssets: any
}

interface GlobalStateDeltaValue {
    action: number,
    bytes?: string
    uint?: number
}

interface GlobalStateDelta {
    key: string
    value: GlobalStateDeltaValue
}

interface ReadableGlobalStateDelta {
    [key: string]: string | number | bigint | undefined
}

function getReadableGlobalState(delta: Array<GlobalStateDelta>) {
  const r = {} as ReadableGlobalStateDelta;

  delta.forEach((d) => {
    const key = Buffer.from(d.key, 'base64').toString('utf8');
    let value = null;

    if (d.value.bytes) {
    // first see if it's a valid address
      const b = new Uint8Array(Buffer.from(d.value.bytes as string, 'base64'));
      value = algosdk.encodeAddress(b);

      // then decode as string
      if (!algosdk.isValidAddress(value)) {
        value = Buffer.from(d.value.bytes as string, 'base64').toString();
      }
    } else {
      value = d.value.uint;
    }

    r[key] = value;
  });

  return r;
}

export default class ARC12 {
  indexer: algosdk.Indexer;

  masterApp: number;

  algodClient: algosdk.Algodv2;

  vaultContract: algosdk.ABIContract;

  masterContract: algosdk.ABIContract;

  constructor(indexer: algosdk.Indexer, algodClient: algosdk.Algodv2, masterApp: number) {
    this.indexer = indexer;
    this.masterApp = masterApp;
    this.algodClient = algodClient;
    this.vaultContract = new algosdk.ABIContract(vaultABI);
    this.masterContract = new algosdk.ABIContract(masterABI);
  }

  async getVault(address: string) {
    const pubKey = algosdk.decodeAddress(address).publicKey;
    try {
      const boxResponse = await this.indexer
        .lookupApplicationBoxByIDandName(this.masterApp, pubKey).do();

      return algosdk.decodeUint64(boxResponse.value, 'safe');
    } catch (e: any) {
      if (e.response.body.message.includes('no application boxes found')) {
        return undefined;
      }

      throw e;
    }
  }

  private async deleteNeeded(vault: number): Promise<boolean> {
    return (await this.indexer.lookupAccountAssets(algosdk.getApplicationAddress(vault)).do())
      .assets.length === 1;
  }

  private async deleteVault(
    atc: algosdk.AtomicTransactionComposer,
    sender: string,
    signer: algosdk.TransactionSigner,
    vault: number,
  ): Promise<algosdk.AtomicTransactionComposer> {
    const res = (await this.indexer.lookupApplications(vault).do());
    const vaultCreator = (getReadableGlobalState(res.application.params['global-state']).creator) as string;

    const appSp = await this.algodClient.getTransactionParams().do();
    appSp.fee = 0;
    appSp.flatFee = true;

    atc.addMethodCall({
      appID: this.masterApp,
      method: algosdk.getMethodByName(this.masterContract.methods, 'deleteVault'),
      methodArgs: [vault, vaultCreator],
      sender,
      suggestedParams: appSp,
      signer,
      boxes: [{ appIndex: this.masterApp, name: algosdk.decodeAddress(sender).publicKey }],
    });

    return atc;
  }

  async reject(
    atc: algosdk.AtomicTransactionComposer,
    sender: string,
    signer: algosdk.TransactionSigner,
    asa: number,
    vault: number,
  ): Promise<algosdk.AtomicTransactionComposer> {
    const asaCreator = (await this.indexer.lookupAssetByID(asa).do()).asset.params.creator;

    const res = (await this.indexer.lookupApplications(vault).do());
    const vaultCreator = (getReadableGlobalState(res.application.params['global-state']).creator) as string;

    const del = await this.deleteNeeded(vault);

    const sp = await this.algodClient.getTransactionParams().do();
    sp.fee = (sp.fee || 1_000) * (del ? 8 : 4); // 8 if delete
    sp.flatFee = true;

    atc.addMethodCall({
      appID: vault,
      method: algosdk.getMethodByName(this.vaultContract.methods, 'reject'),
      methodArgs: [asaCreator, 'Y76M3MSY6DKBRHBL7C3NNDXGS5IIMQVQVUAB6MP4XEMMGVF2QWNPL226CA', asa, vaultCreator],
      sender,
      suggestedParams: sp,
      signer,
      boxes: [{ appIndex: vault, name: algosdk.encodeUint64(asa) }],
    });

    if (del) {
      await this.deleteVault(atc, sender, signer, vault);
    }

    return atc;
  }

  async claim(
    atc: algosdk.AtomicTransactionComposer,
    sender: string,
    signer: algosdk.TransactionSigner,
    asa: number,
    vault: number,
  ): Promise<algosdk.AtomicTransactionComposer> {
    const res = (await this.indexer.lookupApplications(vault).do());
    const vaultCreator = (getReadableGlobalState(res.application.params['global-state']).creator) as string;

    const boxResponse = await this.indexer
      .lookupApplicationBoxByIDandName(vault, algosdk.encodeUint64(asa)).do();

    const asaFunder = algosdk.encodeAddress(boxResponse.value);

    const optInTxn = algosdk.makeAssetTransferTxnWithSuggestedParams(
      sender,
      sender,
      undefined,
      undefined,
      0,
      undefined,
      asa,
      await this.algodClient.getTransactionParams().do(),
    );

    atc.addTransaction({ txn: optInTxn, signer });

    const del = await this.deleteNeeded(vault);
    const appSp = await this.algodClient.getTransactionParams().do();
    appSp.fee = (appSp.fee || 1_000) * (del ? 7 : 3);
    appSp.flatFee = true;

    atc.addMethodCall({
      appID: vault,
      method: algosdk.getMethodByName(this.vaultContract.methods, 'claim'),
      methodArgs: [asa, vaultCreator, asaFunder],
      sender,
      suggestedParams: appSp,
      signer,
      boxes: [{ appIndex: vault, name: algosdk.encodeUint64(asa) }],
    });

    if (del) {
      await this.deleteVault(atc, sender, signer, vault);
    }

    return atc;
  }

  async createVault(
    atc: algosdk.AtomicTransactionComposer,
    sender: string,
    signer: algosdk.TransactionSigner,
    receiver: string,
    master: number,
  ): Promise<algosdk.AtomicTransactionComposer> {
    const suggestedParams = await this.algodClient.getTransactionParams().do();
    const payTxn = algosdk.makePaymentTxnWithSuggestedParams(
      sender,
      algosdk.getApplicationAddress(master),
      347_000,
      undefined,
      undefined,
      suggestedParams,
    );

    const appSp = await this.algodClient.getTransactionParams().do();
    appSp.fee = (appSp.fee || 1_000) * 3;
    appSp.flatFee = true;

    atc.addMethodCall({
      appID: master,
      method: algosdk.getMethodByName(this.masterContract.methods, 'createVault'),
      methodArgs: [receiver, { txn: payTxn, signer }],
      sender,
      suggestedParams: appSp,
      signer,
      boxes: [{ appIndex: master, name: algosdk.decodeAddress(receiver).publicKey }],
    });

    return atc;
  }

  async vaultOptIn(
    atc: algosdk.AtomicTransactionComposer,
    sender: string,
    signer: algosdk.TransactionSigner,
    asa: number,
    vault: number,
  ): Promise<algosdk.AtomicTransactionComposer> {
    const suggestedParams = await this.algodClient.getTransactionParams().do();
    const payTxn = algosdk.makePaymentTxnWithSuggestedParams(
      sender,
      algosdk.getApplicationAddress(vault),
      118_500,
      undefined,
      undefined,
      suggestedParams,
    );

    const appSp = await this.algodClient.getTransactionParams().do();
    appSp.fee = (appSp.fee || 1_000) * 2;
    appSp.flatFee = true;

    atc.addMethodCall({
      appID: vault,
      method: algosdk.getMethodByName(this.vaultContract.methods, 'optIn'),
      methodArgs: [asa, { txn: payTxn, signer }],
      sender,
      suggestedParams: appSp,
      boxes: [{ appIndex: vault, name: algosdk.encodeUint64(asa) }],
      signer,
    });

    return atc;
  }

  async getAssets(address: string): Promise<Assets> {
    const assets: Assets = {
      accountAssets: await this.indexer.lookupAccountAssets(address).do(),
    };
    const vault = await this.getVault(address);

    if (vault) {
      assets.vaultAssets = await this.indexer.lookupAccountAssets(
        algosdk.getApplicationAddress(vault),
      ).do();
    }

    return assets;
  }

  async getHolding(address: string, asa: number): Promise<Holding> {
    const accountAssets: any[] = (await this.indexer.lookupAccountAssets(address)
      .assetId(asa).do()).assets;

    if (accountAssets.length !== 0) {
      return {
        optedIn: true,
      };
    }

    const holding: Holding = { optedIn: false };

    holding.vault = await this.getVault(address);

    if (holding.vault) {
      const appAddr = algosdk.getApplicationAddress(holding.vault);
      const vaultAssets: any[] = (await this.indexer.lookupAccountAssets(appAddr)
        .assetId(asa).do()).assets;

      holding.vaultOptedIn = vaultAssets.length > 0;
    }

    return holding;
  }

  async send(
    atc: algosdk.AtomicTransactionComposer,
    sender: string,
    signer: algosdk.TransactionSigner,
    asa: number,
    receiver: string,
    amount: number,
  ): Promise<{ confirmedRound: number; txIDs: string[]; methodResults: algosdk.ABIResult[]; }> {
    const holding = await this.getHolding(receiver, asa);
    let assetReceiver = receiver;

    if (!holding.optedIn && holding.vault) {
      if (!holding.vaultOptedIn) {
        await this.vaultOptIn(atc, sender, signer, asa, holding.vault);
      }
      assetReceiver = algosdk.getApplicationAddress(holding.vault);
    } else if (!holding.optedIn && !holding.vault) {
      const createAtc = new algosdk.AtomicTransactionComposer();
      await this.createVault(createAtc, sender, signer, receiver, this.masterApp);

      const res = await createAtc.execute(this.algodClient, 3);

      const createdVault = Number(res.methodResults[0].returnValue as algosdk.ABIValue);
      assetReceiver = algosdk.getApplicationAddress(createdVault);
      await this.vaultOptIn(atc, sender, signer, asa, createdVault);
    }

    const tx = algosdk.makeAssetTransferTxnWithSuggestedParams(
      sender,
      assetReceiver,
      undefined,
      undefined,
      amount,
      undefined,
      asa,
      await this.algodClient.getTransactionParams().do(),
    );

    atc.addTransaction({ txn: tx, signer });

    return atc.execute(this.algodClient, 3);
  }
}
