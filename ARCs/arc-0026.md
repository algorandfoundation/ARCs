---
arc: 26
title: URI scheme
description: A specification for encoding Transactions in a URI format.
author: Fabrice Benhamouda (@fabrice102), Ben Guidarelli (@barnjamin)
status: Final
type: Standards Track
category: Interface
sub-category: General
created: 2022-04-21
extended-by: 78, 79
---

## Abstract

This URI specification represents a standardized way for applications and websites to send requests and information through deeplinks, QR codes, etc. It is heavily based on Bitcoin’s <a href="https://github.com/bitcoin/bips/blob/master/bip-0021.mediawiki">BIP-0021</a> and should be seen as derivative of it. The decision to base it on BIP-0021 was made to make it easy and compatible as possible for any other application.

## Specification

### General format

Algorand URIs follow the general format for URIs as set forth in <a href="https://www.rfc-editor.org/rfc/rfc3986">RFC 3986</a>. The path component consists of an Algorand address, and the query component provides additional payment options.

Elements of the query component may contain characters outside the valid range. These must first be encoded according to UTF-8, and then each octet of the corresponding UTF-8 sequence must be percent-encoded as described in RFC 3986.

### ABNF Grammar

```
algorandurn     = "algorand://" algorandaddress [ "?" algorandparams ]
algorandaddress = *base32
algorandparams  = algorandparam [ "&" algorandparams ]
algorandparam   = [ amountparam / labelparam / noteparam / assetparam / otherparam ]
amountparam     = "amount=" *digit
labelparam      = "label=" *qchar
assetparam      = "asset=" *digit
noteparam       = (xnote | note)
xnote           = "xnote=" *qchar
note            = "note=" *qchar
otherparam      = qchar *qchar [ "=" *qchar ]
```

Here, "qchar" corresponds to valid characters of an RFC 3986 URI query component, excluding the "=" and "&" characters, which this specification takes as separators.

The scheme component ("algorand:") is case-insensitive, and implementations must accept any combination of uppercase and lowercase letters. The rest of the URI is case-sensitive, including the query parameter keys.

!!! Caveat
    When it comes to generation of an address' QR,  many exchanges and wallets encodes the address w/o the scheme component (“algorand:”). This is not a URI so it is OK.

### Query Keys

- label: Label for that address (e.g. name of receiver)

- address: Algorand address

- xnote: A URL-encoded notes field value that must not be modifiable by the user when displayed to users.

- note: A URL-encoded default notes field value that the the user interface may optionally make editable by the user.

- amount: microAlgos or smallest unit of asset

- asset: The asset id this request refers to (if Algos, simply omit this parameter)

- (others): optional, for future extensions

### Transfer amount/size

!!! Note
    This is DIFFERENT than Bitcoin’s BIP-0021

If an amount is provided, it MUST be specified in basic unit of the asset. For example, if it’s Algos (Algorand native unit), the amount should be specified in microAlgos. All amounts MUST NOT contain commas nor a period (.) Strictly non negative integers.

e.g. for 100 Algos, the amount needs to be 100000000, for 54.1354 Algos the amount needs to be 54135400.

Algorand Clients should display the amount in whole Algos. Where needed, microAlgos can be used as well. In any case, the units shall be clear for the user.

### Appendix

This section contains several examples

address -

```
algorand://TMTAD6N22HCS2LKH7677L2KFLT3PAQWY6M4JFQFXQS32ECBFC23F57RYX4
```

address with label -

```
algorand://TMTAD6N22HCS2LKH7677L2KFLT3PAQWY6M4JFQFXQS32ECBFC23F57RYX4?label=Silvio
```

Request 150.5 Algos from an address

```
algorand://TMTAD6N22HCS2LKH7677L2KFLT3PAQWY6M4JFQFXQS32ECBFC23F57RYX4?amount=150500000
```

Request 150 units of Asset ID 45 from an address

```
algorand://TMTAD6N22HCS2LKH7677L2KFLT3PAQWY6M4JFQFXQS32ECBFC23F57RYX4?amount=150&asset=45
```

## Rationale

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
