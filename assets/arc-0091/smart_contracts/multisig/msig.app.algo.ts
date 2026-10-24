import {
  Account,
  Contract,
  Global,
  Txn,
  assert,
  itxn,
  uint64,
} from '@algorandfoundation/algorand-typescript';

/**
 * MsigApp - A contract that extends ARC55 functionality for multisig operations
 */
export class MsigApp extends Contract {
  /**
   * Deploy a new On-Chain Msig App.
   * @param admin Address of person responsible for calling `arc55_setup`
   * @returns Msig App Application ID
   */
  createApplication(admin: Account): void {
    // Set admin - if zero address, default to the sender/creator
    if (admin !== Global.zeroAddress) {
      this.arc55_setAdmin(admin);
    } else {
      this.arc55_setAdmin(Txn.sender);
    }
  }

  /**
   * Update the application
   */
  updateApplication(): void {
    this.onlyAdmin();
  }

  /**
   * Destroy the application and return funds to creator address.
   * All transactions must be removed before calling destroy.
   */
  deleteApplication(): void {
    this.onlyAdmin();

    // Send payment to close out the app account
    itxn.payment({
      amount: 0,
      receiver: Global.creatorAddress,
      closeRemainderTo: Global.creatorAddress,
      fee: 0,
    }).submit();
  }

  // Protected helper methods (inherited from ARC55)
  protected onlyAdmin(): void {
    assert(Txn.sender === Global.creatorAddress);
  }

  protected arc55_setAdmin(newAdmin: Account): void {
    // Placeholder - implement if storing admin in state
    // this.arc55_admin.value = newAdmin;
  }
}
