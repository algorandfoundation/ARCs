/* eslint-disable no-undef */
/* eslint-disable class-methods-use-this */

import { Contract } from '../../src/lib/index';

// eslint-disable-next-line no-unused-vars
class WhitelistApplicationRegistry extends Contract {
  whitelist = new BoxMap<[Account, uint16], uint64[]>({ defaultSize: 10 });

  @createApplication
  create(): void {}

  private verifyMBRPayment(payment: PayTxn, preMBR: uint64): void {
    assert(payment.amount === this.app.address.minBalance - preMBR);
    assert(payment.receiver === this.app.address);
  }

  private sendMBRPayment(preMBR: uint64): void {
    sendPayment({
      sender: this.app.address,
      receiver: this.txn.sender,
      amount: preMBR - this.app.address.minBalance,
      fee: 0,
    });
  }

  /**
   * Add collection to whitelist box
   *
   * @param id - The id of the senders's whitelist to add the collection to
   * @param collection - The app ID of the contract for the collection
   * @param payment - The payment transaction to cover the MBR change
   *
   */
  addToWhiteList(id: uint16, collection: uint64, payment: PayTxn): void {
    const preMBR = this.app.address.minBalance;
    const key: [Account, uint16] = [this.txn.sender, id];

    if (this.whitelist.exists(key)) {
      const whitelist = this.whitelist.get(key);
      this.whitelist.delete(key);

      whitelist.push(collection);
      this.whitelist.put(key, whitelist);
    } else {
      const whitelist: uint64[] = [collection];
      this.whitelist.put(key, whitelist);
    }

    this.verifyMBRPayment(payment, preMBR);
  }

  /**
   * Sets a collection whitelist for the sender. Should only be used when adding/removing
   * more than one collection
   *
   * @param id - The id of the sender's whitelist to set
   * @param collections - Array of app IDs that signify the whitelisted collections
   *
   */
  setWhitelist(id: uint16, collections: uint64[]): void {
    const preMBR = this.app.address.minBalance;
    const key: [Account, uint16] = [this.txn.sender, id];

    this.whitelist.delete(key);

    this.whitelist.put(key, collections);

    if (preMBR > this.app.address.minBalance) {
      this.sendMBRPayment(preMBR);
    } else {
      this.verifyMBRPayment(this.txnGroup[this.txn.groupIndex - 1], preMBR);
    }
  }

  /**
   * Deletes a collection whitelist for the sender
   *
   * @param id - The id of the sender's whitelist to delete
   *
   */
  deleteWhitelist(id: uint16): void {
    const preMBR = this.app.address.minBalance;
    const key: [Account, uint16] = [this.txn.sender, id];

    this.whitelist.delete(key);

    this.sendMBRPayment(preMBR);
  }

  /**
   * Deletes a collection from a whitelist for the sender
   *
   * @param id - The id of the sender's whitelist to delete from
   * @param collection - The app ID of the contract for the collection
   * @param index - The index of the collection in the whitelist
   *
   */
  deleteFromWhitelist(id: uint16, collection: uint64, index: uint64): void {
    const preMBR = this.app.address.minBalance;
    const key: [Account, uint16] = [this.txn.sender, id];

    const whitelist = this.whitelist.get(key);
    this.whitelist.delete(key);

    const spliced = whitelist.splice(index, 1);

    this.whitelist.put(key, whitelist);

    assert(spliced[0] === collection);

    this.sendMBRPayment(preMBR);
  }
}