# ARC Repository CI/CD and Release Specification

## 1. Purpose

This document defines the CI/CD, release, and GitHub automation policy for the ARC
repository.

This specification is derived from:

1. [ARC-0000](../ARCs/arc-0000.md)
1. [arckit Specification](../arckit/_docs/specs.md)
1. [arckit Tech Stack](../arckit/_docs/tech-stack.md)

## 2. Goals

The pipeline must:

1. stay simple enough for a solo maintainer and ARC editor to operate confidently;
1. keep the canonical repository gate centered on `arckit validate repo .`;
1. distinguish deterministic machine checks from editorial judgment;
1. help authors keep ARC artifacts current after merge, not only at PR time;
1. automate routine repository hygiene, release packaging, and GitHub routing without
   over-complicating the process.

## 3. Design Principles

The pipeline must follow these principles:

1. **Small and explicit.** Use a small number of clearly named workflows with stable
   responsibilities.
1. **Offline-first.** The required validation gate uses offline `arckit` validation.
1. **Assistive online checks.** External link reachability is useful, but must not
   become the canonical merge gate.
1. **Process-aware.** GitHub-native ARC process requirements are enforced separately
   from `arckit`, because `arckit` does not own GitHub API lookups or repository mutation.
1. **Maintenance-oriented.** CI is both an intake gate and a recurring repository
   maintenance mechanism.
1. **Solo-maintainer friendly.** Prefer one rollup report over noisy issue fan-out,
   and automate reminders rather than editorial decisions.

## 4. Repository Conventions

### 4.1 Protected Branch

The protected branch is `main`. Changes are merged through pull requests only.

### 4.2 Stable Check Surface

The repository should expose these stable required PR checks:

1. `hygiene`
1. `repo-validate-offline`
1. `arc-process-check`
1. `arckit-tool`

The repository should expose this stable non-required PR check:

1. `online-validation`

### 4.3 Labels

The minimum supported labels are:

1. `area:arc`
1. `area:adoption`
1. `area:arckit`
1. `area:github`
1. `kind:draft`
1. `kind:review`
1. `kind:last-call`
1. `kind:final`
1. `kind:adoption`
1. `kind:editorial`
1. `arc-tracking`
1. `maintenance-report`
1. `maintenance:authors-action`
1. `maintenance:editor-action`

## 5. Pull Request Validation

### 5.1 Required Offline Repository Gate

The canonical repository-content gate is:

```text
arckit validate repo .
```

This check is required for relevant pull requests and is the authoritative machine
validation gate for ARC repository artifacts.

That gate includes the repository-scoped vetted adopters registry at
`adoption/vetted-adopters.yaml` and rejects per-ARC adoption actors that are not
lower-kebab-case or are not present in the matching registry category.

That same gate must also reject any `Final` ARC whose canonical adoption summary
has all adoption categories empty.

For ARCs with `implementation-required: true`, that gate must treat ARC front
matter as the authoritative source of `implementation-url` and
`implementation-maintainer`, and it must reject adoption summaries that duplicate
those identity fields under `reference-implementation`.

That same gate must also reject non-canonical `reference-implementation.status`
values; the only supported values are `planned`, `wip`, `shipped`, and `archived`.

That same gate must also reject implementation-required ARCs whose
`implementation-url` does not exactly match the sponsor-specific canonical GitHub
repository path `https://github.com/algorandfoundation/arcN` or
`https://github.com/algorandecosystem/arcN`, where `N` is the unpadded ARC number.

That same gate must also treat ARC front matter as authoritative for `status`,
`sponsor`, and `implementation-required`, and it must reject adoption summaries
that redundantly declare those ARC-owned fields.

That same gate must also reject adoption summaries whose `summary.adoption-readiness`
is `medium` with fewer than 3 adopter entries or `high` with fewer than 5 adopter
entries across all adoption categories.

When present, the repository-root `.arckit.jsonc` is applied implicitly by this
command. Invalid `.arckit.jsonc` content must fail the gate.

During staged metadata migrations, repo-local suppressions may temporarily keep
historical ARC files green in this repository-wide gate.

When ARC Markdown files change, the workflow must also run a changed-file
validation pass that enforces `R:021` even when `.arckit.jsonc` suppresses that
rule at the repository level. This targeted pass must keep unrelated suppressions
intact so historical ARC-specific waivers do not become false failures.

### 5.2 Hygiene Check

The pipeline must include a hygiene check driven by the repository-root
`.pre-commit-config.yaml`.

Its default hook surface must cover:

1. merge conflict markers;
1. line-ending policy violations;
1. final-newline policy;
1. trailing whitespace;
1. YAML syntax and formatting for `.github/**`, `adoption/**`, and `templates/**`;
1. generic Markdown linting through `markdownlint-cli2`.

This check exists outside the CLI so generic Markdown, YAML, and text-file
hygiene is version-pinned once in `pre-commit` rather than duplicated inside
`arckit`.

### 5.3 arckit Tool Check

When `arckit/**` or workflow definitions change, the pipeline must also validate the
tool itself.

The required tool checks are:

1. `gofmt -s` verification;
1. `go vet ./...`;
1. `go test ./...`;
1. `go build ./cmd/arckit`.

When no relevant files changed, the check may complete as a no-op, but its check
name must remain stable for branch protection.

### 5.4 ARC Process Check

The repository must enforce GitHub-native ARC process requirements separately from
`arckit`.

For a pull request that introduces a new numbered ARC file in `ARCs/arc-####.md`,
the process check must:

1. extract the ARC number from the filename;
1. require an existing tracking issue in this repository for that ARC number;
1. require that the tracking issue was opened from the ARC tracking issue template;
1. require that the tracking issue carries the `arc-tracking` label;
1. require that the PR body explicitly references the tracking issue.

For pull requests that materially change ARC status or gate readiness, the process
check must also verify that the corresponding tracking issue still exists and is
referenced.

The process check must fail when the tracking issue is missing, malformed, unlabeled,
or unreferenced.

### 5.5 Tracking Issue Policy

The ARC process remains:

1. proposal discussion begins in GitHub Discussions;
1. an ARC number is assigned at the end of the discussion phase;
1. once the number is assigned, the ARC author opens the tracking issue using the
   repository tracking issue template;
1. only after that may the numbered ARC PR proceed.

Tracking issue creation is a required author action, not an automatic workflow action.

### 5.6 Online PR Validation

PR validation should also include:

```text
pre-commit run lychee --hook-stage manual
```

This check is assistive only.

It may report:

1. external link failures;
1. network-specific failures;
1. `pre-commit` or hook-environment failures that reduce online coverage.

It must not be the canonical blocking merge gate.

## 6. Monthly Maintenance Audit

### 6.1 Purpose

The repository must run a monthly maintenance audit so CI helps maintain existing
ARCs, not only review new changes.

The audit must run:

1. on a monthly schedule;
1. by manual dispatch.

### 6.2 Monthly Audit Scope

The monthly audit must review:

1. offline repository validation;
1. online repository validation;
1. ARC-to-tracking-issue process consistency;
1. inactivity across canonical ARC records.

### 6.3 Deterministic Findings

Deterministic maintenance findings include:

1. offline `arckit` validation failures;
1. missing or invalid vetted adopters registry;
1. missing or invalid required adoption summaries;
1. missing canonical `implementation-url` or `implementation-maintainer` declarations for implementation-required ARCs;
1. non-canonical sponsor-specific implementation repository URLs for implementation-required ARCs;
1. non-canonical reference implementation statuses in adoption summaries;
1. `medium` or `high` adoption readiness declared without enough tracked adopters;
1. `Final` ARCs whose adoption summaries have no tracked adopters;
1. missing local links or asset targets;
1. ARC and tracking issue mismatches that are machine-checkable;
1. missing required tracking issues;
1. missing required PR references;
1. missing required transition metadata that is deterministically enforceable.

These findings require a concrete call to action in the monthly report.

### 6.4 Advisory Online Findings

Online maintenance findings include external link failures discovered by online validation.

These findings must be reported separately from deterministic repository failures and
must not, by themselves, imply a required status transition.

### 6.5 Monthly Report Output

Each monthly run must always produce a workflow summary report.

When action is required, the pipeline must create or update one rollup issue for the
current calendar month, named in the form:

```text
Monthly ARC maintenance report YYYY-MM
```

Re-runs within the same month must update that issue rather than open duplicates.

The rollup issue must contain these sections:

1. `Action required from ARC authors`
1. `Action required from ARC editor`
1. `Suggested status transitions`

Each ARC entry should include:

1. the ARC number and title;
1. the reason it appears in the report;
1. the deterministic evidence or advisory signal;
1. a concrete next step.

When author handles can be parsed from ARC front matter, including YAML sequence-valued
author fields, the report should mention them.

## 7. Inactivity and Status Suggestions

### 7.1 Inactivity Signal

Inactivity must be evaluated from canonical GitHub and repository activity:

1. latest activity on the ARC tracking issue;
1. latest activity on the canonical ARC PR;
1. latest commit touching the ARC file or adoption file.

### 7.2 Suggested `Stagnant`

If an ARC is in `Draft`, `Review`, or `Last Call` and shows no canonical activity
for six months or more, the monthly maintenance report should suggest moving it to
`Stagnant`.

### 7.3 Suggested `Idle`

If an ARC is in `Final` and shows no canonical maintenance activity for six months
or more, the monthly maintenance report should suggest editor review for moving it
to `Idle`.

### 7.4 Existing `Stagnant`

If an ARC is already `Stagnant` and still has no activity after one additional month,
the monthly maintenance report should flag it for editor follow-up under the ARC-0000
process.

### 7.5 Editorial Authority

These outcomes are suggestions only.

The pipeline must not change ARC status automatically. Editors retain responsibility
for the actual transition decision.

## 8. Release Specification

### 8.1 Tag Format

`arckit` releases must be created from tags shaped like:

```text
arckit/vX.Y.Z
```

This keeps releases aligned with the `arckit/` module boundary.

### 8.2 Release Trigger

The release pipeline must trigger only from `arckit/v*` tags created from `main`.

### 8.3 Release Outputs

The release process must build archives for:

1. Linux amd64
1. Linux arm64
1. macOS amd64
1. macOS arm64
1. Windows amd64
1. Windows arm64

The release must publish:

1. archived binaries;
1. a `SHA256SUMS` file;
1. a GitHub Release entry.

GitHub-generated release notes may be used.

### 8.4 Explicit Non-Goals

The v1 release process must not require:

1. GoReleaser;
1. signing;
1. provenance;
1. Homebrew automation;
1. package-manager distribution.

## 9. GitHub Automation

### 9.1 PR Auto-Labeling

PR auto-labeling should be path-based first:

1. `ARCs/**` -> `area:arc`
1. `adoption/**` -> `area:adoption`
1. `arckit/**` -> `area:arckit`
1. `.github/**` -> `area:github`

PR stage labels should be derived from the selected PR template or equivalent PR body
signal, and add exactly one of:

1. `kind:draft`
1. `kind:review`
1. `kind:last-call`
1. `kind:final`
1. `kind:adoption`
1. `kind:editorial`

### 9.2 Issue Routing

Repository issue routing should remain simple:

1. new ARC proposals are redirected to Pre-ARC Discussions;
1. numbered ARCs use the tracking issue template;
1. non-ARC operational issues may use minimal bug or process issue forms;
1. blank issues should be disabled.

### 9.3 Safety Boundary

Jobs that write labels, comments, or maintenance issues must not execute untrusted
PR code.

## 10. Security and Pinning Policy

The CI/CD implementation that follows this specification must:

1. pin GitHub Actions to full commit SHAs;
1. pin Go to exact version `1.26.1` where Go is required;
1. default workflow permissions to read-only;
1. widen permissions only for jobs that must label, comment, or publish releases.

## 11. Deferred for Later

The following are explicitly out of scope for this specification:

1. stale-issue automation;
1. automatic status changes;
1. automatic tracking issue creation;
1. multi-version Go matrices;
1. signing and provenance;
1. broader linter suites beyond the defined hygiene and `arckit` checks.

## 12. Acceptance Criteria

This specification is satisfied when the repository CI/CD design provides all of the
following behaviors:

1. PRs receive stable required checks for hygiene, offline validation, ARC process
   enforcement, and `arckit` tool validation.
1. New numbered ARC PRs fail when the matching tracking issue does not exist or is
   not correctly referenced.
1. Online validation runs on PRs, but does not become the canonical merge gate.
1. A monthly maintenance review produces a workflow summary every month.
1. Actionable maintenance findings create or update one monthly rollup issue with
   author and editor call to action.
1. Six-month inactivity causes a suggested `Stagnant` or `Idle` outcome in the monthly
   report according to ARC status.
1. `arckit` binary releases are defined from `arckit/vX.Y.Z` tags with multi-platform
   archives and checksums.
