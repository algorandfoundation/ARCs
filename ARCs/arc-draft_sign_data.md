---
arc: <to be assigned>
title: Algorand Wallet Structured Arbitrary Data Signing API
description: API function fot signing structured arbitrary data
author: Stefano De Angelis (@deanstef)
discussions-to:  https://forum.algorand.org/t/wallet-council-breakoutgroup-byte-signing/11442
status: Draft
type: Standards Track
category: Interface
created: 2024-02-28
requires:  ARC-1.
---

# Algorand Wallet Structured Arbitrary Data Signing API

> This ARC is inspired  <a href="https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0001.md">ARC-1</a>.

## Abstract

WIP

*ARC-1 provides a standard API to sign Algorand transactions. This ARC provides a standard API for wallets to sign arbitrary data that do not represent and Algorand transaction or other valid..*

## Motivation

WIP

*Signing data is a common and critical operation. Users may need to sign data for multiple reasons (e.g. delegate signatures, DIDs, signing documents, authentication, and more). Algorand wallets need to provide such a feature to unlock users with self-custodial signing features. A standard is needed to avoid security issues. Before signing data, users must be warned about what they are signing and for which purpose. This ARC provides a structured schema for data that has to be signed. It includes data validation and a structured way for wallets to display relevant data.*

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

> Comments like this are non-normative

### Overview

> This section is non-normative

This ARC defines an API for wallets to sign arbitrary data with a `signData(arbData, metadata)` function on a non-empty list of structured `arbData` and `metadata` objects.

A `StdSignData` object is used to represent structured arbitrary data to be signed. Those data require a filed `data` that represents a structured byte array to be signed and a filed `signers` that indicates the public keys to be used to sign.

A `StdSignMetadata` object is used to describe the purpose of the signature request and the type of data to be signed, including the encoding used and the schema.

`data` is structured according to a JSON Schema provided with the metadata.

There are two possible use cases:

1. Sign arbitrary data with one signer that the wallet knows. For example:

```json
{
    "data": "/{.../}",
    "signers": ["xxxx"]
}
{
    "scope": "ARBITRARY",
    "schema": "...",
    "message": "These are just arbitrary bytes",
    "encoding": "..."
}
```

1. Sign arbitrary data with a secondary public key that the wallet knows (recalling the Algorand Rekey) or a multisig address derived by two or more public keys that the wallet knows (recall the Algorand MultiSig), or a hierarchical deterministic (HD) wallet.

```json
{
    "data": "/{.../}",
    "signers": ...,
    "authAddr": "xxxx",
    "msig": {
        "version": 1,
        "threshold": 2,
        "addrs": [
            "xxxxxx",
            "xxxxxx",
            "xxxxxx"
        ]
    },
    "hdPath": {
        "purpose": 44,
        "coinType": 0,
        "account": 0,
        "change": 0,
        "addrIdx": 0
    }
}
{
    "scope": "ARBITRARY",
    "schema": "...",
    "message": "These are just arbitrary bytes",
    "encoding": "..."
}
```

### Interfaces

> Interfaces are defined in TypeScript. All the objects that are defined are valid JSON objects.

This ARC uses interchangeably the terms "throw an error" and "reject a promise with an error".

#### **Interface `SignDataFunction`**

A wallet function `signData` that signs arbitrary data is defined by the interface:

```tsx
export type SignDataFunction = {
    arbData: StdSignData[],
    metadata: StdSignMetadata[],
}
    => Promise<(SignedDataStr)>;
```

- `arbData` is a non-empty array of `StdSigData` objects (defined below)
- `metadata` is an object parameter that provides additional information on the data being signed (defined below)

The `signData` function returns a `SignedDataStr` object or, in case of error, reject the promise with an error object `SignDataError`.

#### Interface `HDWalletMetadata`

An `HDWalletMetadata` specifies the derivation path parameters to derive the keys of an HD wallet from the seed.

> HD derivation levels originally from [BIP-44](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)

```tsx
export interface HDWalletMetadata {
    /**
    * HD Wallet purpose value. First derivation path level. 
    * Hardened derivation is used.
    */
    purpose: number,

    /**
    * HD Wallet coin type value. Second derivation path level.
    * Hardened derivation is used.
    */
    coinType: number,

    /**
    * HD Wallet account number. Third derivation path level.
    * Hardened derivation is used.
    */
    account: number,

    /**
    * HD Wallet change value. Fourth derivation path level.
    * Public derivation is used.
    */
    change: number,

    /**
    * HD Wallet address index value. Fifth derivation path level.
    * Public derivation is used.
    */
    addrIdx: number,
}
```

- `purpose` **SHOULD** be set to `44’` for blockchain accounts.
- `coinType` indicate a derivation subtree for a specific cryptocurrency. It **MUST** be set to `283’` for ALGO according to [SLIP-44](https://github.com/satoshilabs/slips/blob/master/slip-0044.md) registered coin types.

The apostrophe in the numbering convention indicates that hardened derivation is used.

#### Interface `StdData`

The arbitrary data **MUST** be represented as a canonicalized JSON object in accordance with [RFC 8785](https://www.rfc-editor.org/rfc/rfc8785). The JSON **MUST** be valid with respect the schema provided.

```tsx
export type StdData = string;
```

The `StdData` cannot be a valid Algorand object type like a transaction or a logic signature. It **MUST NOT** be prepended with any reserved Algorand domain separator prefix, as defined in the [Algorand specs](https://github.com/algorandfoundation/specs/blob/master/dev/crypto.md#domain-separation) and in the [Go type HashID](https://github.com/algorand/go-algorand/blob/master/protocol/hash.go#L21).

The `StdData` **MUST** be prepended with the `"arcXXXX"` domain separator, so that wallets can distinguish ARC-xxxx arbitrary data to be signed.

#### Interface `Ed25519Pk`

An `Ed25519Pk` object is a 32-byte public key, point of the ed25519 elliptic curve. It is base64 encoded in accordance with [RFC-7468](https://www.ietf.org/rfc/rfc7468.txt). The key has not been transformed into an Algorand address.

```tsx
export type Ed25519Pk = string;
```

> The transformation from a public key to a standard Algorand address is left to the implementation, and it is detailed in the [Public Key to Algorand Address](https://developer.algorand.org/docs/get-details/accounts/#transformation-public-key-to-algorand-address) section of the developer portal.

#### Interface `SignedDataStr`

`SignedDataStr` is the produced 64-byte array Ed25519 digital signature of the `StdData` message.

```tsx
export type SignedDataStr = string;
```

#### Interface `StdSigData`

A `StdSignData` object represents a structured byte array of data to be signed by a wallet.

```tsx
export interface StdSigData {
    /**
    * A non-empty list of structured arbitrary data to be signed.
    */
    data: StdData[];
    
    /**
    * A non-empty list of ed25519 public keys that must sign the data.
    */
    signers: Ed25519Pk[];

    /**
    * Optional delegated public key used to sign data on behalf of a signer.
    * It can be used to sign data with rekeyed Algorand accounts.
    */
    authAddr?: Ed25519Pk;

    /**
    * Optional multisig metadata used to sign the data.
    */
    msig?: MultisigMetadata;

    /**
    * Optional HD wallet metadata used to derive
    * the public key used to sign the data.
    */
    hdPath?: HDWalletMetadata
}
```

The `msig` object is defined in [ARC-1](https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0001.md) and it specifies the metadata of an Algorand multisig address.

#### Enum `ScopeType`

The `ScopeType` enumerates constant strings that specify the specific scope of a byte signature.

This ARC introduces three scopes for signing arbitrary bytes.

| ScopeType | Description |
| --- | --- |
| ARBITRARY | Signature of an array of arbitrary bytes. This is the most generic scope. |
| IDENTITY | Signature of a structured byte array of data used to prove the identity, e.g. proof of the control on a public key. |
| LSIG | Signature of a structured Algorand LogicSig program for delegation.  |

Any extension of this ARC **SHOULD** adopt one of the above `ScopeType` or integrate a new one.

#### Interface `StdSignMetadata`

A `StdSigMetadata` object specifies metadata of a `StdSigData` object that is being signed.

```tsx
export interface StdSigMetadata {
    /**
    * The scope value of the sign data request.
    */
    scope: ScopeType;

    /**
    * Canonical representation of the JSON schema for the signing data.
    */
    schema: string;

    /**
    * Optional message explaining the reason for the signature.
    */
    message?: string;

    /**
    * Optional encoding used to represent the signing data.
    */
    encoding?: string;
}
```

##### Signing Data JSON Schema

The JSON schema for the structured signing data is represented with the following schema.

The `StdData` object **MUST** extend this schema to represent structured arbitrary data being signed.

> The signing data JSON schema is inspired by the schema proposed with [EIP-712: Typed structured data hashing and signing proposal.](https://eips.ethereum.org/EIPS/eip-712)
>
> WIP - *highlight the differences*

```json
{
    "type": "object",
    "properties": {
        "ARCXXXDomain": {
            "type": "string",
            "description": "The ARC-XXXX domain separator",
        },
        "data": {
            "type": "string",
            "description": "The bytes to be signed."
        },
    },
    "required": ["ARCXXXDomain", "data"],
    "additionalProperties": true
}
```

- The `ARCXXXXDomain` object is a string equivalent to `“arcxxxx”` to identify a structured arbitrary data being signed.
- `additionalProperties` can be used to structure more complex arbitrary data to sign.
- The data object is a string of arbitrary bytes. If the schema also provides additional structured objects, then the `data` property **MUST** indicate the SHA256 hash of the canonicalized `additionalProperties`.

An example...

#### Error interface `SignDataError`

The `SignDataError` object follows the same interface as the `SignTxnsError` object defined in [ARC-1](https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0001.md) for transaction signing. It also inherits the same status codes, and in addition

| Status Code | Name | Description |
| --- | --- | --- |
| wip | wip | wip |

### Semantic Requirements

WIP

## Rationale

WIP

## Backwards Compatibility

WIP - Explain how to deal with ARC-47

## Test Cases

N/A

## Reference Implementation

WIP

## Security Considerations

WIP

## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
