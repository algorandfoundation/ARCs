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
- link to the adoption summary artifact,
- confirmation that any newly named adopters were added to `adoption/vetted-adopters.yaml`,
- confirmation that a `Final` ARC still has at least one tracked adopter in its adoption summary.

The ARC front matter is the canonical place for `status`, `sponsor`,
`implementation-required`, `implementation-url`, and
`implementation-maintainer`. Adoption summaries track implementation status and
adoption evidence, not ARC-owned identity metadata.

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

The repository CI/CD and release policy is defined in
[`.github/ci-cd-release-specs.md`](.github/ci-cd-release-specs.md).

### ARC-Kit

[`arckit`](./arckit/README.md) is the ARC repository validator and scaffolding CLI
used by the pull request workflows for ARC-specific validation only.

Generic repository hygiene is handled separately through the repository-root
`.pre-commit-config.yaml`. That shared hook config owns Markdown linting,
whitespace/newline checks, YAML syntax/formatting, and advisory external link
reachability checks.

`arckit` owns ARC-specific metadata, section, reference, and body-link rules,
including rejection of absolute links back into repository content such as ARCs
or assets. External raw HTML anchors are allowed.

The canonical offline repository gate is:

```text
arckit validate repo .
```

This is the authoritative machine validation check for ARC repository artifacts.

When present, the repository-root `.arckit.jsonc` file is applied automatically by
all `arckit validate ...` commands. It supports repo-local suppressions for:

- `ignoreArcs`: skip an ARC number across its ARC, adoption, and asset footprint;
- `ignoreRules`: skip a rule everywhere;
- `ignoreByArc`: skip rules for exact ARC numbers or inclusive ARC ranges.

Invalid `.arckit.jsonc` content fails validation.

During staged ARC validation migrations, `.arckit.jsonc` may temporarily suppress
rules for historical ARC files. The PR workflow still re-enforces `R:021` and
`R:031` through `R:038` for changed ARC Markdown files before merge.

Example:

```jsonc
{
  "ignoreArcs": [42],
  "ignoreRules": ["R:020"],
  "ignoreByArc": {
    "43": ["R:009", "R:013"],
    "50-60": ["R:011"]
  }
}
```

To run the same validation locally, use both layers:

```sh
pre-commit run --all-files
cd arckit
go build ./cmd/arckit
go run ./cmd/arckit validate repo ..
```

The default `pre-commit` hooks are fail-only. They report hygiene violations but
do not rewrite files.

If you want hook-managed autofix locally, run the manual-stage fixers explicitly
and then rerun `pre-commit run --all-files`:

```sh
pre-commit run mixed-line-ending-fix --all-files --hook-stage manual
pre-commit run end-of-file-fix --all-files --hook-stage manual
pre-commit run trailing-whitespace-fix --all-files --hook-stage manual
pre-commit run yamlfmt-fix --all-files --hook-stage manual
```

Useful related commands:

```sh
pre-commit run lychee --all-files --hook-stage manual
cd arckit
go run ./cmd/arckit validate arc ../ARCs/arc-0000.md
go run ./cmd/arckit validate links ../ARCs/arc-0000.md
```

If your pull request changes `arckit/**`, also run the tool validation checks:

```sh
cd arckit
find . -name '*.go' -print0 | xargs -0 gofmt -w -s
go vet ./...
go test ./...
go build ./cmd/arckit
```
