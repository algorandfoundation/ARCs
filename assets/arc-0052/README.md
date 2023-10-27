# ARC52 reference implementation

This implementation is not meant to be used in production. It is a reference implementation for the ARC52 specification.

It shows how the ARC52 specification can be implemented on top of a BIP39 / BIP32-ed25519 / BIP44 wallet. 

note: Bip32-ed25519 is meant to be compatible with our Ledger implementation. Meaning that we only do hard-derivation at the level of the `account` as per BIP44

## Run

```shell
$ yarn
$ yarn test
```

## Output

```shell
 PASS  ./contextual.api.crypto.spec.ts (5.393 s)
  Contextual Derivation & Signing
    (Derivations) Context
      ✓ (OK) ECDH (342 ms)
      ✓ (OK) ECDH, Encrypt and Decrypt (365 ms)
      ✓ Libsodium example ECDH (4 ms)
      Addresses
        Soft Derivations
          ✓ (OK) Derive m'/44'/283'/0'/0/0 Algorand Address Key (152 ms)
          ✓ (OK) Derive m'/44'/283'/0'/0/1 Algorand Address Key (115 ms)
          ✓ (OK) Derive m'/44'/283'/0'/0/2 Algorand Address Key (88 ms)
          ✓ (OK) Derive m'/44'/283'/3'/0/0 Algorand Address Key (98 ms)
        Hard Derivations
          ✓ (OK) Derive m'/44'/283'/1'/0/0 Algorand Address Key (94 ms)
          ✓ (OK) Derive m'/44'/283'/2'/0/1 Algorand Address Key (96 ms)
          Ledger Addresses test Vectors
            ✓ (OK) Derive m'/44'/283'/0'/0/0 Algorand Address (94 ms)
            ✓ (OK) Derive m'/44'/283'/1'/0/0 Algorand Address (90 ms)
            ✓ (OK) Derive m'/44'/283'/2'/0/0 Algorand Address (92 ms)
            ✓ (OK) Derive m'/44'/283'/3'/0/0 Algorand Address (91 ms)
            ✓ (OK) Derive m'/44'/283'/4'/0/0 Algorand Address (100 ms)
            ✓ (OK) Derive m'/44'/283'/5'/0/0 Algorand Address (93 ms)
      Identities
        Soft Derivations
          ✓ (OK) Derive m'/44'/0'/0'/0/0 Identity Key (95 ms)
          ✓ (OK) Derive m'/44'/0'/0'/0/1 Identity Key (94 ms)
          ✓ (OK) Derive m'/44'/0'/0'/0/2 Identity Key (91 ms)
        Hard Derivations
          ✓ (OK) Derive m'/44'/0'/1'/0/0 Identity Key (89 ms)
          ✓ (OK) Derive m'/44'/0'/2'/0/1 Identity Key (95 ms)
      Signing Typed Data
        ✓ (OK) Sign Arbitrary Message against Schem (232 ms)
        ✓ (FAIL) Signing attempt fails because of invalid data against Schema (38 ms)

Test Suites: 1 passed, 1 total
Tests:       22 passed, 22 total

```