---
arc: 10
title: Algorand Wallet Reach Minimum Requirements
description: Minimum requirements for Reach to function with a given wallet.
author: DanBurton (@DanBurton)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/52
status: Deprecated
type: Standards Track
category: Interface
created: 2021-08-09
---

# Algorand Wallet Reach Minimum Requirements

## Abstract

An amalgamation of APIs which comprise the minimum requirements for Reach to be able to function correctly with a given wallet.

## Specification

A group of related functions:

* `enable` (**REQUIRED**)
* `enableNetwork` (**OPTIONAL**)
* `enableAccounts` (**OPTIONAL**)
* `signAndPostTxns` (**REQUIRED**)
* `getAlgodv2Client` (**REQUIRED**)
* `getIndexerClient` (**REQUIRED**)
* `signTxns` (**OPTIONAL**)
* `postTxns` (**OPTIONAL**)

* `enable`: as specified in [ARC-0006](./arc-0006.md#interface-enablefunction).
* `signAndPostTxns`: as specified in [ARC-0008](./arc-0008.md#interface-signandposttxnsfunction).
* `getAlgodv2Client` and `getIndexerClient`: as specified in [ARC-0009](./arc-0009.md#specification).
* `signTxns`: as specified in [ARC-0005](./arc-0005.md#interface-signtxnsfunction) / [ARC-0001](./arc-0001.md#interface-signtxnsfunction).
* `postTxns`: as specified in [ARC-0007](./arc-0007.md#interface-posttxnsfunction).

There are additional semantics for using these functions together.

### Semantic Requirements

* `enable` **SHOULD** be called before calling the other functions and upon refresh of the dApp.
* Calling `enableNetwork` and then `enableAccounts` **MUST** be equivalent to calling `enable`.
* If used instead of `enable`: `enableNetwork` **SHOULD** be called before `enableAccounts` and `getIndexerClient`. Both `enableNetwork` and `enableAccounts` **SHOULD** be called before the other functions.
* If `signAndPostTxns`, `getAlgodv2Client`, `getIndexerClient`, `signTxns`, or `postTxns` are called before `enable` (or `enableAccounts`), they **SHOULD** throw an error object with property `code=4202`. (See Error Standards in [ARC-0001](arc-0001.md#error-standards)).
* `getAlgodv2Client` and `getIndexerClient` **MUST** return connections to the network indicated by the `network` result of `enable`.
* `signAndPostTxns` **MUST** post transactions to the network indicated by the `network` result of `enable`
* The result of `getAlgodv2Client` **SHOULD** only be used to query the network. `postTxns` (if available) and `signAndPostTxns` **SHOULD** be used to send transactions to the network. The `Algodv2Client` object **MAY** be modified to throw exceptions if the caller tries to use it to post transactions.
* `signTxns` and `postTxns` **MAY** or **MAY NOT** be provided. When one is provided, they both **MUST** be provided. In addition, `signTxns` **MAY** display a warning that the transactions are returned to the dApp rather than posted directly to the blockchain.

### Additional requirements regarding LogicSigs

`signAndPostTxns` must also be able to handle logic sigs, and more generally transactions signed by the DApp itself.
In case of logic sigs, callers are expected to sign the logic sig by themselves, rather than expecting the wallet to do so on their behalf.
To handle these cases, we adopt and extend the [ARC-0001](./arc-0001.md#interface-wallettransaction) format for `WalletTransaction`s that do not need to be signed:

```json
{
  "txn": "...",
  "signers": [],
  "stxn": "..."
}
```

* `stxn` is a `SignedTxnStr`, as specified in [ARC-0007](./arc-0007.md#string-specification-signedtxnstr).
* For production wallets, `stxn` **MUST** be checked to match `txn`, as specified in [ARC-0001](./arc-0001.md#semantic-and-security-requirements).

`signAndPostTxns` **MAY** reject when none of the transactions need to be signed by the user.

## Rationale

In order for a wallet to be useable by a DApp, it must support features for account discovery, signing and posting transactions, and querying the network.

To whatever extent possible, the end users of a DApp should be empowered to select their own wallet, accounts, and network to be used with the DApp.
Furthermore, said users should be able to use their preferred network node connection, without exposing their connection details and secrets (such as endpoint URLs and API tokens) to the DApp.

The APIs presented in this document and related documents are sufficient to cover the needed functionality, while protecting user choice and remaining compatible with best security practices.
Most DApps indeed always need to post transactions immediately after signing.
`signAndPostTxns` allows this goal without revealing the signed transactions to the DApp, which prevents surprises to the user: there is no risk the DApp keeps in memory the transactions and post it later without the user knowing it (either to achieve a malicious goal such as forcing double spending, or just because the DApp has a bug).
However, there are cases where `signTxns` and `postTxns` need to be used: for example when multiple users need to coordinate to sign an atomic transfer.

## Reference Implementation

```js
async function main(wallet) {

  // Account discovery
  const enabled = await wallet.enable({genesisID: 'testnet-v1.0'});
  const from = enabled.accounts[0];

  // Querying
  const algodv2 = new algosdk.Algodv2(await wallet.getAlgodv2Client());
  const suggestedParams = await algodv2.getTransactionParams().do();
  const txns = makeTxns(from, suggestedParams);

  // Sign and post
  const res = await wallet.signAndPost(txns);
  console.log(res);

};
```

Where `makeTxns` is comparable to what is seen in [ARC-0001](./arc-0001.md#reference-implementation)'s sample code.

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
