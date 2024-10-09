# ARC-60 Reference Implementation

## Overview

This is a reference implementation of the ARC-60 specification. It is written in TypeScript and uses the Jest testing framework.
The test suite shows the different use cases of the ARC-60 specification.

## Instructions

```bash
$ yarn
$ yarn test
```

## Sample Output

```bash
 PASS  ./arc60wallet.api.spec.ts
  ARC60 TEST SUITE
    rawSign
      ✓ (OK) should sign data correctly (22 ms)
      ✓ (FAILS) should throw error for shorter incorrect length seed (1 ms)
      ✓ (FAILS) should throw error for longer incorrect length seed (1 ms)
    getPublicKey
      ✓ (OK) should return the correct public key
      ✓ (FAILS) should throw error for shorter incorrect length seed
      ✓ (FAILS) should throw error for longer incorrect length seed (1 ms)
    SCOPE == INVALID
      ✓ (FAILS) Tries to sign with invalid scope (18 ms)
    AUTH sign request
      ✓ (OK) Signing AUTH requests (3 ms)
      ✓ (FAILS) Tries to sign with bad json (2 ms)
      ✓ (FAILS) Tries to sign with bad json schema (1 ms)
      ✓ (FAILS) Is missing domain (1 ms)
      ✓ (FAILS) Is missing authenticationData (1 ms)
    Invalid or Unkown Signer
      ✓ (FAILS) Tries to sign with bad signer (1 ms)
    Unknown Encoding
      ✓ (FAILS) Tries to sign with unknown encoding (1 ms)

-------------------|---------|----------|---------|---------|-------------------
File               | % Stmts | % Branch | % Funcs | % Lines | Uncovered Line #s 
-------------------|---------|----------|---------|---------|-------------------
All files          |     100 |      100 |     100 |     100 |                   
 ...0wallet.api.ts |     100 |      100 |     100 |     100 |                   
-------------------|---------|----------|---------|---------|-------------------
Test Suites: 1 passed, 1 total
Tests:       14 passed, 14 total
Snapshots:   0 total
Time:        2.568 s, estimated 3 s
Ran all test suites.
Done in 3.08s.

```