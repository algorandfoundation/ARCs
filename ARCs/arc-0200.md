---
arc: 200
title: Algorand Smart Contract Token Specification
description: Base specification for tokens implemented as smart contracts
author: Nicholas Shellabarger (@temptemp3)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/223
status: Living
type: Standards Track
category: Interface
sub-category: Application
created: 2023-07-03
requires: 3, 4, 22, 28
---

# Algorand Smart Contract Token Specification

## Abstract

This ARC (Algorand Request for Comments) specifies an interface for tokens to be implemented on Algorand as smart contracts. The interface defines a minimal interface required for tokens to be held and transferred, with the potential for further augmentation through additional standard interfaces and custom methods.

## Motivation

Currently, most tokens in the Algorand ecosystem are represented by ASAs (Algorand Standard Assets). However, to provide rich extra functionality, it can be desirable to implement tokens as smart contracts instead. To foster an interoperable token ecosystem, it is necessary that the core interfaces for tokens be standardized.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

### Core Token specification

A smart contract token that is compliant with this standard MUST implement the following interface:

```json
{
  "name": "ARC-200",
  "desc": "Smart Contract Token Base Interface",
  "methods": [
    {
      "name": "arc200_name",
      "desc": "Returns the name of the token",
      "readonly": true,
      "args": [],
      "returns": { "type": "byte[32]", "desc": "The name of the token" }
    },
    {
      "name": "arc200_symbol",
      "desc": "Returns the symbol of the token",
      "readonly": true,
      "args": [],
      "returns": { "type": "byte[8]", "desc": "The symbol of the token" }
    },
    {
      "name": "arc200_decimals",
      "desc": "Returns the decimals of the token",
      "readonly": true,
      "args": [],
      "returns": { "type": "uint8", "desc": "The decimals of the token" }
    },
    {
      "name": "arc200_totalSupply",
      "desc": "Returns the total supply of the token",
      "readonly": true,
      "args": [],
      "returns": { "type": "uint256", "desc": "The total supply of the token" }
    },
    {
      "name": "arc200_balanceOf",
      "desc": "Returns the current balance of the owner of the token",
      "readonly": true,
      "args": [
        {
          "type": "address",
          "name": "owner",
          "desc": "The address of the owner of the token"
        }
      ],
      "returns": {
        "type": "uint256",
        "desc": "The current balance of the holder of the token"
      }
    },
    {
      "name": "arc200_transfer",
      "desc": "Transfers tokens",
      "readonly": false,
      "args": [
        {
          "type": "address",
          "name": "to",
          "desc": "The destination of the transfer"
        },
        {
          "type": "uint256",
          "name": "value",
          "desc": "Amount of tokens to transfer"
        }
      ],
      "returns": { "type": "bool", "desc": "Success" }
    },
    {
      "name": "arc200_transferFrom",
      "desc": "Transfers tokens from source to destination as approved spender",
      "readonly": false,
      "args": [
        {
          "type": "address",
          "name": "from",
          "desc": "The source  of the transfer"
        },
        {
          "type": "address",
          "name": "to",
          "desc": "The destination of the transfer"
        },
        {
          "type": "uint256",
          "name": "value",
          "desc": "Amount of tokens to transfer"
        }
      ],
      "returns": { "type": "bool", "desc": "Success" }
    },
    {
      "name": "arc200_approve",
      "desc": "Approve spender for a token",
      "readonly": false,
      "args": [
        { "type": "address", "name": "spender" },
        { "type": "uint256", "name": "value" }
      ],
      "returns": { "type": "bool", "desc": "Success" }
    },
    {
      "name": "arc200_allowance",
      "desc": "Returns the current allowance of the spender of the tokens of the owner",
      "readonly": true,
      "args": [
        { "type": "address", "name": "owner" },
        { "type": "address", "name": "spender" }
      ],
      "returns": { "type": "uint256", "desc": "The remaining allowance" }
    }
  ],
  "events": [
    {
      "name": "arc200_Transfer",
      "desc": "Transfer of tokens",
      "args": [
        {
          "type": "address",
          "name": "from",
          "desc": "The source of transfer of tokens"
        },
        {
          "type": "address",
          "name": "to",
          "desc": "The destination of transfer of tokens"
        },
        {
          "type": "uint256",
          "name": "value",
          "desc": "The amount of tokens transferred"
        }
      ]
    },
    {
      "name": "arc200_Approval",
      "desc": "Approval of tokens",
      "args": [
        {
          "type": "address",
          "name": "owner",
          "desc": "The owner of the tokens"
        },
        {
          "type": "address",
          "name": "spender",
          "desc": "The approved spender of tokens"
        },
        {
          "type": "uint256",
          "name": "value",
          "desc": "The amount of tokens approve"
        }
      ]
    }
  ]
}
```

Ownership of a token by a zero address indicates that a token is out of circulation indefinitely, or otherwise burned or destroyed.

The methods `arc200_transfer` and `arc200_transferFrom` method MUST error when the balance of `from` is insufficient. In the case of the `arc200_transfer` method, from is implied as the `owner` of the token.
The `arc200_transferFrom` method MUST error unless called by an `spender` approved by an `owner`.
The methods `arc200_transfer` and `arc200_transferFrom` MUST emit a `Transfer` event.
A `arc200_Transfer` event SHOULD be emitted, with `from` being the zero address, when a token is minted.
A `arc200_Transfer` event SHOULD be emitted, with `to` being the zero address, when a token is destroyed.

The `arc200_Approval` event MUST be emitted when an `arc200_approve` or `arc200_transferFrom` method is called successfully.

A value of zero for the `arc200_approve` method and the `arc200_Approval` event indicates no approval.
The `arc200_transferFrom` method and the `arc200_Approval` event indicates the approval value after it is decremented.

The contract MUST allow multiple operators per owner.

All methods in this standard that are marked as `readonly` MUST be read-only as defined by [ARC-22](./arc-0022.md).

## Rationale

This specification is based on <a href="https://eips.ethereum.org/EIPS/eip-20">ERC-20</a>.

### Core Specification

The core specification identical to ERC-20.

## Backwards Compatibility

This standard introduces a new kind of token that is incompatible with tokens defined as ASAs.
Applications that want to index, manage, or view tokens on Algorand will need to handle these new smart tokens as well as the already popular ASA implementation of tokens will need to add code to handle both, and existing smart contracts that handle ASA-based tokens will not work with these new smart contract tokens.

While this is a severe backward incompatibility, smart contract tokens are necessary to provide richer and more diverse functionality for tokens.

## Security Considerations

The fact that anybody can create a new implementation of a smart contract tokens standard opens the door for many of those implementations to contain security bugs.
Additionally, malicious token implementations could contain hidden anti-features unexpected by users.
As with other smart contract domains, it is difficult for users to verify or understand the security properties of smart contract tokens.
This is a tradeoff compared with ASA tokens, which share a smaller set of security properties that are easier to validate to gain the possibility of adding novel features.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
