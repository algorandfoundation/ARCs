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
- The metadata required by a Smart ASA and defining the association between an
  ASA and its controlling Smart Contract.

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
  "returns": { "type": "uint64" }
}
```

Calling `AssetConfig` configures an existing Smart ASA and returns its
identifier.

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

TODO: Note about `AssetCloseTo` missing.
TODO: Note about clawback.

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

### Asset Destroy

```json
{
  "name": "AssetDestroy",
  "args": [{ "type": "asset", "name": "DestroyAsset" }],
  "returns": { "type": "void" }
}
```

Calling `AssetDestroy` destroys a Smart ASA.

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

- If conforming to any ARC that supports additional `properties` (ARC-3,
  ARC-69), then it MUST specify a `arc-xx` key and set the corresponding value
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

## Copyright

Copyright and related rights waived via
[CC0](https://creativecommons.org/publicdomain/zero/1.0/).
