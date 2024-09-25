---
arc: 16
title: Convention for declaring traits of an NFT's
description: This is a convention for declaring traits in an NFT's metadata.
author: Keegan Thompson (@keeganthomp)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/62
status: Final
type: Standards Track
category: ARC
sub-category: Asa
created: 2022-01-04
---

# Standard for declaring traits inside non-fungible NFT's metadata

## Abstract

The goal is to establish a standard for how traits are declared inside a non-fungible NFT's metadata, for example as specified in ([ARC-3](./arc-0003.md)), ([ARC-69](./arc-0069.md)) or ([ARC-72](./arc-0072.md)).

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

> Comments like this are non-normative.

If the property `traits` is provided anywhere in the metadata, it **MUST** adhere to the schema below.
If the NFT is a part of a larger collection and that collection has traits, all the available traits for the collection **MUST** be listed as a property of the `traits` object.
If the NFT does not have a particular trait, it's value **MUST** be "none".

The JSON schema for `traits` is as follows:

```json
{
  "title": "Traits for Non-Fungible Token",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "description": "Traits (attributes) that can be used to calculate things like rarity. Values may be strings or numbers"
    }
  }
}
```

#### Examples

##### Example of an NFT that has traits

```json
{
  "name": "NFT With Traits",
  "description": "NFT with traits",
  "image": "https://s3.amazonaws.com/your-bucket/images/two.png",
  "image_integrity": "sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
  "properties": {
    "creator": "Tim Smith",
    "created_at": "January 2, 2022",
    "traits": {
      "background": "red",
      "shirt_color": "blue",
      "glasses": "none",
      "tattoos": 4,
    }
  }
}
```

##### Example of an NFT that has no traits

```json
{
  "name": "NFT Without Traits",
  "description": "NFT without traits",
  "image": "https://s3.amazonaws.com/your-bucket/images/one.png",
  "image_integrity": "sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
  "properties": {
    "creator": "John Smith",
    "created_at": "January 1, 2022",
  }
}
```

## Rationale

A standard for traits is needed so programs know what to expect in order to calculate things like rarity.

## Backwards Compatibility

If the metadata does not have the field `traits`, each value of `properties` should be considered a trait.

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
