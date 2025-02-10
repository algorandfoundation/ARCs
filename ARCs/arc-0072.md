---
arc: 72
title: Algorand Smart Contract NFT Specification
description: Base specification for non-fungible tokens implemented as smart contracts.
author: William G Hatch (@willghatch)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/162
status: Living
type: Standards Track
category: Interface
sub-category: Application
created: 2023-01-10
requires: 3, 4, 16, 22, 28, 73
---

# Algorand Smart Contract NFT Specification

## Abstract
This specifies an interface for non-fungible tokens (NFTs) to be implemented on Algorand as smart contracts.
This interface defines a minimal interface for NFTs to be owned and traded, to be augmented by other standard interfaces and custom methods.


## Motivation
Currently most NFTs in the Algorand ecosystem are implemented as ASAs.
However, to provide rich extra functionality, it can be desirable to implement NFTs as a smart contract instead.
To foster an interoperable NFT ecosystem, it is necessary that the core interfaces for NFTs be standardized.


## Specification
The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.


### Core NFT specification

A smart contract NFT that is compliant with this standard must implement the interface detection standard defined in [ARC-73](./arc-0073.md).

Additionally, the smart contract MUST implement the following interface:

```json
{
  "name": "ARC-72",
  "desc": "Smart Contract NFT Base Interface",
  "methods": [
    {
      "name": "arc72_ownerOf",
      "desc": "Returns the address of the current owner of the NFT with the given tokenId",
      "readonly": true,
      "args": [
        { "type": "uint256", "name": "tokenId", "desc": "The ID of the NFT" },
      ],
      "returns": { "type": "address", "desc": "The current owner of the NFT." }
    },
    {
      "name": "arc72_transferFrom",
      "desc": "Transfers ownership of an NFT",
      "readonly": false,
      "args": [
        { "type": "address", "name": "from" },
        { "type": "address", "name": "to" },
        { "type": "uint256", "name": "tokenId" }
      ],
      "returns": { "type": "void" }
    },
  ],
  "events": [
    {
      "name": "arc72_Transfer",
      "desc": "Transfer ownership of an NFT",
      "args": [
        {
          "type": "address",
          "name": "from",
          "desc": "The current owner of the NFT"
        },
        {
          "type": "address",
          "name": "to",
          "desc": "The new owner of the NFT"
        },
        {
          "type": "uint256",
          "name": "tokenId",
          "desc": "The ID of the transferred NFT"
        }
      ]
    }
  ]
}
```

Ownership of a token ID by the zero address indicates that ID is invalid.
The `arc72_ownerOf` method MUST return the zero address for invalid token IDs.
The `arc72_transferFrom` method MUST error when `from` is not the owner of `tokenId`.
The `arc72_transferFrom` method MUST error unless called by the owner of `tokenId` or an approved operator as defined by an extension such as the transfer management extension defined in this ARC.
The `arc72_transferFrom` method MUST emit a `arc72_Transfer` event a transfer is successful.
A `arc72_Transfer` event SHOULD be emitted, with `from` being the zero address, when a token is first minted.
A `arc72_Transfer` event SHOULD be emitted, with `to` being the zero address, when a token is destroyed.

All methods in this and other interfaces defined throughout this standard that are marked as `readonly` MUST be read-only as defined by [ARC-22](./arc-0022.md).

The ARC-73 interface selector for this core interface is `0x53f02a40`.


### Metadata Extension

A smart contract NFT that is compliant with this metadata extension MUST implement the interfaces required to comply with the Core NFT Specification, as well as the following interface:

```json
{
  "name": "ARC-72 Metadata Extension",
  "desc": "Smart Contract NFT Metadata Interface",
  "methods": [
    {
      "name": "arc72_tokenURI",
      "desc": "Returns a URI pointing to the NFT metadata",
      "readonly": true,
      "args": [
        { "type": "uint256", "name": "tokenId", "desc": "The ID of the NFT" },
      ],
      "returns": { "type": "byte[256]", "desc": "URI to token metadata." }
    }
  ],
}
```

URIs shorter than the return length MUST be padded with zero bytes at the end of the URI.
The token URI returned SHOULD be an `ipfs://...` URI so the metadata can't expire or be changed by a lapse or takeover of a DNS registration.
The token URI SHOULD NOT be an `http://` URI due to security concerns.
The URI SHOULD resolve to a JSON file following :
- the JSON Metadata File Schema defined in [ARC-3](./arc-0003.md).
- the standard for declaring traits defined in [ARC-16](./arc-0016.md).

Future standards could define new recommended URI or file formats for metadata.

The ARC-73 interface selector for this metadata extension interface is `0xc3c1fc00`.

### Transfer Management Extension

A smart contract NFT that is compliant with this transfer management extension MUST implement the interfaces required to comply with the Core NFT Specification, as well as the following interface:

```json
{
  "name": "ARC-72 Transfer Management Extension",
  "desc": "Smart Contract NFT Transfer Management Interface",
  "methods": [
    {
      "name": "arc72_approve",
      "desc": "Approve a controller for a single NFT",
      "readonly": false,
      "args": [
        { "type": "address", "name": "approved", "desc": "Approved controller address" },
        { "type": "uint256", "name": "tokenId", "desc": "The ID of the NFT" },
      ],
      "returns": { "type": "void" }
    },
    {
      "name": "arc72_setApprovalForAll",
      "desc": "Approve an operator for all NFTs for a user",
      "readonly": false,
      "args": [
        { "type": "address", "name": "operator", "desc": "Approved operator address" },
        { "type": "bool", "name": "approved", "desc": "true to give approval, false to revoke" },
      ],
      "returns": { "type": "void" }
    },
    {
      "name": "arc72_getApproved",
      "desc": "Get the current approved address for a single NFT",
      "readonly": true,
      "args": [
        { "type": "uint256", "name": "tokenId", "desc": "The ID of the NFT" },
      ],
      "returns": { "type": "address", "desc": "address of approved user or zero" }
    },
    {
      "name": "arc72_isApprovedForAll",
      "desc": "Query if an address is an authorized operator for another address",
      "readonly": true,
      "args": [
        { "type": "address", "name": "owner" },
        { "type": "address", "name": "operator" },
      ],
      "returns": { "type": "bool", "desc": "whether operator is authorized for all NFTs of owner" }
    },
  ],
  "events": [
    {
      "name": "arc72_Approval",
      "desc": "An address has been approved to transfer ownership of the NFT",
      "args": [
        {
          "type": "address",
          "name": "owner",
          "desc": "The current owner of the NFT"
        },
        {
          "type": "address",
          "name": "approved",
          "desc": "The approved user for the NFT"
        },
        {
          "type": "uint256",
          "name": "tokenId",
          "desc": "The ID of the NFT"
        }
      ]
    },
    {
      "name": "arc72_ApprovalForAll",
      "desc": "Operator set or unset for all NFTs defined by this contract for an owner",
      "args": [
        {
          "type": "address",
          "name": "owner",
          "desc": "The current owner of the NFT"
        },
        {
          "type": "address",
          "name": "operator",
          "desc": "The approved user for the NFT"
        },
        {
          "type": "bool",
          "name": "approved",
          "desc": "Whether operator is authorized for all NFTs of owner "
        }
      ]
    },
  ]
}
```

The `arc72_Approval` event MUST be emitted when the `arc72_approve` method is called successfully.
The zero address for the `arc72_approve` method and the `arc72_Approval` event indicate no approval, including revocation of previous single NFT controller.
When a `arc72_Transfer` event emits, this also indicates that the approved address for that NFT (if any) is reset to none.
The `arc72_ApprovalForAll` event MUST be emitted when the `arc72_setApprovalForAll` method is called successfully.
The contract MUST allow multiple operators per owner.
The `arc72_transferFrom` method, when its `nftId` argument is owned by its `from` argument, MUST succeed for when called by an address that is approved for the given NFT or approved as operator for the owner.

The ARC-73 interface selector for this transfer management extension interface is `0xb9c6f696`.

### Enumeration Extension

A smart contract NFT that is compliant with this enumeration extension MUST implement the interfaces required to comply with the Core NFT Specification, as well as the following interface:

```json
{
  "name": "ARC-72 Enumeration Extension",
  "desc": "Smart Contract NFT Enumeration Interface",
  "methods": [
    {
      "name": "arc72_balanceOf",
      "desc": "Returns the number of NFTs owned by an address",
      "readonly": true,
      "args": [
        { "type": "address", "name": "owner" },
      ],
      "returns": { "type": "uint256" }
    },
    {
      "name": "arc72_totalSupply",
      "desc": "Returns the number of NFTs currently defined by this contract",
      "readonly": true,
      "args": [],
      "returns": { "type": "uint256" }
    },
    {
      "name": "arc72_tokenByIndex",
      "desc": "Returns the token ID of the token with the given index among all NFTs defined by the contract",
      "readonly": true,
      "args": [
        { "type": "uint256", "name": "index" },
      ],
      "returns": { "type": "uint256" }
    },
  ],
}
```

The sort order for NFT indices is not specified.
The `arc72_tokenByIndex` method MUST error when `index` is greater than `arc72_totalSupply`.

The ARC-73 interface selector for this enumeration extension interface is `0xa57d4679`.


## Rationale

This specification is based on <a href="https://eips.ethereum.org/EIPS/eip-721">ERC-721</a>, with some differences.

### Core Specification

The core specification differs from ERC-721 by:

* removing `safeTransferFrom`, since there is not a test for whether an address on Algorand corresponds to a smart contract
* moving management functionality out of the base specification into an extension
* moving balance query functionality out of the base specification into the enumeration extension

Moving functionality out of the core specification into extensions allows the base specification to be much simpler, and allows extensions for extra capabilities to evolve separately from the core idea of owning and transferring ownership of non-fungible tokens.
It is recommended that NFT contract authors make use of extensions to enrich the capabilities of their NFTs.

### Metadata Extension

The metadata extension differns from the ERC-721 metadata extension by using a fixed-length URI return and removing the `symbol` and `name` operations.  Metadata such as symbol or name can be included in the metadata pointed to by the URI.

### Transfer Management Extension

The transfer management extension is taken from the set of methods and events from the base ERC-721 specification that deal with approving other addresses to transfer ownership of an NFT.
This functionality is important for trusted NFT galleries like OpenSea to list and sell NFTs on behalf of users while allowing the owner to maintain on-chain ownership.
However, this set of functionality is the bulk of the complexity of the ERC-721 standard, and moving it into an extension vastly simplifies the core NFT specification.
Additionally, other interfaces have been proposed to allow for the sale of NFTs in decentralized manners without needing to give transfer control to a trusted third party.

### Enumeration Extension

The enumeration extension is taken from the ERC-721 enumeration extension.
However, it also includes the `arc72_balanceOf` function that is included in the base ERC-721 specification.
This change simplifies the core standard and groups the `arc72_balanceOf` function with related functionality for contracts where supply details are desired.


## Backwards Compatibility

This standard introduces a new kind of NFT that is incompatible with NFTs defined as ASAs.
Applications that want to index, manage, or view NFTs on Algorand will need to handle these new smart NFTs as well as the already popular ASA implementation of NFTs will need to add code to handle both, and existing smart contracts that handle ASA-based NFTs will not work with these new smart contract NFTs.

While this is a severe backwards incompatibility, smart contract NFTs are necessary to provide richer and more diverse functionality for NFTs.


## Security Considerations

The fact that anybody can create a new implementation of a smart contract NFT standard opens the door for many of those implementations to contain security bugs.
Additionally, malicious NFT implementations could contain hidden anti-features unexpected by users.
As with other smart contract domains, it is difficult for users to verify or understand security properties of smart contract NFTs.
This is a tradeoff compared with ASA NFTs, which share a smaller set of security properties that are easier to validate, to gain the possibility of adding novel features.


## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
