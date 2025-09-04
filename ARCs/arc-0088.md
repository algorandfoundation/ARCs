---
arc: 88
title: Ownable Access Control
description: Minimal single-administrator / ownership interface for Algorand smart contracts
author: Ludovit Scholtz (@scholtz)
discussions-to: https://github.com/algorandfoundation/ARCs/discussions
status: Draft
type: Standards Track
category: Interface
sub-category: Access Control
created: 2025-09-03
requires: 4
replaces:
---

# ARC-88: Ownable Access Control

## Abstract

ARC-88 defines a minimal, composable, and standardized ownership / single-administrator pattern for Algorand smart contracts. It provides deterministic method names, ABI types, events, and error codes for acquiring, transferring, renouncing, and querying a canonical `owner` address. The goal is cross-tool interoperability and reduced bespoke patterns.

## Motivation

Many contracts require a single privileged authority (issuer, admin, upgrade controller, treasury). Current implementations vary in method names and semantics (e.g., `set_owner`, `changeAdmin`, `transferOwnership`). A unified ARC improves:

- Indexer & wallet discoverability of privileged accounts
- Security auditing consistency
- Reuse of client libraries
- Simplified composition with higher-level governance layers (timelocks, multisigs)

## Specification

### Owner Definition

`owner` is an ARC-4 `address` stored in application global state (or a dedicated box `owner`). The zero address `0x000...0` indicates the absence of an owner (post-renouncement). Implementations MUST reject privileged calls when owner is zero (unless otherwise noted).

### Initialization

Default rule: If no explicit initialization occurs, the `owner` MUST be set to the application creator (deployment sender) at creation time. This gives a deterministic starting authority.

An `arc88_initialize_owner(new_owner: address)` call in the SAME atomic transaction group as creation MAY override the default and set a different initial owner (e.g., multisig). After initialization (either implicit default or explicit override) further initialize attempts MUST fail with `already_initialized`.

A deferred-zero pattern (starting with owner = zero) is NOT standard under ARC-88; implementers desiring that MUST document divergence. ARC-88 compliant tooling can safely assume a non-zero owner immediately after creation unless event `arc88_ownership_renounced` has occurred.

### Methods (ABI)

All method names are snake*case, prefixed with `arc88*`, and use ARC-4 data types.

1. `arc88_owner() -> (owner: address)` (readonly)
2. `arc88_transfer_ownership(new_owner: address)`
3. `arc88_renounce_ownership()`
4. `arc88_initialize_owner(new_owner: address)` (creation group only OR pre-first-use; fails if already set)
5. `arc88_is_owner(query: address) -> (flag: uint64)` (1 if query == stored owner and owner != zero)

Optional extensions:

- `arc88_pending_owner() -> (pending: address)` plus a two-step pattern (`arc88_transfer_ownership_request` / `arc88_accept_ownership`) for safer transfer; see Appendix A.

### Semantics

- Initial owner is deployer unless overridden via `arc88_initialize_owner` in creation group.
- `arc88_transfer_ownership` MUST fail if caller != current owner OR `new_owner` is zero.
- `arc88_renounce_ownership` MUST fail if caller != current owner. Sets owner to zero.
- `arc88_initialize_owner` MUST fail if owner already set (including implicit default). Only valid before any transfer/renounce and prior to first use, practically restricted to creation group for deterministic indexing.
- `arc88_is_owner` returns 1 only if owner != zero AND query == stored owner.
- Once renounced (owner zero) privileged methods relying on ownership MUST become permanently inaccessible unless contract defines an alternate revival extension (non-standard).

### Error Codes

Implementations SHOULD surface deterministic short codes (uint8) via simulation helpers or revert messages (string) for UI mapping:

- `0x01` not_owner (caller lacks ownership)
- `0x02` zero_address_not_allowed (attempt to set owner to zero in transfer)
- `0x03` already_initialized (initialize after owner set)
- `0x04` no_owner_set (actions disallowed because owner is zero)
- `0x05` pending_transfer_exists (two-step pattern only)
- `0x06` not_pending_owner (accept called by non-pending)

Codes >= `0x20` are reserved for project-specific extensions.

### Events (Logs)

Recommended log schema (tag values illustrative; implementers MAY choose deterministic short tags):

- Tag `0x01` `arc88_ownership_transferred` | previous_owner(address) | new_owner(address)
- Tag `0x02` `arc88_ownership_renounced` | previous_owner(address)
- Tag `0x03` `arc88_ownership_transfer_requested` | previous_owner(address) | pending_owner(address) (two-step only)
- Tag `0x04` `arc88_ownership_transfer_accepted` | previous_owner(address) | new_owner(address) (two-step only)

### Security Considerations

- Front-running: Two-step ownership transfer mitigates risk of transferring to an unintended address.
- Renouncement: Irreversible; UIs MUST strongly warn before invoking.
- Multisig owners: If using an escrow address (LogicSig) or multisig, ensure validity period and fallback plan.
- Upgradeability: If used with upgradable proxy patterns, ensure both proxy and implementation align on ownership semantics to avoid lockout.

### Rationale (Specification)

A minimal interface enables straightforward layering of governance (timelock, DAO vote) where those systems call only the standardized methods. Explicit zero address semantics make post-renouncement state unambiguous.

### Reference Implementation Sketch

Global state keys (or boxes):

- `owner`: bytes (32) storing address (zeroed if none)
- (optional) `pending_owner`: bytes (32) for two-step pattern

Pseudocode (TypeScript / Algorand AVM style):

```
arc88_transfer_ownership(new_owner):
  assert(sender == owner, "not_owner")
  assert(new_owner != ZERO, "zero_address_not_allowed")
  previous = owner
  owner = new_owner
  emit OwnershipTransferred(previous, new_owner)

arc88_renounce_ownership():
  assert(sender == owner, "not_owner")
  previous = owner
  owner = ZERO
  emit OwnershipRenounced(previous)
```

### Backwards Compatibility (Specification)

Projects may already expose non-standard naming. They can add ARC-88 methods alongside legacy ones to migrate clients gradually.

### Test Vectors (Illustrative)

1. Deploy (no init) -> query owner -> expect creator.
2. Deploy with grouped `arc88_initialize_owner(multisig)` -> query owner -> multisig.
3. Transfer to new address -> query -> new owner, old no longer passes `arc88_is_owner`.
4. Renounce -> `arc88_owner` returns zero -> `arc88_transfer_ownership` now fails with `no_owner_set`.
5. Two-step (optional): request -> accept by pending -> event sequence 0x03 then 0x04.

## Rationale

See Appendix A for design trade-offs; choosing single-field owner reduces state and gas vs multi-role patterns while enabling composition.

## Backwards Compatibility

Projects with existing non-standard ownership methods can add ARC-88 methods alongside legacy ones to offer a migration path without breaking clients.

## Security Considerations

Security concerns are captured in section 3.7; centralization risks and renouncement irreversibility are primary issues.

## Appendix A: Two-Step Ownership (Optional)

Additional methods:

- `arc88_transfer_ownership_request(pending: address)` (only owner; cannot be zero)
- `arc88_accept_ownership()` (only pending address)
- `arc88_cancel_ownership_request()` (only owner; clears pending)

State: store `pending_owner`.

Semantics: Upon accept, emit both acceptance event and transfer event; clear pending.

## Reference Implementation

[arc88.algo.ts](https://github.com/scholtz/arc-1400/blob/main/projects/arc-1400/smart_contracts/security_token/arc88.algo.ts)

## Copyright

CC0 1.0 Universal.
