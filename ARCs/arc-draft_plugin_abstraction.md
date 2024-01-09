---
arc: <to be assigned>
title: Plugin-Based Account Abstraction
description: An extendable standard for account abstraction using stateful applciations
author: Joe Polny (@joe-p), Kyle Breeding aka krby.algo (@kylebeee)
discussions-to: <URL>
status: Draft
type: Standards Track
category: ARC
created: 2024-01-08
requires: 4
---

## Abstract
This ARC proposes a standard for using stateful applications and rekey transactions to enable account abstraction on Algorand. The abstracted account is controlled by a single stateful application which is the auth address of the abstracted account. Other applications can be used as plugin to provide additional functionality to the abstracted account.

## Motivation
Manually signing transactions for every dApp interaction can be rather fatiguing for the end-user, which results in a frustrating UX. In some cases, it makes specfific app designs that require a lot of transactions borderline impossible without delegation or an escrow account.

Another common point of friction for end-users in the Algorand ecosystem is ASA opt-in transactions. This is a paticularly high point of friction for onboarding new accounts since they must be funded and then initiate a transaction. This standard can be used to allow mass creation of non-custodial accounts and trigger opt-ins on their behalf.

## Specification

### Definitions
**Abstracted Account** - An account that has functionality beyond a typical keypair-based account.

**Abstracted Account App** - The stateuful application used to control the abstracted account. This app's address is the `auth-addr` of the abstracted account.

**Plugin** - An additional application that adds functionality to the **Abstracted Account App** (and thus the **Abstracted Account**).

**Admin** - An account, sepererate from the **Abstracted Account**, that controls the **Abstracted Account App**. In patiular, this app can initiate rekeys, add plugins, and transfer admin.


### ARC4 Methods

An abstracted app account that adheres to this standard **MUST** implement the following methods

```json
  "methods": [
    {
      "name": "createApplication",
      "desc": "Create an abstracted account application",
      "args": [
        {
          "name": "address",
          "type": "address",
          "desc": "The address of the abstracted account. If zeroAddress, then the address of the contract account will be used"
        },
        {
          "name": "admin",
          "type": "address",
          "desc": "The admin for this app"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "verifyAppAuthAddr",
      "desc": "Verify the abstracted account is rekeyed to this app",
      "args": [],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "rekeyTo",
      "desc": "Rekey the abstracted account to another address. Primarily useful for rekeying to an EOA.",
      "args": [
        {
          "name": "addr",
          "type": "address",
          "desc": "The address to rekey to"
        },
        {
          "name": "flash",
          "type": "bool",
          "desc": "Whether or not this should be a flash rekey. If true, the rekey back to the app address must done in the same txn group as this call"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "rekeyToPlugin",
      "desc": "Temporarily rekey to an approved plugin app address",
      "args": [
        {
          "name": "plugin",
          "type": "application",
          "desc": "The app to rekey to"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "changeAdmin",
      "desc": "Change the admin for this app",
      "args": [
        {
          "name": "newAdmin",
          "type": "account",
          "desc": "The new admin"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "addPlugin",
      "desc": "Add an app to the list of approved plugins",
      "args": [
        {
          "name": "app",
          "type": "application",
          "desc": "The app to add"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "removePlugin",
      "desc": "Remove an app from the list of approved plugins",
      "args": [
        {
          "name": "app",
          "type": "application",
          "desc": "The app to remove"
        }
      ],
      "returns": {
        "type": "void"
      }
    }
  ]
```

### Plugins
TODO

### Wallet and dApp Support
TODO

## Rationale

### App vs Logic Sig
There have similar propsoals for reducing end-user friction, such as [ARC47](./arc-0047.md) which enables safer usage of delegated logic signatures. The major downside of logic signatures is that they are not 

### Plugins
TODO

## Backwards Compatibility
Existing Algorand accounts can transition to an abstracted account by creating a new abstracted account application and setting the address to their current address. This requires them to create a new account to act as the admin.

End-users can use an abstracted account with any dApp provided they rekey the account to an externally owned account.

## Test Cases

TODO: Some functional tests are in this repo https://github.com/joe-p/account_abstraction.git

## Reference Implementation
https://github.com/joe-p/account_abstraction.git

TODO: Migrate to ARC repo, but waiting until development has settled.

## Security Considerations
By adding a plugin to an abstracted account, that plugin can be called by anyone which will initiate a rekey from the abstracted account to the plugin app address. While the plugin must rekey back, there is no safeguards on what the plugin does when it has authority over the abstracted account. As such, extreme diligance must be taken by the end-user to ensure they are adding safe and/or trusted plugins.

## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
