import { Account, assert, BoxMap, Bytes, bytes, clone, contract, Contract, emit, GlobalState, gtxn, op, itxn, Txn, Uint64, uint64, Global } from "@algorandfoundation/algorand-typescript";
import { methodSelector, readonly, Uint8 } from '@algorandfoundation/algorand-typescript/arc4';

type TransactionGroup = {
    // nonce: transaction group identifier
    nonce: uint64,
    // index: transaction position within the group (0-9)
    index: Uint8,
    // signer_index: each signer has their own encrypted version of the transaction
    // this is necessary for encrypted transactions where admin encrypts different ciphertext for each signer using their pubkey
    // signer_index identifies which signer's encrypted version is stored
    signer_index: Uint8
};

type TransactionSignatures = {
    nonce: uint64,
    address: Account
};

    /**
     * Emitted when a new transaction is added to a transaction group
     */
    type TransactionAdded = {
        /* Transaction Group nonce */
        transactionGroup: uint64,
        /* Index of transaction within group */
        transactionIndex: Uint8
        /* Index of signer for encrypted transactions */
        signerIndex: Uint8
    };

    /**
     * Emitted when a transaction has been removed from a transaction group
     */
    type TransactionRemoved = {
        /* Transaction Group nonce */
        transactionGroup: uint64,
        /* Index of transaction within group */
        transactionIndex: Uint8
    };

    /**
     * Emitted when a new signature is added to a transaction group
     */
    type SignatureSet = {
        /* Transaction Group nonce */
        transactionGroup: uint64,
        /* Account of signature */
        signer: Account
    };

    /**
     * Emitted when a signature has been removed from a transaction group
     */
    type SignatureCleared = {
        /* Transaction Group nonce */
        transactionGroup: uint64,
        /* Account of signature */
        signer: Account
    };

//@contract({
//  stateTotals; should be set at deploy time instead of compile time
// the size of the multisig quorum is known and doesn't change per application
//})
export class ARC55 extends Contract {

    // Number of signatures requires
    arc55_threshold = GlobalState<uint64>({ initialValue: Uint64(0) });

    // Incrementing nonce for separating different groups of transactions
    arc55_nonce = GlobalState<uint64>({ initialValue: Uint64(0) });

    // Admin responsible for setup
    arc55_admin = GlobalState<Account>({});

    // Transactions
    arc55_transactions = BoxMap<TransactionGroup, bytes>({ keyPrefix: '' });

    // Signatures
    arc55_signatures = BoxMap<TransactionSignatures, bytes<64>[]>({ keyPrefix: '' });

    protected retGobalState(suffix: uint64, prefix: bytes): GlobalState<bytes> {
      const suffixBytes = op.itob(suffix)
      const key = prefix.concat(suffixBytes)
      return GlobalState<bytes>({ key })
    }

    protected arc55_indexToAddress(index: uint64): GlobalState<Account> {
      return GlobalState<Account>({ key: op.itob(index) })
    }

    protected arc55_addressCount(account: Account): GlobalState<uint64> {
      return GlobalState<uint64>({ key: account.bytes })
    }
    
    /**
     * Check the transaction sender is a signer for the multisig
     */
    protected onlySigner(): void {
      assert(this.arc55_addressCount(Txn.sender).value !== 0);
    }

    /**
     * Check the transaction sender is the admin
     */
    protected onlyAdmin(): void {
      assert(Txn.sender === Global.creatorAddress);
    }

    /**
     * Find out if the transaction sender is the admin
     * @returns True if sender is admin
     */
    protected isAdmin(): boolean {
      return Txn.sender === Global.creatorAddress;
    }


    /**
     * Retrieve the signature threshold required for the multisignature to be submitted
     * @returns Multisignature threshold
     */
    @readonly
    arc55_getThreshold(): uint64 {
      return this.arc55_threshold.value;
    }

    /**
     * Retrieves the admin address, responsible for calling arc55_setup
     * @returns Admin address
     */
    @readonly
    arc55_getAdmin(): Account {
      return this.arc55_admin.value;
    }

    /**
     *
     * @returns Next expected Transaction Group nonce
     */
    @readonly
    arc55_nextTransactionGroup(): uint64 {
      return this.arc55_nonce.value + 1;
    }

    /**
     * Retrieve a transaction from a given transaction group for a specific signer
     * @param transactionGroup Transaction Group nonce
     * @param transactionIndex Index of transaction within group
     * @param signerIndex Index of the signer (determines which encrypted version to retrieve)
     * @returns A single transaction at the specified index for the transaction group nonce and signer
     */
    @readonly
    arc55_getTransaction(transactionGroup: uint64, transactionIndex: Uint8, signerIndex: Uint8): bytes {
      const transactionBox: TransactionGroup = {
        nonce: transactionGroup,
        index: transactionIndex,
        signer_index: signerIndex
      };

      return this.arc55_transactions(transactionBox).value;
    }

    /**
     * Retrieve a list of signatures for a given transaction group nonce and address
     * @param transactionGroup Transaction Group nonce
     * @param signer Account you want to retrieve signatures for
     * @returns Array of signatures
     */
    @readonly
    arc55_getSignatures(transactionGroup: uint64, signer: Account): bytes<64>[] {
      const signatureBox: TransactionSignatures = {
        nonce: transactionGroup,
        address: signer
      };

      return this.arc55_signatures(signatureBox).value;
    }

    /**
     * Find out which address is at this index of the multisignature
     * @param index Account at this index of the multisignature
     * @returns Account at index
     */
    @readonly
    arc55_getSignerByIndex(index: uint64): Account {
      return this.arc55_indexToAddress(index).value;
    }

    /**
     * Check if an address is a member of the multisignature
     * @param address Account to check is a signer
     * @returns True if address is a signer
     */
    @readonly
    arc55_isSigner(address: Account): boolean {
      return this.arc55_addressCount(address).value !== 0;
    }

    /**
     * Calculate the minimum balance requirement for storing a signature
     * @param signaturesSize Size (in bytes) of the signatures to store
     * @returns Minimum balance requirement increase
     */
    @readonly
    arc55_mbrSigIncrease(signaturesSize: uint64): uint64 {
      const currentBalance: uint64 = op.balance(Global.currentApplicationAddress);
      const minimumBalance: uint64 = op.minBalance(Global.currentApplicationAddress);

      // signatureBox costs:
      // + Name: uint64 + address = 8 + 32 = 40
      // + Body: abi + bytes = 2 + signaturesSize
      const mbrSigRequired: uint64 = (2500) + (400 * (40 + 2 + signaturesSize));

      const newMinimumBalance: uint64 = minimumBalance + mbrSigRequired;
      if (currentBalance >= newMinimumBalance) {
        return 0;
      }

      return newMinimumBalance - currentBalance;
    }

    /**
     * Calculate the minimum balance requirement for storing a transaction
     * With signer_index included, each signer stores a different encrypted version
     * @param transactionSize Size (in bytes) of the transaction to store
     * @returns Minimum balance requirement increase
     */
    @readonly
    arc55_mbrTxnIncrease(transactionSize: uint64): uint64 {
      const currentBalance: uint64 = op.balance(Global.currentApplicationAddress);
      const minimumBalance: uint64 = op.minBalance(Global.currentApplicationAddress);

      // transactionBox costs:
      // + Name: uint64 + Uint8 + Uint8 = 8 + 1 + 1 = 10
      // + Body: transactionSize
      const mbrTxnRequired: uint64 = (2500) + (400 * (10 + transactionSize));

      const newMinimumBalance: uint64 = minimumBalance + mbrTxnRequired;
      if (currentBalance >= newMinimumBalance) {
        return 0;
      }

      const result: uint64 = newMinimumBalance - currentBalance;
      return result;
    }


    /**
     * Set the admin address for the On-Chain Msig App
     * @param newAdmin New admin address
     */
    protected arc55_setAdmin(newAdmin: Account): void {
      this.arc55_admin.value = newAdmin;
    }


    /**
     * Setup On-Chain Msig App. This can only be called whilst no transaction groups have been created.
     * @param threshold Initial multisig threshold, must be greater than 0
     * @param addresses Array of addresses that make up the multisig
     */
    arc55_setup(
      threshold: Uint8,
      addresses: Account[]
    ): void {
      assert(!this.arc55_nonce.value);
      this.onlyAdmin();

      assert(threshold.asUint64() > 0);
      // set arc55_threshold, arc55_nonce, and arc55_admin
      this.arc55_threshold.value = Uint64(threshold.asUint64());
      this.arc55_nonce.value = 0;
      this.arc55_admin.value = Txn.sender;

      // If any indexes were previously set, remove all
      // previous addresses before deleting the indexes
      let pIndex: uint64 = 0;
      while (this.arc55_indexToAddress(pIndex).hasValue) {
        const address = this.arc55_indexToAddress(pIndex).value;
        // In puya-ts we need to assert the key exists before deleting
        if (this.arc55_addressCount(address).hasValue) {
          this.arc55_addressCount(address).delete();
          this.arc55_indexToAddress(pIndex).delete();
        }

          pIndex += 1;
      }

      // Store all new addresses as indexes and counts
      let nIndex: uint64 = 0;
      let address: Account;
      while (nIndex < addresses.length) {
        address = addresses[nIndex];

        // Store multisig index as key with address as value
        this.arc55_indexToAddress(nIndex).value = address;

        // Store address as key and counter as value,
        // this is for ease of authentication
        // IF first time, set to 1, else increment by 1
        if (!this.arc55_addressCount(address).hasValue) {
          this.arc55_addressCount(address).value = 0;
        }

        this.arc55_addressCount(address).value += 1;

        nIndex += 1;
      }
    }

    /**
     * Generate a new transaction group nonce for holding pending transactions
     * @returns transactionGroup Transaction Group nonce
     */
    arc55_newTransactionGroup(): uint64 {
      if (!this.isAdmin()) {
        this.onlySigner();
      }

      const n = this.arc55_nextTransactionGroup();
      this.arc55_nonce.value = n;

      return n;
    }

    /**
     * Add a transaction to an existing group. Only one transaction should be included per call.
     * For encrypted transactions, the admin encrypts the same transaction differently for each signer,
     * so each signer stores their own version.
     * @param costs Minimum Balance Requirement for associated box storage costs
     * @param transactionGroup Transaction Group nonce
     * @param index Transaction position within atomic group to add
     * @param signerIndex Index of the signer (determines storage location for encrypted version)
     * @param transaction Transaction to add (encrypted or unencrypted)
     */
    arc55_addTransaction(
      costs: gtxn.PaymentTxn,
      transactionGroup: uint64,
      index: Uint8,
      signerIndex: Uint8,
      transaction: bytes
    ): void {
      if (!this.isAdmin()) {
        this.onlySigner();
      }

      assert(transactionGroup);
      assert(transactionGroup <= this.arc55_nonce.value);

      const transactionBox: TransactionGroup = {
        nonce: transactionGroup,
        index: index,
        signer_index: signerIndex
      };

      // If there are additional addTransactionContinued transactions
      // following this transaction, concatenate all additional data.
      let transactionData = transaction;
      let groupPosition: uint64 = Txn.groupIndex + 1;
      if (groupPosition < Global.groupSize) {
        do {
          if (
            gtxn.ApplicationCallTxn(groupPosition).appId === Txn.applicationId
          && gtxn.ApplicationCallTxn(groupPosition).appArgs(0) === methodSelector("arc55_addTransactionContinued(byte[])void")
          ) {
            transactionData = transactionData.concat(gtxn.ApplicationCallTxn(groupPosition).appArgs(1));
          }
          groupPosition += 1;
        } while (groupPosition < Global.groupSize);
      }

      const mbrTxnIncrease = this.arc55_mbrTxnIncrease(transactionData.length);
      costs.receiver === Global.currentApplicationAddress
      costs.amount >= mbrTxnIncrease

      this.arc55_transactions(transactionBox).value = transactionData;

      emit<TransactionAdded>({
        transactionGroup: transactionGroup,
        transactionIndex: index,
        signerIndex: signerIndex
      })
    }

    arc55_addTransactionContinued(
      transaction: bytes
    ): void {
      if (!this.isAdmin()) {
        this.onlySigner();
      }
    }

    /**
     * Remove transaction from the app for a specific signer. The MBR associated with the transaction will be returned to the transaction sender.
     * @param transactionGroup Transaction Group nonce
     * @param index Transaction position within atomic group to remove
     * @param signerIndex Index of the signer whose encrypted version to remove
     */
    arc55_removeTransaction(
      transactionGroup: uint64,
      index: Uint8,
      signerIndex: Uint8
    ): void {
      if (!this.isAdmin()) {
        this.onlySigner();
      }

      const transactionBox: TransactionGroup = {
        nonce: transactionGroup,
        index: index,
        signer_index: signerIndex
      };

      const txnLength = this.arc55_transactions(transactionBox).length;
      this.arc55_transactions(transactionBox).delete();

      // transactionBox costs:
      // + Name: uint64 + Uint8 + Uint8 = 8 + 1 + 1 = 10
      // + Body: txnLength
      const mbrTxnDecrease: uint64 = (2500) + (400 * (10 + txnLength));


      itxn.payment({
        receiver: Txn.sender,
        amount: mbrTxnDecrease
      }).submit();


      emit<TransactionRemoved>({
        transactionGroup: transactionGroup,
        transactionIndex: index
      })

    }

    /**
     * Set signatures for a particular transaction group. Signatures must be included as an array of byte-arrays
     * @param costs Minimum Balance Requirement for associated box storage costs: (2500) + (400 * (40 + signatures.length))
     * @param transactionGroup Transaction Group nonce
     * @param signatures Array of signatures
     */
    arc55_setSignatures(
      costs: gtxn.PaymentTxn,
      transactionGroup: uint64,
      signatures: bytes<64>[]
    ): void {
      this.onlySigner();

      const mbrSigIncrease = this.arc55_mbrSigIncrease(signatures.length * 64);

      assert(costs.receiver === Global.currentApplicationAddress)
      assert(costs.amount >= mbrSigIncrease)

      const signatureBox: TransactionSignatures = {
        nonce: transactionGroup,
        address: Txn.sender
      };

      // clone signatures to avoid referencing txns directly
      this.arc55_signatures(signatureBox).value = clone(signatures);

      emit<SignatureSet>({
        transactionGroup: transactionGroup,
        signer: Txn.sender
      })
    }

    /**
     * Clear signatures for an address. Be aware this only removes it from the current state of the ledger, and indexers will still know and could use your signature
     * @param transactionGroup Transaction Group nonce
     * @param address Account whose signatures to clear
     */
    arc55_clearSignatures(
      transactionGroup: uint64,
      address: Account
    ): void {
      if (!this.isAdmin()) {
        this.onlySigner();
      }

      const signatureBox: TransactionSignatures = {
        nonce: transactionGroup,
        address: address
      };

      const sigLength: uint64 = this.arc55_signatures(signatureBox).length;
      this.arc55_signatures(signatureBox).delete();

      // signatureBox costs:
      // + Name: uint64 + address = 8 + 32 = 40
      // + Body: sigLength
      const mbrSigDecrease: uint64 = (2500) + (400 * (40 + sigLength));

      itxn.payment({
        receiver: Txn.sender,
        amount: mbrSigDecrease
      }).submit();

      // Emit event
      emit<SignatureCleared>({
        transactionGroup: transactionGroup,
        signer: address
      })

    }
}
