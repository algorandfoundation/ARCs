---
arc: <to be assigned>
title: Method Reference Discovery
description: A standardized way contracts can reveal which references will be needed to call a specific method
author: Joe Polny (@joe-p)
discussions-to: <URL>
status: Draft
type: Standards Track
category: ARC
created: 2023-07-28
requires (*optional): 4
--- 

## Abstract
A contract caller needs to know which resources a contract needs to be availible before calling a method. This ARC proposes a standard way to make that information discoverable.

## Motivation
As of the time of this ARC, it can be hard to know which resources a caller needs to make availible when calling an application. The current solution typically involves proprietary SDKs which can make app usage and composability difficult.

### Simulation Consideration

It should be noted that it is currently planned for the algod simulate endpoint to allow readonly execution of a method without providing in references. The response to this endpoint will return the necessary resources. Once this functionality is availible, this ARC will no longer be needed. The primary intent of this ARC is to serve as an intermediate solution.

## Specification
If an application has a method and the contract wants to make the required resources for calling the method discoverable, it **MUST** implement a readonly method with the same exact signature with an `arcXXXX_` prefix and a return type of `(address[],uint64[],uint64[],(uint64,byte[]))`. 

The return value corresponds to arrays containing the required account, application, asset, and box references respectively.

The argument values provided to this method when called **SHOULD** match the arguments passed to the method for which the callers wants to know the required resources for.

The ARCXXXX method **MUST** be readonly.

## Rationale
The provided method will provide all of the references needed to call the application, which was previously not possible in a standardized way.

## Backwards Compatibility
N/A

## Test Cases
N/A

## Reference Implementation

### ARC200 example

In this example, let's say `arc200_totalSupply()uint256` requires two box refences `"baseSupply"` and `"supplyMultiplier"`, each encoded as `byte[]`.

#### Method Signature

The ARCXXXX method signature would be `arcXXXX_arc200_totalSupply()(address[],uint64[],uint64[],(uint64,byte[]))`

#### Return Value

The ARCXXXX method would return the following value: `[],[],[],[[0, "baseSupply"], [0, "supplyMultiplier"]]`

## Security Considerations
N/A

## Copyright
Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.