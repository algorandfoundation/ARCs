---
arc: 62
title: ASA Circulating Supply
description: Getter method for ASA circulating supply
author: Cosimo Bassi (@cusma)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/302
status: Final
type: Standards Track
category: Interface
sub-category: Asa
created: 2024-06-12
requires: 2, 4, 22
---

## Abstract

This ARC introduces a standard for the definition of circulating supply for Algorand
Standard Assets (ASA) and its client-side retrieval. A reference implementation is
suggested.

## Motivation

Algorand Standard Asset (ASA) `total` supply is _defined_ upon ASA creation.

Creating an ASA on the ledger _does not_ imply its `total` supply is immediately
“minted” or “circulating”. In fact, the semantic of token “minting” on Algorand is
slightly different from other blockchains: it is not coincident with the token units
creation on the ledger.

The Reserve Address, one of the 4 addresses of ASA Role-Based-Access-Control (RBAC),
is conventionally used to identify the portion of `total` supply not yet in circulation.
The Reserve Address has no “privilege” over the token: it is just a “logical” label
used (client-side) to classify an existing amount of ASA as “not in circulation”.

According to this convention, “minting” an amount of ASA units is equivalent to
_moving that amount out of the Reserve Address_.

> ASA may have the Reserve Address assigned to a Smart Contract to enforce specific
> “minting” policies, if needed.

This convention led to a simple and unsophisticated semantic of ASA circulating
supply, widely adopted by clients (wallets, explorers, etc.) to provide standard
information:

```text
circulating_supply = total - reserve_balance
```

Where `reserve_balance` is the ASA balance hold by the Reserve Address.

However, the simplicity of such convention, who fostered adoption across the Algorand
ecosystem, poses some limitations. Complex and sophisticated use-cases of ASA, such
as regulated stable-coins and tokenized securities among others, require more
detailed and expressive definitions of circulating supply.

As an example, an ASA could have “burned”, “locked” or “pre-minted” amounts of token,
not held in the Reserve Address, which _should not_ be considered as “circulating”
supply. This is not possible with the basic ASA protocol convention.

This ARC proposes a standard ABI _read-only_ method (getter) to provide the circulating
supply of an ASA.

## Specification

The keywords "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**",
"**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**"
in this document are to be interpreted as described in <a href="https://datatracker.ietf.org/doc/html/rfc2119">RFC 2119</a>.

> Notes like this are non-normative.

### ABI Method

A compliant ASA, whose circulating supply definition conforms to this ARC, **MUST**
implement the following method on an Application (referred as _Circulating Supply
App_ in this specification):

```json
{
  "name": "arc62_get_circulating_supply",
  "readonly": true,
  "args": [
    {
      "type": "uint64",
      "name": "asset_id",
      "desc": "ASA ID of the circulating supply"
    }
  ],
  "returns": {
    "type": "uint64",
    "desc": "ASA circulating supply"
  },
  "desc": "Get ASA circulating supply"
}
```

The `arc62_get_circulating_supply` **MUST** be a _read-only_ ([ARC-22](./arc-0022.md))
method (getter).

### Usage

Getter calls **SHOULD** be _simulated_.

Any external resources used by the implementation **SHOULD** be discovered and
auto-populated by the simulated getter call.

#### Example 1

> Let the ASA have `total` supply and a Reserve Address (i.e. not set to `ZeroAddress`).
>
> Let the Reserve Address be assigned to an account different from the Circulating
> Supply App Account.
>
> Let `burned` be an external Burned Address dedicated to ASA burned supply.
>
> Let `locked` be an external Locked Address dedicated to ASA locked supply.
>
> The ASA issuer defines the _circulating supply_ as:
>
> ```text
> circulating_supply = total - reserve_balance - burned_balance - locked_balance
> ```
>
> In this case the simulated read-only method call would auto-populate 1 external
> reference for the ASA and 3 external reference accounts (Reserve, Burned and Locked).

#### Example 2

> Let the ASA have `total` supply and _no_ Reserve Address (i.e. set to `ZeroAddress`).
>
> Let `non_circulating_amount` be a UInt64 Global Var defined by the implementation
> of the Circulating Supply App.
>
> The ASA issuer defines the _circulating supply_ as:
>
> ```text
> circulating_supply = total - non_circulating_amount
> ```
>
> In this case the simulated read-only method call would auto-populate just 1 external
> reference for the ASA.

### Circulating Supply Application discovery

> Given an ASA ID, clients (wallet, explorer, etc.) need to discover the related
> Circulating Supply App.

An ASA conforming to this ARC **MUST** specify the Circulating Supply App ID.

> To avoid ecosystem fragmentation this ARC does not propose any new method to specify
> the metadata of an ASA. Instead, it only extends already existing standards.

If the ASA also conforms to any ARC that supports additional `properties` ([ARC-3](./arc-0003.md),
[ARC-19](./arc-0019.md), [ARC-69](./arc-0069.md)),
then it **MUST** include a `arc-62` key and set the corresponding value to a map,
including the ID of the Circulating Supply App as a value for the key `application-id`.

#### Example

```json
{
  //...
  "properties": {
    //...
    "arc-62": {
      "application-id": 123
    }
  }
  //...
}
```

## Rationale

The definition of _circulating supply_ for sophisticated use-cases is usually ASA-specific.
It could involve, for example, complex math or external accounts’ balances, variables
stored in boxes or in global state, etc..

For this reason, the proposed method’s signature does not require any reference
to external resources, a part form the `asset_id` of the ASA for which the circulating
supply is defined.

Eventual external resources can be discovered and auto-populated directly by the
simulated method call.

The rational of this design choice is avoiding fragmentation and integration overhead
for clients (wallets, explorers, etc.).

Clients just need to know:

1. The ASA ID;
1. The Circulating Supply App ID implementing the `arc62_get_circulating_supply`
method for that ASA.

## Backwards Compatibility

Existing ASA willing to conform to this ARC **MUST** specify the Circulating Supply
App ID as [ARC-2](./arc-0002.md) `AssetConfig` transaction note field, as follows:

- The `<arc-number>` **MUST** be equal to `62`;
- The **RECOMMENDED** `<data-format>` are <a href="https://msgpack.org/">MsgPack</a>
(`m`) or <a href="https://www.json.org/json-en.html">JSON</a> (`j`);
- The `<data>` **MUST** specify `application-id` equal to the Circulating Supply
App ID.

> **WARNING**: To preserve the existing ASA RBAC (e.g. Manager Address, Freeze Address,
> etc.) it is necessary to **include all the existing role addresses** in the `AssetConfig`.
> Not doing so would irreversibly disable the RBAC roles!

### Example - JSON without version

```text
arc62:j{"application-id":123}
```

## Reference Implementation

> This section is non-normative.

This section suggests a reference implementation of the Circulating Supply App.

An Algorand-Python example is available [here](../assets/arc-0062).

### Recommendations

An ASA using the reference implementation **SHOULD NOT** assign the Reserve Address
to the Circulating Supply App Account.

A reference implementation **SHOULD** target a version of the AVM that supports
foreign resources pooling (version 9 or greater).

A reference implementation **SHOULD** use 3 external addresses, in addition to the
Reserve Address, to define the not circulating supply.

The **RECOMMENDED** labels for not-circulating balances are: `burned`, `locked`
and `generic`.

> To change the labels of not circulating addresses is sufficient to rename the
> following constants just in `smart_contracts/circulating_supply/config.py`:
> ```python
> NOT_CIRCULATING_LABEL_1: Final[str] = "burned"
> NOT_CIRCULATING_LABEL_2: Final[str] = "locked"
> NOT_CIRCULATING_LABEL_3: Final[str] = "generic"
> ```

### State Schema

A reference implementation **SHOULD** allocate, at least, the following Global State
variables:

- `asset_id` as UInt64, initialized to `0` and set **only once** by the ASA Manager
Address;
- Not circulating address 1 (`burned`) as Bytes, initialized to the Global `Zero Address`
 and set by the ASA Manager Address;
- Not circulating address 2 (`locked`) as Bytes, initialized to the Global `Zero Address`
and set by the ASA Manager Address;
- Not circulating address 3 (`generic`) as Bytes, initialized to the Global `Zero Address`
and set by the ASA Manager Address.

A reference implementation **SHOULD** enforce that, upon setting `burned`, `locked`
and `generic` addresses, the latter already opted-in the `asset_id`.

```json
"state": {
    "global": {
        "num_byte_slices": 3,
        "num_uints": 1
    },
    "local": {
        "num_byte_slices": 0,
        "num_uints": 0
    }
},
"schema": {
    "global": {
        "declared": {
            "asset_id": {
                "type": "uint64",
                "key": "asset_id"
            },
            "not_circulating_label_1": {
                "type": "bytes",
                "key": "burned"
            },
            "not_circulating_label_2": {
                "type": "bytes",
                "key": "locked"
            },
            "not_circulating_label_3": {
                "type": "bytes",
                "key": "generic"
            }
        },
        "reserved": {}
    },
    "local": {
        "declared": {},
        "reserved": {}
    }
},
```

### Circulating Supply Getter

A reference implementation **SHOULD** enforce that the `asset_id` Global Variable
is equal to the `asset_id` argument of the `arc62_get_circulating_supply` getter
method.

> Alternatively the reference implementation could ignore the `asset_id` argument
> and use directly the `asset_id` Global Variable.

A reference implementation **SHOULD** return the ASA _circulating supply_ as:

```text
circulating_supply = total - reserve_balance - burned_balance - locked_balance - generic_balance
```

Where:

- `total` is the total supply of the ASA (`asset_id`);
- `reserve_balance` is the ASA balance hold by the Reserve Address or `0` if the
address is set to the Global `ZeroAddress` or not opted-in `asset_id`;
- `burned_balance` is the ASA balance hold by the Burned Address or `0` if the address
is set to the Global `ZeroAddress` or is not opted-in `asset_id`;
- `locked_balance` is the ASA balance hold by the Locked Address or `0` if the address
is set to the Global `ZeroAddress` or not opted-in `asset_id`;
- `generic_balance` is the ASA balance hold by a Generic Address or `0` if the address
is set to the Global `ZeroAddress` or not opted-in `asset_id`.

## Security Considerations

Permissions over the Circulating Supply App setting and update **SHOULD** be granted
to the ASA Manager Address.

> The ASA trust-model (i.e. who sets the Reserve Address) is extended to the generalized
> ASA circulating supply definition.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
