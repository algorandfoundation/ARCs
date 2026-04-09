# arckit Specification

## 1. Purpose

`arckit` is the canonical local and CI validator for the ARC repository.

It exists to make repository validation deterministic where it can be deterministic,
while keeping human editorial judgment clearly out of the tool.

`arckit` is a single Go CLI that must remain useful with only Go installed.
Generic Markdown, YAML, whitespace, and external-link hygiene is handled by the
repository-root `.pre-commit-config.yaml`, not by `arckit`.

## 2. Product Boundary

`arckit` is responsible for:

1. validating ARC Markdown artifacts;
1. validating adoption summary YAML artifacts;
1. validating repository-wide consistency across ARC, adoption, and asset files;
1. validating the machine-verifiable subset of status transitions;
1. performing conservative ARC front matter formatting fixes;
1. scaffolding new ARC artifacts;
1. emitting stable diagnostics for humans and automation;
1. providing one CLI entrypoint for both local use and CI.

`arckit` is not responsible for:

1. querying or mutating GitHub Discussions, Issues, Pull Requests, or labels;
1. deciding whether an ARC is technically sound or desirable;
1. deciding whether adoption evidence is substantively sufficient;
1. deciding whether a reference implementation is high quality;
1. replacing editor judgment or ARC-0 process decisions;
1. owning generic Markdown, YAML, whitespace, or external-link hygiene;
1. owning the devportal workflows in this repository.

## 3. Design Principles

The CLI must follow these principles:

1. **Native-first.** Core behavior must work with just Go and the repository contents.
1. **Repo-local only.** v1 may read an optional repo-root `.arckit.jsonc`, but must not require user-local config, a `--config` flag, or pluggable policy.
1. **Offline-first.** Native validation must work without network access.
1. **Deterministic.** The same inputs must produce the same diagnostics and exit codes.
1. **Explainable.** Each rule failure must include a stable identifier, severity, message, and hint.
1. **Conservative mutation.** Formatting may normalize safe structure, but must never rewrite semantics.
1. **Single gate.** `arckit validate repo .` is the canonical CI validation command.

## 4. Repository Model

`arckit` assumes these repository conventions:

1. ARC documents live in `ARCs/arc-####.md`.
1. Adoption summaries live in `adoption/arc-####.yaml`.
1. The vetted adopters registry lives in `adoption/vetted-adopters.yaml`.
1. ARC assets, when present, live in `assets/arc-####/`.
1. Optional repo-local validation suppressions live in `.arckit.jsonc` at the repository root.
1. Templates live in `templates/`.
1. The repository contents, not remote GitHub state, are the source of truth for machine checks.

The tool validates repository artifacts only. It may emit reminders about GitHub-native
process steps, but it must not depend on GitHub API access in v1.

## 5. Operating Model

`arckit` has one native validation model.

It must:

1. run all ARC, adoption, repository, and transition validations locally;
1. validate local file links and relative ARC links;
1. auto-discover the optional repo-root `.arckit.jsonc` for `validate` commands only;
1. avoid network requests and external tool resolution;
1. keep generic Markdown, YAML, whitespace, and external-link hygiene outside the CLI.

Native validation remains authoritative for repository semantics.

### 5.1 Repo-Local Ignore Configuration

`arckit` may read one optional `.arckit.jsonc` file from the repository root.

Supported keys are:

- `ignoreArcs`
- `ignoreRules`
- `ignoreByArc`

`ignoreArcs` suppresses one ARC number across its ARC, adoption, asset, and
repo-attributed diagnostics.

`ignoreRules` suppresses one rule globally.

`ignoreByArc` suppresses one or more rules for exact ARC selectors like `0` or `43`
or inclusive ARC ranges like `50-60`.

Migration-specific suppressions are allowed when the repository is moving from one
canonical ARC encoding to another, as long as the steady-state target remains the
documented spec.

Invalid `.arckit.jsonc` content must stop validation with exit code `2`.

## 6. Command-Line Interface

The executable name should be `arckit`.

### 6.1 Required Commands

v1 must provide this command surface:

```text
arckit fmt <path...>
arckit validate arc <arc-file>
arckit validate adoption <adoption-file>
arckit validate links <path...>
arckit validate repo [repo-root]
arckit validate transition <arc-file> --to <status>
arckit init arc --number <n> --title <title> --type <type> [--category <category>] [--sub-category <sub-category>] --sponsor <sponsor>
arckit rules
arckit explain <rule-id>
```

### 6.2 Global Flags

The only global flags required in v1 are:

- `--format <text|json>`
- `--quiet`

The CLI must not expose `--config`, `--severity`, profile flags, or SARIF output
in v1. Repo-local `.arckit.jsonc` loading is implicit.

Validation commands may expose `--ignore-config` to bypass repo-local suppressions
entirely and `--enforce-rule <RULE_ID>` to unsuppress a specific rule while
preserving all other config behavior.

### 6.3 Exit Codes

The CLI must use stable exit codes:

- `0`: command succeeded and no `error` diagnostics were emitted;
- `1`: validation completed and at least one `error` diagnostic was emitted;
- `2`: invalid invocation or runtime failure prevented normal validation.

## 7. Diagnostics

Each diagnostic must include:

1. rule identifier;
1. severity;
1. short title;
1. message;
1. remediation hint;
1. source origin (`native`);
1. file path and source position when applicable.

Supported severities are:

- `error`
- `warning`
- `info`

### 7.1 Rule Identifiers

Rules use stable identifiers in the form `R:NNN`.

Recommended ranges:

- `R:001-R:099`: ARC front matter structure and formatting
- `R:100-R:199`: ARC field semantics and section requirements
- `R:200-R:299`: link and reference rules
- `R:300-R:399`: adoption summary rules
- `R:400-R:499`: repository-wide consistency rules
- `R:500-R:599`: transition rules
- `R:900-R:999`: invocation and runtime diagnostics

### 7.2 JSON Output

JSON output must include at least:

1. command name;
1. diagnostics;
1. summary counts by severity;
1. exit code.

## 8. Validation Commands

### 8.1 `validate arc`

`arckit validate arc <arc-file>` validates a single ARC Markdown document.

It must perform:

1. filename validation;
1. front matter parsing and header order validation;
1. required field checks;
1. conditional field checks;
1. body section presence checks;
1. local link and asset link validation.

### 8.2 `validate adoption`

`arckit validate adoption <adoption-file>` validates one adoption summary YAML file.

It must perform:

1. filename validation;
1. schema and required field validation;
1. enum validation;
1. vetted adopters registry validation;
1. internal consistency validation.

### 8.3 `validate links`

`arckit validate links <path...>` validates links discovered in the provided ARC
files, or directories containing ARC files.

It must always validate:

1. relative local links;
1. relative ARC links;
1. asset links;
1. missing-file and invalid-path conditions.

### 8.4 `validate repo`

`arckit validate repo [repo-root]` is the canonical repository-wide validation command.

It must perform:

1. ARC validation for all discovered ARC files;
1. vetted adopters registry validation;
1. adoption summary validation for all discovered adoption files;
1. repository-wide mapping and reciprocity checks;
1. native local link validation.

This is the single canonical CI gate. Required PR validation jobs should build `arckit`
and run `arckit validate repo .`. When present, the repo-root `.arckit.jsonc` is
applied implicitly.

### 8.5 `validate transition`

`arckit validate transition <arc-file> --to <status>` validates the machine-verifiable
subset of a requested status transition.

Supported target statuses are:

- `Review`
- `Last Call`
- `Final`
- `Idle`

This command must never claim that a transition is fully approved. It validates only
repository evidence and must emit `info` reminders for required manual checks.

## 9. ARC Markdown Rules

### 9.1 File Naming

ARC files must be named `arc-####.md` and must live in `ARCs/`.

The numeric identifier in the filename must match the `arc` front matter field.

### 9.2 Front Matter

ARC files must begin with a YAML front matter block delimited by `---`.

The front matter must decode to one top-level YAML mapping.

Recognized fields, in required order when present, are:

1. `arc`
1. `title`
1. `description`
1. `author`
1. `discussions-to`
1. `status`
1. `type`
1. `category`
1. `sub-category`
1. `created`
1. `updated`
1. `sponsor`
1. `implementation-required`
1. `implementation-url`
1. `implementation-maintainer`
1. `adoption-summary`
1. `last-call-deadline`
1. `idle-since`
1. `requires`
1. `supersedes`
1. `superseded-by`
1. `extends`
1. `extended-by`

Unknown top-level front matter fields are not allowed in v1.

Canonical YAML field shapes:

1. `author`, `updated`, `implementation-maintainer`, `requires`, `supersedes`, `extends`, and `extended-by` must use YAML sequences.
1. `superseded-by` must use a scalar ARC number.
1. `implementation-required` must use a YAML boolean.
1. date fields remain scalar values in `YYYY-MM-DD` form.

### 9.3 Required ARC Fields

Each ARC file must include:

1. `arc`
1. `title`
1. `description`
1. `author`
1. `discussions-to`
1. `status`
1. `type`
1. `created`
1. `sponsor`
1. `implementation-required`

Field requirements:

1. `title` must not include the ARC number.
1. `description` must be a short summary and must not include the ARC number.
1. `created`, `updated`, `last-call-deadline`, and `idle-since` must use `YYYY-MM-DD`.
1. `status` must be one of `Draft`, `Review`, `Last Call`, `Final`, `Stagnant`, `Withdrawn`, `Idle`, `Deprecated`, or `Living`.
1. `type` must be one of `Standards Track` or `Meta`.
1. `sponsor` must be one of `Foundation` or `Ecosystem`.
1. `implementation-required` must be `true` or `false`.
1. `adoption-summary`, when present, must be a relative path under `adoption/`.
1. list-valued ARC metadata must use canonical YAML sequences rather than comma-separated scalars.

### 9.4 Conditional ARC Fields

These conditional rules apply:

1. `implementation-url` and `implementation-maintainer` are required when validating
   transition to `Review`, `Last Call`, or `Final` for an ARC with `implementation-required: true`.
1. `adoption-summary` is required when `status` is `Last Call`, `Final`, `Idle`, or `Deprecated`.
1. `last-call-deadline` is required when `status` is `Last Call` and when validating
   transition to `Final`.
1. `idle-since` is required when `status` is `Idle`.

### 9.5 Required Body Sections

Every ARC file must contain these level-2 sections:

1. `Abstract`
1. `Motivation`
1. `Specification`
1. `Rationale`
1. `Security Considerations`

When `implementation-required: true`, transition validation to `Review` or later also
requires:

1. `Reference Implementation`
1. `Test Cases`

### 9.6 Link and Asset Rules

`arckit` must enforce these repository-semantic link rules:

1. repo-local links must be relative, not root-relative;
1. local link targets must exist;
1. ARC-to-ARC links must target `ARCs/arc-####.md`;
1. asset links inside an ARC must stay under the matching `assets/arc-####/` subtree;
1. the `adoption-summary` field, when present, must resolve to an existing file when
   the adoption summary is required for that ARC status.

External link reachability is handled outside `arckit` by the shared `lychee`
hook in the repository-root `pre-commit` configuration.

## 10. Adoption Summary Rules

### 10.1 File Naming

Adoption summary files must be named `arc-####.yaml` and must live in `adoption/`.

The repository must also contain `adoption/vetted-adopters.yaml`.

### 10.2 Requirement Policy

Adoption summaries are not a blanket requirement for every historical ARC in v1.

They are required when:

1. an ARC is in `Last Call`, `Final`, `Idle`, or `Deprecated`; or
1. transition validation to `Last Call`, `Final`, or `Idle` is being requested.

For all other ARC states, adoption summaries are optional, but if present they must validate.

`arckit init arc` must always generate an adoption summary stub so the repository moves
toward a consistent model without breaking older content.

### 10.3 Required Schema

An adoption summary must support at least this shape:

```yaml
arc: 42
title: Example ARC
status: Review
last-reviewed: 2026-03-26
sponsor: Foundation
implementation-required: true
reference-implementation:
  repository: https://github.com/algorandfoundation/arc42
  maintainers:
    - "@maintainer1"
  status: in_progress
  notes: ""
adoption:
  wallets: []
  explorers: []
  sdk-libraries: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
```

Required top-level fields are:

1. `arc`
1. `title`
1. `status`
1. `last-reviewed`
1. `sponsor`
1. `implementation-required`
1. `adoption`
1. `summary`

`reference-implementation` is required when `implementation-required` is `true`.

### 10.4 Adoption Enums and Entry Shape

Allowed `reference-implementation.status` values are:

- `planned`
- `in_progress`
- `testable`
- `shipped`
- `archived`

The `adoption` section must support these actor classes:

1. `wallets`
1. `explorers`
1. `sdk-libraries`
1. `infra`
1. `dapps-protocols`

Each actor entry must support:

1. `name`
1. `status`
1. `evidence`
1. `notes`

Each actor `name` must:

1. be lower-kebab-case;
1. match an entry in the same category of `adoption/vetted-adopters.yaml`.

Allowed actor `status` values are:

- `planned`
- `in_progress`
- `shipped`
- `declined`
- `unknown`

Allowed `summary.adoption-readiness` values are:

- `low`
- `medium`
- `high`

### 10.5 Internal Consistency

`arckit` must enforce:

1. the filename ARC number matches `arc`;
1. `status` is a valid ARC status;
1. `sponsor` matches the ARC file when both are present;
1. `implementation-required` matches the ARC file when both are present;
1. `reference-implementation.repository`, when present, matches `implementation-url`
   in the ARC file.

## 11. Repository-Wide Rules

`validate repo` must enforce at least:

1. no duplicate ARC numbers across ARC files;
1. one valid `adoption/vetted-adopters.yaml` file exists;
1. no duplicate ARC numbers across adoption files;
1. no orphaned adoption summaries for missing ARC files;
1. no orphaned asset trees for missing ARC files;
1. adoption actor names are present in the matching vetted-adopter category;
1. ARC `adoption-summary` fields, when required, point to the matching adoption file;
1. reciprocal relationship fields are consistent where both files exist:
   - `supersedes` <-> `superseded-by`
   - `extends` <-> `extended-by`

## 12. Transition Rules

### 12.1 Transition to `Review`

The CLI must require:

1. a valid ARC file;
1. current `status: Draft`;
1. all required ARC sections;
1. `Security Considerations` present;
1. `implementation-required` explicitly declared.

If `implementation-required: true`, the CLI must also require:

1. `implementation-url`;
1. `implementation-maintainer`;
1. `Reference Implementation` section;
1. `Test Cases` section.

### 12.2 Transition to `Last Call`

The CLI must require:

1. all machine checks for `Review`;
1. current `status: Review`;
1. a valid adoption summary exists;
1. the ARC `adoption-summary` field points to that file.

If `implementation-required: true`, the CLI must also require:

1. a `reference-implementation` block in the adoption summary;
1. `reference-implementation.status` is one of `in_progress`, `testable`, or `shipped`.

### 12.3 Transition to `Final`

The CLI must require:

1. all machine checks for `Last Call`;
1. current `status: Last Call`;
1. `last-call-deadline` present and valid;
1. the adoption summary contains at least one non-empty evidence entry.

If `implementation-required: true`, the CLI must also require:

1. reference implementation metadata in both ARC and adoption summary;
1. `reference-implementation.status` is `testable` or `shipped`;
1. at least one adoption actor entry has non-empty evidence.

### 12.4 Transition to `Idle`

The CLI must require:

1. current `status: Final`;
1. `idle-since` present and valid;
1. a valid adoption summary exists.

### 12.5 Manual Checks

For every transition command, `arckit` should emit `info` reminders for checks that
remain editorial, including:

1. consensus and dissent handling;
1. adequacy of adoption evidence;
1. quality and maintenance of the reference implementation;
1. editor approval.

## 13. Formatting

`arckit fmt <path...>` is intentionally narrow in v1.

It must:

1. normalize front matter spacing;
1. normalize front matter field ordering;
1. normalize canonical YAML sequence fields without coercing invalid scalar-list legacy encodings;
1. preserve semantic content.
1. leave body whitespace, final-newline policy, and generic Markdown hygiene to
   the repository-root `pre-commit` hooks.

## 14. Scaffolding

`arckit init arc` scaffolds a new ARC locally.

Required inputs:

1. `--number`
1. `--title`
1. `--type`
1. `--sponsor`

The command should also accept optional authoring metadata such as `--author`,
`--description`, and `--implementation-required`.

Outputs:

1. `ARCs/arc-####.md`
1. `adoption/arc-####.yaml`
1. `adoption/vetted-adopters.yaml`, when it does not already exist
1. `assets/arc-####/`

The generated ARC must include an `adoption-summary` field pointing to the generated
adoption stub, and must emit canonical YAML-native list fields for author metadata.

`init arc` must never create remote GitHub artifacts.

## 15. Auxiliary Commands

### 15.1 `rules`

`arckit rules` must list all known rules, including:

1. rule identifier;
1. default severity;
1. title;
1. whether the rule is auto-fixable.

### 15.2 `explain`

`arckit explain <rule-id>` must display:

1. rule identifier;
1. title;
1. severity;
1. description;
1. rationale;
1. remediation guidance;
1. whether the rule can be auto-fixed.

## 16. CI and Release Contract

The repository guidance for v1 is:

1. a root `pre-commit` configuration defines the canonical generic hygiene surface
   and pins the generic tool versions;
1. a PR hygiene workflow runs `pre-commit` on the PR diff;
1. a PR tool workflow runs `gofmt -s` checks, `go vet`, `go test`, and `go build`
   when `arckit/**` changes;
1. a PR repository-validation workflow builds `arckit` and runs `arckit validate repo .`
   when ARC, adoption, template, or tooling files change;
1. a scheduled or manually triggered online workflow runs the shared `lychee`
   hook from `pre-commit` for advisory external link checks;
1. CI pins one exact Go patch release and pins all GitHub Actions to full commit SHAs;
1. releases are produced from tags of the form `arckit/vX.Y.Z` and publish archives plus
   SHA256 checksums.

## 17. Out of Scope for v1

The following are intentionally out of scope:

1. `.arckit.yaml`, user-local configuration, or an explicit `--config` flag;
1. generic Markdown, YAML, whitespace, or external-link hygiene inside `arckit`;
1. GitHub API lookups and repository mutation;
1. SARIF output;
1. multi-version required CI matrices;
1. GoReleaser, Viper, and pluggable rule engines.

## 18. Conformance

An implementation conforms to this specification if it:

1. provides the required command surface or a compatible equivalent;
1. performs native offline validation of ARC, adoption, repository, and transition rules;
1. emits stable diagnostics and exit codes;
1. supports the optional repo-root `.arckit.jsonc` ignore model without adding user-local config;
1. keeps generic hygiene outside the CLI and ARC semantics inside the CLI;
1. keeps CI centered on `arckit validate repo .`;
1. does not require extra local tooling beyond Go for core validation.
