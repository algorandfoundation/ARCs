---
arc: 14
title: Algorand Standard for Authentication
description: Use Algorand accounts to authenticate with third-party services
author: Stefano De Angelis <stefano@algorand.com>
discussions-to: https://github.com/algorandfoundation/ARCs/pull/41
status: Draft
type: Meta
created: 2022-04-05
---

# Algorand Standard for Authentication

> This ARC is intended to be an alternative to PR #41 [Create arc-0014](https://github.com/algorandfoundation/ARCs/pull/41).

## Summary

A standard approach to authenticate users via Algorand accounts.

## Abstract

This document introduces a standard SSO authentication mechanism based on Algorand accounts. It leverages the public-secret key <*PK, SK*> encryption schema used by Algorand to represent users on-chain. This approach fosters the adoption of novel identity and session management systems for Web3 applications.

## Definitions

- System
- Session id
- User
- dApp
- wallet
- Verifier
...

## Motivation

Traditional systems enforce SSO login leveraging on authentication of users with *credentials*, i.e. username and password. Once the identity of a user is verified, the system can establish an authenticated session. Sessions allows users to interact with a system without having to authenticate itsefl anytime. Sessions are represented with a `session-id` (typically a cookie or a JSON Web Token - JWT).

In a blockchain context, users are not identified with credentials. Therefore, traditional authentication mechanisms result impractical. In the Web3 users are identified through their unique blockchain public addresses (accounts for Algorand). To interact with a dApp, users must connect their wallet to the application. Using an authentication layer based on credentials on top of that mechanism might result a redundant and inefficient procedure.

In Web3, dApps and traditional systems will be increasingly more interconnected. It is not difficult to imagine users consuming services both from a dApp and a traditional system simultaneously. In that case, a user should be authenticated first through traditional SSOs, and then connecting their blockchain wallet. A better approach should enforce a single SSO procedure.

This ARC provides the standard to authenticate users leversging on their Algorand accounts, and withour relying on credentials verification.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in [RFC-2119](https://www.ietf.org/rfc/rfc2119.txt).

> Comments like this are non-normative.

### Overview

#### General idea

The mechanism uses blockchain addresses as unique identifiers. Users login into an application on a server/backend by proving they own a blockchain address. The proof is a digital signature that can be validated server-side. If the proof is valid, the server creates a new session (release of a cookie/JWT).

The mechanism works as follows: a user attempts to login into an application with its blockchain address. The backend requests the user to prove her identity by signing a random message with the secret key in control of the blockchain address. The message MUST be unpredictable and only valid for one authentication request. Reproducible messages could be used by an attacker in a replay-attack. Thus, after signing the message the user forwards the result to the backend which validates the signature and eventually instantiates a new authenticated session with the user. The established session is characterized by a session id (i.e. a cookie or JWT).

GRAPH HERE

Diagram 1 summarizes the solution presented above. It illustrates the workflow and the message exchanges between parties. In particular, the user u, owner of the blockchain identity PKu, login into an application hosted by the backend b, thus the authentication mechanism proceeds as follows:

1. u attempts to login to the application with its blockchain address PKu;
2. b retrieves the challenge Cu associated with PKu;
3. b requests u to digitally sign the challenge Cu to prove the ownership of that blockchain address;
4. u signs the challenge and afterwards sends back the tuple <PKu, Sig(Cu)> to b;
5. b retrieves and validates the digital signature Sig(Cu);
6. If the signature is valid, b establishes a new session with u by forwarding a cookie or JWT; if the signature is not valid, b returns NULL (authentication rejected);
7. b updates the challenge Cu associated with PKu for a future session.

The previous section introduced a general mechanism to authenticate users with blockchain identities. This section provides an implementation of such a mechanism using the Algorand blockchain. It considers a User and a Verifier. The User owns an Algorand account which is identified with an Algorand key pair <PKa, SKa>. Algorand transforms traditional 32-bytes cryptographic keys into more readable and user-friendly objects. The public key is represented as Algorand address PKa, whereas the secret key is transformed into a base64 private key SKa. The User accesses her Algorand account via a wallet. The Verifier is any kind of system (i.e. application, dApp, entity) that wants to authenticate the User with her Algorand account. A wallet is any type of Algorand wallet, such as hot wallets like AlgoSigner, MyAlgo Wallet for browser and mobile wallets used through WalletConnect, and cold wallets like the Ledger Nano.

The User presents to the Verifier her Algorand address (PKa) for authentication. The Verifier authenticates the User by asking her to sign an Authentication Message. The digital signature MUST be created with the secret key (SKa) of the Algorand account used for authentication. An Authentication Message is nothing but a sequence of bytes. Users connected to a wallet with an Algorand account can sign the Authentication Message with their SKs, and forge brand new digital signatures.

To sum up, the solution use Algorand cryptographic primitives to achieve the following operations:

- given an Algorand account and an Authentication Message, the user MUST be able to generate a digital signature using the SKa;
- given an Algorand account and an Authentication Message, the verifier MUST be able to verify the digital signature of that message with the public key of the Algorand address (PKa).

### Assumptions

The standard proposed in this document works under the following assumptions:

- Secure communication channels encrypted via SSL/TLS;
- The backend knows the users’ PK;
- For each PK the backend generates a unique challenge;
- Users are the only custodian of the SK;
- Users won’t use multiple PK to login into an application;
- Users won’t change their public address to login into the application;
- The challenge must be changed in each session to avoid replay attack;
- Users MUST use only ed25519 keys to sign messages;
- Multisig, LogicSig not supported;
- Users MUST use non-rekeyed accounts.

### Authentication Message

An Authentication Message is a sequence of bytes representing a message. The Verifier requests users to sign an Authentication Message with her secret key SKa in order to prove the ownership of the Algorand account they want to use for authentication. Such a message MUST include the following information:

- domain name of the service to authenticate;
- description of the service to authenticate;
- Algorand address (PKa) to be authenticated;
- nonce representing a unique/random value generated by the verifier.

The JSON structure for such an object is:

```typescript
interface AuthMessage {
 /** The domain name of the service */
 service: string;
 /** Optional, description of the service */
 desc?: string;
 /** Algorand account to authenticate with*/
 authAcc: string;
 /** Challenge generated server-side */
 nonce: string;
}
```

For example:

```json
{
 "service": "www.servicedomain.com",
 "desc": "Domain offers important services to users",
 "authAcc": "KTGP47G64KCXWJS64W7SGJNKTHE37TYDCI64USXI3XOYE6ZSH4LCI7NIDA",
 "nonce": "1234abcde!%",
}
```

The nonce field MUST be unique for each authentication and MUST not be used more than once to avoid replay attacks.

Users can sign Authentication Messages using their secret key SKa stored into a wallet.

**Problem**: Most of the Algorand wallets can only sign Algorand Transaction objects, and do not offer the possibility to sign arbitrary bytes. To overcome such a limitation, the Authentication Message can be extended to two distinguished objects, namely the Simple Authentication Message, and the Transaction Authentication Message. The former SHOULD be used by any signing mechanisms provided with the functionality of signing random bytes, whereas the latter can be used today with most of the Algorand wallets.

#### Simple Authentication Message

The *Simple Authentication Message* is a sequence of bytes representing an ARC-0014 basic Authentication Message. It SHOULD be used in contexts that support signatures of random bytes. A *Simple Authentication Message* is an Authentication Message prepended with the prefix `ARC-0014-authentication` for domain separation. It **MUST** be represented as the hash SHA-512/256 of the prefix together with an msgpack encoded *Authentication Message*. For example, given an Authentication Message object aut_message, its *Simple Authentication Message* representation would be: `SHA512_256(“ARC-0014-authentication”+msgpacked_auth_message)`.

#### Transaction Authentication Message

A *Transaction Authentication Message* is an *Authentication Message* represented as an Algorand Transaction object. Such a transaction must have the following characteristics:

- it SHOULD be an Algorand Payment Transaction;
- it MUST have a *Simple Authentication Message* into the note field;
- it MUST be invalidated i.e. not executable on any official Algorand network. Malicious users could intercept and execute it causing unexpected consequences. For instance, in case of a MainNet transaction, it could be intercepted by a malicious user and executed, burning some Algos from the sender’s wallet due to the payment of fees.

The fields of a Payment Transaction object which represents a Transaction Authentication Message MUST be initialized as follows

- `amount` = 0;
- `sender/receiver` set with the user’s Algorand account;
- `firstValid/lastValid` = 0;
- `fee` = 0;
- `genesisId` = “ARC-0014-authentication”;
- `genesisHash` = SHA-512/256 of the string “ARC-0014-authentication”;
- `note` = SHA512_256(“ARC-0014-authentication”+msgpacked_auth_message);

The fields genesisId and genesisHash specify respectively the id and hash of the genesis block of the network, and they MUST be initialized to the ARC-0014 values (not conventional for MainNet/TestNet/BetaNet). The note field MUST include a Simple Authentication Message object. Finally, the sender field SHOULD be set to the Algorand account of the user. A detailed description of the transitions fields of Algorand Transactions is available in the transaction reference documentation.

For example:

```json
{
  "txn": {
    "amt": 0,
    "fee": 0,
    "fv": 0,
    "gen": "ARC-0014-authentication",
    "gh": "esN73ktiC1qzkkit8=",
    "lv": 0,
    "rcv": "EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4",
    "snd": "EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4",
    "type": "pay",
    "note": "SGVsbG8gV29ybGQerg1er1dfgt=",
  }
}
```

### How to sign an authentication message?

#### Digital Signature of a Simple Authentication Message

A Simple Authentication Message is a random sequence of bytes. It can be obtained with any open-source cryptographic library (e.g. Python lib PyNaCl) using the secret key of an Algorand account. Algorand account’s secret key SKa represents a base64 concatenation of the traditional 32-bytes secret and public keys . To create a digital signature of a Simple Authentication Message the Algorand secret key MUST be decoded as a standard 32-bytes secret key as follows:

1. base64 decode the Algorand secret key SKa;
2. extract the traditional secret key SK as the first 32 bytes of the decoded key.

#### Digital Signature of a Transaction Authentication Message

The signature of a Transaction Authentication Message is nothing more than a signed transaction object on Algorand. With the SDK method sign() and a secret key, it is possible to create a SignedTransaction object which wraps together the transaction and its  digital signature. For example, the SDK method  unsigned_txn.sign(secret_key) returns a SignedTransaction object with the digital signature of unsigned_transaction produced with secret_key.

Almost any algorand wallet integrates the possibility of signing transactions. For instance, to programmatically sign a transaction with the AlgoSigner wallet, there is an SDK method AlgoSigner.signTxn([TxnObject, …]). This method returns a base64 encoded object which represents the digital signature of a transaction object.

## How to verify a digital signature generated with Algorand keys?

A digital signature generated with the secret key SKa of an Algorand account can be verified with its respective 32-byte public key PKa. The Verifier only needs to decode the public key PK from the Algorand address PKa, and it must know the original Authentication Message. For example, assuming the digital signature Sig(msg) of the Authentication Message msg, the Verifier can validate it using the Algorand SDK as follows:

1. decode the Algorand address into a traditional 32-bytes public key PK;
2. use an open-source cryptographic library to verify the signature Sig(msg) (e.g. Python lib PyNaCl); usually those libraries work with raw bytes and some encoding/decoding is needed:
    - the message msg MUST be bytes encoded with msgpack. For example with the Algorand SDK for Python this is achieved with the method encoding.msgpack_encode(msg);
    - if the message is a Simple Message Transaction, then it MUST be prefixed with the string “ARC-0014-authentication”, otherwise it is a Transaction object and MUST be prefixed with the string “TX” ;
    - the digital signature Sig(msg) must be decoded from base64 with the base64 library of the SDK. For example with the Python SDK this is achieved with the method base64.b64decode(Sig(msg)).

## Standard for Session Id creation

Once the User is authenticated, the Verifier SHOULD exchange a session-id, for example a cookie or a JWT. The way session-ids are implemented is an ARC–0014 user choice. However, in a context of ARC-0014 authentication, it MUST include the following information:

- the Algorand address PKa of the authenticated user;
- the Simple Authentication Message;
the expiration time;
- (optional) device id if the User authenticates via mobile device.

### JWT implementation example

A JWT should be constructed following the RFC 7519 standard. It is composed of three parts, namely the header, payload, and signature. The payload includes general information on the JWT and the signing scheme, the payload contains the claims of the token, and the signature is derived from the header and the payload. The JWT is the concatenation of each part followed by a “.”: JWT = header.payload.signature

For example, an ARC-0014 compliant JWT might be implemented as follows:

Header:

```json
{
 "alg":"SHA512_256",
 "typ": "JWT"
}
```

Payload:

```json
{
 "algo_addr":"PKa",
 "message":"SHA512_256(‘ARC-0014-authentication’+msgpacked_auth_message)",
 "exp":<timestamp>,
 "device":"<device_id>"
}
```

Signature:

`SHA512_256(base64(Header) + “.” + base64(Payload), secret)`

Where the secret is only known by the Verifier.

## Rationale

> The rationale fleshes out the specification by describing what motivated the design and why particular design decisions were made. It should describe alternate designs that were considered and related work, e.g. how the feature is supported in other languages.

## Reference Implementation

> An optional section that contains a reference/example implementation that people can use to assist in understanding or implementing this specification.  If the implementation is too large to reasonably be included inline, then consider adding it as one or more files in `../assets/arc-####/`.

## Security Considerations

> All ARCs must contain a section that discusses the security implications/considerations relevant to the proposed change. Include information that might be important for security discussions, surfaces risks and can be used throughout the life cycle of the proposal. E.g. include security-relevant design decisions, concerns, important discussions, implementation-specific guidance and pitfalls, an outline of threats and risks and how they are being addressed. ARC submissions missing the "Security Considerations" section will be rejected. An ARC cannot proceed to status "Final" without a Security Considerations discussion deemed sufficient by the reviewers.

## Copyright

Copyright and related rights waived via [CC0](https://creativecommons.org/publicdomain/zero/1.0/).
