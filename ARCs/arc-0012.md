---
arc: 12
title: Claimable ASA from vault application
description: A smart signature contract account that can receive & disburse claimable Algorand Smart Assets (ASA) to an intended recipient account.
author: Brian Whippo (@silentrhetoric), Joe Polny (@joe-p)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/131
status: Withdrawn
type: Standards Track
category: ARC
created: 2022-09-05
withdrawal-reason: Not used, will be replaced by ARC-59
---

# Claimable Algorand Standard Assets (ASAs) from vault application

## Abstract

The goal of this standard is to establish a standard in the Algorand ecosytem by which ASAs can be sent to an intended receiver even if their account is not opted in to the ASA.

A on-chain application, called a vault, will be used to custody assets on behalf of a given user, with only that user being able to withdraw assets. A master application will use box storage to keep track of the vault for any given Algorand account.

If integrated into ecosystem technologies including wallets, epxlorers, and dApps, this standard can provide enhanced capabilities around ASAs which are otherwise strictly bound at the protocol level to require opting in to be received. This also enables the ability to "burn" ASAs by sending them to the vault associated with the global Zero Address.

## Motivation

Algorand requires accounts to opt in to receive any ASA, a fact which simultaneously:

1. Grants account holders fine-grained control over their holdings by allowing them to select which assets to allow and preventing receipt of unwanted tokens.
2. Frustrates users and developers when accounting for this requirement especially since other blockchains do not have this requirement.

This ARC lays out a new way to navigate the ASA opt in requirement.

### Contemplated Use Cases

The following use cases help explain how this capability can enhance the possibilities within the Algorand ecosystem.

#### Airdrops

An ASA creator who wants to send their asset to a set of accounts faces the challenge of needing their intended receivers to opt in to the ASA ahead of time, which requires non-trivial communication efforts and precludes the possibility of completing the airdrop as a surprise.  This claimable ASA standard creates the ability to send an airdrop out to individual addresses so that the receivers can opt in and claim the asset at their convenience--or not, if they so choose.

#### Reducing New User On-boarding Friction

An application operator who wants to on-board users to their game or business may want to reduce the friction of getting people started by decoupling their application on-boarding process from the process of funding a non-custodial Algorand wallet, if users are wholly new to the Algorand ecosystem.  As long as the receiver's address is known, an ASA can be sent to them ahead of them having ALGOs in their wallet to cover the minimum balance requirement and opt in to the asset.

#### Token Burning

Similarly to any regular account, the global Zero Address also has a corresponding vault to which one can send a quantity of any ASA to effectively "burn" it, rendering it lost forever.  No one controls the Zero Address, so while it cannot opt into any ASA to receive it directly, it also cannot make any claims from its corresponding vault, which thus functions as an UN-claimable ASAs purgatory account.  By utilizing this approach, anyone can verifiably and irreversibly take a quantity of any ASA out of circulation forever.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.


> Comments like this are non-normative.

### Definitions

- **Claimable ASA**: An Algorand Standard Asset (ASA) which has been transferred to a vault following the standard set forth in this proposal such that only the intended receiver account can claim it at their convenience.
- **Vaultt**: An Algorand application used to hold claimable ASAs for a given account.
- **Master**: An Algorand application used to keep track of all of the vaults created for Algorand accounts.
- **dApp**: A decentralized application frontend, interpreted here to mean an off-chain frontend (a webapp, native app, etc.) that interacts with applications on the blockchain.
- **Explorer**: An off-chain application that allows browsing the blockchain, showing details of transactions.
- **Wallet**: An off-chain application that stores secret keys for on-chain accounts and can display and sign transactions for these accounts.
- **Mainnet ID**: The ID for the application that should be called upon claiming an asset on mainnet
- **Testnet ID**: The ID for the application that should be called upoin claiming an asset on testnet
- **Minimum Balance Requirement (MBR)**: The minimum amount of Algos which must be held by an account on the ledger, which is currently 0.1A + 0.1A per ASA opted into.

### TEAL Smart Contracts

There are two smart contracts being utilized: The [vault](../assets/arc-0012/vault.teal) and the [master](../assets/arc-0012/master.teal).

#### Vault

##### Storage

| Type   | Key        | Value          | Description                                           |
| ------ | ---------- | -------------- | ----------------------------------------------------- |
| Global | “creator”  | Account        | The account that funded the creation of the vault     |
| Global | “master”   | Application ID | The application ID that created the vault             |
| Global | “receiver” | Account        | The account that can claim/reject ASAs from the vault |
| Box    | Asset ID   | Account        | The account that funded the MBR for the given ASA     |

##### Methods

###### Opt-In
* Opts vault into ASA
* Creates box: ASA -> “funder”
  * “funder” being the account that initiates the opt-in
  * “funder” is the one covering the ASA MBR

###### Claim
* Transfers ASA from vault to “receiver”
* Deletes box: ASA -> “funder”
* Returns ASA and box MBR to “funder”

###### Reject
* Sends ASA to ASA creator
* Refunds rejector all fees incurred (thus rejecting is free)
* Deletes box: ASA -> “funder”
* Remaining balance sent to fee sink

#### Master

##### Storage

| Type | Key     | Value          | Description                     |
| ---- | ------- | -------------- | ------------------------------- |
| Box  | Account | Application ID | The vault for the given account |

##### Methods

###### Create Vault
* Creates a vault for a given account (“receiver”)
* Creates box: “receiver” -> vault ID
* App/box MBR funded by vault creator

###### Delete Vault
* Deletes vault app
* Deletes box: “receiver” -> vault ID
* App.box MBR returned to vault creator

###### Verify Axfer
* Verifies asset is going to correct vault for “receiver”

###### getVaultID
* Returns vault ID for “receiver”
* Fails if “receiver” does not have vault

###### getVaultAddr
* Returns vault address for “receiver”
* Fails if “receiver” does not have vault

###### hasVault
* Determines if “receiver” has a vault

## Rationale

This design was created to offer a standard mechanism by which wallets, explorers, and dapps could enable users to send, receive, and find claimable ASAs without requiring any changes to the core protocol.

## Backwards Compatibility

This ARC makes no changes to the consensus protocol and creates no backwards compatibility issues.

## Reference Implementation

### Source code

* <a href="https://github.com/algorandfoundation/ARCs/tree/main/assets/arc-0012/contracts">Contracts</a>
* <a href="https://github.com/algorandfoundation/ARCs/tree/main/assets/arc-0012/arc12-sdk">TypeScript SDK</a>


## Security Considerations

Both applications (The vault and the master have not been audited)

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
