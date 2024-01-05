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
App mutability is an important factor in a rational actors decision making when it comes to interacting with an app and sending funds to a contract account. If a contract is mutable, it is possible that funds could be stolen or lost after the initial program is changed or deleted. Currently, there is no easy way for an end-user to determine the mutability of a contract. It is possible to use static analysis on the decompiled TEAL, but this is not easily accessible and unrealistic for wallets and explorers to implement for every program.

Additionally, there is no standard way for developers to route calls based on their OnCompletion. This can make it hard to determine what is allowed for a specific code path since a check for OnCompletion can be anywhere in the program. The current common practice is to include OnCompletion checks within the methods label/subroutine, but this can lead to a lot of extra bytes in the program.  

## Specification

### TEAL
The TEAL for an application **MUST** start with the following lines:

```
txn ApplicationID
!
pushint 6
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

A client **MUST** parse the decompiled TEAL to ensure that the firt nine lines match the lines specified in the TEAL specification. Prior to parsing, `inctblock` or `bytecblock` opcodes **MUST** be removed from the TEAL because they may come before the opcode in this ARC, but don't affect functionality.

To support future changes where comments may get added to the decompiled TEAL, clients **MUST** only check the beginning of lines to ensure the opcode and its argument(s) match the specification.

The client **MUST** follow the label for a specific action. If that label does not exist in the `switch` opcode OR if it leads to the `err` label, then that action is known to not be supported by the TEAL.

## Rationale
Standardizing the initial opcodes of the TEAL program makes it easy for anyone to parse the TEAL, follow a label for a specific action, and and determine whether that action is supported or not. This makes it possible for tools in the ecosystem, such as explorers and wallets, to tell end-users whether apps they are interacting with are mutable or not.

An alternative approach is to include asserts at the beginning of the program for any OnCompletion that is not supported by the contract, but this leads to a larger program size and potentially opcode cost. This approach also still makes it hard to determine what OnCompletions *are* supported for individual methods.

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
