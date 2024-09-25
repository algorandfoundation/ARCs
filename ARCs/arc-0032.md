---
arc: 32
title: Application Specification
description: A specification for fully describing an Application, useful for Application clients.
author: Benjamin Guidarelli (@barnjamin)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/150
status: Final
type: Standards Track
category: ARC
sub-category: Application
created: 2022-12-01
requires: 4, 21
---

## Abstract

> [!NOTE]
> This specification will be eventually deprecated by the <a href="https://github.com/algorandfoundation/ARCs/pull/258">`ARC-56`</a> specification.

An Application is partially defined by it's [methods](./arc-0004.md) but further information about the Application should be available.  Other descriptive elements of an application may include it's State Schema, the original TEAL source programs, default method arguments, and custom data types.  This specification defines the descriptive elements of an Application that should be available to clients to provide useful information for an Application Client.

## Motivation

As more complex Applications are created and deployed, some consistent way to specify the details of the application and how to interact with it becomes more important.  A specification to allow a consistent and complete definition of an application will help developers attempting to integrate an application they've never worked with before.

## Specification
The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, “RECOMMENDED”, “MAY”, and “OPTIONAL” in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc822.txt">RFC 822</a>..

### Definitions

- [Application Specification](#application-specification): The object containing the elements describing the Application.
- [Source Specification](#source-specification): The object containing a description of the TEAL source programs that are evaluated when this Application is called.
- [Schema Specification](#schema-specification): The object containing a description of the schema required by the Application.
- [Bare Call Specification](#bare-call-specification): The object containing a map of on completion actions to allowable calls for bare methods
- [Hints Specification](#hints-specification): The object containing a map of method signatures to meta data about each method


### Application Specification

The Application Specification is composed of a number of elements that serve to fully describe the Application.

```ts
type AppSpec = {
  // embedded contract fields, see ARC-0004 for more
  contract: ARC4Contract;

  // the original teal source, containing annotations, base64 encoded
  source?: SourceSpec;
  // the schema this application requires/provides
  schema?: SchemaSpec;

  // supplemental information for calling bare methods
  bare_call_config?: CallConfigSpec;
  // supplemental information for calling ARC-0004 ABI methods
  hints: HintsSpec;
  // storage requirements
  state?: StateSpec;
}
```

### Source Specification

Contains the source TEAL files including comments and other annotations.

```ts
// Object containing the original TEAL source files
type SourceSpec = {
  // b64 encoded approval program
  approval: string;
  // b64 encoded clear state program
  clear: string;
}
```

### Schema Specification

The schema of an application is critical to know prior to creation since it is immutable after create. It also helps clients of the application understand the data that is available to be queried from off chain. Individual fields can be referenced from the [default argument](#default-argument) to provide input data to a given ABI method.

While some fields are possible to know ahead of time, others may be keyed dynamically. In both cases the data type being stored MUST be known and declared ahead of time.

```ts
// The complete schema for this application
type SchemaSpec = {
  local: Schema;
  global: Schema;
}

// Schema fields may be declared explicitly or reserved
type Schema = {
  declared: Record<string, DeclaredSchemaValueSpec>;
  reserved: Record<string, ReservedSchemaValueSpec>;
}

// Types supported for encoding/decoding
enum AVMType { uint64, bytes }
// string encoded datatype name defined in arc-4
type ABIType = string;

// Fields that have an explicit key
type DeclaredSchemaValueSpec = {
  type: AVMType | ABIType;
  key: string;
  descr: string;
}

// Fields that have an undetermined key
type ReservedSchemaValueSpec = {
  type: AVMType | ABIType;
  descr: string;
  max_keys: number;
}

```

### Bare call specification

Describes the supported OnComplete actions for bare calls on the contract.

```ts
// describes under what conditions an associated OnCompletion type can be used with a particular method
// NEVER: Never handle the specified on completion type
// CALL: Only handle the specified on completion type for application calls
// CREATE: Only handle the specified on completion type for application create calls
// ALL: Handle the specified on completion type for both create and normal application calls
type CallConfig = 'NEVER' | 'CALL' | 'CREATE' | 'ALL'

type CallConfigSpec = {
    // lists the supported CallConfig for each on completion type, if not specified a CallConfig of NEVER is assumed
    no_op?: CallConfig
    opt_in?: CallConfig
    close_out?: CallConfig
    update_application?: CallConfig
    delete_application?: CallConfig
}
```

### Hints specification

Contains supplemental information about [ARC-0004](./arc-0004.md) ABI methods, each record represents a single method in the [ARC-0004](./arc-0004.md) contract. The record key should be the corresponding ABI signature.

NOTE: Ideally this information would be part of the [ARC-0004](./arc-0004.md) ABI specification.

```ts
type HintSpec = {
  // indicates the method has no side-effects and can be call via dry-run/simulate
  read_only?: bool;
  // describes the structure of arguments, key represents the argument name
  structs?: Record<string, StructSpec>;
  // describes source of default values for arguments, key represents the argument name
  default_arguments?: Record<string, DefaultArgumentSpec>;
  // describes which OnCompletion types are supported
  call_config: CallConfigSpec;
}

// key represents the method signature for an ABI method defined in 'contracts'
type HintsSpec = Record<string, HintSpec>
```

#### Readonly Specification

Indicates the method has no side-effects and can be called via dry-run/simulate

NOTE: This property is made obsolete by [ARC-0022](./arc-0022.md) but is included as it is currently
used by existing reference implementations such as Beaker

#### Struct Specification

Each defined type is specified as an array of `StructElement`s.

The ABI encoding is exactly as if an ABI Tuple type defined the same element types in the same order.
It is important to encode the struct elements as an array since it preserves the order of fields which is critical to encoding/decoding the data properly.

```ts
// Type aliases for readability
type FieldName = string
// string encoded datatype name defined in ARC-0004
type ABIType = string

// Each field in the struct contains a name and ABI type
type StructElement = [FieldName, ABIType]
// Type aliases for readability
type ContractDefinedType = StructElement[]
type ContractDefinedTypeName = string;

// represents a input/output structure
type StructSpec = {
  name: ContractDefinedTypeName
  elements: ContractDefinedType
}
```

For example a `ContractDefinedType` that should provide an array of `StructElement`s

Given the PyTeal:
```py
from pyteal import abi

class Thing(abi.NamedTuple):
  addr: abi.Field[abi.address]
  balance: abi.Field[abi.Uint64]
```
the equivalent ABI type is `(address,uint64)` and an element in the TypeSpec is:

```js
{
// ...
"Thing":[["addr", "address"]["balance","uint64"]],
// ...
}
```

#### Default Argument

Defines how default argument values can be obtained. The `source` field defines how a default value is obtained, the `data` field
contains additional information based on the `source` value.

Valid values for `source` are:
* "constant" - `data` is the value to use
* "global-state" - `data` is the global state key.
* "local-state" - `data` is the local state key
* "abi-method" - `data` is a reference to the ABI method to call. Method should be read only and return a value of the appropriate type

Two scenarios where providing default arguments can be useful:

1. Providing a default value for optional arguments

2. Providing a value for required arguments such as foreign asset or application references without requiring the client to explicitly determine these values when calling the contract


```ts
// ARC-0004 ABI method definition
type ABIMethod = {};

type DefaultArgumentSpec = {
  // Where to look for the default arg value
  source: "constant" | "global-state" | "local-state" | "abi-method"
  // extra data to include when looking up the value
  data: string | bigint | number | ABIMethod
}
```

### State Specifications

Describes the total storage requirements for both global and local storage, this should include both declared and reserved described in SchemaSpec.

NOTE: If the Schema specification contained additional information such that the size could be calculated, then this specification would not be required.

```ts
type StateSchema = {
  // how many byte slices are required
  num_byte_slices: number
  // how many uints are required
  num_uints: number
}

type StateSpec = {
  // schema specification for global storage
  global: StateSchema
  // schema specification for local storage
  local: StateSchema
}
```

### Reference schema

A full JSON schema for application.json can be found in [here](../assets/arc-0032/application.schema.json).

## Rationale
The rationale fleshes out the specification by describing what motivated the design and why particular design decisions were made. It should describe alternate designs that were considered and related work, e.g. how the feature is supported in other languages.

## Backwards Compatibility
All ARCs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ARC must explain how the author proposes to deal with these incompatibilities. ARC submissions without a sufficient backwards compatibility treatise may be rejected outright.

## Test Cases
Test cases for an implementation are mandatory for ARCs that are affecting consensus changes.  If the test suite is too large to reasonably be included inline, then consider adding it as one or more files in `../assets/arc-####/`.

## Reference Implementation

`algokit-utils-py` and `algokit-utils-ts` both provide reference implementations for the specification structure and using the data in an `ApplicationClient`
`Beaker` provides a reference implementation for creating an application.json from a smart contract.

## Security Considerations
All ARCs must contain a section that discusses the security implications/considerations relevant to the proposed change. Include information that might be important for security discussions, surfaces risks and can be used throughout the life cycle of the proposal. E.g. include security-relevant design decisions, concerns, important discussions, implementation-specific guidance and pitfalls, an outline of threats and risks and how they are being addressed. ARC submissions missing the "Security Considerations" section will be rejected. An ARC cannot proceed to status "Final" without a Security Considerations discussion deemed sufficient by the reviewers.

## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
