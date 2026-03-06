---
arc: <to be assigned>
title: Algo X EVM Accounts
description: A LogicSig standard enabling EVM wallets to control Algorand accounts via EIP-712 signed ECDSA secp256k1 signature verification.
author: Tasos Bitsios (@tasosbit)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/<to be assigned>
status: Draft
type: Standards Track
category: Interface
sub-category: Application
created: 2026-03-02
---

# Algo X EVM Accounts

## Abstract

This ARC specifies a LogicSig-based standard that allows any Ethereum (EVM) wallet address to control an Algorand account without requiring the user to manage a separate Algorand private key. A deterministic LogicSig program is compiled per EVM address by substituting the 20-byte Ethereum address as a template variable. To authorize an Algorand transaction, the EVM wallet signs the transaction ID (or atomic group ID) using <a href="https://eips.ethereum.org/EIPS/eip-712">EIP-712</a> typed structured data. The LogicSig verifies the ECDSA (secp256k1) signature on-chain using the AVM's native `ecdsa_pk_recover` and `keccak256` opcodes, and approves the transaction if and only if the recovered Ethereum address matches the template owner. This approach requires no changes to the Algorand protocol and enables users of MetaMask and other EVM wallets to interact with Algorand dApps using their existing credentials.

## Motivation

Users of EVM-compatible wallets represent a large segment of the Web3 ecosystem. Requiring these users to create and manage an additional Algorand wallet with a separate seed phrase significantly raises the barrier to entry for Algorand dApps.

The AVM provides the cryptographic primitives necessary to verify ECDSA secp256k1 signatures natively: `ecdsa_pk_recover` recovers the public key from a signature and `keccak256` produces Ethereum-compatible hashes. Combined with LogicSig template variables, these opcodes make it possible to implement a stateless smart contract that performs full Ethereum signature verification without any on-chain state or application calls.

This ARC defines the standard for how EVM addresses are mapped to Algorand LogicSig addresses, how transactions are signed using EIP-712 typed data, how the signature is encoded in the LogicSig argument, and how the LogicSig verifies the signature. Standardizing these conventions enables interoperability between multiple implementations, wallets, SDKs, and explorers.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

> Comments like this are non-normative.

### Definitions

- **EVM address**: A 20-byte Ethereum account address, canonically represented as a 0x-prefixed 40-character lowercase hex string (e.g., `0x742d35cc6634c0532925a3b844bc9e7595f0beb2`).
- **Algo X EVM Account**: An Algorand LogicSig account whose authorization is controlled by a specific EVM address via the mechanism defined in this ARC.
- **LogicSig address**: The 58-character base32 Algorand address derived from the hash of a compiled LogicSig program.
- **Template variable**: A named placeholder (`TMPL_*`) in a TEAL program that is substituted with a concrete value before compilation to bytecode.
- **EIP-712 digest**: The final 32-byte hash used as the signature payload, computed according to <a href="https://eips.ethereum.org/EIPS/eip-712">EIP-712</a>.
- **Lower-S normalization**: The practice of converting an ECDSA signature's S component to the lower half of the curve order (if `s > n/2`, replace with `n - s`) to ensure a canonical signature form.
- **Type byte**: The first byte of the LogicSig argument (`arg[0]`) that identifies the signature scheme. Defined in this ARC as `0x01` for EVM secp256k1 signatures.

---

### Overview

> This section is non-normative.

The flow from EVM wallet to Algorand authorization is:

1. An off-chain client derives the Algo X EVM Account address for a given EVM address by substituting the 20-byte address as the `TMPL_OWNER` template variable in the canonical LogicSig TEAL source, compiling the result, and deriving the LogicSig address from the compiled program hash.
2. To sign an Algorand transaction (or atomic group), the client constructs an EIP-712 typed data object whose message field contains the 32-byte transaction ID (for single transactions) or group ID (for atomic groups).
3. The EVM wallet signs the EIP-712 typed data using `eth_signTypedData_v4`, producing a 65-byte signature `(R || S || V)`.
4. The client normalizes the signature to lower-S form, prepends the type byte `0x01`, and passes the resulting 66 bytes as `arg[0]` to the LogicSig.
5. The Algorand network evaluates the LogicSig: it verifies the type byte, recomputes the EIP-712 digest, recovers the signer's public key via `ecdsaPkRecover`, derives the Ethereum address, and approves the transaction if and only if the derived address equals the template owner.

---

### Algorand Chain Identifier

All Algo X EVM Account implementations **MUST** use `4160` as the `chainId` in the EIP-712 domain for all Algorand networks (MainNet, TestNet, LocalNet). This constant value is used uniformly across networks because:

- Algorand transactions already provide replay protection through the genesis hash and genesis ID fields embedded in the transaction itself.
- Using a single constant avoids requiring the signer to know which Algorand network they are on, simplifying EVM wallet UX.

The `chainId` value `4160` (`0x1040`) **MUST** be used. Implementations using any other chain ID are not compliant with this ARC.

> `4160` is the decimal representation of `0x1040`, chosen to be recognizable as an Algorand-specific identifier.

---

### LogicSig Template

#### Template Variable

The LogicSig **MUST** contain exactly one template variable:

| Variable | Size | Description |
|----------|------|-------------|
| `TMPL_OWNER` | 20 bytes | The EVM address (normalized to lowercase hex, without `0x` prefix) that controls this LogicSig instance |

The EVM address **MUST** be normalized to lowercase (without `0x` prefix) before substitution into the template. Implementations **MUST NOT** use checksummed (EIP-55) mixed-case encoding as the template value.

#### Precomputed Constants

The LogicSig **MUST** embed the following precomputed constants as byte literals in the TEAL program:

**Domain Separator** (32 bytes):

```
0xcd2715b67ae987618a9e27b3a29c522b1171fd767b2224547d03747eae76adc6
```

Computed as:
```
keccak256(
  keccak256("EIP712Domain(string name,string version,uint256 chainId)")
  || keccak256("Liquid Accounts")
  || keccak256("1")
  || uint256(4160)
)
```

**Message Type Hash** (32 bytes):

```
0xa0d3cab9c111e1025e8e6c24067ada7c8fff46e1696e611a8b2e5049bac4baf6
```

Computed as:
```
keccak256("AlgorandTransaction(bytes32 Transaction ID)")
```

---

### EIP-712 Typed Data

#### Domain

The EIP-712 domain **MUST** be:

```json
{
  "name": "Liquid Accounts",
  "version": "1",
  "chainId": 4160
}
```

The domain type descriptor **MUST** be:

```json
[
  { "name": "name",    "type": "string" },
  { "name": "version", "type": "string" },
  { "name": "chainId", "type": "uint256" }
]
```

No other fields (e.g., `verifyingContract`, `salt`) **SHALL** be included in the domain.

#### Message Type

The primary type **MUST** be `AlgorandTransaction` with the following type definition:

```json
{
  "AlgorandTransaction": [
    { "name": "Transaction ID", "type": "bytes32" }
  ]
}
```

#### Sign Payload

The 32-byte value placed in the `Transaction ID` field **MUST** be:

- The **group ID** of the atomic transaction group, if `groupSize > 1` (i.e., the transaction is part of a multi-transaction group).
- The **transaction ID** of the single transaction, if `groupSize == 1`.

Implementations **MUST** use the group size to select the payload, regardless of group ID existence. Group ID **MUST** be the transaction ID payload when the transaction is part of an atomic group of more than one transaction. Using the group ID for groups ensures that the signature covers the entire atomic group and prevents a signer from being tricked into signing only a subset of a group.

> It is valid for a single transaction to have a group ID attached. In this case it should be ignored and the Transaction ID must be used.

> The group ID is a 32-byte value computed by the Algorand SDK from the set of transactions in the group (in canonical msgpack encoding). It is available as `txn.Group` / `Global.GroupID` in TEAL once the group is assembled.

The typed data message **MUST** be constructed as:

```json
{
  "Transaction ID": "0x<32-byte payload as lowercase hex>"
}
```

#### Digest Computation

The EIP-712 digest is:

```
messageHash = keccak256(MESSAGE_TYPE_HASH || payload)
digest      = keccak256(0x1901 || DOMAIN_SEPARATOR || messageHash)
```

where `||` denotes byte concatenation and `payload` is the 32-byte sign payload defined above.

---

### LogicSig Argument Format

The first argument (`arg[0]`) passed to the LogicSig **MUST** be exactly 66 bytes, structured as:

```
arg[0] = Type (1 byte) || R (32 bytes) || S (32 bytes) || V (1 byte)
```

| Offset | Length | Description |
|--------|--------|-------------|
| 0 | 1 | Type byte. **MUST** be `0x01` for EVM secp256k1 signatures. |
| 1 | 32 | R component of the ECDSA signature, big-endian. |
| 33 | 32 | S component of the ECDSA signature, big-endian (lower-S normalized). |
| 65 | 1 | V component. **MUST** be `0x1b` (27) or `0x1c` (28). |

The S component **MUST** be lower-S normalized before encoding. If the signature returned by the EVM wallet has `s > n/2` (where `n` is the secp256k1 curve order `0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141`), the implementation **MUST** replace `s` with `n - s` and flip `V` (`0x1b` ↔ `0x1c`). The AVM's `ecdsa_pk_recover` opcode only accepts lower-S signatures and will reject non-normalized signatures.

#### Type Byte Registry

The type byte is defined to support future multi-scheme LogicSig compositions. Implementations that do not recognize a type byte **MUST** reject the transaction (by returning `0`/failure from the LogicSig program). The following type bytes are defined:

| Type Byte | Scheme | Status |
|-----------|--------|--------|
| `0x00` | Reserved | — |
| `0x01` | EVM secp256k1 (EIP-712), defined in this ARC | Active |

All other type byte values are reserved for future ARC definitions.

---

### LogicSig Verification Algorithm

The LogicSig program **MUST** implement the following verification logic:

1. Read `arg[0]` (66 bytes).
2. Assert that `arg[0][0] == 0x01` (type byte check). Fail if not.
3. Extract `R = arg[0][1:33]`, `S = arg[0][33:65]`, `V = arg[0][65]`.
4. Compute `recoveryId = V - 27`. The value **MUST** be `0` or `1`.
5. Determine the sign payload:
   - If `Global.GroupSize == 1`: payload = `Txn.TxID` (32 bytes)
   - If `Global.GroupSize > 1`: payload = `Global.GroupID` (32 bytes)
6. Compute `messageHash = keccak256(MESSAGE_TYPE_HASH || payload)`.
7. Compute `digest = keccak256(0x1901 || DOMAIN_SEPARATOR || messageHash)`.
8. Execute `(pubkeyX, pubkeyY) = ecdsaPkRecover(Secp256k1, digest, recoveryId, R, S)`.
9. Compute `recoveredAddress = keccak256(pubkeyX || pubkeyY)[12:32]` (last 20 bytes).
10. Approve the transaction if and only if `recoveredAddress == TMPL_OWNER`.

---

### Address Derivation

The Algo X EVM Account address for a given EVM address is derived as follows:

1. Normalize the EVM address to lowercase hex without `0x` prefix (20 bytes, 40 hex characters).
2. Substitute the normalized bytes as `TMPL_OWNER` in the canonical LogicSig TEAL source.
3. Compile the resulting TEAL program to produce a bytecode program `P`.
4. The Algorand address is `ChecksumBase32(SHA512_256("Program" || P))`, following the standard Algorand address derivation for LogicSig accounts.

Each EVM address maps deterministically to exactly one Algorand address. The mapping is:
- **Injective**: distinct EVM addresses yield distinct Algorand addresses.
- **Stateless**: no on-chain registration or state is required to establish the mapping.
- **Offline-computable**: the address can be computed from the EVM address alone without any network access.

> In practice, clients use the Algod `compileTeal` endpoint with the template variable substituted to obtain the compiled bytecode, then derive the address using the standard `algosdk.LogicSigAccount` address derivation.

---

### SDK Interface

Implementations of this standard **SHOULD** expose the following interface for interoperability with dApp developers and wallet adapters:

```typescript
interface LiquidEvmSdkInterface {
  /**
   * Derive the Algorand LogicSig address for a given EVM address.
   * MUST be deterministic: same input always produces the same output.
   */
  getAddress(params: { evmAddress: string }): Promise<string>

  /**
   * Return the 32-byte payload that the EVM wallet must sign.
   * For grouped transactions: the group ID.
   * For a single transaction: the transaction ID.
   */
  getSignPayload(txnGroup: Transaction[]): Uint8Array

  /**
   * Sign one or more Algorand transactions for the given EVM address,
   * invoking the provided callback to obtain the EIP-712 signature.
   */
  signTxn(params: {
    evmAddress: string
    txns: Transaction[]
    signMessage: (typedData: SignTypedDataParams) => Promise<string>
  }): Promise<Uint8Array[]>
}

interface SignTypedDataParams {
  domain: {
    name: "Liquid Accounts"
    version: "1"
    chainId: 4160
  }
  types: {
    EIP712Domain: Array<{ name: string; type: string }>
    AlgorandTransaction: [{ name: "Transaction ID"; type: "bytes32" }]
  }
  primaryType: "AlgorandTransaction"
  message: { "Transaction ID": `0x${string}` }
}
```

The `signMessage` callback **MUST** be called with the full EIP-712 typed data object (domain, types, primaryType, message) so that the EVM wallet can display a human-readable signing prompt. Implementations **MUST NOT** pre-hash the payload before calling `signMessage`.

---

### Wallet Adapter Requirements

Wallet adapters implementing this ARC **MUST**:

1. Present the EVM wallet's `eth_signTypedData_v4` (or equivalent) prompt for every transaction signature request.
2. Derive the Algorand account address from the connected EVM address using the address derivation procedure defined above.
3. Use the group ID as the sign payload for atomic transaction groups of more than 1 transaction, and the transaction ID for single transactions.
4. Normalize the returned signature to lower-S form before encoding it as the LogicSig argument.
5. Surface the connected EVM address (as wallet account metadata) alongside the derived Algorand address so that explorers and dApps can display both identifiers to the user.

Wallet adapters **SHOULD**:

1. Cache the compiled LogicSig bytecode for each EVM address to avoid redundant compilation requests.
2. Support multiple simultaneous EVM accounts, providing one Algo X EVM Account address per EVM address.
3. Request that the EVM wallet switch to `chainId: 4160` (the Algorand network identifier) before presenting signing prompts, to display correct chain context in the wallet UI. **TODO confirm if this is required.**

---

## Rationale

### LogicSig vs. Smart Contract

A LogicSig (stateless program) was chosen over a stateful smart contract for account authorization because:

- **No opt-in or minimum balance overhead**: A LogicSig account is a standard Algorand account. It does not require opting in to an application or holding additional ALGO for application state.
- **Minimal fee**: LogicSig authorization adds only the base transaction fee with no inner transaction overhead.
- **Deterministic address**: The LogicSig address is purely a function of the program, allowing off-chain derivation without any on-chain registration step.
- **No admin key**: There is no privileged operator or upgrade mechanism. Once compiled, the program is immutable and the owner address cannot be changed.

### EIP-712 over Raw Transaction Hash

EIP-712 typed structured data was chosen for the signing payload because:

- **Human-readable prompts**: EVM wallets display the structured type and value to the user, making it clear what they are signing, rather than presenting an opaque 32-byte hash.
- **Domain separation**: The domain separator (embedding the application name, version, and chain ID) prevents signatures created for other purposes from being replayed against Algo X EVM Accounts.
- **Broad wallet support**: `eth_signTypedData_v4` is supported by all major EVM wallets (MetaMask, Ledger, WalletConnect v2, etc.).

### Group ID for Atomic Groups

Signing the group ID rather than individual transaction IDs for atomic groups ensures that a single EVM wallet signature authorizes the complete atomic transfer. This minimizes the user signatures needed for an atomic group and prevents the signer from being misled about the full scope of what they are authorizing.

### Constant Chain ID 4160

Using a single chain ID for all Algorand networks facilitates having a deterministic Algorand address for all networks.

If distinct chain IDs were used per Algorand network, this would yield different pre-computed hashes, which would in turn yield different addresses per network. This must be avoided as it would break user expectations.

Replay protection across Algorand networks is already provided by the Algorand transactions' genesis hash field.

### Type Byte for Future Extensibility

The type byte prefix on `arg[0]` reserves the ability to extend this standard to additional signature schemes (e.g., WebAuthn/Passkey using secp256r1 as type `0x02`) within the same LogicSig framework without breaking existing implementations. A LogicSig can branch on the type byte to support multiple authentication schemes, enabling composed authorization policies such as "EVM **or** Passkey" in a single account.

### Lower-S Normalization

The AVM's `ecdsa_pk_recover` opcode requires that the signature's S component be in the lower half of the secp256k1 curve order. EVM wallets may produce signatures with S in either half. Normalizing to lower-S off-chain (and adjusting V accordingly) is a standard practice (used in Bitcoin and most Ethereum tooling) that does not affect the recovered public key.

## Backwards Compatibility

This ARC introduces a new standard that has no conflict with any existing Algorand accounts, applications, or ARCs. Algo X EVM Account addresses are standard Algorand addresses that can receive ALGO and ASAs, be used as transaction senders, and interact with any Algorand application. Existing Algorand tooling requires no changes to support these accounts.

## Test Cases

The following test vectors verify correct implementation of the address derivation and signature verification.

### EIP-712 Digest

Given:
- `DOMAIN_SEPARATOR = 0xcd2715b67ae987618a9e27b3a29c522b1171fd767b2224547d03747eae76adc6`
- `MESSAGE_TYPE_HASH = 0xa0d3cab9c111e1025e8e6c24067ada7c8fff46e1696e611a8b2e5049bac4baf6`
- `payload = 0x` + (32 zero bytes, for illustration)

Expected:
```
messageHash = keccak256(MESSAGE_TYPE_HASH || payload)
            = keccak256(0xa0d3cab9c111e1025e8e6c24067ada7c8fff46e1696e611a8b2e5049bac4baf6
                        0000000000000000000000000000000000000000000000000000000000000000)

digest = keccak256(0x1901 || DOMAIN_SEPARATOR || messageHash)
```

Full end-to-end test vectors (EVM private key → signature → LogicSig approval) are available in the reference implementation test suite at `../assets/arc-draft/logicsig.e2e.spec.ts`.

### Lower-S Normalization

Given a signature `(R, S, V)` where `S > n/2`:
- Input: `V = 0x1b`, `S = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFD74DCE442A3CCDA477FA4BCA9DF6C5555`
- Expected normalized: `V = 0x1c`, `S = 0x0000000000000000000000000000000145d1f8a40b7bc5f4402da1e2f0c9ebec` (i.e., `n - S`)

## Reference Implementation

The canonical reference implementation is the `liquid-accounts-evm` TypeScript package, which provides:

- **LogicSig TEAL source**: the parameterized TEAL program with `TMPL_OWNER`
- **`LiquidEvmSdk`**: address derivation, transaction signing, and signer factory
- **`parseEvmSignature`**: lower-S normalization and type-byte encoding
- **`buildTypedData`**: EIP-712 typed data construction

Source: <a href="https://github.com/tasos-bitsios/liquid-accounts">https://github.com/tasos-bitsios/liquid-accounts</a>

> References to "Liquid Accounts" will be updated when the name is finalized.

### Signing a Single Transaction (TypeScript)

```typescript
import { LiquidEvmSdk, buildTypedData } from 'liquid-accounts-evm'
import { AlgorandClient, AlgoAmount } from '@algorandfoundation/algokit-utils'

const algorand = AlgorandClient.fromEnvironment()
const sdk = new LiquidEvmSdk({ algorand })

const evmAddress = '0x742d35cc6634c0532925a3b844bc9e7595f0beb2'

// Derive Algorand address
const algoAddress = await sdk.getAddress({ evmAddress })

// Sign a payment transaction
const txn = await algorand.createTransaction.payment({
  sender: algoAddress,
  receiver: '...',
  amount: AlgoAmount.Algo(1),
})

const signed = await sdk.signTxn({
  evmAddress,
  txns: [txn],
  signMessage: async (typedData) => {
    // Pass directly to eth_signTypedData_v4 or any EIP-712-compatible signer
    return window.ethereum.request({
      method: 'eth_signTypedData_v4',
      params: [evmAddress, JSON.stringify(typedData)],
    })
  },
})

await algorand.client.algod.sendRawTransaction(signed).do()
```

### Signing an Atomic Group (TypeScript)

```typescript
const { transactions: [txn1, txn2] } = await algorand.newGroup()
  .addPayment({ sender: algoAddress, receiver: addr2, amount: AlgoAmount.Algo(1) })
  .addPayment({ sender: algoAddress, receiver: addr3, amount: AlgoAmount.Algo(2) })
  .buildTransactions()

// Both transactions in the group share the same group ID.
// The SDK automatically uses the group ID as the sign payload.
const signed = await sdk.signTxn({ evmAddress, txns: [txn1, txn2], signMessage })
```

### LogicSig Verification (Pseudocode)

```
function verify(arg0: bytes[66], txn: Transaction, owner: bytes[20]) -> bool:
  if arg0[0] != 0x01: return false          // type byte check
  R = arg0[1:33]
  S = arg0[33:65]
  V = arg0[65]
  recoveryId = V - 27                        // 0 or 1

  if groupSize == 1:
    payload = txn.txID                       // 32 bytes
  else:
    payload = global.groupID                 // 32 bytes

  msgHash = keccak256(MESSAGE_TYPE_HASH || payload)
  digest  = keccak256(0x1901 || DOMAIN_SEPARATOR || msgHash)

  (pubX, pubY) = ecdsaPkRecover(Secp256k1, digest, recoveryId, R, S)
  recovered = keccak256(pubX || pubY)[12:32]  // last 20 bytes

  return recovered == owner
```

## Security Considerations

### Signature Scope and Replay Prevention

**Transaction binding**: The EIP-712 message binds the signature to a specific transaction ID or group ID. The signature cannot be reused to authorize a different transaction.

**Cross-group replay**: For atomic groups, signing the group ID rather than individual transaction IDs ensures that the signer authorizes the complete group. An adversary cannot extract a subset of a group and submit it as a standalone transaction or as part of a different group.

**Cross-network replay**: Algorand transactions already contain a genesis hash field that binds them to a specific network. A transaction signed for TestNet cannot be submitted to MainNet. The EIP-712 domain's chain ID (`4160`) is therefore constant across Algorand networks without weakening replay protection.

**Cross-protocol replay**: The EIP-712 domain separator (containing the name `"Liquid Accounts"`, version `"1"`, and chain ID `4160`) prevents signatures created for other EIP-712 applications from being accepted by this LogicSig, and vice versa.

### Signature Malleability

ECDSA signatures have a known malleability: for any valid `(R, S, V)`, the pair `(R, n - S, flipped V)` is also a valid signature for the same message under the same key. The AVM only accepts lower-S signatures. All compliant implementations **MUST** normalize signatures to lower-S form before encoding them as the LogicSig argument. Failing to do so will result in a transaction rejection if the wallet produces an upper-S signature.

### Template Immutability

Once a LogicSig is compiled with a specific `TMPL_OWNER`, the owner address is permanently embedded in the program bytecode. The Algo X EVM Account address cannot be transferred to a different EVM address without creating a new LogicSig. This is by design: the LogicSig is analogous to a deterministic wallet derived from the EVM private key.

### Loss of EVM Private Key

If the user loses their EVM private key, they lose access to the associated Algo X EVM Account. There is no recovery mechanism. Users **SHOULD** be clearly informed that their Algorand account's security depends entirely on the security of their EVM wallet. dApp frontends **SHOULD** encourage users to use hardware wallets or accounts with social recovery for Algo X EVM Account management.

### EVM Wallet UI Trust

The EIP-712 signing prompt displayed by the EVM wallet is under the wallet's control. The `Transaction ID` field will be displayed as a 32-byte hex value which is not human-interpretable. Users are trusting that the dApp requesting the signature is authorized and that the transaction has not been tampered with between display and signing.

dApp implementations **SHOULD** provide clear UI context about what is being signed (transaction type, amounts, recipients) before invoking the EVM signing prompt, and **SHOULD** offer a transaction review step so users can inspect the transaction before it reaches the wallet prompt.

### LogicSig Program Hash Forgery

The Algorand address derived from a LogicSig is the hash of the compiled program. Two different programs cannot produce the same address (SHA-512/256 preimage resistance). There is no known attack that allows an adversary to forge a program with a specific owner address that hashes to the same address.

### Dependency on AVM ECDSA Opcodes

The security of this standard depends on the correctness of the AVM's `ecdsa_pk_recover` implementation. This opcode has been available in the AVM since consensus version 6 and is used in production by other Algorand applications.

### Third-Party EVM RPC Endpoints

This ARC does not specify or mandate any EVM-compatible RPC server for Algorand. Implementations that provide an EVM JSON-RPC interface to allow EVM wallets to query ALGO balances using an Ethereum-compatible interface **SHOULD** clearly communicate to users that such an interface is not a consensus-level protocol and that it does not affect on-chain security.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
