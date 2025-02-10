---
arc: 15
title: Encrypted Short Messages
description: Scheme for encryption/decryption that allows for private messages.
author: Stéphane Barroso (@sudoweezy), Paweł Pierścionek (@urtho)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/*
status: Deprecated
type: Standards Track
category: Interface
created: 2022-11-21
requires: 4
---

# Encrypted Short Messages

## Abstract

The goal of this convention is to have a standard way for block explorers, wallets, exchanges, marketplaces, and more generally, client software to send, read & delete short encrypted messages.

## Specification
The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

> Comments like this are non-normative.

### Account's message Application

To receive a message, an Account **MUST** create an application that follows this convention:

- A Local State named `public_key` **MUST** contain an *NACL Public Key (Curve 25519)* key
- A Local State named `arc` **MUST** contain the value `arc15-nacl-curve25519`


- A Box `inbox` where:
  - Keys is an ABI encoded of the tuple `(address,uint64)` containing the address of the sender and the round when the message is sent
  - Value is an encoded  **text**

> With this design, for each round, the sender can only write one message per round.
> For the same round, an account can receive multiple messages if distinct sender sends them

### ABI Interface

The associated smart contract **MUST** implement the following ABI interface:
```json
{
  "name": "ARC_0015",
  "desc": "Interface for an encrypted messages application",
  "methods": [
    {
      "name": "write",
      "desc": "Write encrypted text to the box inbox",
      "args": [
        { "type": "byte[]", "name": "text", "desc": "Encrypted text provided by the sender." }
      ],
      "returns": { "type": "void" }
    },
    {
      "name": "authorize",
      "desc": "Authorize an addresses to send a message",
      "args": [
        { "type": "byte[]", "name": "address_to_add", "desc": "Address of a sender" },
        { "type": "byte[]", "name": "info", "desc": "information about the sender" }
      ],
      "returns": { "type": "void" }
    },
    {
      "name": "remove",
      "desc": "Delete the encrypted text sent by an account on a particular round. Send the MBR used for a box to the Application's owner.",
      "args": [
        { "type": "byte[]", "name": "address", "desc": "Address of the sender"},
        { "type": "uint64", "name": "round", "desc": "Round when the message was sent"}
      ],
      "returns": { "type": "void" }
    },
    {
      "name": "set_public_key",
      "desc": "Register a NACL Public Key (Curve 25519) to the global value public_key",
      "args": [
        { "type": "byte[]", "name": "public_key", "desc": "NACL Public Key (Curve 25519)" }
      ],
      "returns": { "type": "void" }
    }
  ]
}
```
> Warning: The remove method only removes the box used for a message, but it is still possible to access it by looking at the indexer.

## Rationale
Algorand blockchain unlocks many new use cases - anonymous user login to dApps and classical WEB2.0 solutions being one of them. For many use-cases, anonymous users still require asynchronous event notifications, and email seems to be the only standard option at the time of the creation of this ARC. With wallet adoption of this standard, users will enjoy real-time encrypted A2P (application-to-person) notifications without having to provide their email addresses and without any vendor lock-in.

There is also a possibility to do a similar version of this ARC with one App which will store every message for every Account.

Another approach was to use the note field for messages, but with box storage available, it was a more practical and secure design.

## Reference Implementation

The following codes are not audited and are only here for information purposes.
It **MUST** not be used in production.

Here is an example of how the code can be run in python :
[main.py](../assets/arc-0015/main.py).
> The delete method is only for test purposes, it is not part of the ABI for an `ARC-15` Application.

An example the application created using Beaker can be found here :
[application.py](../assets/arc-0015/application.py).


## Security Considerations
Even if the message is encrypted, it will stay on the blockchain.
If the secret key used to decrypt is compromised at one point, every related message IS at risk.

## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
