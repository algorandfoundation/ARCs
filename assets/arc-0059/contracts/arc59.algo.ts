/* eslint-disable max-classes-per-file */

// eslint-disable-next-line import/no-unresolved, import/extensions
import { Contract } from '@algorandfoundation/tealscript';

type SendAssetInfo = {
  /**
   * The total number of inner transactions required to send the asset through the router.
   * This should be used to add extra fees to the app call
   */
  itxns: uint64;
  /** The total MBR the router needs to send the asset through the router. */
  mbr: uint64;
  /** Whether the router is already opted in to the asset or not */
  routerOptedIn: boolean;
  /** Whether the receiver is already directly opted in to the asset or not */
  receiverOptedIn: boolean;
  /** The amount of ALGO the receiver would currently need to claim the asset */
  receiverAlgoNeededForClaim: uint64;
};

class ControlledAddress extends Contract {
  @allow.create('DeleteApplication')
  new(): Address {
    sendPayment({
      rekeyTo: this.txn.sender,
    });

    return this.app.address;
  }
}

export class ARC59 extends Contract {
  inboxes = BoxMap<Address, Address>();

  /**
   * Deploy ARC59 contract
   *
   */
  createApplication(): void {}

  /**
   * Opt the ARC59 router into the ASA. This is required before this app can be used to send the ASA to anyone.
   *
   * @param asa The ASA to opt into
   */
  arc59_optRouterIn(asa: AssetID): void {
    sendAssetTransfer({
      assetReceiver: this.app.address,
      assetAmount: 0,
      xferAsset: asa,
    });
  }

  /**
   * Gets the existing inbox for the receiver or creates a new one if it does not exist
   *
   * @param receiver The address to get or create the inbox for
   * @returns The inbox address
   */
  arc59_getOrCreateInbox(receiver: Address): Address {
    if (this.inboxes(receiver).exists) return this.inboxes(receiver).value;

    const inbox = sendMethodCall<typeof ControlledAddress.prototype.new>({
      onCompletion: OnCompletion.DeleteApplication,
      approvalProgram: ControlledAddress.approvalProgram(),
      clearStateProgram: ControlledAddress.clearProgram(),
    });

    this.inboxes(receiver).value = inbox;

    return inbox;
  }

  /**
   *
   * @param receiver The address to send the asset to
   * @param asset The asset to send
   *
   * @returns Returns the following information for sending an asset:
   * The number of itxns required, the MBR required, whether the router is opted in, whether the receiver is opted in,
   * and how much ALGO the receiver would need to claim the asset
   */
  arc59_getSendAssetInfo(receiver: Address, asset: AssetID): SendAssetInfo {
    const routerOptedIn = this.app.address.isOptedInToAsset(asset);
    const receiverOptedIn = receiver.isOptedInToAsset(asset);
    const info: SendAssetInfo = {
      itxns: 1,
      mbr: 0,
      routerOptedIn: routerOptedIn,
      receiverOptedIn: receiverOptedIn,
      receiverAlgoNeededForClaim: 0,
    };

    if (receiverOptedIn) return info;

    const algoNeededToClaim = receiver.minBalance + globals.assetOptInMinBalance + globals.minTxnFee;

    // Determine how much ALGO the receiver needs to claim the asset
    if (receiver.balance < algoNeededToClaim) {
      info.receiverAlgoNeededForClaim += algoNeededToClaim - receiver.balance;
    }

    // Add mbr and transaction for opting the router in
    if (!routerOptedIn) {
      info.mbr += globals.assetOptInMinBalance;
      info.itxns += 1;
    }

    if (!this.inboxes(receiver).exists) {
      // Two itxns to create inbox (create + rekey)
      // One itxns to send MBR
      // One itxn to opt in
      info.itxns += 4;

      // Calculate the MBR for the inbox box
      const preMBR = globals.currentApplicationAddress.minBalance;
      this.inboxes(receiver).value = globals.zeroAddress;
      const boxMbrDelta = globals.currentApplicationAddress.minBalance - preMBR;
      this.inboxes(receiver).delete();

      // MBR = MBR for the box + min balance for the inbox + ASA MBR
      info.mbr += boxMbrDelta + globals.minBalance + globals.assetOptInMinBalance;

      return info;
    }

    const inbox = this.inboxes(receiver).value;

    if (!inbox.isOptedInToAsset(asset)) {
      // One itxn to opt in
      info.itxns += 1;

      if (!(inbox.balance >= inbox.minBalance + globals.assetOptInMinBalance)) {
        // One itxn to send MBR
        info.itxns += 1;

        // MBR = ASA MBR
        info.mbr += globals.assetOptInMinBalance;
      }
    }

    return info;
  }

  /**
   * Send an asset to the receiver
   *
   * @param receiver The address to send the asset to
   * @param axfer The asset transfer to this app
   * @param additionalReceiverFunds The amount of ALGO to send to the receiver/inbox in addition to the MBR
   *
   * @returns The address that the asset was sent to (either the receiver or their inbox)
   */
  arc59_sendAsset(axfer: AssetTransferTxn, receiver: Address, additionalReceiverFunds: uint64): Address {
    verifyAssetTransferTxn(axfer, {
      assetReceiver: this.app.address,
    });

    // If the receiver is opted in, send directly to their account
    if (receiver.isOptedInToAsset(axfer.xferAsset)) {
      sendAssetTransfer({
        assetReceiver: receiver,
        assetAmount: axfer.assetAmount,
        xferAsset: axfer.xferAsset,
      });

      if (additionalReceiverFunds !== 0) {
        sendPayment({
          receiver: receiver,
          amount: additionalReceiverFunds,
        });
      }

      return receiver;
    }

    const inboxExisted = this.inboxes(receiver).exists;
    const inbox = this.arc59_getOrCreateInbox(receiver);

    if (additionalReceiverFunds !== 0) {
      sendPayment({
        receiver: inbox,
        amount: additionalReceiverFunds,
      });
    }

    if (!inbox.isOptedInToAsset(axfer.xferAsset)) {
      let inboxMbrDelta = globals.assetOptInMinBalance;
      if (!inboxExisted) inboxMbrDelta += globals.minBalance;

      // Ensure the inbox has enough balance to opt in
      if (inbox.balance < inbox.minBalance + inboxMbrDelta) {
        sendPayment({
          receiver: inbox,
          amount: inboxMbrDelta,
        });
      }

      // Opt the inbox in
      sendAssetTransfer({
        sender: inbox,
        assetReceiver: inbox,
        assetAmount: 0,
        xferAsset: axfer.xferAsset,
      });
    }

    // Transfer the asset to the inbox
    sendAssetTransfer({
      assetReceiver: inbox,
      assetAmount: axfer.assetAmount,
      xferAsset: axfer.xferAsset,
    });

    return inbox;
  }

  /**
   * Claim an ASA from the inbox
   *
   * @param asa The ASA to claim
   */
  arc59_claim(asa: AssetID): void {
    const inbox = this.inboxes(this.txn.sender).value;

    sendAssetTransfer({
      sender: inbox,
      assetReceiver: this.txn.sender,
      assetAmount: inbox.assetBalance(asa),
      xferAsset: asa,
      assetCloseTo: this.txn.sender,
    });

    sendPayment({
      sender: inbox,
      receiver: this.txn.sender,
      amount: inbox.balance - inbox.minBalance,
    });
  }

  /**
   * Reject the ASA by closing it out to the ASA creator. Always sends two inner transactions.
   * All non-MBR ALGO balance in the inbox will be sent to the caller.
   *
   * @param asa The ASA to reject
   */
  arc59_reject(asa: AssetID) {
    const inbox = this.inboxes(this.txn.sender).value;

    sendAssetTransfer({
      sender: inbox,
      assetReceiver: asa.creator,
      assetAmount: inbox.assetBalance(asa),
      xferAsset: asa,
      assetCloseTo: asa.creator,
    });

    sendPayment({
      sender: inbox,
      receiver: this.txn.sender,
      amount: inbox.balance - inbox.minBalance,
    });
  }

  /**
   * Get the inbox address for the given receiver
   *
   * @param receiver The receiver to get the inbox for
   *
   * @returns Zero address if the receiver does not yet have an inbox, otherwise the inbox address
   */
  arc59_getInbox(receiver: Address): Address {
    return this.inboxes(receiver).exists ? this.inboxes(receiver).value : globals.zeroAddress;
  }

  /** Claim any extra algo from the inbox */
  arc59_claimAlgo() {
    const inbox = this.inboxes(this.txn.sender).value;

    assert(inbox.balance - inbox.minBalance !== 0);

    sendPayment({
      sender: inbox,
      receiver: this.txn.sender,
      amount: inbox.balance - inbox.minBalance,
    });
  }
}
