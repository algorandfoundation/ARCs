---
arc: 36
title: Convention for declaring filters of an NFT
description: This is a convention for declaring filters in an NFT metadata
author: Stéphane Barroso (@sudoweezy)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/163
status: Final
type: Standards Track
category: ARC
sub-category: Asa
created: 2023-03-10
---

# Standard for declaring filters inside non-fungible token metadata

## Abstract

The goal is to establish a standard for how filters are declared inside a non-fungible (NFT) metadata.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

> Comments like this are non-normative.

If the property `filters` is provided anywhere in the metadata of an nft, it **MUST** adhere to the schema below.
If the nft is a part of a larger collection and that collection has filters, all the available filters for the collection **MUST** be listed as a property of the `filters` object.
If the nft does not have a particular filter, it's value **MUST** be "none".

The JSON schema for `filters` is as follows:

```json
{
  "title": "Filters for Non-Fungible Token",
  "type": "object",
  "properties": {
    "filters": {
      "type": "object",
      "description": "Filters can be used to filter nfts of a collection.  Values must be an array of strings or numbers."
    }
  }
}
```

#### Examples

##### Example of an NFT that has traits & filters

```json
{
  "name": "NFT With Traits & filters",
  "description": "NFT with traits & filters",
  "image": "https://s3.amazonaws.com/your-bucket/images/two.png",
  "image_integrity": "sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
  "properties": {
    "creator": "Tim Smith",
    "created_at": "January 2, 2022",
    "traits": {
      "background": "yellow",
      "head": "curly"
    },
    "filters": {
      "xp": 120,
      "state": "REM"
    }
  }
}
```

## Rationale

A standard for filters is needed so programs know what to expect in order to filter things without using rarity.

## Backwards Compatibility

If `filters` wants to be added on top of fields [ARC-16](./arc-0016.md) `traits` and `filters` should be inside the `properties` object. (eg: [Example above](./arc-0036.md#example-of-an-nft-that-has-traits--filters))

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
