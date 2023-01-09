---
arc: 23
title: Append the hash of the contract description to the compiled application's bytes
status: Draft
---

# Append the hash of the contract description to the compiled application bytes

## Abstract

The following document introduces a convention for appending informations to the compiled application's bytes (application, ABI as described in [ARC-4](https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0004.md), ..).

The goal of this convention is to standardize the process of verifying and interacting with smart contracts.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in [RFC-2119](https://www.ietf.org/rfc/rfc2119.txt).

> Comments like this are non-normative.

### Encoding

The encoded byte is the `arc23: ` followed by the a <a href="https://github.com/multiformats/cid">CID</a> representing a folder of files, it can be accessed using <a href="https://docs.ipfs.tech/">IPFS</a>:
```JSON
{"arc23": CID}
```

### Appending

The encoded byte are appended to the compiled application as a [bytecblock](https://developer.algorand.org/docs/get-details/dapps/avm/teal/opcodes/#bytecblock-bytes) containing 1 byte constant which is the encoded object.
> The reason to use `bytecblock` is that adding a `bytecblock` opcode at the end of a TEAL program does not change the semantics of the program, as long as: opcodes are properly aligned, there is no jump after the last position (that would make the program fail without `bytecblock`), and there is enough space left to add the opcode.
The size of the compiled application + the bytecblock should not exceed the maximum size of a compiled application according to the latest consensus parameters supported by the compiler.

## Rationale

By appending the ipfs cid of the folder containing informations about the application, any user with access to the blockchain could easily verify the Application, the ABI of the application and interact with it.

## Reference Implementation
The following codes are not audited and are only here for information purposes.
It **MUST** not be used in production.

Here is an example of a python script that can generate the hash and append it to the compiled application, according to arc-23:
[main.py](../assets/arc-0023/main.py).
The output is here [output.txt](../assets/arc-0023/output.txt)

An Folder containing: 
- example of the application [Application.py](../assets/arc-0023/Folder/Application.py).
- example of the Abi that follow [ARC-4](arc-0004.md) [Abi.json](../assets/arc-0023/Folder/Abi.json).


Files are accessible through ipfs commmand:
```console
$ ipfs cat bafybeihdzhfifv46p27ee6weqq6h7ydttzwni7gvdogr3mg6rwscdyzici/Abi.json
$ ipfs cat bafybeihdzhfifv46p27ee6weqq6h7ydttzwni7gvdogr3mg6rwscdyzici/Application.py
```
## Security Considerations


## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.