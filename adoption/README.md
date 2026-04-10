# ARC Adoption Summaries

This directory contains machine-readable adoption summaries for numbered ARCs.

They are operational artifacts used to record:

- implementation status,
- ecosystem adoption signals,
- evidence used for status transitions.

This directory also contains the canonical vetted adopters registry:

- `vetted-adopters.yaml`

Every adopter `name` used in an ARC adoption summary must:

- be lower-kebab-case;
- appear in the matching category of `vetted-adopters.yaml`.

The ARC Markdown front matter is authoritative for `status`, `sponsor`, and
`implementation-required`. Adoption-summary validation derives those values from
the matching ARC file rather than repeating them in YAML.

When the matching ARC is in `Final` status, at least one adoption category must
contain at least one adopter entry. A `Final` ARC with an all-empty adoption
summary is invalid.

When the matching ARC front matter sets `implementation-required: true`:

- the ARC front matter is authoritative for `implementation-url` and `implementation-maintainer`;
- the adoption summary tracks only `reference-implementation.status` and `reference-implementation.notes`.

## Example

```yaml
arc: 44
title: Transition Ready ARC
last-reviewed: 2026-04-09
reference-implementation:
  status: shipped
  notes: Canonical implementation status used for conformance tracking.
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

The canonical implementation repository URL and maintainer list for this ARC would
be declared in the ARC Markdown front matter, together with the ARC `status`,
`sponsor`, and `implementation-required` values, not repeated in this file.

An ARC expected to be adopted by ecosystem implementers should keep this file current
throughout `Draft`, `Review`, `Last Call`, `Final`, `Idle`, and `Deprecated` states.
