---
arc: 00XX
title: Smart ASA
description: An ARC for an ASA controlled by an Algorand Smart Contract
author:
  - Cosimo Bassi (@cusma)
  - Adriano Di Luzio (@aldur)
  - John Jannotti (@jannotti)
discussions-to: <URL>
status: Draft
type: Standards Track
category (*only required for Standards Track): Interface
created: 2022-04-27
requires: ARC-4
---

## Abstract

A "Smart ASA" is an Algorand Standard Asset (ASA) controlled by a Smart Contract
that exposes methods to create, configure, transfer, freeze, and destroy it.

This ARC defines the ABI interface of such smart contract, the required
metadata, and suggests a reference implementation.

## Motivation

The Algorand Standard Asset (ASA) is an excellent building block for on-chain
applications. It is battle-tested and widely adopted and supported by SDKs,
wallets, and dApps.

However, the ASA lacks in flexibility and configurability. For instance, once
issued it can't be re-configured (its unit name, decimals, maximum supply).
Also, it is freely transferable (unless frozen). This prevents developers from
specifying additional business logic to be checked while transferring it (think
of royalties or [vesting](https://en.wikipedia.org/wiki/Vesting)).

Enforcing transfer conditions require freezing the asset and transferring it
through a clawback operation -- which results in a process that is opaque to
users and wallets and a bad experience for the users.

The Smart ASA defined by this ARC extends the ASA to increase its expressiveness
and its flexibility. Developers can now introduce additional business logic
around its operations. Wallets, dApps, and SDKs can recognize Smart ASAs and
adjust their flows and user experiences accordingly.

## Specification

The keywords "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be
interpreted as described in RFC 2119.

The following sections describe:

- The ABI interface for a controlling Smart Contract (the Smart Contract that
  controls a Smart ASA).
- The metadata required to denote a Smart ASA and define the association between
  an ASA and its controlling Smart Contract.

### ABI Interface

The ABI interface specified here draws inspiration from the transaction
[reference](https://developer.algorand.org/docs/get-details/asa/#asset-functions)
of an Algorand Standard Asset (ASA).

To provide a unified and familiar interface between the Algorand Standard Asset
and the Smart ASA, method names and parameters have been adapted to the ABI
types but left otherwise unchanged.

#### Asset Creation

```json
{
  "name": "AssetCreate",
  "args": [
    { "type": "uint64", "name": "Total" },
    { "type": "uint32", "name": "Decimals" },
    { "type": "bool", "name": "DefaultFrozen" },
    { "type": "string", "name": "UnitName" },
    { "type": "string", "name": "AssetName" },
    { "type": "string", "name": "URL" },
    { "type": "[]byte", "name": "MetaDataHash" },
    { "type": "address", "name": "ManagerAddr" },
    { "type": "address", "name": "ReserveAddr" },
    { "type": "address", "name": "FreezeAddr" },
    { "type": "address", "name": "ClawbackAddr" }
  ],
  "returns": { "type": "uint64" }
}
```

Calling `AssetCreate` creates a new Smart ASA and returns the identifier of the
ASA. The [metadata section](#metadata) describes its required properties.

> Upon a call to `AssetCreate`, a reference implementation SHOULD:
>
> - Mint an Algorand Standard Asset (ASA) that MUST specify the properties defined
>   in the [metadata section](#metadata). The `Manager`, `Reserve` and `Freeze`
>   addresses SHOULD be set to the account of the controlling smart contract. The
>   remaining fields are left to the implementation, which MAY set `Total` to `2 ** 64 - 1`
>   to enable dynamically increasing the circulating supply of the
>   asset. `AssetName` and `UnitName` MAY be set to `SMART-ASA` and `S-ASA`, to
>   mark that this ASA is Smart and has a controlling application.
> - Persist the `Total`, `Decimals`, `DefaultFrozen`, etc. fields for later
>   use/retrieval.
> - Return the ID of the created ASA.
>
> It is RECOMMENDED for calls to this method to be permissioned, e.g. to only
> approve transactions issued by the controlling smart contract creator.

#### Asset Configuration

```json
{
  "name": "AssetConfig",
  "args": [
    { "type": "asset", "name": "ConfigAsset" },
    { "type": "uint64", "name": "Total" },
    { "type": "uint32", "name": "Decimals" },
    { "type": "bool", "name": "DefaultFrozen" },
    { "type": "string", "name": "UnitName" },
    { "type": "string", "name": "AssetName" },
    { "type": "string", "name": "URL" },
    { "type": "[]byte", "name": "MetaDataHash" },
    { "type": "address", "name": "ManagerAddr" },
    { "type": "address", "name": "ReserveAddr" },
    { "type": "address", "name": "FreezeAddr" },
    { "type": "address", "name": "ClawbackAddr" }
  ],
  "returns": { "type": "void" }
}
```

Calling `AssetConfig` configures an existing Smart ASA and returns its
identifier.

> Upon a call to `AssetConfig`, a reference implementation SHOULD:
>
> - Fail if `ConfigAsset` does not correspond to an ASA controlled by this smart
>   contract.
> - Update the persisted `Total`, `Decimals`, `DefaultFrozen`, etc. fields for later
>   use/retrieval.
>
> It is RECOMMENDED for calls to this method to be permissioned (see
> `AssetCreate`).
>
> The business logic associated to the update of the other parameters is left to
> the implementation. An implementation that maximizes similarities with ASAs,
> SHOULD NOT allow modifying the `ClawbackAddr` or `FreezeAddr` after they
> have been set to the special value `ZeroAddress`.
>
> The implementation MAY provide flexibility on the fields of an ASA that
> cannot be updated after initial configuration. For instance, it MAY update the
> `Total` parameter to enable minting of new units or restricting the maximum
> supply; when doing so, the implementation SHOULD ensure that the updated
> `Total` is not lower than the current circulating supply of the asset.

### Asset Transfer

```json
{
  "name": "AssetTransfer",
  "args": [
    { "type": "asset", "name": "XferAsset" },
    { "type": "uint64", "name": "AssetAmount" },
    { "type": "account", "name": "AssetSender" },
    { "type": "account", "name": "AssetReceiver" }
  ],
  "returns": { "type": "void" }
}
```

Calling `AssetTransfer` transfers a Smart ASA.

> Upon a call to `AssetTransfer`, a reference implementation SHOULD:
>
> - Fail if `XferAsset` does not correspond to an ASA controlled by this smart
>   contract.
> - Succeed if the `Sender` of the transaction is the `AssetSender` and
>   `AssetSender` and `AssetReceiver` are not in a frozen state (see
>   [below](#asset-freeze)).
> - Succeed if the `Sender` of the transaction corresponds to the `ClawbackAddr`,
>   as persisted by the controlling smart contract. This enables clawback
>   operations on the Smart ASA.
>
> Internally, the controlling smart contract SHOULD issue a clawback inner
> transaction that transfers the `AssetAmount` from `AssetSender` to
> `AssetReceiver`. The inner transaction will fail on the usual conditions (e.g.
> not enough balance).
>
> Note that the method interface does not specify `AssetCloseTo`, because
> holders of a Smart ASA will need two transactions (RECOMMENDED in an Atomic
> Transfer) to close their position:
>
> - A call to this method to transfer their outstanding balance (possibly as a
>   `CloseOut` operation if the controlling smart contract required opt in); and
> - an additional transaction to opt out of the ASA.

### Asset Freeze

```json
{
  "name": "AssetFreeze",
  "args": [
    { "type": "asset", "name": "FreezeAsset" },
    { "type": "account", "name": "FreezeAccount" },
    { "type": "bool", "name": "AssetFrozen" }
  ],
  "returns": { "type": "void" }
}
```

Calling `AssetFreeze` prevents an account from transferring a Smart ASA.

> Upon a call to `AssetFreeze`, a reference implementation SHOULD:
>
> - Fail if `FreezeAsset` does not correspond to an ASA controlled by this smart
>   contract.
> - Succeed iff the `Sender` of the transaction corresponds to the `FreezeAddr`,
>   as persisted by the controlling smart contract.
>
> The controlling smart contract SHOULD persist the pair `FreezeAccount` and
> `FreezeAccount` (for instance by setting `frozen` flag in the local storage of
> the `FreezeAccount`). See the [security considerations
> section](#security-considerations) for how to ensure that Smart ASA holders
> cannot reset their `frozen` flag by clearing out their state at the controlling smart
> contract.

### Asset Destroy

```json
{
  "name": "AssetDestroy",
  "args": [{ "type": "asset", "name": "DestroyAsset" }],
  "returns": { "type": "void" }
}
```

Calling `AssetDestroy` destroys a Smart ASA.

> Upon a call to `AssetDestroy`, a reference implementation SHOULD:
>
> - Fail if `DestroyAsset` does not correspond to an ASA controlled by this smart
>   contract.
>
> It is RECOMMENDED for calls to this method to be permissioned (see
> `AssetCreate`).
>
> The controlling smart contract SHOULD perform an asset destroy operation on
> the ASA with ID `DestroyAsset`. The operation will fail if the asset is still
> in circulation.

### Getters

TODO.

#### Full ABI Spec

```javascript
// TODO
{
}
```

### Metadata

#### ASA Metadata

The ASA underlying a Smart ASA:

- MUST be `DefaultFrozen`.
- MUST specify the ID of the controlling Smart Contract (see below); and
- MUST set the `ClawbackAddr` to the account of such Smart Contract.

#### Specifying the controlling Smart Contract

A Smart ASA MUST specify the ID of its controlling Smart Contract as follows:

- If conforming to any ARC that supports additional `properties` ([ARC-3](./arc-0003.md),
  [ARC-69](./arc-0069.md)), then it MUST specify a `arc-xx` key and set the corresponding value
  to a map, specifying as the value for `application-id` the ID of the
  controlling Smart Contract.
- TODO: URL encoding.
- TODO: Reserve.

The metadata **MUST** be immutable.

Example:

```javascript
//...
"properties":{
  //...
  "arc-xx":{
    "application-id": 123,
  }
}
//...
```

### OptIn and OptOut

TODO.

## Rationale

TODO.

## Backwards Compatibility

TODO.

TODO: ARC-18.

## Reference Implementation

https://github.com/aldur/tc-asa

## Security Considerations

Non-normative: Keep in mind that the rules governing a Smart ASA are only in
place as long as:

- The ASA remains frozen,
- its `ClawbackAddr` is set as specified in the [metadata section](#metadata),
- the controlling Smart Contract is not updatable, not deletable, not rekeyable.

TODO: Opt-in, clear state and freeze.

## Copyright

Copyright and related rights waived via
[CC0](https://creativecommons.org/publicdomain/zero/1.0/).
