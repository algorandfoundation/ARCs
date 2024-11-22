---
arc: 20
title: Smart ASA
description: An ARC for an ASA controlled by an Algorand Smart Contract
author: Cosimo Bassi (@cusma), Adriano Di Luzio (@aldur), John Jannotti (@jannotti)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/109
status: Final
type: Standards Track
category: Interface
sub-category: Asa
created: 2022-04-27
requires: 4, 22
---

## Abstract

A "Smart ASA" is an Algorand Standard Asset (ASA) controlled by a Smart Contract
that exposes methods to create, configure, transfer, freeze, and destroy the
asset.

This ARC defines the ABI interface of such Smart Contract, the required
metadata, and suggests a reference implementation.

## Motivation

The Algorand Standard Asset (ASA) is an excellent building block for on-chain
applications. It is battle-tested and widely supported by SDKs, wallets, and
dApps.

However, the ASA lacks in flexibility and configurability. For instance, once
issued it can't be re-configured (its unit name, decimals, maximum supply).
Also, it is freely transferable (unless frozen). This prevents developers from
specifying additional business logic to be checked while transferring it (think
of royalties or vesting https://en.wikipedia.org/wiki/Vesting).

Enforcing transfer conditions require freezing the asset and transferring it
through a clawback operation -- which results in a process that is opaque to
users and wallets and a bad experience for the users.

The Smart ASA defined by this ARC extends the ASA to increase its expressiveness
and its flexibility. By introducing this as a standard both developers, users
(marketplaces, wallets, dApps, etc.) and SDKs can confidently and consistently
recognize Smart ASAs and adjust their flows and user experiences accordingly.

## Specification

The keywords "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be
interpreted as described in RFC
2119 https://datatracker.ietf.org/doc/html/rfc2119.

The following sections describe:

- The ABI interface for a controlling Smart Contract (the Smart Contract that
  controls a Smart ASA).
- The metadata required to denote a Smart ASA and define the association between
  an ASA and its controlling Smart Contract.

### ABI Interface

The ABI interface specified here draws inspiration from the transaction
reference https://developer.algorand.org/docs/get-details/asa/#asset-functions
of an Algorand Standard Asset (ASA).

To provide a unified and familiar interface between the Algorand Standard Asset
and the Smart ASA, method names and parameters have been adapted to the ABI
types but left otherwise unchanged.

#### Asset Creation

```json
{
  "name": "asset_create",
  "args": [
    { "type": "uint64", "name": "total" },
    { "type": "uint32", "name": "decimals" },
    { "type": "bool", "name": "default_frozen" },
    { "type": "string", "name": "unit_name" },
    { "type": "string", "name": "name" },
    { "type": "string", "name": "url" },
    { "type": "byte[]", "name": "metadata_hash" },
    { "type": "address", "name": "manager_addr" },
    { "type": "address", "name": "reserve_addr" },
    { "type": "address", "name": "freeze_addr" },
    { "type": "address", "name": "clawback_addr" }
  ],
  "returns": { "type": "uint64" }
}
```

Calling `asset_create` creates a new Smart ASA and returns the identifier of the
ASA. The [metadata section](#metadata) describes its required properties.

> Upon a call to `asset_create`, a reference implementation SHOULD:
>
> - Mint an Algorand Standard Asset (ASA) that MUST specify the properties
>   defined in the [metadata section](#metadata). In addition:
>   - The `manager`, `reserve` and `freeze` addresses SHOULD be set to the
>     account of the controlling Smart Contract.
>   - The remaining fields are left to the implementation, which MAY set `total`
>     to `2 ** 64 - 1` to enable dynamically increasing the max circulating
>     supply of the Smart ASA.
>   - `name` and `unit_name` MAY be set to `SMART-ASA` and `S-ASA`, to
>     denote that this ASA is Smart and has a controlling application.
> - Persist the `total`, `decimals`, `default_frozen`, etc. fields for later
>   use/retrieval.
> - Return the ID of the created ASA.
>
> It is RECOMMENDED for calls to this method to be permissioned, e.g. to only
> approve transactions issued by the controlling Smart Contract creator.

#### Asset Configuration

```json
[
  {
    "name": "asset_config",
    "args": [
      { "type": "asset", "name": "config_asset" },
      { "type": "uint64", "name": "total" },
      { "type": "uint32", "name": "decimals" },
      { "type": "bool", "name": "default_frozen" },
      { "type": "string", "name": "unit_name" },
      { "type": "string", "name": "name" },
      { "type": "string", "name": "url" },
      { "type": "byte[]", "name": "metadata_hash" },
      { "type": "address", "name": "manager_addr" },
      { "type": "address", "name": "reserve_addr" },
      { "type": "address", "name": "freeze_addr" },
      { "type": "address", "name": "clawback_addr" }
    ],
    "returns": { "type": "void" }
  },
  {
    "name": "get_asset_config",
    "readonly": true,
    "args": [{ "type": "asset", "name": "asset" }],
    "returns": {
      "type": "(uint64,uint32,bool,string,string,string,byte[],address,address,address,address)",
      "desc": "`total`, `decimals`, `default_frozen`, `unit_name`, `name`, `url`, `metadata_hash`, `manager_addr`, `reserve_addr`, `freeze_addr`, `clawback`"
    }
  }
]
```

Calling `asset_config` configures an existing Smart ASA.

> Upon a call to `asset_config`, a reference implementation SHOULD:
>
> - Fail if `config_asset` does not correspond to an ASA controlled by this smart
>   contract.
> - Succeed iff the `sender` of the transaction corresponds to the `manager_addr`
>   that was previously persisted for `config_asset` by a previous call to this
>   method or, if never caller, to `asset_create`.
> - Update the persisted `total`, `decimals`, `default_frozen`, etc. fields for later
>   use/retrieval.
>
> The business logic associated to the update of the other parameters is left to
> the implementation. An implementation that maximizes similarities with ASAs,
> SHOULD NOT allow modifying the `clawback_addr` or `freeze_addr` after they
> have been set to the special value `ZeroAddress`.
>
> The implementation MAY provide flexibility on the fields of an ASA that
> cannot be updated after initial configuration. For instance, it MAY update the
> `total` parameter to enable minting of new units or restricting the maximum
> supply; when doing so, the implementation SHOULD ensure that the updated
> `total` is not lower than the current circulating supply of the asset.

Calling `get_asset_config` reads and returns the `asset`'s configuration as specified in:

- The most recent invocation of `asset_config`; or
- if `asset_config` was never invoked for `asset`, the invocation of
  `asset_create` that originally created it.

> Upon a call to `get_asset_config`, a reference implementation SHOULD:
>
> - Fail if `asset` does not correspond to an ASA controlled by this smart
>   contract (see `asset_config`).
> - Return `total`, `decimals`, `default_frozen`, `unit_name`, `name`, `url`,
>   `metadata_hash`, `manager_addr`, `reserve_addr`, `freeze_addr`, `clawback`
>   as persisted by `asset_create` or `asset_config`.

#### Asset Transfer

```json
{
  "name": "asset_transfer",
  "args": [
    { "type": "asset", "name": "xfer_asset" },
    { "type": "uint64", "name": "asset_amount" },
    { "type": "account", "name": "asset_sender" },
    { "type": "account", "name": "asset_receiver" }
  ],
  "returns": { "type": "void" }
}
```

Calling `asset_transfer` transfers a Smart ASA.

> Upon a call to `asset_transfer`, a reference implementation SHOULD:
>
> - Fail if `xfer_asset` does not correspond to an ASA controlled by this smart
>   contract.
> - Succeed if:
>   - the `sender` of the transaction is the `asset_sender` and
>   - `xfer_asset` is not in a frozen state
>     (see [Asset Freeze below](#asset-freeze)) and
>   - `asset_sender` and `asset_receiver` are not in a frozen state
>     (see [Asset Freeze below](#asset-freeze))
> - Succeed if the `sender` of the transaction corresponds to the `clawback_addr`,
>   as persisted by the controlling Smart Contract. This enables clawback
>   operations on the Smart ASA.
>
> Internally, the controlling Smart Contract SHOULD issue a clawback inner
> transaction that transfers the `asset_amount` from `asset_sender` to
> `asset_receiver`. The inner transaction will fail on the usual conditions (e.g.
> not enough balance).
>
> Note that the method interface does not specify `asset_close_to`, because
> holders of a Smart ASA will need two transactions (RECOMMENDED in an Atomic
> Transfer) to close their position:
>
> - A call to this method to transfer their outstanding balance (possibly as a
>   `CloseOut` operation if the controlling Smart Contract required opt in); and
> - an additional transaction to close out of the ASA.

#### Asset Freeze

```json
[
  {
    "name": "asset_freeze",
    "args": [
      { "type": "asset", "name": "freeze_asset" },
      { "type": "bool", "name": "asset_frozen" }
    ],
    "returns": { "type": "void" }
  },
  {
    "name": "account_freeze",
    "args": [
      { "type": "asset", "name": "freeze_asset" },
      { "type": "account", "name": "freeze_account" },
      { "type": "bool", "name": "asset_frozen" }
    ],
    "returns": { "type": "void" }
  }
]
```

Calling `asset_freeze` prevents any transfer of a Smart ASA. Calling
`account_freeze` prevents a specific account from transferring or receiving a
Smart ASA.

> Upon a call to `asset_freeze` or `account_freeze`, a reference implementation
> SHOULD:
>
> - Fail if `freeze_asset` does not correspond to an ASA controlled by this smart
>   contract.
> - Succeed iff the `sender` of the transaction corresponds to the `freeze_addr`,
>   as persisted by the controlling Smart Contract.
>
> In addition:
>
> - Upon a call to `asset_freeze`, the controlling Smart Contract SHOULD persist
>   the tuple `(freeze_asset, asset_frozen)` (for instance, by setting a `frozen`
>   flag in _global_ storage).
> - Upon a call to `account_freeze` the controlling Smart Contract SHOULD persist
>   the tuple `(freeze_asset, freeze_account, asset_frozen)` (for instance by
>   setting a `frozen` flag in the _local_ storage of the `freeze_account`). See
>   the [security considerations section](#security-considerations) for how to
>   ensure that Smart ASA holders cannot reset their `frozen` flag by clearing out
>   their state at the controlling Smart Contract.

```json
[
  {
    "name": "get_asset_is_frozen",
    "readonly": true,
    "args": [{ "type": "asset", "name": "freeze_asset" }],
    "returns": { "type": "bool" }
  },
  {
    "name": "get_account_is_frozen",
    "readonly": true,
    "args": [
      { "type": "asset", "name": "freeze_asset" },
      { "type": "account", "name": "freeze_account" }
    ],
    "returns": { "type": "bool" }
  }
]
```

The value return by `get_asset_is_frozen` (respectively,
`get_account_is_frozen`) tells whether any account (respectively
`freeze_account`) can transfer or receive `freeze_asset`. A `false` value
indicates that the transfer will be rejected.

> Upon a call to `get_asset_is_frozen`, a reference implementation SHOULD
> retrieve the tuple `(freeze_asset, asset_frozen)` as stored on `asset_freeze`
> and return the value corresponding to `asset_frozen`.
> Upon a call to `get_account_is_frozen`, a reference implementation SHOULD
> retrieve the tuple `(freeze_asset, freeze_account, asset_frozen)` as
> stored on `account_freeze` and return the value corresponding to
> `asset_frozen`.

#### Asset Destroy

```json
{
  "name": "asset_destroy",
  "args": [{ "type": "asset", "name": "destroy_asset" }],
  "returns": { "type": "void" }
}
```

Calling `asset_destroy` destroys a Smart ASA.

> Upon a call to `asset_destroy`, a reference implementation SHOULD:
>
> - Fail if `destroy_asset` does not correspond to an ASA controlled by this smart
>   contract.
>
> It is RECOMMENDED for calls to this method to be permissioned (see
> `asset_create`).
>
> The controlling Smart Contract SHOULD perform an asset destroy operation on
> the ASA with ID `destroy_asset`. The operation will fail if the asset is still
> in circulation.

#### Circulating Supply

```json
{
  "name": "get_circulating_supply",
  "readonly": true,
  "args": [{ "type": "asset", "name": "asset" }],
  "returns": { "type": "uint64" }
}
```

Calling `get_circulating_supply` returns the circulating supply of a Smart ASA.

> Upon a call to `get_circulating_supply`, a reference implementation SHOULD:
>
> - Fail if `asset` does not correspond to an ASA controlled by this smart
>   contract.
> - Return the circulating supply of `asset`, defined by the difference between
>   the ASA `total` and the balance held by its `reserve_addr` (see [Asset
>   Creation](#asset-creation)).

#### Full ABI Spec

```json
{
  "name": "arc-0020",
  "methods": [
    {
      "name": "asset_create",
      "args": [
        {
          "type": "uint64",
          "name": "total"
        },
        {
          "type": "uint32",
          "name": "decimals"
        },
        {
          "type": "bool",
          "name": "default_frozen"
        },
        {
          "type": "string",
          "name": "unit_name"
        },
        {
          "type": "string",
          "name": "name"
        },
        {
          "type": "string",
          "name": "url"
        },
        {
          "type": "byte[]",
          "name": "metadata_hash"
        },
        {
          "type": "address",
          "name": "manager_addr"
        },
        {
          "type": "address",
          "name": "reserve_addr"
        },
        {
          "type": "address",
          "name": "freeze_addr"
        },
        {
          "type": "address",
          "name": "clawback_addr"
        }
      ],
      "returns": {
        "type": "uint64"
      }
    },
    {
      "name": "asset_config",
      "args": [
        {
          "type": "asset",
          "name": "config_asset"
        },
        {
          "type": "uint64",
          "name": "total"
        },
        {
          "type": "uint32",
          "name": "decimals"
        },
        {
          "type": "bool",
          "name": "default_frozen"
        },
        {
          "type": "string",
          "name": "unit_name"
        },
        {
          "type": "string",
          "name": "name"
        },
        {
          "type": "string",
          "name": "url"
        },
        {
          "type": "byte[]",
          "name": "metadata_hash"
        },
        {
          "type": "address",
          "name": "manager_addr"
        },
        {
          "type": "address",
          "name": "reserve_addr"
        },
        {
          "type": "address",
          "name": "freeze_addr"
        },
        {
          "type": "address",
          "name": "clawback_addr"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "get_asset_config",
      "readonly": true,
      "args": [
        {
          "type": "asset",
          "name": "asset"
        }
      ],
      "returns": {
        "type": "(uint64,uint32,bool,string,string,string,byte[],address,address,address,address)",
        "desc": "`total`, `decimals`, `default_frozen`, `unit_name`, `name`, `url`, `metadata_hash`, `manager_addr`, `reserve_addr`, `freeze_addr`, `clawback`"
      }
    },
    {
      "name": "asset_transfer",
      "args": [
        {
          "type": "asset",
          "name": "xfer_asset"
        },
        {
          "type": "uint64",
          "name": "asset_amount"
        },
        {
          "type": "account",
          "name": "asset_sender"
        },
        {
          "type": "account",
          "name": "asset_receiver"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "asset_freeze",
      "args": [
        {
          "type": "asset",
          "name": "freeze_asset"
        },
        {
          "type": "bool",
          "name": "asset_frozen"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "account_freeze",
      "args": [
        {
          "type": "asset",
          "name": "freeze_asset"
        },
        {
          "type": "account",
          "name": "freeze_account"
        },
        {
          "type": "bool",
          "name": "asset_frozen"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "get_asset_is_frozen",
      "readonly": true,
      "args": [
        {
          "type": "asset",
          "name": "freeze_asset"
        }
      ],
      "returns": {
        "type": "bool"
      }
    },
    {
      "name": "get_account_is_frozen",
      "readonly": true,
      "args": [
        {
          "type": "asset",
          "name": "freeze_asset"
        },
        {
          "type": "account",
          "name": "freeze_account"
        }
      ],
      "returns": {
        "type": "bool"
      }
    },
    {
      "name": "asset_destroy",
      "args": [
        {
          "type": "asset",
          "name": "destroy_asset"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "get_circulating_supply",
      "readonly": true,
      "args": [
        {
          "type": "asset",
          "name": "asset"
        }
      ],
      "returns": {
        "type": "uint64"
      }
    }
  ]
}
```

### Metadata

#### ASA Metadata

The ASA underlying a Smart ASA:

- MUST be `DefaultFrozen`.
- MUST specify the ID of the controlling Smart Contract (see below); and
- MUST set the `ClawbackAddr` to the account of such Smart Contract.

The metadata **MUST** be immutable.

#### Specifying the controlling Smart Contract

A Smart ASA MUST specify the ID of its controlling Smart Contract.

If the Smart ASA also conforms to any ARC that supports additional `properties`
([ARC-3](./arc-0003.md), [ARC-69](./arc-0069.md)), then it MUST include a
`arc-20` key and set the corresponding value to a map, including the ID of the
controlling Smart Contract as a value for the key `application-id`. For example:

```javascript
{
  //...
  "properties": {
    //...
    "arc-20": {
      "application-id": 123
    }
  }
  //...
}
```

> To avoid ecosystem fragmentation this ARC does NOT propose any new method to
> specify the metadata of an ASA. Instead, it only extends already
> existing standards.

### Handling opt in and close out

A Smart ASA MUST require users to opt to the ASA and MAY require them to opt in
to the controlling Smart Contract. This MAY be performed at two separate times.

The reminder of this section is non-normative.

> Smart ASAs SHOULD NOT require users to opt in to the controlling Smart
> Contract, unless the implementation requires storing information into their
> local schema (for instance, to implement [freezing](#asset-freeze); also see
> [security considerations](#security-considerations)).
>
> Clients MAY inspect the local state schema of the controlling Smart Contract
> to infer whether opt in is required.
>
> If a Smart ASA requires opt in, then clients SHOULD prevent users from closing
> out the controlling Smart Contract unless they don't hold a balance for any of
> the ASAs controlled by the Smart Contract.

## Rationale

This ARC builds on the strengths of the ASA to enable a Smart Contract to
control its operations and flexibly re-configure its configuration.

The rationale is to have a "Smart ASA" that is as widely adopted as the ASA both
by the community and by the surrounding ecosystem. Wallets, dApps, and
marketplaces:

- Will display a user's Smart ASA balance out-of-the-box (because of the
  underlying ASA).
- SHOULD recognize Smart ASAs and inform the users accordingly by displaying the
  name, unit name, URL, etc. from the controlling Smart Contract.
- SHOULD enable users to transfer the Smart ASA by constructing the appropriate
  transactions, which call the ABI methods of the controlling Smart Contract.

With this in mind, this standard optimizes for:

- Community adoption, by minimizing the [ASA metadata](#metadata) that need to
  be set and the requirements of a conforming implementation.
- Developer adoption, by re-using the familiar ASA transaction reference in the
  methods' specification.
- Ecosystem integration, by minimizing the amount of work that a wallet, dApp or
  service should perform to support the Smart ASA.

## Backwards Compatibility

Existing ASAs MAY adopt this standard if issued or re-configured to match the
requirements in the [metadata section](#metadata).

This requires:

- The ASA to be `DefaultFrozen`.
- Deploying a Smart Contract that will manage, control and operate on the
  asset(s).
- Re-configuring the ASA, by setting its `ClawbackAddr` to the account of the
  controlling Smart Contract.
- Associating the ID of the Smart Contract to the ASA (see
  [metadata](#metadata)).

### [ARC-18](./arc-0018.md)

Assets implementing [ARC-18](./arc-0018.md) MAY also be compatible with this ARC
if the Smart Contract implementing royalties enforcement exposes the ABI methods
specified here. The corresponding ASA and their metadata are compliant with this
standard.

## Reference Implementation

A reference implementation is available here:
https://github.com/algorandlabs/smart-asa.

## Security Considerations

Keep in mind that the rules governing a Smart ASA are only in place as long as:

- The ASA remains frozen;
- the `ClawbackAddr` of the ASA is set to a controlling Smart Contract, as
  specified in the [metadata section](#metadata);
- the controlling Smart Contract is not updatable, nor deletable, nor
  re-keyable.

### Local State

If your controlling Smart Contract implementation writes information to a user's
local state, keep in mind that users can close out the application and (worse)
clear their state at all times. This requires careful considerations.

For instance, if you determine a user's [freeze](#asset-freeze) state by reading
a flag from their local state, you should consider the flag _set_ and the user
_frozen_ if the corresponding local state key is _missing_.

For a `default_frozen` Smart ASA this means:

- Set the `frozen` flag (to `1`) at opt in.
- Explicitly verify that a user's `frozen` flag is not set (is `0`) before
  approving transfers.
- If the key `frozen` is missing from the user's local state, then considered
  the flag to be set and reject all transfers.

This prevents users from resetting their `frozen` flag by clearing their state
and then opting into the controlling Smart Contract again.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
