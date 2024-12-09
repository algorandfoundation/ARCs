---
arc: 22
title: Add `read-only` annotation to ABI methods
description: Convention for creating methods which don't mutate state
author: ori-shem-tov (@ori-shem-tov)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/125
status: Final
type: Standards Track
category: Interface
sub-category: Application
created: 2022-03-16
requires: 4
---

# Extend [ARC-4](./arc-0004.md) to add `read-only` annotation to methods

The following document introduces a convention for creating methods (as described in [ARC-4](./arc-0004.md)) which don't mutate state.

## Abstract

The goal of this convention is to allow smart contract developers to distinguish between methods which mutate state and methods which don't by introducing a new property to the `Method` descriptor.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

> Comments like this are non-normative.

### Read-only functions

A `read-only` function is a function with no side-effects. In particular, a `read-only` function **SHOULD NOT** include:
- local/global state modifications
- calls to non `read-only` functions
- inner-transactions

It is **RECOMMENDED** for a `read-only` function to not access transactions in a group or metadata of the group.

> The goal is to allow algod to easily execute `read-only` functions without broadcasting a transaction

In order to support this annotation, the following `Method` descriptor is suggested:
```typescript
interface Method {
  /** The name of the method */
  name: string;
  /** Optional, user-friendly description for the method */
  desc?: string;
  /** Optional, is it a read-only method (according to ARC-22) */
  readonly?: boolean
  /** The arguments of the method, in order */
  args: Array<{
    /** The type of the argument */
    type: string;
    /** Optional, user-friendly name for the argument */
    name?: string;
    /** Optional, user-friendly description for the argument */
    desc?: string;
  }>;
  /** Information about the method's return value */
  returns: {
    /** The type of the return value, or "void" to indicate no return value. */
    type: string;
    /** Optional, user-friendly description for the return value */
    desc?: string;
  };
}
```
## Rationale

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
