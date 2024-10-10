---
arc: 28
title: Algorand Event Log Spec
description: A methodology for structured logging by Algorand dapps.
author: Dan Burton (@DanBurton)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/144
status: Final
type: Standards Track
category: ARC
sub-category: Application
created: 2022-07-18
requires: 4
---

# Algorand Event Log Spec

## Abstract

Algorand dapps can use the <a href="https://developer.algorand.org/docs/get-details/dapps/avm/teal/opcodes/#log">`log`</a>  primitive to attach information about an application call. This ARC proposes the concept of Events, which are merely a way in which data contained in these logs may be categorized and structured.

In short: to emit an Event, a dapp calls `log` with ABI formatting of the log data, and a 4-byte prefix to indicate which Event it is.

## Specification

Each kind of Event emitted by a given dapp has a unique 4-byte identifier. This identifier is derived from its name and the structure of its contents, like so:

### Event Signature

An Event Signature is a utf8 string, comprised of: the name of the event, followed by an open paren, followed by the comma-separated names of the data types contained in the event (Types supported are the same as in [ARC-4](./arc-0004.md#types)), followed by a close paren. This follows naming conventions similar to ABI signatures, but does not include the return type.

### Deriving the 4-byte prefix from the Event Signature

To derive the 4-byte prefix from the Event Signature, perform the `sha512/256` hash algorithm on the signature, and select the first 4 bytes of the result.

This is the same process that is used by the [ABI Method Selector ](./arc-0004.md#method-selector) as specified in ARC-4.

### Argument Encoding

The arguments to a tuple **MUST** be encoded as if they were a single [ARC-4](./arc-0004.md) tuple (opposed to concatenating the encoded values together). For example, an event signature `foo(string,string)` would contain the 4-byte prefix and a `(string,string)` encoded byteslice.

### ARC-4 Extension

#### Event

An event is represented as follow:

```typescript
interface Event {
  /** The name of the event */
  name: string;
  /** Optional, user-friendly description for the event */
  desc?: string;
  /** The arguments of the event, in order */
  args: Array<{
    /** The type of the argument */
    type: string;
    /** Optional, user-friendly name for the argument */
    name?: string;
    /** Optional, user-friendly description for the argument */
    desc?: string;
  }>;
}
```

#### Method

This ARC extends ARC-4 by adding an array events of type `Event[]` to the `Method` interface. Concretely, this give the following extended Method interface:

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
  /** All of the events that the method use */
  events: Event[];
  /** Information about the method's return value */
  returns: {
    /** The type of the return value, or "void" to indicate no return value. */
    type: string;
    /** Optional, user-friendly description for the return value */
    desc?: string;
  };
}
```

#### Contract
> Even if events are already inside `Method`, the contract **MUST** provide an array of `Events` to improve readability.

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
  /** All of the events that the contract contains */
  events: Event[];
}
```

## Rationale

Event logging allows a dapp to convey useful information about the things it is doing.  Well-designed Event logs allow observers to more easily interpret the history of interactions with the dapp. A structured approach to Event logging could also allow for indexers to more efficiently store and serve queryable data exposed by the dapp about its history.

## Reference Implementation

### Sample interpretation of Event log data

An exchange dapp might emit a `Swapped` event with two `uint64` values representing quantities of currency swapped. The event signature would be: `Swapped(uint64,uint64)`.

Suppose that dapp emits the following log data (seen here as base64 encoded): `HMvZJQAAAAAAAAAqAAAAAAAAAGQ=`.

Suppose also that the dapp developers have declared that it follows this spec for Events, and have published the signature `Swapped(uint64,uint64)`.

We can attempt to parse this log data to see if it is one of these events, as follows. (This example is written in JavaScript.)

First, we can determine the expected 4-byte prefix by following the spec above:

```js
> { sha512_256 } = require('js-sha512')
> sig = 'Swapped(uint64,uint64)'
'Swapped(uint64,uint64)'
> hash = sha512_256(sig)
'1ccbd9254b9f2e1caf190c6530a8d435fc788b69954078ab937db9b5540d9567'
> prefix = hash.slice(0,8) // 8 nibbles = 4 bytes
'1ccbd925'
```

Next, we can inspect the data to see if it matches the expected format:
4 bytes for the prefix, 8 bytes for the first uint64, and 8 bytes for the next.

```js
> b = Buffer.from('HMvZJQAAAAAAAAAqAAAAAAAAAGQ=', 'base64')
<Buffer 1c cb d9 25 00 00 00 00 00 00 00 2a 00 00 00 00 00 00 00 64>
> b.slice(0,4).toString('hex')
'1ccbd925'
> b.slice(4, 12)
<Buffer 00 00 00 00 00 00 00 2a>
> b.slice(12,20)
<Buffer 00 00 00 00 00 00 00 64>
```

We see that the 4-byte prefix matches the signature for `Swapped(uint64,uint64)`, and that the rest of the data can be interpreted using the types declared for that signature. We interpret the above Event data to be: `Swapped(0x2a,0x64)`, meaning `Swapped(42,100)`.

## Security Considerations

As specify in ARC-4, methods which have a `return` value MUST NOT emit an event after they log their `return` value.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
