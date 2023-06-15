---
arc: <to be assigned>
title: Remove ABI Reference Types
description: Removes reference types for the ABI specification
author: Joe Polny (@joe-p)
discussions-to: <URL>
status: Draft
type: Standards Track
category: ARC
created: 2023-07-15
requires: [ARC-0004](./arc-0004.md)
---

## Abstract
The ABI has three reference types: `Asset`, `Account`, `Application`. These reference types should be removed from the ABI specification.

## Motivation
Foreign references for a given method call are not always arguments to the method itself and different implementations of an interface might require different references. If a implementation of an interface in one contract requires different references than another implementation, it's method signature (thus method selector) changes which mitigates the benefits of implementing a standard interface.

Additionally, in AVM9, group resource sharing means a given transaction might not have all of the required resources in its own foreign reference arrays.

Informing callers on what references a given method will need can be covered in it's respective [ARC-0032](./arc-0032.md) description.

## Specification
The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

The following changes **MUST** be adhered to by compliant contracts and clients.

* `Asset` becomes an alias for `uint64`, with the value being the asset's unique ID

* `Account` is no longer an ABI type. `Address` should be used instead.

* `Application` becomes an alias for `uint64`, with the value being the application's unique ID

Methods **MUST** explicitly use the `Asset`, `Address`, or `Application` if it is in the arguments. Otherwise the argument **MUST NOT** be in the method signature.

## Rationale
This removes client-specific information from [ARC-0004](./arc-0004.md), makes it easier to write generic interfaces, and is more flexible for AVM9 resource sharing.

## Backwards Compatibility
[ARC-0004](./arc-0004.md) clients and contracts are not backwards compatible due to the changes in `Asset` and `Application` encoding and omission of the `Account` type.

## Test Cases
N/A

## Reference Implementation
An [ARC-0004](./arc-0004.md) method `foo(Account)void` will become `foo(Address)void` if, and only if, the on-chain logic explitictly uses the passed in address. If the on-chain logic does not explicitly use the passed in address (for example, the address is read from box storage), the method signature will become `foo()void`.

## Security Considerations
None

## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
