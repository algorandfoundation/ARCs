---
arc: xx
title: Algorand standard for mutable collection metadata as an ASA.
description: This is a standard for defining mutable, arbitrarily large collections of ASAs.
author: Evan Maltz/Kinn Foundation @evan_maltz
status: Draft
type: Standards Track
category: ARC
created: 2022-08-12
discussions-to: https://github.com/algorandfoundation/ARCs/issues/25
---

# Standard for defining mutable collections of ASAs

## Summary

This document introduces the standard for declaring collection metadata, which represent an arbitrary group of indices.

## Abstract

This ARC introduces a standard for defining a collection ASA (cASA) linked to ARC19 metadata. This standard defines a collection as a mutable set of indices, providing flexibility. By linking to ARC19 metadata on IPFS, the set can be arbitrarily large, even including the complete set of all indices. Indices can be application IDs, ASA IDs, or addresses. 

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in [RFC-2119](https://www.ietf.org/rfc/rfc2119.txt).

> Comments like this are non-normative.

To be considered a collection, the cASA **MUST** use the unit name `CLXN` and **MUST** contain a property `assets`. The `assets` property **MUST** contain an array with at least one index. A collection name and description are **RECOMMENDED**. All other fields are **OPTIONAL** and additional, custom fields **MAY** be included. The cASA **SHOULD** be minted using ARC19 to allow for immutability, but **MAY** be minted in the ARC3 format if immutability is preferred.

The JSON schema is as follows:

```json
{
  "name": "Name of the collection",
  "assets": [],
  "description": "Description of purpose and/or contents of the collection"
}
```

#### Examples

##### Example of a cASA containing ASA indices

```json
{
  "name": "A cASA of ASAs",
  "assets": [417708610, 470842789, 230946361],
  "description": "My favorite tokens."
}
```

##### Example of a cASA with mixed indices and custom fields

```json
{
  "name": "Cell NFTs",
  "assets": [289741461, 333143277, 374937094, 290027123, 463778159],
  "description": "A mix of ASAs and application indices of NFTs about cells.",
  "NFD": "cells.algo",
  "register": "KinnDAO Registered Collection"
}
```

## Rationale

NFT collections are currently only defined off chain, and many sites have their own index of collections that require independent verification. This standard allows anyone to define an NFT collection with optional verification. Anyone can assemble their own collection list from data on chain without relying on an intermediary or gatekeeper, representing a common reference. Creators can mint a cASA to group their own creations, and any user can mint a collection from multiple creators for arbitrary categorization. An ASA can be part of multiple collections or none at all. Using ARC19 as the base format allows cASAs to be mutable: new assets can be added to or removed from the list dynamically. A consistent unit name ('CLXN') allows for easy asset search by unit name. Allowing collections of applications or addresses provides further flexibility and new functionality. Using this specification applies no additional requirements to the creation of assets, or the existing standards for doing so, which allows for backwards compatibility with existing minted items.

## Copyright

Copyright and related rights waived via [CC0](https://creativecommons.org/publicdomain/zero/1.0/).
