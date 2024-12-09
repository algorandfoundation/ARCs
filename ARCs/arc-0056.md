---
arc: 56
title: Extended App Description
description: Adds information to the ABI JSON description
author: Joe Polny (@joe-p), Rob Moore (@robdmoore)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/258
status: Final
type: Standards Track
category: ARC
created: 2023-11-14
requires: 4, 22, 28
---

## Abstract

This ARC takes the existing JSON description of a contract as described in [ARC-4](./arc-0004.md) and adds more fields for the purpose of client interaction

## Motivation

The data provided by ARC-4 is missing a lot of critical information that clients should know when interacting with an app. This means ARC-4 is insufficient to generate type-safe clients that provide a superior developer experience.

On the other hand, [ARC-32](./arc-0032.md) provides the vast majority of useful information that can be used to <a href="https://github.com/algorandfoundation/algokit-cli/blob/main/docs/features/generate.md#1-typed-clients">generate typed clients</a>, but requires a separate JSON file on top of the ARC-4 json file, which adds extra complexity and cognitive overhead.

## Specification

### Contract Interface

Every application is described via the following interface which is an extension of the `Contract` interface described in [ARC-4](./arc-0004.md).

```ts
/** Describes the entire contract. This interface is an extension of the interface described in ARC-4 */
interface Contract {
  /** The ARCs used and/or supported by this contract. All contracts implicitly support ARC4 and ARC56 */
  arcs: number[];
  /** A user-friendly name for the contract */
  name: string;
  /** Optional, user-friendly description for the interface */
  desc?: string;
  /**
   * Optional object listing the contract instances across different networks.
   * The key is the base64 genesis hash of the network, and the value contains
   * information about the deployed contract in the network indicated by the
   * key. A key containing the human-readable name of the network MAY be
   * included, but the corresponding genesis hash key MUST also be defined
   */
  networks?: {
    [network: string]: {
      /** The app ID of the deployed contract in this network */
      appID: number;
    };
  };
  /** Named structs used by the application. Each struct field appears in the same order as ABI encoding. */
  structs: { [structName: StructName]: StructField[] };
  /** All of the methods that the contract implements */
  methods: Method[];
  state: {
    /** Defines the values that should be used for GlobalNumUint, GlobalNumByteSlice, LocalNumUint, and LocalNumByteSlice when creating the application  */
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
    /** Mapping of human-readable names to StorageKey objects */
    keys: {
      global: { [name: string]: StorageKey };
      local: { [name: string]: StorageKey };
      box: { [name: string]: StorageKey };
    };
    /** Mapping of human-readable names to StorageMap objects */
    maps: {
      global: { [name: string]: StorageMap };
      local: { [name: string]: StorageMap };
      box: { [name: string]: StorageMap };
    };
  };
  /** Supported bare actions for the contract. An action is a combination of call/create and an OnComplete */
  bareActions: {
    /** OnCompletes this method allows when appID === 0 */
    create: ("NoOp" | "OptIn" | "DeleteApplication")[];
    /** OnCompletes this method allows when appID !== 0 */
    call: (
      | "NoOp"
      | "OptIn"
      | "CloseOut"
      | "UpdateApplication"
      | "DeleteApplication"
    )[];
  };
  /** Information about the TEAL programs */
  sourceInfo?: {
    /** Approval program information */
    approval: ProgramSourceInfo;
    /** Clear program information */
    clear: ProgramSourceInfo;
  };
  /** The pre-compiled TEAL that may contain template variables. MUST be omitted if included as part of ARC23 */
  source?: {
    /** The approval program */
    approval: string;
    /** The clear program */
    clear: string;
  };
  /** The compiled bytecode for the application. MUST be omitted if included as part of ARC23 */
  byteCode?: {
    /** The approval program */
    approval: string;
    /** The clear program */
    clear: string;
  };
  /** Information used to get the given byteCode and/or PC values in sourceInfo. MUST be given if byteCode or PC values are present */
  compilerInfo?: {
    /** The name of the compiler */
    compiler: "algod" | "puya";
    /** Compiler version information */
    compilerVersion: {
      major: number;
      minor: number;
      patch: number;
      commitHash?: string;
    };
  };
  /** ARC-28 events that MAY be emitted by this contract */
  events?: Array<Event>;
  /** A mapping of template variable names as they appear in the TEAL (not including TMPL_ prefix) to their respective types and values (if applicable) */
  templateVariables?: {
    [name: string]: {
      /** The type of the template variable */
      type: ABIType | AVMType | StructName;
      /** If given, the base64 encoded value used for the given app/program */
      value?: string;
    };
  };
  /** The scratch variables used during runtime */
  scratchVariables?: {
    [name: string]: {
      slot: number;
      type: ABIType | AVMType | StructName;
    };
  };
}
```

### Method Interface

Every method in the contract is described via a `Method` interface. This interface is an extension of the one defined in [ARC-4](./arc-0004.md).

```ts
/** Describes a method in the contract. This interface is an extension of the interface described in ARC-4 */
interface Method {
  /** The name of the method */
  name: string;
  /** Optional, user-friendly description for the method */
  desc?: string;
  /** The arguments of the method, in order */
  args: Array<{
    /** The type of the argument. The `struct` field should also be checked to determine if this arg is a struct. */
    type: ABIType;
    /** If the type is a struct, the name of the struct */
    struct?: StructName;
    /** Optional, user-friendly name for the argument */
    name?: string;
    /** Optional, user-friendly description for the argument */
    desc?: string;
    /** The default value that clients should use. */
    defaultValue?: {
      /** Where the default value is coming from
       * - box: The data key signifies the box key to read the value from
       * - global: The data key signifies the global state key to read the value from
       * - local: The data key signifies the local state key to read the value from (for the sender)
       * - literal: the value is a literal and should be passed directly as the argument
       * - method: The utf8 signature of the method in this contract to call to get the default value. If the method has arguments, they all must have default values. The method **MUST** be readonly so simulate can be used to get the default value.
       */
      source: "box" | "global" | "local" | "literal" | "method";
      /** Base64 encoded bytes, base64 ARC4 encoded uint64, or UTF-8 method selector */
      data: string;
      /** How the data is encoded. This is the encoding for the data provided here, not the arg type. Undefined if the data is method selector */
      type?: ABIType | AVMType;
    };
  }>;
  /** Information about the method's return value */
  returns: {
    /** The type of the return value, or "void" to indicate no return value. The `struct` field should also be checked to determine if this return value is a struct. */
    type: ABIType;
    /** If the type is a struct, the name of the struct */
    struct?: StructName;
    /** Optional, user-friendly description for the return value */
    desc?: string;
  };
  /** an action is a combination of call/create and an OnComplete */
  actions: {
    /** OnCompletes this method allows when appID === 0 */
    create: ("NoOp" | "OptIn" | "DeleteApplication")[];
    /** OnCompletes this method allows when appID !== 0 */
    call: (
      | "NoOp"
      | "OptIn"
      | "CloseOut"
      | "UpdateApplication"
      | "DeleteApplication"
    )[];
  };
  /** If this method does not write anything to the ledger (ARC-22) */
  readonly?: boolean;
  /** ARC-28 events that MAY be emitted by this method */
  events?: Array<Event>;
  /** Information that clients can use when calling the method */
  recommendations?: {
    /** The number of inner transactions the caller should cover the fees for */
    innerTransactionCount?: number;
    /** Recommended box references to include */
    boxes?: {
      /** The app ID for the box */
      app?: number;
      /** The base64 encoded box key */
      key: string;
      /** The number of bytes being read from the box */
      readBytes: number;
      /** The number of bytes being written to the box */
      writeBytes: number;
    };
    /** Recommended foreign accounts */
    accounts?: string[];
    /** Recommended foreign apps */
    apps?: number[];
    /** Recommended foreign assets */
    assets?: number[];
  };
}
```

### Event Interface

[ARC-28](./arc-0028.md) events are described using an extension of the original interface described in the ARC, with the addition of an optional struct field for arguments

```ts
interface Event {
  /** The name of the event */
  name: string;
  /** Optional, user-friendly description for the event */
  desc?: string;
  /** The arguments of the event, in order */
  args: Array<{
    /** The type of the argument. The `struct` field should also be checked to determine if this arg is a struct. */
    type: ABIType;
    /** Optional, user-friendly name for the argument */
    name?: string;
    /** Optional, user-friendly description for the argument */
    desc?: string;
    /** If the type is a struct, the name of the struct */
    struct?: StructName;
  }>;
}
```

### Type Interfaces

The types defined in [ARC-4](./arc-0004.md) may not fully described the best way to use the ABI values as intended by the contract developers. These type interfaces are intended to supplement ABI types so clients can interact with the contract as intended.

```ts
/** An ABI-encoded type */
type ABIType = string;

/** The name of a defined struct */
type StructName = string;

/** Raw byteslice without the length prefixed that is specified in ARC-4 */
type AVMBytes = "AVMBytes";

/** A utf-8 string without the length prefix that is specified in ARC-4 */
type AVMString = "AVMString";

/** A 64-bit unsigned integer */
type AVMUint64 = "AVMUint64";

/** A native AVM type */
type AVMType = AVMBytes | AVMString | AVMUint64;

/** Information about a single field in a struct */
interface StructField {
  /** The name of the struct field */
  name: string;
  /** The type of the struct field's value */
  type: ABIType | StructName | StructField[];
}
```

### Storage Interfaces

These interfaces properly describe how app storage is access within the contract

```ts
/** Describes a single key in app storage */
interface StorageKey {
  /** Description of what this storage key holds */
  desc?: string;
  /** The type of the key */
  keyType: ABIType | AVMType | StructName;
  /** The type of the value */
  valueType: ABIType | AVMType | StructName;
  /** The bytes of the key encoded as base64 */
  key: string;
}

/** Describes a mapping of key-value pairs in storage */
interface StorageMap {
  /** Description of what the key-value pairs in this mapping hold */
  desc?: string;
  /** The type of the keys in the map */
  keyType: ABIType | AVMType | StructName;
  /** The type of the values in the map */
  valueType: ABIType | AVMType | StructName;
  /** The base64-encoded prefix of the map keys*/
  prefix?: string;
}
```

### SourceInfo Interface

These interfaces give clients more information about the contract's source code.

```ts
interface ProgramSourceInfo {
  /** The source information for the program */
  sourceInfo: SourceInfo[];
  /** How the program counter offset is calculated
   * - none: The pc values in sourceInfo are not offset
   * - cblocks: The pc values in sourceInfo are offset by the PC of the first op following the last cblock at the top of the program
   */
  pcOffsetMethod: "none" | "cblocks";
}

interface SourceInfo {
  /** The program counter value(s). Could be offset if pcOffsetMethod is not "none" */
  pc: Array<number>;
  /** A human-readable string that describes the error when the program fails at the given PC */
  errorMessage?: string;
  /** The TEAL line number that corresponds to the given PC. RECOMMENDED to be used for development purposes, but not required for clients */
  teal?: number;
  /** The original source file and line number that corresponds to the given PC. RECOMMENDED to be used for development purposes, but not required for clients */
  source?: string;
}
```

### Template Variables

Template variables are variables in the TEAL that should be substitued prior to compilation. The usage of the variable **MUST** appear in the TEAL starting with `TMPL_`. Template variables **MUST** be an argument to either `bytecblock` or `intcblock`. If a program has template variables, `bytecblock` and `intcblock` **MUST** be the first two opcodes in the program (unless one is not used).

#### Example

```js
#pragma version 10
bytecblock 0xdeadbeef TMPL_FOO
intcblock 0x12345678 TMPL_BAR
```

### Dynamic Template Variables

When a program has a template variable with a dynamic length, the `pcOffsetMethod` in `ProgramSourceInfo` **MUST** be `cblocks`. The `pc` value in each `SourceInfo` **MUST** be the pc determined at compilation minus the last `pc` value of the last `cblock` at compilation.

When a client is leveraging a source map with `cblocks` as the `pcOffsetMethod`, it **MUST** determine the `pc` value by parsing the bytecode to get the PC value of the first op following the last `cblock` at the top of the program. See the reference implementation section for an example of how to do this.

## Rationale

ARC-32 essentially addresses the same problem, but it requires the generation of two separate JSON files and the ARC-32 JSON file contains the ARC-4 JSON file within it (redundant information). The goal of this ARC is to create one JSON schema that is backwards compatible with ARC-4 clients, but contains the relevant information needed to automatically generate comprehensive client experiences.

### State

Describes all of the state that MAY exist in the app and how one should decode values. The schema provides the required schema when creating the app.

### Named Structs

It is common for high-level languages to support named structs, which gives names to the indexes of elements in an ABI tuple. The same structs should be useable on the client-side just as they are used in the contract.

### Action

This is one of the biggest deviation from ARC-32, but provides a much simpler interface to describe and understand what any given method can do.

## Backwards Compatibility

The JSON schema defined in this ARC should be compatible with all ARC-4 clients, provided they don't do any strict schema checking for extraneous fields.

## Test Cases

NA

## Reference Implementation

### Calculating cblock Offsets

Below is an example of how to determine the TEAL/source line for a PC from an algod error message when the `pcOffsetMethod` is `cblocks`.

```ts
/** An ARC56 JSON file */
import arc56Json from "./arc56.json";

/** The bytecblock opcode */
const BYTE_CBLOCK = 38;
/** The intcblock opcode */
const INT_CBLOCK = 32;

/**
 * Get the offset of the last constant block at the beginning of the program
 * This value is used to calculate the program counter for an ARC56 program that has a pcOffsetMethod of "cblocks"
 *
 * @param program The program to parse
 * @returns The PC value of the opcode after the last constant block
 */
function getConstantBlockOffset(program: Uint8Array) {
  const bytes = [...program];

  const programSize = bytes.length;
  bytes.shift(); // remove version

  /** The PC of the opcode after the bytecblock */
  let bytecblockOffset: number | undefined;

  /** The PC of the opcode after the intcblock */
  let intcblockOffset: number | undefined;

  while (bytes.length > 0) {
    /** The current byte from the beginning of the byte array */
    const byte = bytes.shift()!;

    // If the byte is a constant block...
    if (byte === BYTE_CBLOCK || byte === INT_CBLOCK) {
      const isBytecblock = byte === BYTE_CBLOCK;

      /** The byte following the opcode is the number of values in the constant block */
      const valuesRemaining = bytes.shift()!;

      // Iterate over all the values in the constant block
      for (let i = 0; i < valuesRemaining; i++) {
        if (isBytecblock) {
          /** The byte following the opcode is the length of the next element */
          const length = bytes.shift()!;
          bytes.splice(0, length);
        } else {
          // intcblock is a uvarint, so we need to keep reading until we find the end (MSB is not set)
          while ((bytes.shift()! & 0x80) !== 0) {
            // Do nothing...
          }
        }
      }

      if (isBytecblock) bytecblockOffset = programSize - bytes.length - 1;
      else intcblockOffset = programSize - bytes.length - 1;

      if (bytes[0] !== BYTE_CBLOCK && bytes[0] !== INT_CBLOCK) {
        // if the next opcode isn't a constant block, we're done
        break;
      }
    }
  }

  return Math.max(bytecblockOffset ?? 0, intcblockOffset ?? 0);
}

/** The error message from algod */
const algodError =
  "Network request error. Received status 400 (Bad Request): TransactionPool.Remember: transaction ZR2LAFLRQYFZFV6WVKAPH6CANJMIBLLH5WRTSWT5CJHFVMF4UIFA: logic eval error: assert failed pc=162. Details: app=11927, pc=162, opcodes=log; intc_0 // 0; assert";

/** The PC of the error */
const pc = Number(algodError.match(/pc=(\d+)/)![1]);

// Parse the ARC56 JSON to determine if the PC values are offset by the constant blocks
if (arc56Json.sourceInfo.approval.pcOffsetMethod === "cblocks") {
  /** The program can either be cached locally OR retrieved via the algod API */
  const program = new Uint8Array([
    10, 32, 3, 0, 1, 6, 38, 3, 64, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48,
    48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48,
    48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48,
    48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 32, 48, 48, 48,
    48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48,
    48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 3, 102, 111, 111, 40, 41, 34, 42,
    49, 24, 20, 129, 6, 11, 49, 25, 8, 141, 12, 0, 85, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 71, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 136, 0, 3, 129, 1, 67, 138, 0,
    0, 42, 176, 34, 68, 137, 136, 0, 3, 129, 1, 67, 138, 0, 0, 42, 40, 41, 132,
    137, 136, 0, 3, 129, 1, 67, 138, 0, 0, 0, 137, 128, 4, 21, 31, 124, 117,
    136, 0, 13, 73, 21, 22, 87, 6, 2, 76, 80, 80, 176, 129, 1, 67, 138, 0, 1,
    34, 22, 137, 129, 1, 67, 128, 4, 184, 68, 123, 54, 54, 26, 0, 142, 1, 255,
    240, 0, 128, 4, 154, 113, 210, 180, 128, 4, 223, 77, 92, 59, 128, 4, 61,
    135, 13, 135, 128, 4, 188, 11, 23, 6, 54, 26, 0, 142, 4, 255, 135, 255, 149,
    255, 163, 255, 174, 0,
  ]);

  /** Get the offset of the last constant block */
  const offset = getConstantBlockOffset(program);

  /** Find the source info object that corresponds to the error's PC */
  const sourceInfoObject = arc56Json.sourceInfo.approval.sourceInfo.find((s) =>
    s.pc.includes(pc - offset)
  )!;

  /** Get the TEAL line and source line that corresponds to the error */
  console.log(
    `Error at PC ${pc} corresponds to TEAL line ${sourceInfoObject.teal} and source line ${sourceInfoObject.source}`
  );
}
```

## Security Considerations

The type values used in methods **MUST** be correct, because if they were not then the method would not be callable. For state, however, it is possible to have an incorrect type encoding defined. Any significant security concern from this possibility is not immediately evident, but it is worth considering.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.

```

```
