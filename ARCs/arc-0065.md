---
arc: 65
title: AVM Run Time Errors In Program
description: Informative AVM run time errors based on program bytecode
author: Cosimo Bassi (@cusma), Tasos Bitsios (@tasosbit), Steve Ferrigno (@nullun)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/315
status: Final
type: Standards Track
category: ARC
created: 2024-10-09
---

## Abstract

This document introduces a convention for rising informative run time errors on
the Algorand Virtual Machine (AVM) directly from the program bytecode.

## Motivation

The AVM does not offer native opcodes to catch and raise run time errors.

The lack of native error handling semantics could lead to fragmentation of tooling
and frictions for AVM clients, who are unable to retrieve informative and useful
hints about the occurred run time failures.

This ARC formalizes a convention to rise AVM run time errors based just on the program
bytecode.

## Specification

The keywords "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**",
"**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**"
in this document are to be interpreted as described in <a href="https://datatracker.ietf.org/doc/html/rfc2119">RFC 2119</a>.

> Notes like this are non-normative.

### Error format

> The AVM programs bytecode have limited sized. In this convention, the errors are
> part of the bytecode, therefore it is good to mind errors' formatting and sizing.

> Errors consist of a _code_ and an optional _short message_.

Errors **MUST** be prefixed either with:

- `ERR:` for custom errors;
- `AER:` reserved for future ARC standard errors.

Errors **MUST** use `:` as domain separator.

It is **RECOMMENDED** to use `UTF-8` for the error bytes string encoding.

It is **RECOMMENDED** to use _short_ error messages.

It is **RECOMMENDED** to use <a href="https://en.wikipedia.org/wiki/Camel_case/">camel case</a>
for alphanumeric error codes.

It is **RECOMMENDED** to avoid error byte strings of _exactly_ 8 or 32 bytes.

### In Program Errors

When a program wants to emit informative run time errors, directly from the bytecode,
it **MUST**:

1. Push to the stack the bytes string containing the error;
1. Execute the `log` opcode to use the bytes from the top of the stack;
1. Execute the `err` opcode to immediately terminate the program.

Upon a program run time failure, the Algod API response contains both the failed
_program counter_ (`pc`) and the `logs` array with the _errors_.

The program **MAY** return multiple errors in the same failed execution.

The errors **MUST** be retrieved by:

1. Decoding the `base64` elements of the `logs` array;
1. Validating the decoded elements against the error regexp.

### Error examples

> Error conforming this specification are always prefixed with `ERR:`.

Error with a _numeric code_: `ERR:042`.

Error with an _alphanumeric code_: `ERR:BadRequest`.

Error with a _numeric code_ and _short message_: `ERR:042:AFunnyError`.

### Program example

The following program example raises the error `ERR:001:Invalid Method` for any
application call to methods different from `m1()void`.

```teal
#pragma version 10

txn ApplicationID
bz end

method "m1()void"
txn ApplicationArgs 0
match method1
byte "ERR:001:Invalid Method"
log
err

method1:
b end

end:
int 1
```

Full Algod API response of a failed execution:

```json
{
    "data": {
        "app-index":1004,
        "eval-states": [
            {
                "logs": ["RVJSOjAwMTpJbnZhbGlkIE1ldGhvZA=="]
            }
        ],
        "group-index":0,
        "pc":41
    },
    "message":"TransactionPool.Remember: transaction ESI4GHAZY46MCUCLPBSB5HBRZPGO6V7DDUM5XKMNVPIRJK6DDAGQ: logic eval error: err opcode executed. Details: app=1004, pc=41"
}
```

The `logs` array contains the `base64` encoded error `ERR:001:Invalid Method`.

The `logs` array **MAY** contain elements that are not errors (as specified by the
regexp).

It is **NOT RECOMMENDED** to use the `message` field to retrieve errors.

### AVM Compilers

AVM compilers (and related tools) **SHOULD** provide two error compiling options:

1. The one specified in this ARC as **default**;
1. The one specified in [ARC-56](./arc-0056.md) as fallback, if compiled bytecode
size exceeds the AVM limits.

> Compilers **MAY** optimize for program bytecode size by storing the error prefixes
in the `bytecblock` and concatenating the error message at the cost of some extra
opcodes.

## Rationale

This convention for AVM run time errors presents the following PROS and CONS.

**PROS:**
- No additional artifacts required to return informative run time errors;
- Errors are directly returned in the Algod API response, which can be filtered
with the specified error regexp.

**CONS:**
- Errors consume program bytecode size.

## Security Considerations

> Not applicable.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
