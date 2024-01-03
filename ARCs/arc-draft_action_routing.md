---
arc: <to be assigned>
title: Application Action Routing
description: Standardized handling of OnCompletion and ApplicationID actions 
author: Joe Polny (@joe-p)
discussions-to: <URL>
status: Draft
type: Standards Track
category: ARC
created: 2024-01-02
---

## Abstract
This ARC provides a standard way to handle an application call depending on the combination of the `OnCompletion` and `ApplicationID`.

## Motivation
Currently there is no way for someone to tell whether an app can be updated or deleted without doing a full static analysis on the TEAL.

## Specification

### TEAL
The TEAL for an application **MUST** start with the following lines:

```
txn ApplicationID
!
int 6
*
txn OnCompletion
+
switch call_NoOp call_OptIn call_CloseOut NOT_IMPLEMENTED call_UpdateApplication call_DeleteApplication create_NoOp create_OptIn NOT_IMPLEMENTED NOT_IMPLEMENTED NOT_IMPLEMENTED create_DeleteApplication
NOT_IMPLEMENTED:
err
```

For every action (combination of `ApplicationID === 0` and `OnCompletion`) there is a label that corresponds to that action. If that action is not implemented in the contract, its label should be replaced by a label that simply contains `err`. In the above TEAL, this is the `NOT_IMPLEMENTED` label.

The label containing the `err` opcode **MUST** come right after the `switch` opcode.

The specific label names in the pre-compiled TEAL **MAY** be different, but it is **RECOMMENDED** to use the label names here for consistency across tooling.

The `switch` statement in a contract **SHOULD** only include labels until the last supported action is reached. For example, if an application only supports `NoOp` creations, the switch can simply be `switch create_NoOp`. All other actions will simply go to the next opcode, which **MUST** be `err` as specified above.

### Client Parsing

A client **MUST** parse the decompiled TEAL to ensure that the firt nine lines match the lines specified in the TEAL specification. `intcblock` and `bytecblock` opcodes **MUST** be ignored if they are the first or second opcodes executed.

The client **MUST** follow the label for a specific action. If that label does not exist in the `switch` opcode OR if it leads to the `err` label, then that action is known to not be supported by the TEAL.

## Rationale
Standardizing the initial opcodes of the TEAL program makes it easy for anyone to parse the TEAL, follow a label for a specific action, and and determine whether that action is supported or not. 

## Backwards Compatibility
N/A

## Test Cases
N/A

## Reference Implementation
TODO: Client parsing example

## Security Considerations
If a client improperly implements parsing, they may mislead users about the mutability of the contract. It is imperative that `intcblock` and `bytecblock` are handled properly and the following opcodes are exactly as specificed in this ARC.

## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
