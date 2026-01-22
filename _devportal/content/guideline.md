---
title: ARC Category Guidelines
description: ARCs by categories
sidebar:
  label: Guildelines
  order: 2
---

Welcome to the Guideline. Here you'll find information on which ARCs to use for your project.
## General ARCs

### ARC 0 - ARC Purpose and Guidelines

#### What is an ARC?
ARC stands for Algorand Request for Comments. An ARC is a design document providing information to the Algorand community or describing a new feature for Algorand or its processes or environment.
The ARC should provide a concise technical specification and a rationale for the feature.
The ARC author is responsible for building consensus within the community and documenting dissenting opinions.
We intend ARCs to be the primary mechanisms for proposing new features and collecting community technical input on an issue.
We maintain ARCs as text files in a versioned repository. Their revision history is the historical record of the feature proposal.

### ARC 26 - URI scheme

This URI specification represents a standardized way for applications and websites to send requests and information through deeplinks, QR codes, etc. It is heavily based on Bitcoin’s <a href="https://github.com/bitcoin/bips/blob/master/bip-0021.mediawiki">BIP-0021</a> and should be seen as derivative of it. The decision to base it on BIP-0021 was made to make it easy and compatible as possible for any other application.

### ARC 65 - AVM Run Time Errors In Program

This document introduces a convention for rising informative run time errors on
the Algorand Virtual Machine (AVM) directly from the program bytecode.

### ARC 78 - URI scheme, keyreg Transactions extension

This URI specification represents an extension to the base Algorand URI encoding standard ([ARC-26](/arc-standards/arc-0026)) that specifies encoding of key registration transactions through deeplinks, QR codes, etc.

### ARC 79 - URI scheme, App NoOp call extension

NoOp calls are Generic application calls to execute the Algorand smart contract ApprovalPrograms.
This URI specification proposes an extension to the base Algorand URI encoding standard ([ARC-26](/arc-standards/arc-0026)) that specifies encoding of application NoOp transactions into <a href="https://www.rfc-editor.org/rfc/rfc3986">RFC 3986</a> standard URIs.

### ARC 82 - URI scheme blockchain information

This URI specification defines a standardized method for querying application and asset data on Algorand.
It enables applications, websites, and QR code implementations to construct URIs that allow users to retrieve data such as application state and asset metadata in a structured format.
This specification is inspired by [ARC-26](/arc-standards/arc-0026) and follows similar principles, with adjustments specific to read-only queries for applications and assets.

### ARC 83 - xGov Council - Application Process

The goal of this ARC is to clearly define the process for running for an xGov Council seat.

### ARC 86 - xGov status and voting power

This ARC defines the Expert Governor (xGov) status and voting power in the Algorand Expert Governance.

### ARC 90 - URI scheme

This ARC defines a unified Algorand URI scheme that covers payment transactions, key registration, application NoOp calls, and read-only blockchain queries. It expands on earlier URI specifications to support deeplinks, QR codes, and other contexts where structured URIs communicate transaction intent or state queries.

## Asa ARCs

### ARC 3 - Conventions Fungible/Non-Fungible Tokens

The goal of these conventions is to make it simpler for block explorers, wallets, exchanges, marketplaces, and more generally, client software to display the properties of a given ASA.

### ARC 16 - Convention for declaring traits of an NFT's

The goal is to establish a standard for how traits are declared inside a non-fungible NFT's metadata, for example as specified in ([ARC-3](/arc-standards/arc-0003)), ([ARC-69](/arc-standards/arc-0069)) or ([ARC-72](/arc-standards/arc-0072)).

### ARC 19 - Templating of NFT ASA URLs for mutability

This ARC describes a template substitution for URLs in ASAs, initially for ipfs:// scheme URLs allowing mutable CID replacement in rendered URLs.
The proposed template-XXX scheme has substitutions like:
```
template-ipfs://{ipfscid:<version>:<multicodec>:<field name containing 32-byte digest, ie reserve>:<hash type>}[/...]
```
This will allow modifying the 32-byte 'Reserve address' in an ASA to represent a new IPFS content-id hash. Changing of the reserve address via an asset-config transaction will be all that is needed to point an ASA URL to new IPFS content. The client reading this URL, will compose a fully formed IPFS Content-ID based on the version, multicodec, and hash arguments provided in the ipfscid substitution.

### ARC 20 - Smart ASA

A "Smart ASA" is an Algorand Standard Asset (ASA) controlled by a Smart Contract
that exposes methods to create, configure, transfer, freeze, and destroy the
asset.
This ARC defines the ABI interface of such a Smart Contract, the required
metadata, and suggests a reference implementation.

### ARC 36 - Convention for declaring filters of an NFT

The goal is to establish a standard for how filters are declared inside a non-fungible (NFT) metadata.

### ARC 62 - ASA Circulating Supply

This ARC introduces a standard for the definition of circulating supply for Algorand
Standard Assets (ASA) and its client-side retrieval. A reference implementation is
suggested.

### ARC 69 - ASA Parameters Conventions, Digital Media

The goal of these conventions is to make it simpler to display the properties of a given ASA. This ARC differs from [ARC-3](/arc-standards/arc-0003) by focusing on optimization for fetching of digital media, as well as the use of onchain metadata. Furthermore, since asset configuration transactions are used to store the metadata, this ARC can be applied to existing ASAs.
While mutability helps with backwards compatibility and other use cases, like leveling up an RPG character, some use cases call for immutability. In these cases, the ASA manager MAY remove the manager address, after which point the Algorand network won't allow anyone to send asset configuration transactions for the ASA. This effectively makes the latest valid [ARC-69](/arc-standards/arc-0069) metadata immutable.

### ARC 71 - Non-Transferable ASA

The goal is to make it simpler for block explorers, wallets, exchanges, marketplaces, and more generally, client software to identify & interact with a Non-transferable ASA (NTA).
This defines an interface extending [ARC-3](/arc-standards/arc-0003) & [ARC-69](/arc-standards/arc-0069) non fungible ASA to create Non-transferable ASA. Before issuance, both parties (issuer and receiver), have to agree on who has (if any) the authorization to burn this ASA.
> This spec is compatible with [ARC-19](/arc-standards/arc-0019) to create an updatable Non-transferable ASA.

### ARC 89 - ASA Metadata Registry

This ARC defines the interface and the implementation of a singleton Application
that provides Algorand Standard Assets metadata through the Algod API or the AVM.

## Application ARCs

### ARC 4 - Application Binary Interface (ABI)

This document introduces conventions for encoding method calls,
including argument and return value encoding, in Algorand Application
call transactions.
The goal is to allow clients, such as wallets and
dapp frontends, to properly encode call transactions based on a description
of the interface. Further, explorers will be able to show details of
these method invocations.
#### Definitions
* **Application:** an Algorand Application, aka "smart contract",
  "stateful contract", "contract", or "app".
* **HLL:** a higher level language that compiles to TEAL bytecode.
* **dapp (frontend)**: a decentralized application frontend, interpreted here to
  mean an off-chain frontend (a webapp, native app, etc.) that interacts with
  Applications on the blockchain.
* **wallet**: an off-chain application that stores secret keys for on-chain
  accounts and can display and sign transactions for these accounts.
* **explorer**: an off-chain application that allows browsing the blockchain,
  showing details of transactions.

### ARC 18 - Royalty Enforcement Specification

A specification to describe a set of methods that offer an API to enforce Royalty Payments https://en.wikipedia.org/wiki/Royalty_payment to a Royalty Receiver given a policy describing the royalty shares, both on primary and secondary sales.
This is an implementation of an [ARC-20](/arc-standards/arc-0020) specification and other methods may be implemented in the same contract according to that specification.

### ARC 21 - Round based datafeed oracles on Algorand

The following document introduces conventions for building round based datafeed oracles on Algorand using the ABI defined in [ARC-4](/arc-standards/arc-0004)

### ARC 22 - Add `read-only` annotation to ABI methods

The goal of this convention is to allow smart contract developers to distinguish between methods which mutate state and methods which don't by introducing a new property to the `Method` descriptor.

### ARC 23 - Sharing Application Information

The following document introduces a convention for appending information (stored in various files) to the compiled application's bytes.
The goal of this convention is to standardize the process of verifying and adding this information.
The encoded information byte string is `arc23` followed by the IPFS CID v1 of a folder containing the files with the information.
The minimum required file is `contract.json` representing the contract metadata (as described in [ARC-4](/arc-standards/arc-0004)), and as extended by future potential ARCs).

### ARC 28 - Algorand Event Log Spec

Algorand dapps can use the <a href="https://developer.algorand.org/docs/get-details/dapps/avm/teal/opcodes/#log">`log`</a>  primitive to attach information about an application call. This ARC proposes the concept of Events, which are merely a way in which data contained in these logs may be categorized and structured.
In short: to emit an Event, a dapp calls `log` with ABI formatting of the log data, and a 4-byte prefix to indicate which Event it is.

### ARC 32 - Application Specification

> [!NOTE]
> This specification will be eventually deprecated by the <a href="https://github.com/algorandfoundation/ARCs/pull/258">`ARC-56`</a> specification.
An Application is partially defined by it's [methods](/arc-standards/arc-0004) but further information about the Application should be available.  Other descriptive elements of an application may include it's State Schema, the original TEAL source programs, default method arguments, and custom data types.  This specification defines the descriptive elements of an Application that should be available to clients to provide useful information for an Application Client.

### ARC 54 - ASA Burning App

This ARC provides TEAL which would deploy a application that can be used for burning Algorand Standard Assets. The goal is to have the apps deployed on the public networks using this TEAL to provide a standardized burn address and app ID.

### ARC 56 - Extended App Description

This ARC takes the existing JSON description of a contract as described in [ARC-4](/arc-standards/arc-0004) and adds more fields for the purpose of client interaction

### ARC 72 - Algorand Smart Contract NFT Specification

This specifies an interface for non-fungible tokens (NFTs) to be implemented on Algorand as smart contracts.
This interface defines a minimal interface for NFTs to be owned and traded, to be augmented by other standard interfaces and custom methods.

### ARC 73 - Algorand Interface Detection Spec

This ARC specifies an interface detection interface based on <a href="https://eips.ethereum.org/EIPS/eip-165">ERC-165</a>.
This interface allows smart contracts and indexers to detect whether a smart contract implements a particular interface based on an interface selector.

### ARC 74 - NFT Indexer API

This specifies a REST interface that can be implemented by indexing services to provide data about NFTs conforming to the [ARC-72](/arc-standards/arc-0072) standard.

### ARC 87 - Key Name Specification

Adopt a standard key name specification for complex data. 
This defines key names that can be used to represent JSON,
Blobs, or other structures that do not fit neatly into the state

### ARC 200 - Algorand Smart Contract Token Specification

This ARC (Algorand Request for Comments) specifies an interface for tokens to be implemented on Algorand as smart contracts. The interface defines a minimal interface required for tokens to be held and transferred, with the potential for further augmentation through additional standard interfaces and custom methods.

## Explorer ARCs

### ARC 2 - Algorand Transaction Note Field Conventions

The goal of these conventions is to make it simpler for block explorers and indexers to parse the data in the note fields and filter transactions of certain dApps.

## Wallet ARCs

### ARC 1 - Algorand Wallet Transaction Signing API

The goal of this API is to propose a standard way for a dApp to request the signature of a list of transactions to an Algorand wallet. This document also includes detailed security requirements to reduce the risks of users being tricked to sign dangerous transactions. As the Algorand blockchain adds new features, these requirements may change.

### ARC 5 - Wallet Transaction Signing API (Functional)

ARC-1 defines a standard for signing transactions with security in mind. This proposal is a strict subset of ARC-1 that outlines only the minimum functionality required in order to be useable.
Wallets that conform to ARC-1 already conform to this API.
Wallets conforming to [ARC-5](/arc-standards/arc-0005) but not ARC-1 **MUST** only be used for testing purposes and **MUST NOT** used on MainNet.
This is because this ARC-5 does not provide the same security guarantees as ARC-1 to protect properly wallet users.

### ARC 25 - Algorand WalletConnect v1 API

WalletConnect https://walletconnect.com/ is an open protocol to communicate securely between mobile wallets and decentralized applications (dApps) using QR code scanning (desktop) or deep linking (mobile). It’s main use case allows users to sign transactions on web apps using a mobile wallet.
This document aims to establish a standard API for using the WalletConnect v1 protocol on Algorand, leveraging the existing transaction signing APIs defined in [ARC-1](/arc-standards/arc-0001).

### ARC 27 - Provider Message Schema

Building off of the work of the previous ARCs relating to; provider transaction signing ([ARC-0005][arc-0005]), provider address discovery ([ARC-0006][arc-0006]), provider transaction network posting ([ARC-0007][arc-0007]) and provider transaction signing & posting ([ARC-0008][arc-0008]), this proposal aims to comprehensively outline a common message schema between clients and providers.
Furthermore, this proposal extends the aforementioned methods to encompass new functionality such as:
* Extending the message structure to target specific networks, thereby supporting multiple AVM (Algorand Virtual Machine) chains.
* Adding a new method that disables clients on providers.
* Adding a new method to discover provider capabilities, such as what networks and methods are supported.
This proposal serves as a formalization of the message schema and leaves the implementation details to the prerogative of the clients and providers.
<sup>[Back to top ^][top]</sup>

### ARC 35 - Algorand Offline Wallet Backup Protocol

This document outlines the high-level requirements for a wallet-agnostic backup protocol that can be used across all wallets on the Algorand ecosystem.

### ARC 47 - Logic Signature Templates

This standard allows wallets to sign known logic signatures and clearly tell the user what they are signing.

### ARC 55 - On-Chain storage/transfer for Multisig

This ARC proposes the utilization of on-chain smart contracts to facilitate the storage and transfer of Algorand multisignature metadata, transactions, and corresponding signatures for the respective multisignature sub-accounts.

### ARC 59 - ASA Inbox Router

The goal of this standard is to establish a standard in the Algorand ecosystem by which ASAs can be sent to an intended receiver even if their account is not opted in to the ASA.
A wallet custodied by an application will be used to custody assets on behalf of a given user, with only that user being able to withdraw assets. A master application will be used to map inbox addresses to user address. This master application can route ASAs to users performing whatever actions are necessary.
If integrated into ecosystem technologies including wallets, explorers, and dApps, this standard can provide enhanced capabilities around ASAs which are otherwise strictly bound at the protocol level to require opting in to be received.

### ARC 60 - Algorand Wallet Arbitrary Signing API

This ARC proposes a standard for arbitrary data signing. It is designed to be a simple and flexible standard that can be used in a wide variety of applications.

