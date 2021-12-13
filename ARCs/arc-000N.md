---
arc: <to be assigned>
title: Smart Signature Template Validation
description: Description and implementation of a method to allow 
author: Ben Guidarelli<@barnjamin>
discussions-to: TODO 
status: Draft
type: Standards Track
category:  ARC>
created: December 13, 2021
requires (*optional): ARC-0004 
---

## Abstract

This ARC describes the processes for validating a Smart Signature Template. Using the compiled bytecode of the Smart Signature and some additional metadata, a Smart Contract is able to validate that a given Address is an instance of the Template.

## Motivation

A Smart Signature (SmartSig) has some advantages compared to a Smart Contract. For example it may use up to 20k ops per invocation (compared to 700 in a smart contract). It may also be used as an account that can opt into an application and obeys pre-defined rules (e.g. doesn't close out unexpectedly). 

If the SmartSig is constant (i.e. not a Template with potentially different variables) the Address of the Sender may be used and compared to some predefined address.

However, if a Template is used (e.g. in the case of a unique SmartSig per Account), there is no straight forward method to validate that the SmartSig adheres to the logic expected. 

## Specification

The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, “RECOMMENDED”, “MAY”, and “OPTIONAL” in this document are to be interpreted as described in RFC 2119.

### Terms

- Smart Signature Template - A Smart Signature source file that contains template variables to be populated for a specific instance
- Template Variable - A variable to be substituted later, specifying the name of the variable. The Template Variable must start with `TMPL_` and the type is inferred from the op used (e.g. `pushbytes TMPL_ASSET_NAME` => `bytes` )
- Blanked Template - An assembled Smart Signature Template with the Template Variables set to their 0 value
- Assembler - The software that converts a TEAL source file to a binary file representing the bytecode of the contract.


### Description 

In order for a Smart Contract to perform validation on a Template, two pieces of information should be known ahead of time:

1. The positions of the Template Variables within the compiled contract
2. The bytecode of the compiled SmartSig with all Template Variables set to the 0 value (0 for uint64, "" for bytes)

> Note: Integers in the assembled contract are encoded as [`uvarint`](https://www.sqlite.org/src4/doc/trunk/www/varint.wiki)

With this information, the Smart Contract performing the validation may accept a number of arguments corressponding to the order and type of the Template Variables in the SmartSig. It may then encode the variables properly (uint64=>varuint) and insert them in the bytecode at its start position.

### Preparation 

During assembly of a SmartSig Template, the assembler should note the bytecode position and type of any Template Variables. These value of these Template Variables should be set to the zero value for it's type. 

The output of the assembly step should include 2 artifacts:

1. A JSON file containing the list of template variables 
ex:
```js
{
    "template_variables":[
        {
            "name":"TMPL_ASSET_ID",
            "type": 1,
            "position":12
        },
        {
            "name":"TMPL_CREATOR_ADDR",
            "type": 2,
            "position":25
        },
    ],
    ...
}
```
2. A binary file containing the bytecode of the SmartSig Template

These two artifact contain enough information to produce an ABI style method, accepting as arguments all the template variables specified in the Template. It can use this information to produce the address of the expected SmartSig.

The ABI method signature **MUST** specify the types corresponding to those in the `template_variables` array. They **MAY** use type aliases as specified in [ARC-0004](TODO) but the encoding should use the two stack types (uint64/bytes).
`populate(tmpl_var_1_type, tmpl_var_2_type, ...)byte[]`

### Populate

This section contains some pseudocode to be implemented by a contract author or provided as part of this ARC

> Note: For each variable inserted, the position in the bytcode will need to be offset by the length of the variable added in bytes.

#### pseudocode:
```
encode(val, typ):
    if typ == int:
        return encode_uint64(val)
    else:
        return encode_bytes(val)

encode_bytes(val):
    return concat(encode_uint64(len(val)), val)

encode_uint64(val):
    return uvarint(val)

populate(uint64,address)byte[]:
    bytecode = []byte("0xDEADBEEF...") # the bytecode of the compiled, blanked, template
    positions = [12, 25] # index in the bytecode of the template varaiables, taken from the assemble step
    offset = 0

    for x in range(len(positions)/8):
        encoded = encode_bytes(arg[x])

        bytecode = concat(
            extract(bytecode, 0, position[x]+offset),
            encoded,
            substr(bytecode, position[x]+offset, len(bytecode))
        )

        offset += len(encoded)

    return bytecode
```


### Validate

With the Populated bytecode, the contract may call Sha512_265("Program"||populated_bytecode) to produce the expected Address.  If the Address in the transaction matches the address produced, the contract is considered valid we know it is an instance of the Template. 

## Rationale

The design for this ARC was inspired by the need to safely allow additional storage for a smart contract or offload compute from a Smart Contract to a Smart Signature.

## Backwards Compatibility

There are no backwards incompatible changes for this specification.

## Test Cases

// TODO: ??

## Reference Implementation

// TODO: https://github.com/barnjamin/rareaf/blob/price-tokens/src/contracts/python/tcv.py

## Security Considerations

Risks include, incorrect population logic or subtle issues with the bytecode used or misunderstanding of how this _should_ be used.

Any of those may result in incorrectly approving a Smart Signature Template and catestrophic loss of funds is possible. 

// TODO:

## Copyright
Copyright and related rights waived via [CC0](https://creativecommons.org/publicdomain/zero/1.0/).