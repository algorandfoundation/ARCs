---
arc: 44
title: ABI Extension AVM 9+
description: Application Binary Interface Extension for AVM 9+
author: Jannotti (@jannotti), Jason Paulos (@jasonpaulos)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/44
status: Final
type: Standards Track
category: Interface
created: 2021-07-29
---

# Algorand Transaction Calling Conventions

## Abstract

This document introduces an [ARC-4](arc-0004.md) extension for encoding method calls, including argument and return value encoding, in Algorand Application call transactions using AVM > 9.
The goal is to allow clients, such as wallets and
dapp frontends, to properly encode call transactions based on a description 
of the interface. Further, explorers will be able to show details of
these method invocations.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

> Comments like this are non-normative.

Interfaces are defined in TypeScript. All the objects that are defined 
are valid JSON objects, and all JSON `string` types are UTF-8 encoded.


#### Types

The following types are supported in the Algorand ABI.
* `uint<N>`: An `N`-bit unsigned integer, where `8 <= N <= 512` and `N % 8 = 0`. When this type is used as part of a method signature, `N` must be written as a base 10 number without any leading zeros.
* `byte`: An alias for `uint8`.
* `address`: Used to represent a 32-byte Algorand address. This is equivalent to `byte[32]`.
* `string`: A variable-length byte array (`byte[]`) assumed to contain UTF-8 encoded content.
  
* `ubigint`: A variable-length byte array (`byte[]`) assumed to contain a variable-width unsigned integer.
* `timestamp`: An alias for `uint64` assumed to represent a Unix epoch timestamp

* reference types `account`, `asset`, `application`, `box`: **MUST NOT** be used as the return type.
For encoding purposes they are an alias for `uint8`. See section "Reference Types" below.

For encoding purposes they are an alias for `uint8`. See section "Reference Types" below.

Additional special use types are defined in [Reference Types](#reference-types)
and [Transaction Types](#transaction-types).

#### Static vs Dynamic Types

For encoding purposes, the types are divided into two categories: static and dynamic.

The dynamic types are:

*  `<type>[]` for any `type`
    * This includes `string` and `ubigint` types since they are aliases for `byte[]`.



#### Encoding Rules

Let `len(a)` be the number of bytes in the binary string `a`. The
returned value shall be considered to have the ABI type `uint16`.
Let `enc` be a mapping from values of the ABI types to binary
strings. This mapping defines the encoding of the ABI.
For any ABI value `x`, we recursively define `enc(x)` to be as follows:

* If `x` is a unsigned variable-width integer, `ubigint` 
  * `enc(x)` is encoded as `byte[]` containing the big-endian encoding of `x`. `enc(x)` SHOULD be as succint as possible without 0-padding beyond the highest MSB set.

Other aliased types' encodings are already covered:

- `string` is an alias for `byte[]`
- `address` is an alias for `byte[32]`
- `byte` is an alias for `uint8`
- `timestamp` is an alias for `uint64`
- each of the reference types is an alias for `uint8`

### Reference Types
Four special types are supported _only_ as the type of an argument.
They _can_ be embedded in arrays and tuples.
* `account` represents an Algorand account, stored in the Accounts (`apat`) array
* `asset` represents an Algorand Standard Asset (ASA), stored in the Foreign Assets (`apas`) array
* `application` represents an Algorand Application, stored in the Foreign Apps (`apfa`) array
* `box` represents a Box owned by an Algorand Application, stored in the Boxes (`apbx`) array

Some AVM opcodes require specific values to be placed in the "foreign arrays" of the Application call transaction. These four types allow methods to describe these requirements. To encode method calls that have these types as arguments, the value in question is placed in the Accounts (`apat`), Foreign Assets (`apas`), Foreign Apps (`apfa`) or Boxes (`apbx`) arrays, respectively, and a `uint8` containing the index of the value in the appropriate array is encoded in the normal location for this argument.

## Rationale

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.