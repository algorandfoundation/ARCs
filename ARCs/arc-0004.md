---
arc: 4
title: Application Binary Interface (ABI)
description: Conventions for encoding method calls in Algorand Application
author: Jannotti (@jannotti), Jason Paulos (@jasonpaulos)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/44
status: Final
type: Standards Track
category: Interface
sub-category: Application
created: 2021-07-29
---

# Algorand Transaction Calling Conventions

## Abstract

This document introduces conventions for encoding method calls,
including argument and return value encoding, in Algorand Application
call transactions.
The goal is to allow clients, such as wallets and
dapp frontends, to properly encode call transactions based on a description
of the interface. Further, explorers will be able to show details of
these method invocations.

### Definitions

* **Application:** an Algorand Application, aka "smart contract",
  "stateful contract", "contract", or "app".
* **HLL:** a higher level language that compiles to TEAL bytecode.
* **dapp (frontend)**: a decentralized application frontend, interpreted here to
  mean an off-chain frontend (a webapp, native app, etc.) that interacts with
  Applications on the blockchain.
* **wallet**: an off-chain application that stores secret keys for on-chain
  accounts and can display and sign transactions for these accounts.
* **explorer**: an off-chain application that allows browsing the blockchain,
  showing details of transactions.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

> Comments like this are non-normative.

Interfaces are defined in TypeScript. All the objects that are defined
are valid JSON objects, and all JSON `string` types are UTF-8 encoded.

### Overview

This document makes recommendations for encoding method invocations as
Application call transactions, and for describing methods for
access by higher-level entities.  Encoding recommendations are
intended to be minimal, intended only to allow interoperability among
Applications.  Higher level recommendations are intended to enhance
user-facing interfaces, such as high-level languages, dapps, and
wallets.  Applications that follow the recommendations described here are
called _[ARC-4](./arc-0004.md) Applications_.

### Methods

A method is a section of code intended to be invoked externally with
an Application call transaction. A method must have a name, it may
take a list of arguments as input when it is invoked, and it may
return a single value (which may be a tuple) when it finishes
running. The possible types for arguments and return values are
described later in the [Encoding](#encoding) section.

Invoking a method involves creating an Application call transaction to
specifically call that method. Methods are different from internal
subroutines that may exist in a contract, but are not externally
callable. Methods may be invoked by a top-level Application call
transaction from an off-chain caller, or by an Application call inner
transaction created by another Application.


#### Method Signature

A method signature is a unique identifier for a method. The signature
is a string that consists of the method's name, an open parenthesis, a
comma-separated list of the types of its arguments, a closing
parenthesis, and the method's return type, or `void` if it does not
return a value. The names of the arguments **MUST NOT** be included in a
method's signature, and **MUST NOT** contain any whitespace.

For example, `add(uint64,uint64)uint128` is the method signature for a
method named `add` which takes two uint64 parameters and returns a
uint128. Signatures are encoded in ASCII.

For the benefit of universal interoperability (especially in HLLs),
names **MUST** satisfy the regular expression `[_A-Za-z][A-Za-z0-9_]*`.
Names starting with an underscore are reserved and **MUST** only
be used as specified in this ARC or future ABI-related ARC.


#### Method Selector

Method signatures contain all the information needed to identify a
method, however the length of a signature is unbounded. Rather than
consume program space with such strings, a method selector is used to
identify methods in calls. A method selector is the first four bytes of
the SHA-512/256 hash of the method signature.

For example, the method selector for a method named `add` which takes
two uint64 parameters and returns a uint128 can be computed as
follows:

```
Method signature: add(uint64,uint64)uint128
SHA-512/256 hash (in hex): 8aa3b61f0f1965c3a1cbfa91d46b24e54c67270184ff89dc114e877b1753254a
Method selector (in hex): 8aa3b61f
```


#### Method Description

A method description provides further information about a method
beyond its signature. This description is encoded in JSON and consists
of a method's name, description (optional), arguments (their types, and optional names and
descriptions), and return type and optional description for the
return type. From this structure, the method's signature and selector
can be calculated. The Algorand SDKs provide convenience functions
to calculate signatures and selectors from such JSON files.

These details will enable high-level languages and dapps/wallets to
properly encode arguments, call methods, and decode return
values. This description can populate UIs in dapps, wallets, and
explorers with description of parameters, as well as populate
information about methods in IDEs for HLLs.

The JSON structure for such an object is:

```typescript
interface Method {
  /** The name of the method */
  name: string;
  /** Optional, user-friendly description for the method */
  desc?: string;
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

For example:

```json
{
  "name": "add",
  "desc": "Calculate the sum of two 64-bit integers",
  "args": [
    { "type": "uint64", "name": "a", "desc": "The first term to add" },
    { "type": "uint64", "name": "b", "desc": "The second term to add" }
  ],
  "returns": { "type": "uint128", "desc": "The sum of a and b" }
}
```


### Interfaces

An Interface is a logically grouped set of methods. All method selectors in an
Interface **MUST** be unique. Method names **MAY** not be unique, as long as
the corresponding method selectors are different. Method names in Interfaces
**MUST NOT** begin with an underscore.

An Algorand Application *implements* an Interface if it supports
all of the methods from that Interface. An Application **MAY** implement
zero, one, or multiple Interfaces.

Interface designers **SHOULD** try to prevent collisions of method selectors
between Interfaces that are likely to be implemented together by the same
Application.

> For example, an Interface `Calculator` providing addition and subtraction
> of integer methods and an Interface `NumberFormatting` providing formatting
> methods for numbers into strings are likely to be used together.
> Interface designers should ensure that all the methods in `Calculator` and
> `NumberFormatting` have distinct method selectors.


#### Interface Description

An Interface description is a JSON object containing the JSON
descriptions for each of the methods in the Interface.

The JSON structure for such an object is:

```typescript
interface Interface {
  /** A user-friendly name for the interface */
  name: string;
  /** Optional, user-friendly description for the interface */
  desc?: string;
  /** All of the methods that the interface contains */
  methods: Method[];
}
```

Interface names **MUST** satisfy the regular expression `[_A-Za-z][A-Za-z0-9_]*`.
Interface names starting with `ARC` are reserved to interfaces defined in ARC.
Interfaces defined in `ARC-XXXX` (where `XXXX` is a 0-padded number) **SHOULD**
start with `ARC_XXXX`.

For example:

```json
{
  "name": "Calculator",
  "desc": "Interface for a basic calculator supporting additions and multiplications",
  "methods": [
    {
      "name": "add",
      "desc": "Calculate the sum of two 64-bit integers",
      "args": [
        { "type": "uint64", "name": "a", "desc": "The first term to add" },
        { "type": "uint64", "name": "b", "desc": "The second term to add" }
      ],
      "returns": { "type": "uint128", "desc": "The sum of a and b" }
    },
    {
      "name": "multiply",
      "desc": "Calculate the product of two 64-bit integers",
      "args": [
        { "type": "uint64", "name": "a", "desc": "The first factor to multiply" },
        { "type": "uint64", "name": "b", "desc": "The second factor to multiply" }
      ],
      "returns": { "type": "uint128", "desc": "The product of a and b" }
    }
  ]
}
```

### Contracts

A Contract is a declaration of what an Application implements. It includes
the complete list of the methods implemented by the related Application.
It is similar to an Interface, but it may include further details about the
concrete implementation, as well as implementation-specific methods that
do not belong to any Interface. All methods in a Contract **MUST** be
unique; specifically, each method **MUST** have a unique method selector.

Method names in Contracts **MAY** begin with underscore, but these
names are reserved for use by this ARC and future extensions of this ARC.

#### OnCompletion Actions and Creation

In addition to the set of methods from the Contract's definition,
a Contract **MAY** allow Application calls with zero arguments, also
known as bare Application calls. Since method invocations with zero
arguments still encode the method selector as the first Application
call argument, bare Application calls are always distinguishable
from method invocations.

The primary purpose of bare Application calls is to allow the
execution of an OnCompletion (`apan`) action which requires no
inputs and has no return value. A Contract **MAY** allow this for
all of the OnCompletion actions listed below, for only a subset of
them, or for none at all. Great care should be taken when allowing
these operations.

Allowed OnCompletion actions:
* 0: NoOp
* 1: OptIn
* 2: CloseOut
* 4: UpdateApplication
* 5: DeleteApplication

Note that OnCompletion action 3, ClearState, is **NOT** allowed to
be invoked as a bare Application call.

> While ClearState is a valid OnCompletion action, its behavior differs
> significantly from the other actions. Namely, an Application running during
> ClearState which wishes to have any effect on the state of the chain
> must never fail, since due to the unique behavior about ClearState
> failure, doing so would revert any effect made by that Application.
> Because of this, Applications running during ClearState are
> incentivized to never fail. Accepting any user input, whether that
> is an ABI method selector, method arguments, or even relying on the
> absence of Application arguments to indicate a bare Application call,
> is therefore a dangerous operation, since there is no way to enforce
> properties or even the existence of data that is supplied by the user.

If a Contract elects to allow bare Application calls for some
OnCompletion actions, then that Contract **SHOULD** also allow
any of its methods to be called with those OnCompletion actions,
as long as this would not cause undesirable or nonsensical behavior.

> The reason for this is because if it's acceptable to allow an
> OnCompletion action to take place in isolation inside of a bare
> Application call, then it's most likely acceptable to allow
> the same action to take place at the same time as an ABI method
> call. And since the latter can be accomplished in just one
> transaction, it can be more efficient.

If a Contract requires an OnCompletion action to take inputs or
to return a value, then the **RECOMMENDED** behavior of the Contract
is to not allow bare Application calls for that OnCompletion
action. Rather, the Contract should have one or more methods that
are meant to be called with the appropriate OnCompletion action set
in order to process that action.

A Contract **MUST NOT** allow any of its methods to be called with the
ClearState OnCompletion action.

> To reinforce an earlier point, it is unsafe for a ClearState program
> to read any user input, whether that is a method argument or even
> relying on a certain method selector to be present. This behavior
> makes it unsafe to use ABI calling conventions during ClearState.

If an Application is called with greater than zero Application call
arguments (i.e. **NOT** a bare Application call) and the OnCompletion
action is **NOT** ClearState, the Application **MUST**
always treat the first argument as a method selector and invoke
the specified method. This behavior **MUST** be followed for all
OnCompletion actions, except for ClearState. This applies to Application
creation transactions as well, where the supplied Application ID is 0.

Similar to OnCompletion actions, if a Contract requires its
creation transaction to take inputs or to return a value, then
the **RECOMMENDED** behavior of the Contract should be to not
allow bare Application calls for creation. Rather, the Contract
should have one or more methods that are meant to be called
in order to create the Contract.

#### Contract Description

A Contract description is a JSON object containing the JSON
descriptions for each of the methods in the Contract.

The JSON structure for such an object is:

```typescript
interface Contract {
  /** A user-friendly name for the contract */
  name: string;
  /** Optional, user-friendly description for the interface */
  desc?: string;
  /**
   * Optional object listing the contract instances across different networks
   */
  networks?: {
    /**
     * The key is the base64 genesis hash of the network, and the value contains
     * information about the deployed contract in the network indicated by the
     * key
     */
    [network: string]: {
      /** The app ID of the deployed contract in this network */
      appID: number;
    }
  }
  /** All of the methods that the contract implements */
  methods: Method[];
}
```

Contract names **MUST** satisfy the regular expression `[_A-Za-z][A-Za-z0-9_]*`.

The `desc` fields of the Contract and the methods inside the Contract
**SHOULD** contain information that is not explicitly encoded in the other fields,
such as support of bare Application calls, requirement of specific
OnCompletion action for specific methods, and methods to call for creation
(if creation cannot be done via a bare Application call).

For example:

```json
{
  "name": "Calculator",
  "desc": "Contract of a basic calculator supporting additions and multiplications. Implements the Calculator interface.",
  "networks": {
    "wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8=": { "appID": 1234 },
    "SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=": { "appID": 5678 },
  },
  "methods": [
    {
      "name": "add",
      "desc": "Calculate the sum of two 64-bit integers",
      "args": [
        { "type": "uint64", "name": "a", "desc": "The first term to add" },
        { "type": "uint64", "name": "b", "desc": "The second term to add" }
      ],
      "returns": { "type": "uint128", "desc": "The sum of a and b" }
    },
    {
      "name": "multiply",
      "desc": "Calculate the product of two 64-bit integers",
      "args": [
        { "type": "uint64", "name": "a", "desc": "The first factor to multiply" },
        { "type": "uint64", "name": "b", "desc": "The second factor to multiply" }
      ],
      "returns": { "type": "uint128", "desc": "The product of a and b" }
    }
  ]
}
```


### Method Invocation

In order for a caller to invoke a method, the caller and the method
implementation (callee) must agree on how information will be passed
to and from the method. This ABI defines a standard for where this
information should be stored and for its format.

This standard does not apply to Application calls with the ClearState
OnCompletion action, since it is unsafe for ClearState programs to rely
on user input.

#### Standard Format

The method selector must be the first Application call argument (index 0),
accessible as `txna ApplicationArgs 0` from TEAL (except for bare Application
calls, which use zero application call arguments).

If a method has 15 or fewer arguments, each argument **MUST** be placed in
order in the following Application call argument slots (indexes 1 through
15). The arguments **MUST** be encoded as defined in the [Encoding](#encoding)
section.

Otherwise, if a method has 16 or more arguments, the first 14 **MUST** be
placed in order in the following Application call argument slots (indexes 1
through 14), and the remaining arguments **MUST** be encoded as a tuple
in the final Application call argument slot (index 15). The arguments must
be encoded as defined in the [Encoding](#encoding) section.

If a method has a non-void return type, then the return value of the method
**MUST** be located in the final logged value of the method's execution,
using the `log` opcode. The logged value **MUST** contain a specific 4 byte
prefix, followed by the encoding of the return value as defined in the
[Encoding](#encoding) section. The 4 byte prefix is defined as the first 4
bytes of the SHA-512/256 hash of the ASCII string `return`. In hex, this is
`151f7c75`.

> For example, if the method `add(uint64,uint64)uint128` wanted to return the
> value 4160, it would log the byte array `151f7c7500000000000000000000000000001040`
> (shown in hex).

#### Implementing a Method

An ARC-4 Application implementing a method:

1. **MUST** check if `txn NumAppArgs` equals 0. If true, then this is a
bare Application call. If the Contract supports bare Application calls
for the current transaction parameters (it **SHOULD** check the OnCompletion
action and whether the transaction is creating the application), it **MUST**
handle the call appropriately and either approve or reject the transaction.
The following steps **MUST** be ignored in this case. Otherwise, if
the Contract does not support this bare application call, the Contract
**MUST** reject the transaction.

2. **MUST** examine `txna ApplicationArgs 0` to identify the selector
of the method being invoked. If the contract does not implement a method with
that selector, the Contract **MUST** reject the transaction.

3. **MUST** execute the actions required to implement the method being
invoked. In general, this works by branching to the body of the method
indicated by the selector.

4. The code for that method **MAY** extract the arguments it needs, if
any, from the application call arguments as described in the [Encoding](#encoding)
section. If the method has more than 15 arguments and the contract
needs to extract an argument beyond the 14th, it **MUST** decode
`txna ApplicationArgs 15` as a tuple to access the arguments contained in it.

5. If the method is non-void, the Application **MUST** encode the
return value as described in the [Encoding](#encoding) section and then
`log` it with the prefix `151f7c75`. Other values **MAY** be logged
before the return value, but other values **MUST NOT** be logged after
the return value.


#### Calling a Method from Off-Chain

To invoke an ARC-4 Application, an off-chain system, such as a dapp or wallet,
would first obtain the Interface or Contract description JSON object
for the app. The client may now:

1. Create an Application call transaction with the following parameters:
    1. Use the ID of the desired Application whose program code
       implements the method being invoked, or 0 if they wish to
       create the Application.
    2. Use the selector of the method being invoked as the first
       Application call argument.
    3. Encode all arguments for the method, if any, as described in
       the [Encoding](#encoding) section. If the method has more than 15 arguments,
       encode all arguments beyond (but not including) the 14th
       as a tuple into the final Application call argument.
2. Submit this transaction and wait until it successfully commits to
   the blockchain.
3. Decode the return value, if any, from the ApplyData's log
   information.

Clients **MAY** ignore the return value.

An exception to the above instructions is if the app supports bare
Application calls for some transaction parameters, and the client
wishes to invoke this functionality. Then the client may simply create
and submit to the network an Application call transaction with the ID of
the Application (or 0 if they wish to create the application) and the
desired OnCompletion value set. Application arguments **MUST NOT** be
present.

### Encoding

This section describes how ABI types can be represented as byte strings.

Like the <a href="https://docs.soliditylang.org/en/v0.8.6/abi-spec.html">EthereumABI</a>, this
encoding specification is designed to have the following two
properties:


1. The number of non-sequential "reads" necessary to access a value is
   at most the depth of that value inside the encoded array
   structure. For example, at most 4 reads are needed to retrieve a
   value at `a[i][k][l][r]`.
2. The encoding of a value or array element is not interleaved with
   other data and it is relocatable, i.e. only relative “addresses”
   (indexes to other parts of the encoding) are used.


#### Types

The following types are supported in the Algorand ABI.

* `uint<N>`: An `N`-bit unsigned integer, where `8 <= N <= 512` and `N % 8 = 0`. When this type is
used as part of a method signature, `N` must be written as a base 10 number without any leading zeros.
* `byte`: An alias for `uint8`.
* `bool`: A boolean value that is restricted to either 0 or 1. When encoded, up to 8 consecutive
`bool` values will be packed into a single byte.
* `ufixed<N>x<M>`: An `N`-bit unsigned fixed-point decimal number with precision `M`, where
`8 <= N <= 512`, `N % 8 = 0`, and `0 < M <= 160`, which denotes a value `v` as `v / (10^M)`. When
this type is used as part of a method signature, `N` and `M` must be written as base 10 numbers
without any leading zeros.
* `<type>[<N>]`: A fixed-length array of length `N`, where `N >= 0`. `type` can be any other type.
When this type is used as part of a method signature, `N` must be written as a base 10 number without
any leading zeros, _unless_ `N` is zero, in which case only a single 0 character should be used.
* `address`: Used to represent a 32-byte Algorand address. This is equivalent to `byte[32]`.
* `<type>[]`: A variable-length array. `type` can be any other type.
* `string`: A variable-length byte array (`byte[]`) assumed to contain UTF-8 encoded content.
* `(T1,T2,…,TN)`: A tuple of the types `T1`, `T2`, …, `TN`, `N >= 0`.
* reference types `account`, `asset`, `application`: **MUST NOT** be used as the return type.
For encoding purposes they are an alias for `uint8`. See section "Reference Types" below.

Additional special use types are defined in [Reference Types](#reference-types)
and [Transaction Types](#transaction-types).

#### Static vs Dynamic Types

For encoding purposes, the types are divided into two categories: static and dynamic.

The dynamic types are:

*  `<type>[]` for any `type`
    * This includes `string` since it is an alias for `byte[]`.
* `<type>[<N>]` for any dynamic `type`
* `(T1,T2,...,TN)` if `Ti` is dynamic for some `1 <= i <= N`

All other types are static. For a static type, all encoded values of
that type have the same length, irrespective of their actual value.


#### Encoding Rules

Let `len(a)` be the number of bytes in the binary string `a`. The
returned value shall be considered to have the ABI type `uint16`.

Let `enc` be a mapping from values of the ABI types to binary
strings. This mapping defines the encoding of the ABI.

For any ABI value `x`, we recursively define `enc(x)` to be as follows:

* If `x` is a tuple of `N` types, `(T1,T2,...,TN)`, where `x[i]` is the value at index `i`, starting at 1:
    * `enc(x) = head(x[1]) ... head(x[N]) tail(x[1]) ... tail(x[N])`
    * Let `head` and `tail` be mappings from values in this tuple to binary strings. For each `i` such that `1 <= i <= N`, these mappings are defined as:
        * If `Ti` (the type of `x[i]`) is static:
            * If `Ti` is `bool`:
                * Let `after` be the largest integer such that all `T(i+j)` are `bool`, for `0 <= j <= after`.
                * Let `before` be the largest integer such that all `T(i-j)` are `bool`, for `0 <= j <= before`.
                * If `before % 8 == 0`:
                    * `head(x[i]) = enc(x[i]) | (enc(x[i+1]) >> 1) | ... | (enc(x[i + min(after,7)]) >> min(after,7))`, where `>>` is bitwise right shift which pads with 0, `|` is bitwise or, and `min(x,y)` returns the minimum value of the integers `x` and `y`.
                    * `tail(x[i]) = ""` (the empty string)
                * Otherwise:
                    * `head(x[i]) = ""` (the empty string)
                    * `tail(x[i]) = ""` (the empty string)
            * Otherwise:
                * `head(x[i]) = enc(x[i])`
                * `tail(x[i]) = ""` (the empty string)
        * Otherwise:
            * `head(x[i]) = enc(len( head(x[1]) ... head(x[N]) tail(x[1]) ... tail(x[i-1]) ))`
            * `tail(x[i]) = enc(x[i])`
* If `x` is a fixed-length array `T[N]`:
    * `enc(x) = enc((x[0], ..., x[N-1]))`, i.e. it’s encoded as if it were an `N` element tuple where every element is type `T`.
* If `x` is a variable-length array `T[]` with `k` elements:
    * `enc(x) = enc(k) enc([x[0], ..., x[k-1]])`, i.e. it’s encoded as if it were a fixed-length array of `k` elements, prefixed with its length, `k` encoded as a `uint16`.
* If `x` is an `N`-bit unsigned integer, `uint<N>`:
    * `enc(x)` is the `N`-bit big-endian encoding of `x`.
* If `x` is an `N`-bit unsigned fixed-point decimal number with precision `M`, `ufixed<N>x<M>`:
    * `enc(x) = enc(x * 10^M)`, where `x * 10^M` is interpreted as a `uint<N>`.
* If `x` is a boolean value `bool`:
    * `enc(x)` is a single byte whose **most significant bit** is either 1 or 0, if `x` is true or false respectively. All other bits are 0. Note: this means that a value of true will be encoded as `0x80` (`10000000` in binary) and a value of false will be encoded as `0x00`. This is in contrast to most other encoding schemes, where a value of true is encoded as `0x01`.

Other aliased types' encodings are already covered:
- `string` and `address` are aliases for `byte[]` and `byte[32]` respectively
- `byte` is an alias for `uint8`
- each of the reference types is an alias for `uint8`

#### Reference Types

Three special types are supported _only_ as the type of an argument.
They _can_ be embedded in arrays and tuples.

* `account` represents an Algorand account, stored in the Accounts (`apat`) array
* `asset` represents an Algorand Standard Asset (ASA), stored in the Foreign Assets (`apas`) array
* `application` represents an Algorand Application, stored in the Foreign Apps (`apfa`) array

Some AVM opcodes require specific values to be placed in the "foreign
arrays" of the Application call transaction. These three types allow
methods to describe these requirements. To encode method calls that
have these types as arguments, the value in question is placed in the
Accounts (`apat`), Foreign Assets (`apas`), or Foreign Apps (`apfa`)
arrays, respectively, and a `uint8` containing the index of the value
in the appropriate array is encoded in the normal location for this
argument.

Note that the Accounts and Foreign Apps arrays have an implicit value
at index 0, the Sender of the transaction or the called Application,
respectively. Therefore, indexes of any additional values begin at 1.
Additionally, for efficiency, callers of a method that wish to pass the
transaction Sender as an `account` value or the called Application as
an `application` value **SHOULD** use 0 as the index of these values
and not explicitly add them to Accounts or Foreign Apps arrays.

When passing addresses, ASAs, or apps that are _not_ required to be
accessed by such opcodes, ARC-4 Contracts **SHOULD** use the base
types for passing these types: `address` for accounts and `uint64`
for asset or Application IDs.

#### Transaction Types

Some apps require that they are invoked as part of a larger
transaction group, containing specific additional transactions.  Seven
additional special types are supported (only) as argument types to
describe such requirements.

* `txn` represents any Algorand transaction
* `pay` represents a PaymentTransaction (algo transfer)
* `keyreg` represents a KeyRegistration transaction (configure
  consensus participation)
* `acfg` represent a AssetConfig transaction (create, configure, or
  destroy ASAs)
* `axfer` represents an AssetTransfer transaction (ASA transfer)
* `afrz` represents an AssetFreezeTx transaction (freeze or unfreeze
  ASAs)
* `appl` represents an ApplicationCallTx transaction (create/invoke a Application)

Arguments of these types are encoded as consecutive transactions in
the same transaction group as the Application call, placed in the
position immediately preceding the Application call. Unlike "foreign"
references, these special types are not encoded in ApplicationArgs as
small integers "pointing" to the associated object.  In fact, they
occupy no space at all in the Application Call transaction
itself. Allowing explicit references would create opportunities for
multiple transaction "values" to point to the same transaction in the
group, which is undesirable. Instead, the locations of the
transactions are implied entirely by the placement of the transaction
types in the argument list.

For example, to invoke the method `deposit(string,axfer,pay,uint32)void`,
a client would create a transaction group containing, in this order:
1. an asset transfer
2. a payment
3. the actual Application call

When encoding the other (non-transaction) arguments, the client
**MUST** act as if the transaction arguments were completely absent
from the method signature. The Application call would contain the method
selector in ApplicationArgs[0], the first (string) argument in
ApplicationArgs[1], and the fourth (uint32) argument in ApplicationArgs[2].


ARC-4 Applications **SHOULD** be constructed to allow their invocations to be
combined with other contract invocations in a single atomic group if
they can do so safely. For example, they **SHOULD** use `gtxns` to examine
the previous index in the group for a required `pay` transaction,
rather than hardcode an index with `gtxn`.

In general, an ARC-4 Application method with `n` transactions as arguments **SHOULD**
only inspect the `n` previous transactions. In particular, it **SHOULD NOT**
inspect transactions after and it **SHOULD NOT** check the size of a transaction
group (if this can be done safely).
In addition, a given method **SHOULD** always expect the same
number of transactions before itself. For example, the method
`deposit(string,axfer,pay,uint32)void` is always preceded by two transactions.
It is never the case that it can be called only with one asset transfer
but no payment transfer.

> The reason for the above recommendation is to provide minimal
> composability support while preventing obvious dangerous attacks.
> For example, if some apps expect payment transactions after them
> while other expect payment transaction before them, then the same
> payment may be counted twice.

## Rationale

## Security Considerations

None.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
