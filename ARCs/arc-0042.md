---
arc: 42
title: xGov Pilot - Integration
description: Integration of xGov Process
author: Stéphane Barroso (@SudoWeezy)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/204
status: Deprecated
type: Informational
created: 2023-06-01
---

## Abstract

This ARC aims to explain how the xGov process can be integrated within dApps.

## Motivation

By leveraging the xGov decentralization, it can improve the overall efficiency of this initiative.

## Specification
The keywords "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

### How to register

#### How to find the xGov Escrow address

The xGov Escrow address can be extracted using this endpoint: `https://governance.algorand.foundation/api/periods/active/`.

```json
{
  ...
  "xgov_escrow_address": "string",
  ...
}
```

#### Registration
Governors should specify the xGov-related fields. Specifically, governors can sign up to be xGovs by designating as beneficiaries the xGov escrow address (that changes from one governance period to the next). They can also designate an xGov-controller address that would participate on their behalf in xGov votes via the optional parameter "xGv":"aaa". Namely, the Notes field has the form.

af/gov1:j{"com":nnn,"mmm1":nnn1,"mmm2":nnn2,"bnf":"XYZ","xGv":"ABC"}
Where:

"com":nnn is the Algo commitment;
"mmm":nnn is a commitment for LP-token with asset-ID mmm;
"bnf":"XYZ" designates the address "XYZ" as the recipient of rewards ("XYZ" must equal the xGov escrow in order to sign up as an xGov);
The optional "xGv":"ABC" designates address "ABC" as the xGov-controller of this xGov account.

#### Goal example

goal clerk send -a 0 -f ALDJ4R2L2PNDGQFSP4LZY4HATIFKZVOKTBKHDGI2PKAFZJSWC4L3UY5HN4 -t RFKCBRTPO76KTY7KSJ3HVWCH5HLBPNBHQYDC52QH3VRS2KIM7N56AS44M4 -n

‘af/gov1:j{“com”:1000000,“12345":2,“67890”:30,“bnf”:“DRWUX3L5EW7NAYCFL3NWGDXX4YC6Y6NR2XVYIC6UNOZUUU2ERQEAJHOH4M”,“xGv”:“ALDJ4R2L2PNDGQFSP4LZY4HATIFKZVOKTBKHDGI2PKAFZJSWC4L3UY5HN4”}’


### How to Interact with the Voting Application
#### How to get the Application ID
Every vote will be a different ID, but search for all apps created by the used account and look at the global state to see if is_bootstrapped is 1.

#### ABI

The ABI is available <a href="https://github.com/algorandfoundation/nft_voting_tool/blob/main/src/algorand/smart_contracts/artifacts/VotingRoundApp/contract.json">here </a>.
A working test example of how to call application's method is here:
https://github.com/algorandfoundation/nft_voting_tool/blob/main/src/algorand/smart_contracts/tests/voting.spec.ts

## Rationale
This integration will improve the usage of the process.

## Backwards Compatibility
None


## Security Considerations
None

## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
