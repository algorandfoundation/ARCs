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

> This ARC is inspired by <a href="https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0001.md">ARC-1</a>.

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

A `StdSignData` object represents structured arbitrary data that must be signed. Those data include a field `data` that is an array of bytes and a field `signers` for the signing keys.

A `StdSignMetadata` object describes the signature scope and the type of data being signed, including their encoding and the JSON schema.

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

1. Sign arbitrary data with a secondary public key that the wallet knows (recalling the Algorand Rekey) or a multisig address derived by two or more public keys that the wallet knows (recalling the Algorand MultiSig), or a hierarchical deterministic (HD) wallet. For example:

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
    => Promise<(SignedData)>;
```

- `arbData` is a non-empty array of `StdSigData` objects (defined below)
- `metadata` is an object parameter that provides additional information on the data being signed (defined below)

The `signData` function returns a `SignedData` object or, in case of error, reject the promise with an error object `SignDataError`.

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

The arbitrary data **MUST** be represented as a canonicalized JSON object in accordance with <a href="https://www.rfc-editor.org/rfc/rfc8785">RFC 8785</a>. The JSON **MUST** be valid with respect the schema provided.

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

`SignedDataStr` is the produced 64-byte array Ed25519 digital signature of the `StdData` object.

```tsx
export type SignedDataStr = string;
```

#### Interface `SignedData`

A `SignedData` object is a list of `SignedDataStr` representing the signed data. The list must contain all the signatures in case of a multisig.

```tsx
export type SignedData = SignedDataStr[];
```

#### Interface `StdSignData`

A `StdSignData` object represents a structured byte array of data to be signed by a wallet.

```tsx
export interface StdSignData {
    /**
    * The structured arbitrary data to be signed.
    */
    data: StdData;
    
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

A `StdSignMetadata` object specifies metadata of a `StdSignData` object that is being signed.

```tsx
export interface StdSignMetadata {
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

If the optional parameter `encoding` is not specified, the `StdData` object should be UTF-8 encoded following the <a href="https://www.rfc-editor.org/rfc/rfc8785">RFC 8785</a>.

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

The `SignDataError` object extends the `SignTxnsError` defined in [ARC-1](https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0001.md)

```ts
export interface SignDataError extends Error {
  code: number;
  data?: any;
  failingSignData: (StdSigData | null)[];
}
```

`SignDataError` uses the same error codes of `SignTxnsError` as well as the following codes:

| Status Code | Name | Description |
| --- | --- | --- |
| wip | wip | wip |

### Semantic Requirements

The call `signData(arbData, metadata)` **MUST** either return an array `ret` of signed data of the same length of `arbData` or reject the call throwing an error `err`.

> Following [ARC-1](https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0001.md) terminology, in this ARC the term **Rejecting** means throwing an error `err.code=4300`.

Upon calling `signData(arbData, metadata)`:

- the length of `arbData` and `metadata` **MUST** be equal, otherwise the wallet **MUST** reject the call with a `4300` error.
- for each element `i` of `arbData`:
  - the corresponding metadata **MUST** be `metadata[i]`.
  - if the encoding `metadata[i].encoding` is present, it **MUST** be used to decode the `arbData[i].data`.
  - the decoded `arbData[i].data` **MUST** be validated with respect to the JSON schema `metadata[i].data`. If the validation fails, the call **MUST** be rejected with a `4300` error.
  - the wallet **MUST** ask users for signing confirmation. It **MUST** display the `metadata[i].message` if present, and the structured data being signed following to the `metadata[i].schema`
    - if the user approves, the `arbData[i].data` **MUST** be signed by `arbData[i].signers` and `ret[i]` MUST be set to the corresponding `SignedDataStr`.
    - if the user rejects, the call **MUST** fail with error code `4001`.

Note that if any `arbData[i]` cannot be signed for any reason, the wallet **MUST** throw an error, such that

- `err.message` **SHOULD** indicate the reason of the error (e.g. specify that `arbData[i].data` is not a valid JSON object according to `metadata[i].schema`)
- `err.failingSignData` **SHOULD** return the `StdSignData` object that caused the error, otherwise `null`.

#### Semantic of `StdSignData`

- `data`:
  - it **MUST** be a valid StdData object, otherwise the wallet MUST reject.
  - the encoding **MUST** be equal to the value specified by `metadata` if any, otherwise it **MUST** be UTF-8 encoded.
  - if `data` cannot be decoded into a canonicalized JSON object, the wallet **MUST** reject.
  - if the decoded `data` does not comply with the JSON schema in `metadata`, the wallet **MUST** reject.

- `signers`:

  - it **MUST** be a list of valid Ed25519Pk objects, otherwise the wallet **MUST** rejct.
  - the wallet **MAY** transform the `Ed25519Pk` into a valid `AlgorandAddress`.
  > From <a href="https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0001.md">ARC-1</a>, an `AlgorandAddress` is represented by a 58-character base32 string. It includes the checksum.
  - if `signers` length is greater than 1:
    - the wallet **MUST** reject if `msig` is not specified.
    - the wallet **MUST** reject if signers[0] is not equal to the corresponding `msig`.
    > For example, in case of Algorand MultiSig the `msig` address resolves by hashing the MultiSig metadata and addresses, as detailed with the <a href="https://github.com/algorandfoundation/specs/blob/master/dev/ledger.md#multisignature">Multisignature specs</a>.
    - the wallet **MUST** reject if signers is not a subset of `msig.addrs`
    - the wallet **MUST** try to return a `SignedData` object with all the `SignedDataStr` corresponding to `signers[i]` with `i>0`. If it cannot it SHOULD throw a `4001` error.

  - if `signers` length is 1:
    - if `msig` is specified, the wallet **MUST** reject.
    - if `authAddr` is specified the wallet **MUST** reject if `signers[0]` is not equal to `authAddr`.
    - if `hdPath` is specified, the wallet **MUST** reject if `signers[0]` is not equal to the derived public key with the `hdPath` parameters.
    - In all cases, the wallet **MUST** only try to return a `SignedData` object with one `SignedDataStr` for `signers[0]`.

- `authAddr`:
  - The wallet **MAY** not support this field. In that case, it **MUST** throw a 4200 error.
  - if specified, it **MUST** be a valid `Ed25519Pk` object, otherwise the wallet **MUST** reject.
  - is specified and supported, the wallet **MUST** try to return a `SignedData` object that includes a `SignedDataStr` for `authAddr`.

- `msig`:
  - The wallet **MAY** not support this field. In that case, it **MUST** throw a 4200 error.
  - if specified, it **MUST** be a valid `MultisigMetadata` object, otherwise the wallet **MUST** reject.
  - if specified and supported, the wallet **MUST** verify that the `msig` address corresponds to `signers[0]`.
  - If specified and supported, the wallet **MUST** try to return a `SignedData` object with all the `SignedDataStr` it can provide and that the wallet user agrees to provide. If the wallet can produce more signatures than the requested threshold (`msig.threshold`), it **MAY** only provide a `SignedData` object with `msig.threshold` signatures. It is also possible that the wallet cannot provide at least `msig.threshold` signatures (either because the user prevented signing with some keys or because the wallet does not know enough keys). In that case, the returned `SignedData` object will contain only the signatures the wallet can produce. However, the wallet **MUST** provide at least one `SignedDataStr` or throw an error.

- `hdPath`:
  - The wallet **MAY** not support this field. In that case, it **MUST** throw a `4200` error.
  - if specified, it **MUST** be a valid `HDWalletMetadata` object, otherwise the wallet **MUST** reject.
  - if specified and supported, the wallet **MUST** verify that the derivation path resolves to a public key corresponding to `signers[0]`.
  - if specified and supported, the waller **MUST** try to return a `SignedData` object that includes a `SignedDataStr` for `signers[0]`.

#### Semantic of `StdSignMetadata`

WIP

#### General Validation

Every input of the `signData(arbData, metadata)` must be validated. The validation:

- **SHALL NOT** rely on TypeScript typing as this can be bypassed. Types **MUST** be manually verified.

_- **SHALL NOT** assume the Algorand SDK does any validation, as the Algorand SDK `signBytes` method does not check inputs structure. The only exception for the above rule is for de-serialization of transactions. Once de-serialized, every field of the transaction must be manually validated._

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
