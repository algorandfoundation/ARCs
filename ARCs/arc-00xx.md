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
interpreted as described in [RFC
2119](https://datatracker.ietf.org/doc/html/rfc2119).

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
>   in the [metadata section](#metadata). In addition:
>   - The `Manager`, `Reserve` and `Freeze` addresses SHOULD be set to the
>     account of the controlling Smart Contract.
>   - The remaining fields are left to the implementation, which MAY set `Total`
>     to `2 ** 64 - 1` to enable dynamically increasing the circulating supply
>     of the asset.
>   - `AssetName` and `UnitName` MAY be set to `SMART-ASA` and `S-ASA`, to denote
>     that this ASA is Smart and has a controlling application.
> - Persist the `Total`, `Decimals`, `DefaultFrozen`, etc. fields for later
>   use/retrieval.
> - Return the ID of the created ASA.
>
> It is RECOMMENDED for calls to this method to be permissioned, e.g. to only
> approve transactions issued by the controlling Smart Contract creator.

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
>   as persisted by the controlling Smart Contract. This enables clawback
>   operations on the Smart ASA.
>
> Internally, the controlling Smart Contract SHOULD issue a clawback inner
> transaction that transfers the `AssetAmount` from `AssetSender` to
> `AssetReceiver`. The inner transaction will fail on the usual conditions (e.g.
> not enough balance).
>
> Note that the method interface does not specify `AssetCloseTo`, because
> holders of a Smart ASA will need two transactions (RECOMMENDED in an Atomic
> Transfer) to close their position:
>
> - A call to this method to transfer their outstanding balance (possibly as a
>   `CloseOut` operation if the controlling Smart Contract required opt in); and
> - an additional transaction to close out of the ASA.

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
>   as persisted by the controlling Smart Contract.
>
> The controlling Smart Contract SHOULD persist the pair `(FreezeAccount, AssetFrozen)`
> (for instance by setting `frozen` flag in the local storage of the `FreezeAccount`).
> See the [security considerations section](#security-considerations) for how to ensure
> that Smart ASA holders cannot reset their `frozen` flag by clearing out their state
> at the controlling Smart Contract.

TODO: Add getter to check if someone is frozen.

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
> The controlling Smart Contract SHOULD perform an asset destroy operation on
> the ASA with ID `DestroyAsset`. The operation will fail if the asset is still
> in circulation.

### Getters

TODO: Prose.

```json
[
  {
    "name": "getTotal",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "uint64" }
  },
  {
    "name": "getDecimals",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "uint32" }
  },
  {
    "name": "getDefaultFrozen",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "bool" }
  },
  {
    "name": "getUnitName",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "string" }
  },
  {
    "name": "getAssetName",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "string" }
  },
  {
    "name": "getURL",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "string" }
  },
  {
    "name": "getMetaDataHash",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "[]byte" }
  },
  {
    "name": "getManagerAddr",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "address" }
  },
  {
    "name": "getReserveAddr",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "address" }
  },
  {
    "name": "getFreezeAddr",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "address" }
  },
  {
    "name": "getClawbackAddr",
    "args": [{ "type": "asset", "name": "Asset" }],
    "returns": { "type": "address" }
  }
]
```

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

> To avoid ecosystem fragmentation this ARC does NOT propose any
> new method to specify the metadata of an ASA. Instead, it extends already
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
> out the controlling Smart Contract unless they don't hold an ASA balance.

## Rationale

This ARC builds on the strengths of the ASA to enable a Smart Contract to
control its operations and flexibly re-configure its configuration.

The rationale is to have a "Smart ASA" that is as widely adopted as the ASA both
by the community and by the surrounding ecosystem. Wallets and dApps:

- Will display a user's Smart ASA balance out-of-the-box (because of the
  underlying ASA).
- SHOULD recognize Smart ASAs and inform the users accordingly.
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

### ARC-18

Assets implementing [ARC-18](./arc-0018.md) MAY also be compatible with this ARC
if the Smart Contract implementing royalties enforcement exposes the ABI methods
specified here. The corresponding ASA and their metadata are compliant with this
standard.

## Reference Implementation

[https://github.com/aldur/tc-asa](https://github.com/aldur/tc-asa)

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
a flag into their local state, you should:

- Set the flag to `frozen` at opt in;
- explicitly verify that a user's `frozen` flag is `0` before approving
  transfers.
- If missing, the flag should be considered set and prevent transfers.

This prevents users from removing their `frozen` flag by clearing their state
and then opting into the controlling Smart Contract again.

## Copyright

Copyright and related rights waived via
[CC0](https://creativecommons.org/publicdomain/zero/1.0/).
