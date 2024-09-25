---
arc: 2
title: Algorand Transaction Note Field Conventions
description: Conventions for encoding data in the note field at application-level
author: Fabrice Benhamouda (@fabrice102), Stéphane Barroso (@SudoWeezy), Cosimo Bassi (@cusma)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/2
status: Final
type: Standards Track
category: ARC
sub-category: Explorer
created: 2021-07-06
---

# Algorand Transaction Note Field Conventions

## Abstract

The goal of these conventions is to make it simpler for block explorers and indexers to parse the data in the note fields and filter transactions of certain dApps.

## Specification

Note fields should be formatted as follows:

for dApps
```
<dapp-name>:<data-format><data>
```

for ARCs
```
arc<arc-number>:<data-format><data>
```

where:
* `<dapp-name>` is the name of the dApp:
    * Regexp to satisfy: `[a-zA-Z0-9][a-zA-Z0-9_/@.-]{4-31}`
      In other words, a name should:
         * only contain alphanumerical characters or `_`, `/`, `-`, `@`, `.`
         * start with an alphanumerical character
         * be at least 5 characters long
         * be at most 32 characters long
    * Names starting with `a/` and `af/` are reserved for the Algorand protocol and the Algorand Foundation uses.

* `<arc-number>` is the number of the ARC:
    * Regexp to satisfy: `\b(0|[1-9]\d*)\b`
      In other words, an arc-number should:
        * Only contain a digit number, without any padding

* `<data-format>` is one of the following:
    * `m`: <a href="https://msgpack.org">MsgPack</a>
    * `j`: <a href="https://json.org">JSON</a>
    * `b`: arbitrary bytes
    * `u`: utf-8 string
* `<data>` is the actual data in the format specified by `<data-format>`

**WARNING**: Any user can create transactions with arbitrary data and may impersonate other dApps. In particular, the fact that a note field start with `<dapp-name>` does not guarantee that it indeed comes from this dApp. The value `<dapp-name>` cannot be relied upon to ensure provenance and validity of the `<data>`.

**WARNING**: Any user can create transactions with arbitrary data, including ARC numbers, which may not correspond to the intended standard. An ARC number included in a note field does not ensure compliance with the corresponding standard. The value of the ARC number cannot be relied upon to ensure the provenance and validity of the <data>.

### Versioning

This document suggests the following convention for the names of dApp with multiple versions: `mydapp/v1`, `mydapp/v2`, ... However, dApps are free to use any other convention and may include the version inside the `<data>` part instead of the `<dapp-name>` part.

## Rationale

The goal of these conventions is to facilitate displaying notes by block explorers and filtering of transactions by notes. However, the note field **cannot be trusted**, as any user can create transactions with arbitrary note fields. An external mechanism needs to be used to ensure the validity and provenance of the data. For example:

* Some dApps may only send transactions from a small set of accounts controlled by the dApps. In that case, the sender of the transaction should be checked.
* Some dApps may fund escrow accounts created from some template TEAL script. In that case, the note field may contain the template parameters and the escrow account address should be checked to correspond to the resulting TEAL script.
* Some dApps may include a signature in the `<data>` part of the note field. The `<data>` may be an MsgPack encoding of a structure of the form:
    ```json
    {
        "d": ... // actual data
        "sig": ... // signature of the actual data (encoded using MsgPack)
    }
    ```
    In that case, the signature should be checked.

The conventions were designed to support multiple use cases of the notes. Some dApps may just record data on the blockchain without using any smart contracts. Such dApps typically would use JSON or MsgPack encoding.

On the other hands, dApps that need reading note fields from smart contracts most likely would require easier-to-parse formats of data, which would most likely consist in application-specific byte strings.

Since `<dapp-name>:` is a prefix of the note, transactions for a given dApp can easily be filtered by the <a href="https://github.com/algorand/indexer">indexer</a> ().

The restrictions on dApp names were chosen to allow most usual names while avoiding any encoding or displaying issues. The maximum length (32) matches the maximum length of ASA on Algorand, while the minimum length (5) has been chosen to limit collisions.

## Reference Implementation

> This section is non-normative.

Consider [ARC-20](./arc-0020.md), that provides information about Smart ASA's Application.

Here a potential note indicating that the Application ID is 123:

* JSON without version:
    ```
    arc20:j{"application-id":123}
    ```

Consider a dApp named `algoCityTemp` that stores temperatures from cities on the blockchain.

Here are some potential notes indicating that Singapore's temperature is 35 degree Celsius:
* JSON without version:
    ```
    algoCityTemp:j{"city":"Singapore","temp":35}
    ```
* JSON with version in the name:
    ```
    algoCityTemp/v1:j{"city":"Singapore","temp":35}
    ```
* JSON with version in the name with index lookup:
    ```
    algoCityTemp/v1/35:j{"city":"Singapore","temp":35}
    ```
* JSON with version in the data:
    ```
    algoCityTemp:j{"city":"Singapore","temp":35,"ver":1}
    ```
* UTF-8 string without version:
    ```
    algoCityTemp:uSingapore|35
    ```
* Bytes where the temperature is encoded as a signed 1-byte integer in the first position:
    ```
    algoCityTemp:b#Singapore
    ```
    (`#` is the ASCII character for 35.)
* MsgPack corresponding to the JSON example with version in the name. The string is encoded in base64 as it contains characters that cannot be printed in this document. But the note should contain the actual bytes and not the base64 encoding of them:
    ```
    YWxnb0NpdHlUZW1wL3YxOoKkY2l0ealTaW5nYXBvcmWkdGVtcBg=
    ```

## Security Considerations
> Not Applicable

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
