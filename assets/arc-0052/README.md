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
 PASS  ./contextual.api.crypto.spec.ts
  Contextual Derivation & Signing
    ✓ (OK) Root Key (2 ms)
    (JS Library) Reference Implementation alignment with known BIP32-Ed25519 JS LIB
      ✓ (OK) BIP32-Ed25519 derive key m'/44'/283'/0'/0/0 (135 ms)
      ✓ (OK) BIP32-Ed25519 derive key m'/44'/283'/0'/0/1 (120 ms)
      ✓ (OK) BIP32-Ed25519 derive PUBLIC key m'/44'/283'/1'/0/1 (284 ms)
      ✓ (OK) BIP32-Ed25519 derive PUBLIC key m'/44'/0'/1'/0/2 (277 ms)
    (Derivations) Context
      ✓ (OK) ECDH (4 ms)
      ✓ (OK) ECDH, Encrypt and Decrypt (5 ms)
      ✓ Libsodium example ECDH (8 ms)
      Addresses
        Soft Derivations
          ✓ (OK) Derive m'/44'/283'/0'/0/0 Algorand Address Key (1 ms)
          ✓ (OK) Derive m'/44'/283'/0'/0/1 Algorand Address Key (1 ms)
          ✓ (OK) Derive m'/44'/283'/0'/0/2 Algorand Address Key (2 ms)
        Hard Derivations
          ✓ (OK) Derive m'/44'/283'/1'/0/0 Algorand Address Key (3 ms)
          ✓ (OK) Derive m'/44'/283'/2'/0/1 Algorand Address Key (2 ms)
          ✓ (OK) Derive m'/44'/283'/3'/0/0 Algorand Address Key (1 ms)
      Identities
        Soft Derivations
          ✓ (OK) Derive m'/44'/0'/0'/0/0 Identity Key (1 ms)
          ✓ (OK) Derive m'/44'/0'/0'/0/1 Identity Key (2 ms)
          ✓ (OK) Derive m'/44'/0'/0'/0/2 Identity Key (1 ms)
        Hard Derivations
          ✓ (OK) Derive m'/44'/0'/1'/0/0 Identity Key (2 ms)
          ✓ (OK) Derive m'/44'/0'/2'/0/1 Identity Key (1 ms)
      Signing Typed Data
        ✓ (OK) Sign Arbitrary Message against Schem (54 ms)
        ✓ (FAIL) Signing attempt fails because of invalid data against Schema (33 ms)
        Reject Regular Transaction Signing. IF TAG Prexies are present signing must fail
          ✓ (FAIL) [TX] Tag
          ✓ (FAIL) [MX] Tag (1 ms)
          ✓ (FAIL) [Program] Tag
          ✓ (FAIL) [progData] Tag (1 ms)
          Reject tags present in the encoded payload
            ✓ (FAIL) [TX] Tag (2 ms)
            ✓ (FAIL) [MX] Tag
            ✓ (FAIL) [Program] Tag (1 ms)
            ✓ (FAIL) [progData] Tag


```

## BIP39 / BIP32-ed25519 / BIP44 Test Vectors

All keys are in extended format: [kl][kr][chaincode]

Public key = kl * Ed25519GeneratorPoint

- `BIP39 mnemonic`: _salon zoo engage submit smile frost later decide wing sight chaos renew lizard rely canal coral scene hobby scare step bus leaf tobacco slice_

- `root key (hex)`: a8ba80028922d9fcfa055c78aede55b5c575bcd8d5a53168edf45f36d9ec8f4694592b4bc892907583e22669ecdf1b0409a9f3bd5549f2dd751b51360909cd05b4b67277d74d4ddb3688daeeb02075482ceb812db8a5757c9e792d14ec791554

### BIP44 paths

#### Child Private Derivation

- `m'/44'/283'/0'/0/0`: 70db1fe0dd722955f44b57b26eee09ac282172559eb2bc6b408f69e2e9ec8f461b8fbc75f0e2a9f454d7de35a812c97fa08b984df342389f4d1f9aec5637e67468c11d826d16f7f505d27bd52c314caea79a7b9b6e87f5aa7e20f6cf06ac51f9
  - corresponding public key: 7915e7ecbaad1dc9bc22a9e496686687f1a8cb4895b7ca46f86d64dd56c6cd97 (kl * Ed25519GeneratorPoint)
  
- `m'/44'/283'/0'/0/1`: 58678408f4f88d7ec700be1090940bedd584c21dc7f8b00028b69e01edec8f4645779d1c90ecc5168945024e4201552a349a825518badb8d4b017bf1dcce4dac910cb8dbcd10680c7754ca101089f4c9da61a2dff79e0301d9a1fb7656fb8c73
  - public key: 054a6881d8809c348a402d67ba2feedcd8e3145f40f21a6bbd0de09c30c78d0a


- `m'/44'/283'/0'/0/2`: 88ce207945dfecea41acbb2a9b4563268d3bee03ea2e199af502c500e7ec8f467b95d890e8c7e5aa79d9905f3c6794e1fc3bb9478552b7f24c751b94dd3becd76699130de29273b2ae742991b7daa0adae6fb656038b25895d4af3c85769a64d
  - public key: 8cea8052cfa1fd8cec0b4fad6241a91f2edbfe9f072586f243839174e40a25ef

- `m'/44'/283'/1'/0/0`: 18ada9291ae27f7415e99f4485d81493e68952c8b1af4f335cc52962ecec8f464c106ef1f56fa64b2a13631a68a6d6c9902eec52e2c6226a3ed953ae94bc8a613303bbf403f94b5eb189eed703b3985e12c6726d6b06d4ed8aac3d5b258e3c8c
  - public key: 04f3ba279aa781ab4f8f79aaf6cf91e3d7ff75429064dc757001a30e10c628df

- `m'/44'/283'/2'/0/0`: 209dc0b3458ee1f7484189bf584c02807f1a5726168aff5ea14fc117f3ec8f46b68ea14ca84c0da34aa4990416c6da01b14fbe30c5c225238515a93a0a28a33dbecbe8e64a514d44c8c0a3775844d5aa8e18ea2182321f1ff87061e49adc4a77
  - public key: 400d78302258dc7b3cb56d1a09f85b018e8100865ced0b5cda474c26bbc07c30


- `m'/44'/283'/3'/0/0`: 3008165b92e6b29d6a3f28b593f15dcf6ce4c9a5a1aadc3a43ca068ce7ec8f46b8251db83f68dedaee1054ec545257f95ba00e33566d99926bef8743b77b42b8867472d1f1886c88ec36991f14a333454003236b5375d4a8bf4f01b2ff85ec9d
  - public key: cf8d28a3d41bc656acbfeadb64d06054142c97bee6a987c11d934f84853df866

- `m'/44'/0'/0'/0/0`: 08b10dfcfc37a0cc998ec3305c6c32d4412de73f3f9633342248d14aeeec8f461997b1136524a5d3499d933c6f6e739d38fbf7404a183f4d835c6fe105cf44d016634694e4087b2d547c8b550a0f053d2fba990ba7e5e58186963393994d06a9
  - public key: 28804f08d8c145e172c998fe75058237b8181846ca763894ae3eefea6ab88352

- `m'/44'/0'/0'/0/1`: 80d6dd1cefb42fc93187739ad102e5dfd533b54b847b0693661bfee9edec8f466cab918344a8e7e9ae196d94cce301aa1d7360f550b25f3dc7468a006e423e3de80e43cd2588a9996b8156d3dc233ca31470f7d49261edd2e13d8c4c0096163c
  - public key: fc8d1c79edd406fa415cb0a76435eb83b6f8af72cd9bd673753471470205057a

- `m'/44'/0'/0'/0/2`: 7847139bee38330c809a56f3d2261bfbd7da7c1c88999c38bb12175cefec8f467617b7dba8330b644eae1c24a2d6b7211f5f36d9919d9a240777fa32d382feb3debfe21afa0de5021342f3bfafe18a91e11c441ab98ff5bcbba2dfba3190ce6f
  - public key: f7e317899420454886fe79f24e25af0bbb1856c440b14829674015e5fc2ad28a

