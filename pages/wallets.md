---
layout: page
title: Algorand Wallet Compatiblity Matrix
permalink: /wallets
---
## Status
Allowed Status are:
- Supported
- Not Supported
- Planned
- EXPERIMENTAL
- UNKNOWN
- DEPRECATED

## Wallet ARCs

| ARCS           | Daffi         | Defly         | Exodus        | Pera Wallet   |
| -------------- | ------------- | ------------- | ------------- | ------------- |
| [1][ARC-1]     | UNKNOWN       | Supported     | Supported     | Supported     |
| [5][ARC-5]     | UNKNOWN       | Supported     | Supported     | Supported     |
| [6][ARC-6]     | UNKNOWN       | Not Supported | Not Supported | Not Supported |
| [7][ARC-7]     | UNKNOWN       | Not Supported | Supported     | Not Supported |
| [8][ARC-8]     | UNKNOWN       | Not Supported | Supported     | Not Supported |
| [9][ARC-9]     | UNKNOWN       | Not Supported | Supported     | Not Supported |
| [10][ARC-10]   | UNKNOWN       | Not Supported | Supported     | Not Supported |
| [11][ARC-11]   | UNKNOWN       | Not Supported | Supported     | Not Supported |
| [25][ARC-25]   | UNKNOWN       | Supported     | Supported     | Supported     |
| [35][ARC-35]   | UNKNOWN       | Supported     | Not Supported | Planned       |

[ARC-1]: ../ARCs/arc-0001.md "Algorand Wallet Transaction Signing API"
[ARC-5]: ../ARCs/arc-0005.md "Wallet Transaction Signing API (Functional)"
[ARC-6]: ../ARCs/arc-0006.md "Algorand Wallet Address Discovery API"
[ARC-7]: ../ARCs/arc-0007.md "Algorand Wallet Post Transactions API"
[ARC-8]: ../ARCs/arc-0008.md "Algorand Wallet Sign and Post API"
[ARC-9]: ../ARCs/arc-0009.md "Algorand Wallet Algodv2 and Indexer API"
[ARC-10]: ../ARCs/arc-0010.md "Algorand Wallet Reach Minimum Requirements"
[ARC-11]: ../ARCs/arc-0011.md "Algorand Wallet Reach Browser Spec"
[ARC-25]: ../ARCs/arc-0025.md "Algorand WalletConnect v1 API"
[ARC-35]: ../ARCs/arc-0035.md "Algorand Offline Wallet Backup Protocol"

## NFT & Token ARCs

| ARCS           | Daffi         | Defly         | Exodus        | Pera Wallet   |
| -------------- | ------------- | ------------- | ------------- | ------------- |
| [3][ARC-3]     | UNKNOWN       | UNKNOWN       | Supported     | UNKNOWN       |
| [16][ARC-16]   | UNKNOWN       | UNKNOWN       | Supported     | UNKNOWN       |
| [19][ARC-19]   | UNKNOWN       | UNKNOWN       | Supported     | UNKNOWN       |
| [20][ARC-20]   | UNKNOWN       | UNKNOWN       | Not Supported | UNKNOWN       |
| [69][ARC-69]   | UNKNOWN       | UNKNOWN       | Supported     | UNKNOWN       |
| [72][ARC-72]   | UNKNOWN       | UNKNOWN       | Not Supported | UNKNOWN       |
| [200][ARC-200] | UNKNOWN       | UNKNOWN       | Not Supported | UNKNOWN       |

[ARC-3]: ../ARCs/arc-0003.md "Conventions Fungible/Non-Fungible Tokens"
[ARC-16]: ../ARCs/arc-0016.md "Convention for declaring traits of an NFT's"
[ARC-19]: ../ARCs/arc-0019.md "Templating of NFT ASA URLs for mutability"
[ARC-20]: ../ARCs/arc-0020.md "Smart ASA"
[ARC-69]: ../ARCs/arc-0069.md "ASA Parameters Conventions, Digital Media"
[ARC-72]: ../ARCs/arc-0072.md "Algorand Smart Contract NFT Specification"
[ARC-200]: ../ARCs/arc-0200.md "Algorand Smart Contract Token Specification"


## BIPS

| BIPS           | Daffi         | Defly         | Exodus        | Pera Wallet   |
| -------------- | ------------- | ------------- | ------------- | ------------- |
| [32][BIP-32]   | UNKNOWN       | Not supported | Supported     | Not supported |
| [39][BIP-39]   | UNKNOWN       | Planned       | Supported     | Planned       |
| [44][BIP-44]   | UNKNOWN       | Not supported | Supported     | Not supported |

[BIP-32]: https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki "Hierarchical Deterministic Wallets"
[BIP-39]: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki "Mnemonic code for generating deterministic keys"
[BIP-44]: https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki "Multi-Account Hierarchy for Deterministic Wallets"

**Disclaimer:** This website is under constant modification.
