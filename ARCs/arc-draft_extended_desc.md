---
arc: <to be assigned>
title: Extended App Description
description: Adds more information to the ARC4 JSON description
author: Joe Polny (@joe-p)
discussions-to: <URL>
status: Draft
type: Standards Track
category: ARC
created: 2023-11-14
requires: 4
---


## Abstract
This ARC takes the existing JSON description of a contract as described in ARC4 and adds more fields for the purpose of client interaction

## Motivation
The data provided by [ARC4](./arc-0004.md) is missing a lot of critical information that clients should know when interacting with an app. This means ARC4 is insufficient to generate type-safe clients that provide a superior developer experience.

On the other hand, [ARC32](https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0032.md) provides the vast majority of useful information that can be used to [generate typed clients](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/features/generate.md#1-typed-clients), but requires a separate JSON file on top of the ARC4 json file, which adds extra complexity and cognitive overhead.

## Specification
```ts
/** Mapping of named structs to the ABI type of their fields */
interface StructFields {
  [name: string]: string | StructFields;
}

/** Describes a single key in app storage */
interface StorageKey {
  /** The type of the key. Can be ABI type or named struct */
  keyType: string;
  /** The type of the value. Can be ABI type or named struct */
  valueType: string;
  /** The key itself, as a byte array */
  key: number[];
}

interface StorageMap {
  /** The type of the key. Can be ABI type or named struct */
  keyType: string;
  /** The type of the value. Can be ABI type or named struct */
  valueType: string;
  /** The prefix of the map, as a string */
  prefix: string;
}

interface Method {
  /** The name of the method */
  name: string;
  /** Optional, user-friendly description for the method */
  desc?: string;
  /** The arguments of the method, in order */
  args: Array<{
    /** The type of the argument */
    type: string;
    /** If the type is a struct, the name of the struct */
    struct?: string;
    /** Optional, user-friendly name for the argument */
    name?: string;
    /** Optional, user-friendly description for the argument */
    desc?: string;
  }>;
  /** Information about the method's return value */
  returns: {
    /** The type of the return value, or "void" to indicate no return value. */
    type: string;
    /** If the type is a struct, the name of the struct */
    struct?: string;
    /** Optional, user-friendly description for the return value */
    desc?: string;
  };
  /** an action is a combination of call/create and an OnComplete */
  actions: {
    /** OnCompeltes this method allows when appID === 0 */
    create: ('NoOp' | 'OptIn' | 'DeleteApplication')[];
    /** OnCompeltes this method allows when appID !== 0 */
    call: ('NoOp' | 'OptIn' | 'CloseOut' | 'ClearState' | 'UpdateApplication' | 'DeleteApplication')[];
  };
  readonly: boolean;
}

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
    };
  };
  /** Named structs use by the application */
  structs: StructFields;
  /** All of the methods that the contract implements */
  methods: Method[];
  state: {
    schema: {
      global: {
        ints: number;
        bytes: number;
      };
      local: {
        ints: number;
        bytes: number;
      };
    };
    /** Describes single key-value pairs in the application's state */
    keys: {
      global: StorageKey[];
      local: StorageKey[];
      box: StorageKey[];
    };
    /** Describes key-value maps in the application's state */
    maps: {
      global: StorageMap[];
      local: StorageMap[];
      box: StorageMap[];
    };
  };
}
```

## Rationale
ARC32 essentially addresses the same problem, but it requires the generation of two separate JSON files and the ARC32 JSON file contains the ARC4 JSON file within it (redundant information). The goal of this ARC is to create one JSON schema that is backwards compatible with ARC4 clients, but contains the relevant information needed to automatically generate comprehensive client experiences.

### State

Describes all of the state that MAY exist in the app and how one should decode values. The schema provides the required schema when creating the app. 

### Named Structs

It is common for high-level languages to support named structs, which gives names to the indexes of elements in an ABI tuple. The same structs should be useable on the client-side just as they are used in the contract.

### Action

This is one of the biggest deviation from ARC32, but provides a much simpler interface to describe and understand what any given method can do. 

## Backwards Compatibility
The JSON schema defined in this ARC should be compatible with all ARC4 clients, provided they don't do any strict schema checking for extraneous fields.

## Test Cases
NA

## Reference Implementation
TODO

## Security Considerations
The type values used in methods are guranteed to be correct, because if they were not then the method would not be callable. For state, however, it is possible to have an incorrect type encoding defined. Any significant security concern from this possibility is not immediately evident, but it is worth considering.  

## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
