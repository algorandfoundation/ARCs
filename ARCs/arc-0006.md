---
arc: 6
title: Algorand Wallet Address Discovery API
description: API function, enable, which allows the discovery of accounts
author: DanBurton (@DanBurton)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/52
status: Deprecated
type: Standards Track
category: Interface
created: 2021-08-09
---

# Algorand Wallet Address Discovery API

## Abstract

A function, `enable`, which allows the discovery of accounts.
Optional functions, `enableNetwork` and `enableAccounts`, which handle the multiple capabilities of `enable` separately.
This document requires nothing else, but further semantic meaning is prescribed to these functions in [ARC-0010](arc-0010.md#semantic-requirements) which builds off of this one and a few others.
The caller of this function is usually a dApp.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

> Comments like this are non-normative.

### Interface `EnableFunction`

```ts
export type AlgorandAddress = string;
export type GenesisHash = string;

export type EnableNetworkFunction = (
  opts?: EnableNetworkOpts
) => Promise<EnableNetworkResult>;

export type EnableAccountsFunction = (
  opts?: EnableAccountsOpts
) => Promise<EnableAccountsResult>;

export type EnableFunction = (
  opts?: EnableOpts
) => Promise<EnableResult>;

export type EnableOpts = (
  EnableNetworkOpts & EnableAccountsOpts
);

export interface EnableNetworkOpts {
  genesisID?: string;
  genesisHash?: GenesisHash;
};

export interface EnableAccountsOpts {
  accounts?: AlgorandAddress[];
};


export type EnableResult = (
  EnableNetworkResult & EnableAccountsResult
);

export interface EnableNetworkResult {
  genesisID: string;
  genesisHash: GenesisHash;
}

export interface EnableAccountsResult {
  accounts: AlgorandAddress[];
}

export interface EnableError extends Error {
  code: number;
  data?: any;
}
```

An `EnableFunction` with optional input argument `opts:EnableOpts` **MUST** return a value `ret:EnableResult` or **MUST** throw an exception object of type `EnableError`.

#### String specification: `GenesisID` and `GenesisHash`

A `GenesisID` is an ascii string

A `GenesisHash` is base64 string representing a 32-byte genesis hash.

#### String specification: `AlgorandAddress`

Defined as in [ARC-0001](./arc-0001.md#interface-algorandaddress):

> An Algorand address is represented by a 58-character base32 string. It includes includes the checksum.

#### Error Standards

`EnableError` follows the same rules as `SignTxnsError` from [ARC-0001](./arc-0001.md#error-interface-signtxnserror) and uses the same status error codes.

### Interface `WalletAccountManager`
```ts
export interface WalletAccountManager {
  switchAccount: (addr: AlgorandAddress) => Promise<void>
  switchNetwork: (genesisID: string) => Promise<void>
  onAccountSwitch: (hook: (addr: AlgorandAddress) => void)
  onNetworkSwitch: (hook: (genesisID: string, genesisHash: GenesisHash) => void)
}
```

Wallets SHOULD expose `switchAccount` function to allow an app to switch an account to another one managed by the wallet. The `switchAccount` function should return a promise which will be fulfilled when the wallet will effectively switch an account.
The function must thrown an `Error` exception when the wallet can't execute the switch (for example, the provided address is not managed by the wallet or when the address is not a valid Algorand address).

Similarly, wallets SHOULD expose `switchNetwork` function to instrument a wallet to switch to another network.
The function must thrown an `Error` exception when the wallet can't execute the switch (for example, when the provided genesis ID is not recognized by the wallet).

Very often, webapp have their own state with information about the user (provided by the account address) and a network. For example, a webapp can list all compatible Smart Contracts for a given network.
For descent integration with a wallet, we must be able to react in a webapp on the account and network switch from the wallet interface. For that we define 2 functions which MUST be exposed by wallets: `onAccountSwitch` and `onNetworkSwitch`. These function will register a hook and will call it whenever a user switches respectively an account or network from the wallet interface.

### Semantic requirements

This ARC uses interchangeably the terms "throw an error" and "reject a promise with an error".

#### First call to `enable`

Regarding a first call by a caller to `enable(opts)` or `enable()` (where `opts` is `undefined`), with potential promised return value `ret`:

When `genesisID` and/or `genesisHash` is specified in `opts`:

- The call `enable(opts)` **MUST** either throw an error or return an object `ret` where `ret.genesisID` and `ret.genesisHash` match `opts.genesisID` and `opts.genesisHash` (i.e., `ret.genesisID` is identical to `opts.genesisID` if `opts.genesisID` is specified, and `ret.genesisHash` is identical to `opts.genesisHash` if `opts.genesisHash` is specified).
- The user **SHOULD** be prompted for permission to acknowledge control of accounts on that specific network (defined by `ret.genesisID` and `ret.genesisHash`).
- In the case only `opts.genesisID` is provided, several networks may match this ID and the user **SHOULD** be prompted to select the network they wish to use.

When neither `genesisID` nor `genesisHash` is specified in `opts`:

- The user **SHOULD** be prompted to select the network they wish to use.
- The call `enable(opts)` **MUST** either throw an error or return an object `ret` where `ret.genesisID` and `ret.genesisHash` **SHOULD** represent the user's selection of network.
- The function **MAY** throw an error if it does not support user selection of network.

When `accounts` is specified in `opts`:

- The call `enable(opts)` **MUST** either throw an error or return an object `ret` where `ret.accounts` is an array that starts with all the same elements as `opts.accounts`, in the same order.
- The user **SHOULD** be prompted for permission to acknowledge their control of the specified accounts. The wallet **MAY** allow the user to provide more accounts than those listed. The wallet **MAY** allow the user to select fewer accounts than those listed, in which the wallet **MUST** return an error which **SHOULD** be a user rejected error and contain the rejected accounts in `data.accounts`.

When `accounts` is not specified in `opts`:

- The user **SHOULD** be prompted to select the accounts they wish to reveal on the selected network.
- The call `enable(opts)` **MUST** either throw an error or return an object `ret` where `ret.accounts` is a empty or non-empty array.
- If `ret.accounts` is not empty, the caller **MAY** assume that `ret.accounts[0]` is the user's "currently-selected" or "default" account, for DApps that only require access to one account.

> Empty `ret.accounts` array are used to allow a DApp to get access to an Algorand node but not to signing capabilities.

#### Network

In addition to the above rules, in all cases, if `ret.genesisID` is one of the official network `mainnet-v1.0`, `testnet-v1.0`, or `betanet-v1.0`, `ret.genesisHash` **MUST** match the genesis hash of those networks

| Genesis ID     | Genesis Hash                                   |
| -------------- | ---------------------------------------------- |
| `mainnet-v1.0` | `wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8=` |
| `testnet-v1.0` | `SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=` |
| `betanet-v1.0` | `mFgazF+2uRS1tMiL9dsj01hJGySEmPN28B/TjjvpVW0=` |

When using a genesis ID that is not one of the above, the caller **SHOULD** always provide a `genesisHash`.
This is because a `genesisID` does not uniquely define a network in that case.
If a caller does not provide a `genesisHash`, multiple calls to `enable` may return a different network with the same `genesisID` but a different `genesisHash`.

#### Identification of the caller

The `enable` function **MAY** remember the choices of the user made by a specific caller and use them everytime the same caller calls the function.
The function **MUST** ensure that the caller can be securely identified.
In particular, by default, the function **MUST NOT** allow webapps on the http protocol to call it, as such webapps can easily be modified by a man-in-the-middle attacker.
In the case of callers that are https websites, the caller **SHOULD** be identified by its fully qualified domain name.

The function **MAY** offer the user some "developer mode" or "advanced" options to allow calls from insecure dApps.
In that case, the fact that the caller is insecure and/or the fact that the wallet in "developer mode" **MUST** be clearly displayed by the wallet.

#### Multiple calls to `enable`

The same caller **MAY** call multiple time the `enable` function.
When the caller is a dApp, every time a dApp is refreshed, it actually **SHOULD** call the `enable()` function.

The `enable` function **MAY NOT** return the same value every time it is called, even when called with the exact same argument `opts`.
The caller **MUST NOT** assume that the `enable` function will always return the same value, and **MUST** properly handle changes of available accounts and/or changes of network.

For example, a user may want to change network or accounts for a dApp.
That is why, upon refresh, the dApp **SHOULD** automatically switch network and perform all required changes.
Examples of required changes include but are not limited to change of the list of accounts, change of statuses of the account (e.g., opted in or not), change of the balances of the accounts.

### `enableNetwork` and `enableAccounts`

It may be desirable for a dapp to perform network queries prior to requesting that the user enable an account for use with the dapp. Wallets may provide the functionality of `enable` in two parts: `enableNetwork` for network discovery, and `enableAccounts` for account discovery, which together are the equivalent of calling `enable`.

## Rationale

This API puts power in the user's hands to choose a preferred network and account to use when interacting with a dApp.

It also allows dApp developers to suggest a specific network, or specific accounts, as appropriate.
The user still maintains the ability to reject the dApp's suggestions, which corresponds to rejecting the promise returned by `enable()`.

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
