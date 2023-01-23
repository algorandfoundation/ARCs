/* eslint-disable no-undef */
/* eslint-disable max-classes-per-file */

// eslint-disable-next-line import/no-unresolved, import/extensions
import { Contract } from '@joe-p/tealscript';

// eslint-disable-next-line no-unused-vars
class Vault extends Contract {
  creator = new GlobalValue<Account>({ key: 'creator' });

  master = new GlobalValue<Application>({ key: 'master' });

  receiver = new GlobalValue<Account>({ key: 'receiver' });

  funderMap = new BoxMap<Asset, Account>({ defaultSize: 32 });

  private closeAcct(vaultCreator: Account): void {
    assert(vaultCreator === this.creator.get());

    /// Send the MBR to the vault creator
    sendPayment({
      receiver: vaultCreator,
      amount: globals.currentApplicationAddress.minBalance,
      fee: 0,
      /// Any remaining balance is sent the receiver for the vault
      closeRemainderTo: this.txn.sender,
    });

    const deleteVaultTxn = this.txnGroup[this.txn.groupIndex + 1];
    /// Ensure Master.deleteVault is being called for this vault
    assert(deleteVaultTxn.applicationID === this.master.get());
    assert(deleteVaultTxn.applicationArgs[0] === method('deleteVault(application,account)void'));
    assert(deleteVaultTxn.applications[1] === this.app);
  }

  @createApplication
  create(receiver: Account, sender: Account): void {
    this.creator.put(sender);
    this.receiver.put(receiver);
    this.master.put(globals.callerApplicationID);
  }

  reject(asaCreator: Account, feeSink: Account, asa: Asset, vaultCreator: Account): void {
    assert(this.txn.sender === this.receiver.get());
    assert(feeSink === addr('Y76M3MSY6DKBRHBL7C3NNDXGS5IIMQVQVUAB6MP4XEMMGVF2QWNPL226CA'));
    const preMbr = globals.currentApplicationAddress.minBalance;

    /// Send asset back to creator since they are guranteed to be opted in
    sendAssetTransfer({
      assetReceiver: asaCreator,
      xferAsset: asa,
      assetAmount: 0,
      assetCloseTo: asaCreator,
      fee: 0,
    });

    this.funderMap.delete(asa);

    const mbrAmt = preMbr - globals.currentApplicationAddress.minBalance;

    /// Send MBR to fee sink
    sendPayment({
      receiver: feeSink,
      amount: mbrAmt - this.txn.fee,
      fee: 0,
    });

    /// Send fee back to sender
    sendPayment({
      receiver: this.txn.sender,
      amount: this.txn.fee,
      fee: 0,
    });

    if (globals.currentApplicationAddress.totalAssets === 0) this.closeAcct(vaultCreator);
  }

  optIn(asa: Asset, mbrPayment: PayTxn): void {
    assert(!this.funderMap.exists(asa));
    assert(mbrPayment.sender === this.txn.sender);
    assert(mbrPayment.receiver === globals.currentApplicationAddress);

    const preMbr = globals.currentApplicationAddress.minBalance;

    this.funderMap.put(asa, this.txn.sender);

    /// Opt vault into asa
    sendAssetTransfer({
      assetReceiver: globals.currentApplicationAddress,
      assetAmount: 0,
      fee: 0,
      xferAsset: asa,
    });

    assert(mbrPayment.amount === globals.currentApplicationAddress.minBalance - preMbr);
  }

  claim(asa: Asset, creator: Account, asaMbrFunder: Account): void {
    assert(this.funderMap.exists(asa));
    assert(asaMbrFunder === this.funderMap.get(asa));
    assert(this.txn.sender === this.receiver.get());
    assert(this.creator.get() === creator);

    const initialMbr = globals.currentApplicationAddress.minBalance;

    this.funderMap.delete(asa);

    /// Transfer all of the asset to the receiver
    sendAssetTransfer({
      assetReceiver: this.txn.sender,
      fee: 0,
      assetAmount: globals.currentApplicationAddress.assetBalance(asa),
      xferAsset: asa,
      assetCloseTo: this.txn.sender,
    });

    /// Send MBR to the funder
    sendPayment({
      receiver: asaMbrFunder,
      amount: initialMbr - globals.currentApplicationAddress.minBalance,
      fee: 0,
    });

    if (globals.currentApplicationAddress.totalAssets === 0) this.closeAcct(creator);
  }

  @deleteApplication
  delete(): void {
    assert(!globals.currentApplicationAddress.hasBalance);
    assert(this.txn.sender === globals.creatorAddress);
  }
}

// eslint-disable-next-line no-unused-vars
class Master extends Contract {
  vaultMap = new BoxMap<Account, Application>({ defaultSize: 8 });

  @createApplication
  create(): void {
    assert(this.txn.applicationID === new Application(0));
  }

  createVault(receiver: Account, mbrPayment: PayTxn): Application {
    assert(!this.vaultMap.exists(receiver));
    assert(mbrPayment.receiver === globals.currentApplicationAddress);
    assert(mbrPayment.sender === this.txn.sender);
    assert(mbrPayment.closeRemainderTo === globals.zeroAddress);

    const preCreateMBR = globals.currentApplicationAddress.minBalance;

    /// Create the vault
    sendMethodCall<[Account, Account], void>({
      name: 'create',
      OnCompletion: 'NoOp',
      fee: 0,
      methodArgs: [receiver, this.txn.sender],
      clearStateProgram: this.app.clearStateProgram,
      approvalProgram: Vault,
      globalNumByteSlice: 2,
      globalNumUint: 1,
    });

    const vault = this.itxn.createdApplicationID;

    /// Fund the vault with account MBR
    sendPayment({
      receiver: vault.address,
      amount: globals.minBalance,
      fee: 0,
    });

    this.vaultMap.put(receiver, vault);

    // eslint-disable-next-line max-len
    assert(mbrPayment.amount === (globals.currentApplicationAddress.minBalance - preCreateMBR) + globals.minBalance);

    return vault;
  }

  verifyAxfer(receiver: Account, vaultAxfer: AssetTransferTxn, vault: Application): void {
    assert(this.vaultMap.exists(receiver));

    assert(this.vaultMap.get(receiver) === vault);
    assert(vaultAxfer.assetReceiver === vault.address);
    assert(vaultAxfer.assetCloseTo === globals.zeroAddress);
  }

  hasVault(receiver: Account): uint64 {
    return this.vaultMap.exists(receiver);
  }

  getVaultId(receiver: Account): Application {
    return this.vaultMap.get(receiver);
  }

  getVaultAddr(receiver: Account): Account {
    return this.vaultMap.get(receiver).address;
  }

  deleteVault(vault: Application, creator: Account): void {
    /// The fee needs to be 0 because all of the fees need to paid by the vault call
    /// This ensures the sender will be refunded for all fees if they are rejecting the last ASA
    assert(this.txn.fee === 0);
    assert(vault === this.vaultMap.get(this.txn.sender));

    const vaultCreator = vault.global('creator') as Account;
    assert(vaultCreator === creator);

    const preDeleteMBR = globals.currentApplicationAddress.minBalance;

    /// Call delete on the vault
    sendMethodCall<[], void>({
      applicationID: vault,
      OnCompletion: 'DeleteApplication',
      name: 'delete',
      fee: 0,
    });

    this.vaultMap.delete(this.txn.sender);

    /// Send the MBR back to the vault creator
    sendPayment({
      receiver: vaultCreator,
      amount: preDeleteMBR - globals.currentApplicationAddress.minBalance,
      fee: 0,
    });
  }
}
