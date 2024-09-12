---
arc: 74
title: NFT Indexer API
description: REST API for reading data about Application's NFTs.
author: William G Hatch (@willghatch)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/170
status: Final
type: Standards Track
category: Interface
sub-category: Application
created: 2023-02-17
requires: 72
---


## Abstract
This specifies a REST interface that can be implemented by indexing services to provide data about NFTs conforming to the [ARC-72](arc-0072.md) standard.

## Motivation
While most data is available on-chain, reading and analyzing on-chain logs to get a complete and current picture about NFT ownership and history is slow and impractical for many uses.
This REST interface standard allows analysis of NFT contracts to be done in a centralized manner to provide fast, up-to-date responses to queries, while allowing users to pick from any indexing provider.


## Specification

This specification defines two REST endpoints: `/nft-index/v1/tokens` and `/nft-index/v1/transfers`.
Both endpoints respond only to `GET` requests, take no path parameters, and consume no input.
But both accept a variety of query parameters.

### `GET /nft-indexer/v1/tokens`

Produces `application/json`.

Optional Query Parameters:

| Name | Schema | Description |
| --- | --- | --- |
| round | integer | Include results for the specified round.  For performance reasons, this parameter may be disabled on some configurations. |
| next | string | Token for the next page of results.  Use the `next-token` provided by the previous page of results. |
| limit | integer | Maximum number of results to return.  There could be additional pages even if the limit is not reached. |
| contractId | integer | Limit results to NFTs implemented by the given contract ID. |
| tokenId | integer | Limit results to NFTs with the given token ID. |
| owner | address | Limit results to NFTs owned by the given owner. |
| mint-min-round | integer | Limit results to NFTs minted on or after the given round. |
| mint-max-round | integer | Limit results to NFTs minted on or before the given round. |

When successful, returns a response with code 200 and an object with the schema:

| Name | Required? | Schema | Description |
| --- | --- | --- | --- |
| tokens | required | <Token> array | Array of Token objects that fit the query parameters, as defined below. |
| current-round | required | integer | Round at which the results were computed. |
| next-token | optional | string | Used for pagination, when making another request provide this token as the `next` parameter. |

The `Token` object has the following schema:

| Name | Required? | Schema | Description |
| --- | --- | --- | --- |
| owner | required | address | The current owner of the NFT. |
| contractId | required | integer | The ID of the ARC-72 contract that defines the NFT. |
| tokenId | required | integer | The tokenID of the NFT, which along with the contractId addresses a unique ARC-72 token. |
| mint-round | optional | integer | The round at which the NFT was minted (IE the round at which it was transferred from the zero address to the first owner). |
| metadataURI | optional | string | The URI given for the token by the `metadataURI` API of the contract, if applicable. |
| metadata | optional | object | The result of resolving the `metadataURI`, if applicable and available. |

When unsuccessful, returns a response with code 400 or 500 and an object with the schema:

| Name | Required? | Schema |
| --- | --- | --- |
| data | optional | object |
| message | required | string |

### `GET /nft-indexer/v1/transfers`

Produces `application/json`.

Optional Query Parameters:

| Name | Schema | Description |
| --- | --- | --- |
| round | integer | Include results for the specified round.  For performance reasons, this parameter may be disabled on some configurations. |
| next | string | Token for the next page of results.  Use the `next-token` provided by the previous page of results. |
| limit | integer | Maximum number of results to return.  There could be additional pages even if the limit is not reached. |
| contractId | integer | Limit results to NFTs implemented by the given contract ID. |
| tokenId | integer | Limit results to NFTs with the given token ID. |
| user | address | Limit results to transfers where the user is either the sender or receiver. |
| from | address | Limit results to transfers with the given address as the sender. |
| to | address | Limit results to transfers with the given address as the receiver. |
| min-round | integer | Limit results to transfers that were executed on or after the given round. |
| max-round | integer | Limit results to transfers that were executed on or before the given round. |

When successful, returns a response with code 200 and an object with the schema:

| Name | Required? | Schema | Description |
| --- | --- | --- | --- |
| transfers | required | <Transfer> array | Array of Transfer objects that fit the query parameters, as defined below. |
| current-round | required | integer | Round at which the results were computed. |
| next-token | optional | string | Used for pagination, when making another request provide this token as the `next` parameter. |

The `Transfer` object has the following schema:

| Name | Required? | Schema | Description |
| --- | --- | --- | --- |
| contractId | required | integer | The ID of the ARC-72 contract that defines the NFT. |
| tokenId | required | integer | The tokenID of the NFT, which along with the contractId addresses a unique ARC-72 token. |
| from | required | address | The sender of the transaction. |
| to | required | address | The receiver of the transaction. |
| round | required | integer | The round of the transfer. |

When unsuccessful, returns a response with code 400 or 500 and an object with the schema:

| Name | Required? | Schema |
| --- | --- | --- |
| data | optional | object |
| message | required | string |

## Rationale
This standard was designed to feel similar to the Algorand indexer API, and uses the same query parameters and results where applicable.

## Backwards Compatibility
This standard presents a versioned REST interface, allowing future extensions to change the interface in incompatible ways while allowing for the old service to run in tandem.


## Security Considerations
All data available through this indexer API is publicly available.

## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
