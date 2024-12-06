---
arc: 79
title: URI scheme, App NoOp call extension
description: A specification for encoding NoOp Application call Transactions in a URI format.
author: MG (@emg110)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/319
status: Final
type: Standards Track
category: Interface
sub-category: General
created: 2024-09-11
extends: 26
---

## Abstract
NoOp calls are Generic application calls to execute the Algorand smart contract ApprovalPrograms.

This URI specification proposes an extension to the base Algorand URI encoding standard ([ARC-26](./arc-0026.md)) that specifies encoding of application NoOp transactions into <a href="https://www.rfc-editor.org/rfc/rfc3986">RFC 3986</a> standard URIs.

## Specification

### General format

As in [ARC-26](./arc-0026.md), URIs follow the general format for URIs as set forth in <a href="https://www.rfc-editor.org/rfc/rfc3986">RFC 3986</a>. The path component consists of an Algorand address, and the query component provides additional transaction parameters.

Elements of the query component may contain characters outside the valid range. These are encoded differently depending on their expected character set. The text components (note, xnote) must first be encoded according to UTF-8, and then each octet of the corresponding UTF-8 sequence **MUST** be percent-encoded as described in RFC 3986. The binary components (args, refs, etc.) **MUST** be encoded with base64url as specified in <a href="https://www.rfc-editor.org/rfc/rfc4648.html#section-5">RFC 4648 section 5</a>.

### ABNF Grammar

```
algorandurn     = "algorand://" algorandaddress [ "?" noopparams ]
algorandaddress = *base32
noopparams      = noopparam [ "&" noopparams ]
noopparam       = [ typeparam / appparam /  methodparam / argparam / boxparam / assetarrayparam / accountarrayparam / apparrayparam / feeparam / otherparam ]
typeparam       = "type=appl"
appparam        = "app=" *digit
methodparam     = "method=" *qchar
boxparam        = "box=" *qbase64url
argparam        = "arg=" (*qchar | *digit)
feeparam        = "fee=" *digit
accountparam    = "account=" *base32
assetparam      = "asset=" *digit
otherparam      = qchar *qchar [ "=" *qchar ]
```

- "qchar" corresponds to valid characters of an RFC 3986 URI query component, excluding the "=" and "&" characters, which this specification takes as separators.
- "qbase64url" corresponds to valid characters of "base64url" encoding, as defined in <a href="https://www.rfc-editor.org/rfc/rfc4648.html#section-5">RFC 4648 section 5</a>
- All params from the base [ARC-26](./arc-0026.md) standard, are supported and usable if fit the NoOp application call context (e.g. note)
- As in the base [ARC-26](./arc-0026.md) standard, the scheme component ("algorand:") is case-insensitive, and implementations **MUST** accept any combination of uppercase and lowercase letters. The rest of the URI is case-sensitive, including the query parameter keys.

### Query Keys

- address: Algorand address of transaction sender

- type: fixed to "appl". Used to disambiguate the transaction type from the base [ARC-26](./arc-0026.md) standard and other possible extensions

- app: The first reference is set to specify the called application (Algorand Smart Contract) ID and is mandatory. Additional references are optional and will be used in the Application NoOp call's foreign applications array.

- method: Specify the full method expression (e.g "example_method(uint64,uint64)void").

- arg: specify args used for calling NoOp method, to be encoded within URI.

- box: Box references to be used in Application NoOp method call box array.

- asset: Asset reference to be used in Application NoOp method call foreign assets array.

- account: Account or nfd address to be used in Application NoOp method call foreign accounts array.

- fee: Optional. An optional static fee to set for the transaction in microAlgos.

- (others): optional, for future extensions

Note 1: If the fee is omitted , it means that Minimum Fee is preferred to be used for transaction.

### Template URI vs actionable URI

If the URI is constructed so that other dApps, wallets or protocols could use it with their runtime Algorand entities of interest, then :

- The placeholder account/app address in URI **MUST** be ZeroAddress ("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAY5HFKQ"). Since ZeroAddress cannot initiate any action this approach is considered non-vulnerable and secure.


### Example

Call claim(uint64,uint64)byte[] method on contract 11111111 paying a fee of 10000 micro ALGO from an specific address

```
algorand://TMTAD6N22HCS2LKH7677L2KFLT3PAQWY6M4JFQFXQS32ECBFC23F57RYX4?type=appl&app=11111111&method=claim(uint64,uint64)byte[]&arg=20000&arg=474567&asset=45&fee=10000
```

Call the same claim(uint64,uint64)byte[] method on contract 11111111 paying a default 1000 micro algo fee

```
algorand://TMTAD6N22HCS2LKH7677L2KFLT3PAQWY6M4JFQFXQS32ECBFC23F57RYX4?type=appl&app=11111111&method=claim(uint64,uint64)byte[]&arg=20000&arg=474567&asset=45&app=22222222&app=33333333
```



## Rationale

Algorand application NoOp method calls cover the majority of application transactions in Algorand and have a wide range of use-cases.
For use-cases where the runtime knows exactly what the called application needs in terms of arguments and transaction arrays and there are no direct interactions, this extension will be required since ARC-26 standard does not currently support application calls.

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
