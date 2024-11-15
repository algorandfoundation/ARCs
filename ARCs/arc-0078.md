---
arc: 78
title: URI scheme, keyreg Transactions extension
description: A specification for encoding Key Registration Transactions in a URI format.
author: Tasos Bitsios (@tasosbit)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/314
status: Final
type: Standards Track
category: Interface
sub-category: General
created: 2024-10-02
extends: 26
---

## Abstract

This URI specification represents an extension to the base Algorand URI encoding standard ([ARC-26](./arc-0026.md)) that specifies encoding of key registration transactions through deeplinks, QR codes, etc.

## Specification

### General format

As in [ARC-26](./arc-0026.md), URIs follow the general format for URIs as set forth in <a href="https://www.rfc-editor.org/rfc/rfc3986">RFC 3986</a>. The path component consists of an Algorand address, and the query component provides additional transaction parameters.

Elements of the query component may contain characters outside the valid range. These are encoded differently depending on their expected character set. The text components (note, xnote) must first be encoded according to UTF-8, and then each octet of the corresponding UTF-8 sequence must be percent-encoded as described in RFC 3986. The binary components (votekey, selkey, etc.) must be encoded with base64url as specified in <a href="https://www.rfc-editor.org/rfc/rfc4648.html#section-5">RFC 4648 section 5</a>.

### Scope

This ARC explicitly supports the two major subtypes of key registration transactions:

- Online keyreg transcation
  - Declares intent to participate in consensus and configures required keys
- Offline keyreg transaction
  - Declares intent to stop participating in consensus

The following variants of keyreg transactions are not defined:

- Non-participating keyreg transcation
  - This transaction subtype is considered deprecated
- Heartbeat keyreg transaction
  - This transaction subtype will be included in the future block incentives protocol. The protocol specifies that this transaction type must be submitted by a node in response to a programmatic "liveness challenge". It is not meant to be signed or submitted by an end user.

### ABNF Grammar

```
algorandurn     = "algorand://" algorandaddress [ "?" keyregparams ]
algorandaddress = *base32
keyregparams    = keyregparam [ "&" keyregparams ]
keyregparam     = [ typeparam / votekeyparam / selkeyparam / sprfkeyparam / votefstparam / votelstparam / votekdparam / noteparam / feeparam / otherparam ]
typeparam       = "type=keyreg"
votekeyparam    = "votekey=" *qbase64url
selkeyparam     = "selkey=" *qbase64url
sprfkeyparam    = "sprfkey=" *qbase64url
votefstparam    = "votefst=" *qdigit
votelstparam    = "votelst=" *qdigit
votekdparam     = "votekdkey=" *qdigit
noteparam       = (xnote | note)
xnote           = "xnote=" *qchar
note            = "note=" *qchar
fee             = "fee=" *qdigit
otherparam      = qchar *qchar [ "=" *qchar ]
```

- "qbase64url" corresponds to valid characters of "base64url" encoding, as defined in <a href="https://www.rfc-editor.org/rfc/rfc4648.html#section-5">RFC 4648 section 5</a>
- "qchar" corresponds to valid characters of an RFC 3986 URI query component, excluding the "=" and "&" characters, which this specification takes as separators.

As in the base [ARC-26](./arc-0026.md) standard, the scheme component ("algorand:") is case-insensitive, and implementations must accept any combination of uppercase and lowercase letters. The rest of the URI is case-sensitive, including the query parameter keys.

### Query Keys

- address: Algorand address of transaction sender. Required.

- type: fixed to "keyreg". Used to disambiguate the transaction type from the base [ARC-26](./arc-0026.md) standard and other possible extensions. Required.

- votekeyparam: The vote key parameter to use in the transaction. Encoded with <a href="https://www.rfc-editor.org/rfc/rfc4648.html#section-5">base64url</a> encoding. Required for keyreg online transactions.

- selkeyparam: The selection key parameter to use in the transaction. Encoded with <a href="https://www.rfc-editor.org/rfc/rfc4648.html#section-5">base64url</a> encoding. Required for keyreg online transactions.

- sprfkeyparam: The state proof key parameter to use in the transaction. Encoded with <a href="https://www.rfc-editor.org/rfc/rfc4648.html#section-5">base64url</a> encoding. Required for keyreg online transactions.

- votefstparam: The first round on which the voting keys will valid. Required for keyreg online transactions.

- votelstparam: The last round on which the voting keys will be valid. Required for keyreg online transactions.

- votekdparam: The key dilution key parameter to use. Required for keyreg online transactions.

- xnote: As in [ARC-26](./arc-0026.md). A URL-encoded notes field value that must not be modifiable by the user when displayed to users. Optional.

- note: As in [ARC-26](./arc-0026.md). A URL-encoded default notes field value that the the user interface may optionally make editable by the user. Optional.

- fee: Optional. A static fee to set for the transaction in microAlgos. Useful to signal intent to receive participation incentives (e.g. with a 2,000,000 microAlgo transaction fee.) Optional.

- (others): optional, for future extensions

### Appendix

This section contains encoding examples. The raw transaction object is presented along with the resulting [ARC-78](./arc-0078.md) URI encoding.

#### Encoding keyreg online transactioon with minimum fee

The following raw keyreg transaction:

```
{
  "txn": {
    "fee": 1000,
    "fv": 1345,
    "gh:b64": "kUt08LxeVAAGHnh4JoAoAMM9ql/hBwSoiFtlnKNeOxA=",
    "lv": 2345,
    "selkey:b64": "+lfw+Y04lTnllJfncgMjXuAePe8i8YyVeoR9c1Xi78c=",
    "snd:b64": "+gJAXOr2rkSCdPQ5DEBDLjn+iIptzLxB3oSMJdWMVyQ=",
    "sprfkey:b64": "3NoXc2sEWlvQZ7XIrwVJjgjM30ndhvwGgcqwKugk1u5W/iy/JITXrykuy0hUvAxbVv0njOgBPtGFsFif3yLJpg==",
    "type": "keyreg",
    "votefst": 1300,
    "votekd": 100,
    "votekey:b64": "UU8zLMrFVfZPnzbnL6ThAArXFsznV3TvFVAun2ONcEI=",
    "votelst": 11300
  }
}
```

Will result in this ARC-78 encoded URI:

```
algorand://7IBEAXHK62XEJATU6Q4QYQCDFY475CEKNXGLYQO6QSGCLVMMK4SLVTYLMY?
type=keyreg
&selkey=-lfw-Y04lTnllJfncgMjXuAePe8i8YyVeoR9c1Xi78c
&sprfkey=3NoXc2sEWlvQZ7XIrwVJjgjM30ndhvwGgcqwKugk1u5W_iy_JITXrykuy0hUvAxbVv0njOgBPtGFsFif3yLJpg
&votefst=1300
&votekd=100
&votekey=UU8zLMrFVfZPnzbnL6ThAArXFsznV3TvFVAun2ONcEI
&votelst=11300
```

Note: newlines added for readability.

Note the difference between base64 encoding in the raw object and base64url encoding in the URI parameters. For example, the selection key parameter `selkey` that begins with `+lfw+` in the raw object is encoded in base64url to `-lfw-`.

Note: Here, the fee is omitted from the URI (due to being set to the minimum 1,000 microAlgos.) When the fee is omitted, it is left up to the application or wallet to decide. This is for demonstrative purposes - the ARC-78 standard does not require this behavior.

#### Encoding keyreg offline transactioon

The following raw keyreg transaction:

```
{
  "txn": {
    "fee": 1000,
    "fv": 1776240,
    "gh:b64": "kUt08LxeVAAGHnh4JoAoAMM9ql/hBwSoiFtlnKNeOxA=",
    "lv": 1777240,
    "snd:b64": "+gJAXOr2rkSCdPQ5DEBDLjn+iIptzLxB3oSMJdWMVyQ=",
    "type": "keyreg"
  }
}
```

Will result in this ARC-78 encoded URI:

```
algorand://7IBEAXHK62XEJATU6Q4QYQCDFY475CEKNXGLYQO6QSGCLVMMK4SLVTYLMY?type=keyreg
```

This offline keyreg transaction encoding is the smallest compatible ARC-78 representation.


#### Encoding keyreg online transactioon with custom fee and note

The following raw keyreg transaction:

```
{
  "txn": {
    "fee": 2000000,
    "fv": 1345,
    "gh:b64": "kUt08LxeVAAGHnh4JoAoAMM9ql/hBwSoiFtlnKNeOxA=",
    "lv": 2345,
    "note:b64": "Q29uc2Vuc3VzIHBhcnRpY2lwYXRpb24gZnR3",
    "selkey:b64": "+lfw+Y04lTnllJfncgMjXuAePe8i8YyVeoR9c1Xi78c=",
    "snd:b64": "+gJAXOr2rkSCdPQ5DEBDLjn+iIptzLxB3oSMJdWMVyQ=",
    "sprfkey:b64": "3NoXc2sEWlvQZ7XIrwVJjgjM30ndhvwGgcqwKugk1u5W/iy/JITXrykuy0hUvAxbVv0njOgBPtGFsFif3yLJpg==",
    "type": "keyreg",
    "votefst": 1300,
    "votekd": 100,
    "votekey:b64": "UU8zLMrFVfZPnzbnL6ThAArXFsznV3TvFVAun2ONcEI=",
    "votelst": 11300
  }
}
```

Will result in this ARC-78 encoded URI:

```
algorand://7IBEAXHK62XEJATU6Q4QYQCDFY475CEKNXGLYQO6QSGCLVMMK4SLVTYLMY?
type=keyreg
&selkey=-lfw-Y04lTnllJfncgMjXuAePe8i8YyVeoR9c1Xi78c
&sprfkey=3NoXc2sEWlvQZ7XIrwVJjgjM30ndhvwGgcqwKugk1u5W_iy_JITXrykuy0hUvAxbVv0njOgBPtGFsFif3yLJpg
&votefst=1300
&votekd=100
&votekey=UU8zLMrFVfZPnzbnL6ThAArXFsznV3TvFVAun2ONcEI
&votelst=11300
&fee=2000000
&note=Consensus%2Bparticipation%2Bftw
```

Note: newlines added for readability.

## Rationale

The present aims to provide a standardized way to encode key registration transactions in order to enhance the user experience of signing key registration transactions in general, and in particular in the use case of an Algorand node runner that does not have their spending keys resident on their node (as is best practice.)

The parameter names were chosen to match the corresponding names in encoded key registration transactions.

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
