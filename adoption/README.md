# ARC Adoption Summaries

This directory contains machine-readable adoption summaries for numbered ARCs.

They are operational artifacts used to record:

- implementation ownership,
- implementation status,
- ecosystem adoption signals,
- evidence used for status transitions.

This directory also contains the canonical vetted adopters registry:

- `vetted-adopters.yaml`

Every adopter `name` used in an ARC adoption summary must:

- be lower-kebab-case;
- appear in the matching category of `vetted-adopters.yaml`.

When an ARC adoption summary is in `Final` status, at least one of its adoption
categories must contain at least one adopter entry. A `Final` summary with all
adoption categories empty is invalid.

## Example

```yaml
arc: 44
title: Transition Ready ARC
status: Final
last-reviewed: 2026-04-09
sponsor: Foundation
implementation-required: true
reference-implementation:
  repository: https://github.com/example/arc-0044
  maintainers:
    - "@maintainer"
  status: shipped
  notes: Canonical reference implementation used for conformance testing.
adoption:
  wallets:
    - name: example-wallet
      status: shipped
      evidence: https://example.com/wallet-proof
      notes: Wallet supports the full ARC flow in production.
  explorers:
    - name: example-explorer
      status: shipped
      evidence: https://example.com/explorer-proof
      notes: Explorer renders ARC-specific data in public UI.
  sdk-libraries:
    - name: example-sdk
      status: in_progress
      evidence: https://example.com/sdk-proof
      notes: SDK support is merged and pending the next release.
  infra:
    - name: example-indexer
      status: shipped
      evidence: https://example.com/indexer-proof
      notes: Infra service exposes the ARC data through its public API.
  dapps-protocols:
    - name: example-dapp
      status: planned
      evidence: https://example.com/dapp-proof
      notes: Integration is scheduled for the next product milestone.
summary:
  adoption-readiness: high
  blockers:
    - Need one more independent SDK implementation.
  notes: Adoption is strong enough for Final, with ongoing follow-up in SDK tooling.
```

All adopter names in the example above must already exist in
`adoption/vetted-adopters.yaml` under the matching category.

An ARC expected to be adopted by ecosystem implementers should keep this file current
throughout `Draft`, `Review`, `Last Call`, `Final`, `Idle`, and `Deprecated` states.
