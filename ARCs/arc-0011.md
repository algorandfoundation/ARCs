---
arc: 11
title: Algorand Wallet Reach Browser Spec
description: Convention for DApps to discover Algorand wallets in browser
author: DanBurton (@DanBurton)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/52
status: Deprecated
type: Standards Track
category: Interface
created: 2021-08-09
---

# Algorand Wallet Reach Browser Spec

## Abstract

A common convention for DApps to discover Algorand wallets in browser code: `window.algorand`.
A property `algorand` attached to the `window` browser object, with all the features defined in [ARC-0010](./arc-0010.md#specification).

## Specification

```ts
interface WindowAlgorand {
  enable: EnableFunction;
  enableNetwork?: EnableNetworkFunction;
  enableAccounts?: EnableAccountsFunction;
  signAndPostTxns: SignAndPostTxnsFunction;
  getAlgodv2Client: GetAlgodv2ClientFunction;
  getIndexerClient: GetIndexerClientFunction;
  signTxns?: SignTxnsFunction;
  postTxns?: SignTxnsFunction;
}
```

With the specifications and semantics for each function as stated in [ARC-0010](./arc-0010.md#specification).

## Rationale

DApps should be unopinionated about which wallet they are used with. End users should be able to inject their wallet of choice into the DApp. Therefore, in browser contexts, we reserve `window.algorand` for this purpose.

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
