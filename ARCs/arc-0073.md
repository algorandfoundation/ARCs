---
arc: 73
title: Algorand Interface Detection Spec
description: A specification for smart contracts and indexers to detect interfaces of smart contracts.
author: William G Hatch (@willghatch)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/165
status: Final
type: Standards Track
category: Interface
sub-category: Application
created: 2023-01-10
requires: 4, 22, 28
---

# Algorand Interface Detection Spec

## Abstract

This ARC specifies an interface detection interface based on <a href="https://eips.ethereum.org/EIPS/eip-165">ERC-165</a>.
This interface allows smart contracts and indexers to detect whether a smart contract implements a particular interface based on an interface selector.


## Motivation

[ARC-4](arc-0004.md) applications have associated Contract or Interface description JSON objects that allow users to call their methods.
However, these JSON objects are communicated outside of the consensus network.
Therefore indexers can not reliably identify contract instances of a particular interface, and smart contracts have no way to detect whether another contract supports a particular interface.
An on-chain method to detect interfaces allows greater composability for smart contracts, and allows indexers to automatically detect implementations of interfaces of interest.


## Specification
The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.


### How Interfaces are Identified

The specification for interfaces is defined by [ARC-4](./arc-0004.md).
This specification extends ARC-4 to define the concept of an interface selector.
We define the interface selector as the XOR of all selectors in the interface.
Selectors in the interface include selectors for methods, selectors for events as defined by [ARC-28](./arc-0028.md), and selectors for potential future kinds of interface components.

As an example, consider an interface that has two methods and one event, `add(uint64,uint64)uint128`, `add3(uint64,uint64,uint64)uint128`, and `alert(uint64)`.
The method selector for the `add` method is the first 4 bytes of the method signature's SHA-512/256 hash.
The SHA-512/256 hash of `add(uint64,uint64)uint128` is `0x8aa3b61f0f1965c3a1cbfa91d46b24e54c67270184ff89dc114e877b1753254a`, so its method selector is `0x8aa3b61f`.
The SHA-512/256 hash of `add3(uint64,uint64,uint64)uint128` is `0xa6fd1477731701dd2126f24facf3492d470cf526e7d4d849fea33d102b45f03d`, so its method selector is `0xa6fd1477`
The SHA-512/256 hash of `alert(uint64)` is `0xc809efe9fd45417226d52b605658b83fff27850a01efeea30f694d1e112d5463`, so its method selector is `0xc809efe9`
The interface selector is defined as the bitwise exclusive or of all method and event selectors, so the interface selector is `0x8aa3b61f XOR 0xa6fd1477 XOR 0xc809efe9`, which is `0xe4574d81`.

### How a Contract will Publish the Interfaces it Implements for Detection

In addition to out-of-band JSON contract or interface description data, a contract that is compliant with this specification shall implement the following interface:

```json
{
  "name": "ARC-73",
  "desc": "Interface for interface detection",
  "methods": [
    {
      "name": "supportsInterface",
      "desc": "Detects support for an interface specified by selector.",
      "readonly": true,
      "args": [
        { "type": "byte[4]", "name": "interfaceID", "desc": "The selector of the interface to detect." },
      ],
      "returns": { "type": "bool", "desc": "Whether the contract supports the interface." }
    }
  ]
}
```

The `supportsInterface` method must be `readonly` as specified by [ARC-22](./arc-0022.md).

The implementing contract must have a `supportsInterface` method that returns:

* `true` when `interfaceID` is `0x4e22a3ba` (the selector for [ARC-73](./arc-0073.md), this interface)
* `false` when `interfaceID` is `0xffffffff`
* `true` for any other `interfaceID` the contract implements
* `false` for any other `interfaceID`


## Rationale

This specification is nearly identical to the related specification for Ethereum, <a href="https://eips.ethereum.org/EIPS/eip-165">ERC-165</a>, merely adapted to Algorand.


## Security Considerations

It is possible that a malicious contract may lie about interface support.
This interface makes it easier for all kinds of actors, inclulding malicious ones, to interact with smart contracts that implement it.


## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
