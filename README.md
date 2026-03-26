# Algorand Request for Comments (ARCs)

This is the official repository for _Algorand Request for Comments_ (ARCs).

Responsibilities:

- defining the canonical process / style guidelines / artifacts for ARCs,
- enforcing the process and artifacts correctness,
- storing ARCs artifacts and history.

## Canonical Venues

Before submitting a new ARC, please refer to the [ARC-0](./ARCs/arc-0000.md).

### 1. Pre-ARC discussion

Canonical venue: GitHub Discussions

Used for:
- validating whether a new ARC is needed,
- checking duplication with existing ARCs,
- collecting early ecosystem feedback,
- determining sponsor classification,
- deciding whether a reference implementation is required.

### 2. ARC text

Canonical venue: one GitHub Pull Request per numbered ARC

Used for:
- normative ARC content,
- revisions to the specification,
- status field updates,
- corrections and clarifications.

### 3. ARC tracking

Canonical venue: one GitHub Issue per numbered ARC

Used for:
- gate decisions,
- checklist completion,
- status transition records,
- links to the reference implementation,
- link to the adoption summary artifact.

### 4. Reference implementation

Canonical venue: dedicated implementation repository

Used for:
- code,
- tests,
- examples,
- conformance artifacts,
- implementation issue tracking.

When `sponsor: Foundation`, the canonical implementation repository lives under
the <a href="https://github.com/algorandfoundation">Algorand Foundation organization</a>.

When `sponsor: Ecosystem`, the canonical implementation repository lives under
the <a href="https://github.com/algorandecosystem">Algorand Ecosystem organization</a>.

### 5. External discussion channels

Mailing lists, Discord, forums, calls, and social channels are non-canonical.

Any material conclusion reached outside GitHub must be copied into the relevant
Pre-ARC discussion or ARC tracking issue.

## Pull Request Templates

Use the appropriate PR template for the type of change:

- New ARC draft: [`arc-draft.md`](./.github/PULL_REQUEST_TEMPLATE/arc-draft.md)
- Transition to Review: [`arc-review.md`](./.github/PULL_REQUEST_TEMPLATE/arc-review.md)
- Transition to Last Call: [`arc-last-call.md`](./.github/PULL_REQUEST_TEMPLATE/arc-last-call.md)
- Transition to Final: [`arc-final.md`](./.github/PULL_REQUEST_TEMPLATE/arc-final.md)
- Adoption or implementation update: [`arc-adoption-or-implementation.md`](./.github/PULL_REQUEST_TEMPLATE/arc-adoption-or-implementation.md)
- Editorial-only change: [`arc-edit.md`](./.github/PULL_REQUEST_TEMPLATE/arc-edit.md)

## Pull Request Validation

TBD
