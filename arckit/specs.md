# ARC Kit CLI Specification

## 1. Purpose

`arckit` is the canonical local and CI validator for the ARC repository.

It is intended to be used both:

1. locally by authors, editors, and maintainers while drafting or updating ARC artifacts;
1. in repository CI to enforce repository rules, formatting, structural integrity,
and machine-verifiable process gates.

`arckit` exists to make the ARC process deterministic where it can be deterministic,
while clearly separating machine-verifiable checks from human editorial judgment.

## 2. Scope

`arckit` is responsible for:

1. formatting ARC Markdown files;
1. linting and validating ARC Markdown files against ARC repository rules;
1. validating ARC adoption summary YAML files;
1. validating cross-artifact consistency across the repository;
1. validating the machine-verifiable subset of ARC status-transition gates;
1. scaffolding new ARC artifacts;
1. producing diagnostics suitable for both humans and automation.

`arckit` is not responsible for:

1. creating or modifying GitHub Discussions, Issues, or Pull Requests;
1. deciding whether a proposal is technically sound or desirable;
1. deciding whether adoption is sufficient beyond the machine-verifiable evidence
recorded in repository artifacts;
1. replacing editor judgment where ARC-0 requires human review.

## 3. Design Principles

The CLI **MUST** satisfy the following principles:

1. **Repository-first**. The repository contents are the source of truth.
1. **Offline-first**. Core validation **MUST** work without network access.
1. **Deterministic**. Identical inputs **MUST** produce identical outputs.
1. **Explainable**. Every failing rule **MUST** produce a stable rule identifier, 
severity, message, and remediation hint.
1. **Minimal surprise**. Safe formatting fixes **MAY** be automatic; semantic changes
**MUST NOT** be automatic.
1. **CI-compatible**. The CLI **MUST** support machine-readable output and stable
exit codes.
1. **Migration-aware**. The CLI **SHOULD** support compatibility modes when ARC-0
process text and repository metadata are temporarily out of sync.

## 4. Repository Model

The CLI assumes the following canonical artifact model:

1. a numbered ARC Markdown document in `ARCs/arc-####.md`;
1. a numbered adoption summary YAML document in `adoption/arc-####.yaml`;
1. optional auxiliary files in `assets/arc-####/`;
1. a canonical tracking issue for each numbered ARC;
1. a pre-ARC discussion before ARC numbering.

The CLI validates repository artifacts only. It does not query or mutate canonical
GitHub records unless explicitly run in online mode.

## 5. Operating Modes

### 5.1 Offline Mode

Offline mode is the default.

In offline mode, the CLI **MUST NOT** perform network requests. It validates only
the contents of local files and repository structure.

Offline mode **MUST** support all formatting checks, Markdown checks, YAML schema
checks, cross-file consistency checks, and transition checks that rely only on the
repository state.

### 5.2 Online Mode

Online mode is optional and opt-in.

In online mode, the CLI **MAY** validate network-dependent properties, including:

1. whether a GitHub user or organization exists;
1. whether a GitHub repository exists and is public;
1. whether a `discussions-to` or implementation URL resolves;
1. whether remote GitHub resource types match expectations.

A failed network lookup in online mode **MUST** be distinguishable from a semantic
validation failure.

## 6. Command-Line Interface

The CLI executable name **SHOULD** be `arckit`.

### 6.1 Commands

The CLI **MUST** provide the following commands:

```text
arckit fmt <path...>
arckit validate arc <arc-file>
arckit validate adoption <adoption-file>
arckit validate repo [repo-root]
arckit validate transition <arc-file> --to <status>
arckit init arc --number <n> --title <title> --type <type> --sponsor <sponsor>
arckit rules
arckit explain <rule-id>
```

### 6.2 Common Options

All commands **SHOULD** support the following common options where applicable:

- `--config <path>`: path to configuration file.
- `--offline`: force offline mode.
- `--online`: enable online mode.
- `--format <text|json|sarif>`: diagnostic output format.
- `--quiet`: suppress non-error output.
- `--verbose`: emit detailed execution context.
- `--severity <minimum>`: minimum severity to emit.

### 6.3 Exit Codes

The CLI **MUST** use stable process exit codes:

- `0`: success, no validation errors at or above the selected severity threshold.
- `1`: validation failed.
- `2`: invalid invocation, configuration error, or runtime failure.

## 8. Diagnostics Model

Each diagnostic emitted by the CLI **MUST** include:

1. rule identifier;
1. severity;
1. short title;
1. human-readable message;
1. remediation hint;
1. source location, if applicable.

### 8.1 Rule Identifier Format

Rules are identified by stable IDs in the form `R:NNN`, where `NNN` is a zero-padded
integer in the range `001` to `999`.

### 8.2 Rule Ranges

The CLI **SHOULD** reserve rule ranges by domain:

- `R:001-R:099`: front-matter structure and formatting
- `R:100-R:199`: front-matter semantics
- `R:200-R:299`: body structure and style
- `R:300-R:399`: link and reference rules
- `R:400-R:499`: adoption summary rules
- `R:500-R:599`: cross-artifact repository rules
- `R:600-R:699`: status-transition rules
- `R:700-R:799`: initialization and scaffolding rules
- `R:800-R:899`: online-only integrity rules
- `R:900-R:999`: internal/configuration/runtime rules

### 8.3 Severity Levels

The CLI **MUST** support these severities:

- `error`
- `warning`
- `info`

Repository CI **SHOULD** fail on `error` diagnostics.

## 9. Formatting Specification

`arckit fmt` is responsible for safe formatting only.

### 9.1 Scope of Formatting

`arckit fmt` **MAY** automatically fix:

1. front-matter spacing normalization;
1. front-matter fields ordering;
1. front-matter field value normalization;

`arckit fmt` **MUST NOT** automatically change:

1. header semantics;
1. link destinations;
1. section order when semantic ambiguity exists;
1. ARC numbering;
1. adoption data;
1. any metadata value other than whitespace normalization.

### 9.2 Relationship with External Tools

The formatter integrates with `markdownlint` and other formatters, but `arckit`
remains the canonical interface presented to local users and CI.

Thr `arckit fmt` **MUST** invoke a pinned `markdownlint-cli2` autofix pass before
applying ARC-specific safe fixes.

The `arckit validate` **MUST** invoke the same `markdownlint` toolchain in check
mode and then run ARC-specific validations.

## 10. ARC Markdown Validation

`arckit validate arc` validates a single ARC Markdown file in isolation.

### 10.1 File Naming

ARC files **MUST** be named `arc-####.md`, where `####` is the zero-padded ARC number.

Examples:

- `arc-0000.md`
- `arc-0042.md`
- `arc-1234.md`

### 10.2 Front-Matter Block

An ARC file **MUST** begin with a YAML front-matter block delimited by `---`.

Unknown headers are invalid unless explicitly allowed by configuration.

Duplicate headers are invalid.

### 10.3 Recognized Front-Matter Fields

The CLI **MUST** recognize the following fields in the following order:

1. `arc`
1. `title`
1. `description`
1. `author`
1. `discussions-to`
1. `status`
1. `last-call-deadline`
1. `type`
1. `category`
1. `sub-category`
1. `created`
1. `updated`
1. `requires`
1. `supersedes`
1. `superseded-by`
1. `extends`
1. `extended-by`
1. `sponsor`
1. `implementation-required`
1. `implementation-url`
1. `implementation-maintainer`
1. `adoption-summary`
1. `idle-since`

Fields not present in a document are omitted; ordering rules apply to fields that
are present.

### 10.4 Header Formatting Rules

The CLI **MUST** enforce:

1. exactly one space after `:` in each header line;
1. no leading or trailing whitespace in header values;
1. comma-separated list fields with exactly one space after each comma;
1. no empty list items;
1. no duplicated numeric list items;
1. no duplicated string list items unless explicitly allowed by a future profile.

### 10.5 Field Semantics

#### 10.5.1 `arc`

- **Type**: integer
- **Constraints**: non-negative
- **Additional rule**: when the filename is `arc-####.md`, the numeric value **MUST**
match the filename number.

#### 10.5.2 `title`

- **Type**: string
- **Constraints**: 2 to 44 characters inclusive
- **Rules**:
  - must not contain the ARC number in any form;
  - must not begin or end with whitespace.

The CLI **SHOULD NOT** reject a title merely because it contains words like `standard`,
since that is stylistic rather than machine-verifiable.

#### 10.5.3 `description`

- **Type**: string
- **Constraints**: 2 to 140 characters inclusive
- **Rules**:
  - must not contain the ARC number in any form;
  - must not begin or end with whitespace.

#### 10.5.4 `author`

- **Type**: comma-separated list
- **Allowed entry forms**:
  - `Full Name (@handle)`
  - `Full Name <email@example.com> (@handle)`
  - `@handle`

The CLI **MUST** enforce:

1. at least one entry includes a GitHub handle;
1. every GitHub handle begins with `@`;
1. handles contain only characters valid in GitHub usernames;
1. email addresses, when present, are syntactically valid.

In online mode, the CLI **MAY** verify that referenced GitHub handles exist.

#### 10.5.5 `discussions-to`

- **Type**: URL
- **Constraints**: valid absolute HTTPS URL

The CLI **MUST** apply state-aware validation:

1. `discussions-to` **SHOULD** point to the canonical discussion URL;
1. `discussions-to` **MUST NOT** point to a pull request.

#### 10.5.6 `status`

Allowed values are:

- `Draft`
- `Review`
- `Last Call`
- `Final`
- `Stagnant`
- `Withdrawn`
- `Idle`
- `Deprecated`
- `Living`

The `status` field **MUST** contain exactly one value.

#### 10.5.7 `last-call-deadline`

- **Type**: date
- **Required** only when `status` is `Last Call`
- **Forbidden** otherwise

#### 10.5.8 `type`

Allowed values are:

- `Standards Track`
- `Meta`

The `type` field **MUST** contain exactly one value.

#### 10.5.9 `category`

`category` is required only when `type` is `Standards Track` and forbidden otherwise.

Allowed values are:

- `Interface`
- `Data`
- `Cryptography`
- `Protocol`
- `Governance`

The `category` field **MUST** contain at most one value.

#### 10.5.10 `sub-category`

`sub-category` is optional.

Allowed values are:

- `General`
- `ASA`
- `Application`
- `LSig`
- `Event`
- `Library`
- `Identity`
- `Explorer`
- `Wallet`

Unless a future profile explicitly enables multi-valued classification, `sub-category`
**MUST** contain at most one value.

#### 10.5.11 Date Fields

The following fields are dates or date lists:

- `created`
- `updated`
- `last-call-deadline`
- `idle-since`

The CLI **MUST** enforce ISO 8601 calendar-date format: `YYYY-MM-DD`.

The CLI **MUST** also enforce chronological consistency where machine-verifiable:

1. `updated` dates, when present, must be greater than or equal to `created`;
1. `last-call-deadline`, when present, must be greater than or equal to `created`;
1. `idle-since`, when present, must be greater than or equal to `created`.

#### 10.5.12 ARC Relationship Fields

The following fields are numeric ARC reference lists:

- `requires`
- `supersedes`
- `superseded-by`
- `extends`
- `extended-by`

The CLI **MUST** enforce:

1. integers only;
1. ascending numeric order;
1. no duplicates.

Cross-file reciprocity for these fields is validated by `validate repo`, not by
`validate arc`.

#### 10.5.13 `sponsor`

Allowed values are:

- `Foundation`
- `Ecosystem`

#### 10.5.14 `implementation-required`

Allowed values are:

- `true`
- `false`

#### 10.5.15 `implementation-url`

- **Type**: URL
- **Required** when `implementation-required` is `true` and the ARC is validated
for `Review` or later
- **Forbidden** only if a repository profile explicitly requires strict absence
when `implementation-required` is `false`

The CLI **MUST** enforce:

1. valid HTTPS URL;
1. GitHub-hosted repository URL;
1. repository name of the form `arc<N>`, where `<N>` is the non-zero-padded ARC
number.

The CLI **MUST** also enforce sponsor-to-organization mapping:

1. `Foundation` sponsor implies `github.com/algorandfoundation/arc<N>`;
1. `Ecosystem` sponsor implies `github.com/algorandecosystem/arc<N>`.

In online mode, the CLI **MAY** verify that the repository exists and is public.

#### 10.5.16 `implementation-maintainer`

- **Type**: comma-separated list of GitHub handles or organization/team references
- **Purpose**: identifies the owner responsible for the reference implementation

If `implementation-required` is `true`, `implementation-mainteiner` **MUST** be
present in ARC files and **MUST** be present for `validate transition --to Review`.

#### 10.5.18 `adoption-summary`

- **Type**: relative path
- **Constraints**:
  - must point to `adoption/arc-####.yaml` or `./adoption/arc-####.yaml` according 
  to repository profile;
  - must match the ARC number.

#### 10.5.19 `idle-since`

- **Type**: date
- **Required** only when `status` is `Idle`
- **Forbidden** otherwise

### 10.6 Conditional Header Rules

The CLI **MUST** enforce:

1. `category` required only for `Standards Track` and forbidden for `Meta`;
1. `last-call-deadline` required only for `Last Call`;
1. `idle-since` required only for `Idle`;
1. `implementation-url` and `implementation-maintainer` required for transition
validation to `Review` or later when `implementation-required` is `true`.

### 10.7 Body Section Requirements

The CLI **MUST** validate level-2 section structure.

#### 10.7.1 Required Level-2 Sections

Every ARC **MUST** contain these sections in this order:

1. `Abstract`
1. `Motivation`
1. `Specification`
1. `Rationale`
1. `Security Considerations`
1. `Copyright`

#### 10.7.2 Conditionally Required Level-2 Sections

The CLI **MUST** require:

1. `Backwards Compatibility` when the front matter or body indicates backward incompatibility;
1. `Reference Implementation` when `implementation-required` is `true`;
1. `Test Cases` when `implementation-required` is `true`.

#### 10.7.3 Allowed Level-2 Sections

Allowed level-2 sections are:

1. `Abstract`
1. `Motivation`
1. `Specification`
1. `Rationale`
1. `Backwards Compatibility`
1. `Reference Implementation`
1. `Test Cases`
1. `Security Considerations`
1. `Copyright`

Unknown level-2 sections are invalid unless allowed by repository configuration.

### 10.8 Link and Reference Rules

The CLI **MUST** enforce the following link rules:

1. ARC-to-ARC links **MUST** be relative;
1. the first body reference to each ARC number **MUST** be a hyperlink;
1. subsequent references to the same ARC number **MAY** be plain text;
1. external links in normative prose **SHOULD** use HTML `<a href="...">...</a>`
form when that is the repository convention;
1. root-relative links are invalid;
1. absolute GitHub blob links to ARC documents are invalid.

### 10.9 Textual Restrictions

The CLI **MUST** enforce only machine-verifiable textual restrictions.

It **MUST** reject:

1. malformed ARC references in `title` or `description`;
1. ARC numbers in `title` or `description`;
1. malformed relative ARC links.

It **SHOULD NOT** reject purely stylistic wording choices unless explicitly configured.

### 10.10 Namespace Rules for Interface/Event Conventions

If the ARC declares `category: Interface` and a supported `sub-category` that requires 
namespacing conventions, the CLI **MAY** validate examples or referenced assets
for ARC-number namespaces such as `arc<N>_methodName` or `arc<N>_eventName`.

This validation is profile-specific and **SHOULD** default to `warning` unless the
repository has standardized these checks.

## 11. Adoption Summary Validation

`arckit validate adoption` validates a single adoption summary YAML file.

### 11.1 File Naming

Adoption summary files **MUST** be named `arc-####.yaml` and reside in the `adoption/`
directory.

### 11.2 Schema

The adoption summary schema **MUST** support at least the following fields:

```yaml
arc: 42
title: Example ARC
status: Review
last-reviewed: 2026-03-26
sponsor: Foundation
implementation-required: true
reference-implementation:
  repository: https://github.com/algorandfoundation/arc42
  owner: @owner
  maintainer:
    - @maintainer1
  status: in_progress
  notes: ""
adoption:
  wallets: []
  explorers: []
  sdk-libraries: []
  indexers-infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
```

### 11.3 Required Fields

The CLI **MUST** require:

1. `arc`
1. `title`
1. `status`
1. `last-reviewed`
1. `sponsor`
1. `implementation-required`
1. `reference-implementation`

### 11.4 `reference-implementation` Rules

The CLI **MUST** enforce:

1. `repository` present when `implementation-required` is `true`;
1. `maintainer` present when `implementation-required` is `true`;
1. `status` value in:
   - `planned`
   - `in_progress`
   - `testable`
   - `shipped`
   - `archived`

### 11.5 Adoption Actor Classes

The `adoption` section **MUST** support these actor classes:

1. `wallets`
1. `explorers`
1. `sdk-libraries`
1. `infra`
1. `dapps-protocols`

Each actor entry **MUST** support:

1. `name`
1. `status`
1. `evidence`
1. `notes`

Allowed actor `status` values are:

- `planned`
- `in_progress`
- `shipped`
- `declined`
- `unknown`

### 11.7 Internal Consistency

The CLI **MUST** enforce:

1. `arc` matches the numeric identifier in the filename;
1. `status` is a valid ARC status;
1. `sponsor` matches the ARC file sponsor when both files are available;
1. `implementation-required` matches the ARC file when both files are available.

## 12. Repository-Wide Validation

`arckit validate repo` validates the repository as a whole.

### 12.1 Repository Layout Rules

The CLI **MUST** validate the following layout invariants:

1. ARC files reside in `ARCs/`;
1. adoption summaries reside in `adoption/`;
1. auxiliary ARC assets, when present, reside in `assets/arc-####/`;
1. no orphaned adoption summaries exist for missing ARC files unless explicitly allowed;
1. no duplicate ARC numbers exist across files.

### 12.2 ARC-to-Adoption Mapping

For every numbered ARC, the CLI **MUST** enforce:

1. exactly one matching adoption summary exists;
1. the ARC front matter `adoption-summary` points to the matching file;
1. the adoption summary `arc` field matches the ARC number.

### 12.3 Relationship Reciprocity

For reciprocal relationship fields, the CLI **MUST** enforce consistent pairs where
both referenced files are present:

1. if ARC A `supersedes` ARC B, then ARC B `superseded-by` **MUST** include ARC A;
1. if ARC A `extends` ARC B, then ARC B `extended-by` **MUST** include ARC A.

### 12.4 Implementation Repository Mapping

When `implementation-required` is `true`, the CLI **MUST** enforce:

1. the implementation repository URL matches the ARC number;
1. sponsor and GitHub organization mapping are consistent;
1. the adoption summary repository URL, if present, matches the ARC front matter
URL.

### 12.5 Tracking-Issue Awareness

Because tracking issues are canonical but live on GitHub, the CLI **SHOULD** support
a repository convention for recording the tracking issue number in a local manifest
or front matter extension.

If such a repository convention is enabled, `validate repo` **MUST** enforce it.

If no local source of truth exists for tracking issue identifiers, the CLI **MUST
NOT** fail offline validation solely because GitHub issue linkage cannot be verified.

## 13. Transition Validation

`arckit validate transition <arc-file> --to <status>` validates the machine-verifiable
subset of a requested status transition.

This command **MUST NOT** claim that a transition is fully approved. It validates
only repository evidence.

### 13.1 Supported Target Statuses

The CLI **MUST** support transition validation to at least:

- `Review`
- `Last Call`
- `Final`
- `Idle`

### 13.2 Transition to `Review`

To validate a transition to `Review`, the CLI **MUST** require:

1. valid ARC front matter;
1. required body sections present;
1. `status` currently `Draft`;
1. `implementation-required` declared;
1. `Security Considerations` present.

If `implementation-required` is `true`, the CLI **MUST** also require:

1. `implementation-url` present and valid;
1. `implementation-maintainer` present;
1. `Reference Implementation` section present;
1. `Test Cases` section present.

The CLI **SHOULD** emit `info` diagnostics for manual editorial checks that remain
outside machine validation.

### 13.3 Transition to `Last Call`

To validate a transition to `Last Call`, the CLI **MUST** require:

1. the ARC satisfies all `Review` machine checks;
1. a valid adoption summary exists;
1. `adoption-summary` path matches the local file;
1. `status` currently `Review` unless profile allows dry-run validation;
1. no missing implementation metadata when implementation is required.

If `implementation-required` is `true`, the CLI **MUST** also require:

1. adoption summary contains a reference implementation block;
1. reference implementation status is one of `in_progress`, `testable`, or `shipped`.

The CLI **SHOULD** emit manual-check reminders for unresolved normative questions
and adoption intent, because those are not fully machine-verifiable.

### 13.4 Transition to `Final`

To validate a transition to `Final`, the CLI **MUST** require:

1. the ARC satisfies all `Last Call` machine checks;
1. `last-call-deadline` present and valid if the ARC is or was in `Last Call`;
1. an adoption summary exists and is valid;
1. adoption summary records at least one non-empty evidence entry.

If `implementation-required` is `true`, the CLI **MUST** also require:

1. reference implementation repository metadata present in both ARC and adoption
summary;
1. reference implementation status is `testable` or `shipped`;
1. at least one adoption actor entry has non-empty evidence.

The CLI **MUST** report that editor approval and qualitative adoption assessment
remain manual.

### 13.5 Transition to `Idle`

To validate a transition to `Idle`, the CLI **MUST** require:

1. current status `Final` unless profile allows dry-run validation;
1. `idle-since` present and valid;
1. adoption summary exists.

The determination that maintenance or adoption has actually dropped is editorial
and **MUST** remain a manual judgment.

## 14. Rules for `init arc`

`arckit init arc` scaffolds a new numbered ARC locally.

### 14.1 Inputs

The command **MUST** support at least:

- `--number <n>`
- `--title <title>`
- `--type <type>`
- `--sponsor <sponsor>`

### 14.2 Outputs

The command **MUST** create:

1. `ARCs/arc-####.md`
1. `adoption/arc-####.yaml`
1. `assets/arc-####/` if requested

### 14.3 Generated Content

Generated ARC content **MUST**:

1. use the repository's ARC templates;
1. initialize required front matter;
1. initialize required level-2 sections;
1. initialize `adoption-summary` with the correct relative path;
1. initialize implementation metadata fields according to selected flags.

Generated adoption content **MUST**:

1. use the repository adoption template;
1. set the numeric ARC identifier;
1. initialize actor-class arrays;
1. initialize summary fields.

`init arc` **MUST NOT** create remote GitHub artifacts.

## 15. Configuration

The CLI **SHOULD** support repository configuration via a file such as `.arckit.yaml`.

### 15.1 Configurable Settings

Configuration **MAY** define:

1. whether unknown headers are allowed;
1. whether additional level-2 sections are allowed;
1. whether online checks are enabled in CI;
1. which path spellings are accepted for `adoption-summary`;
1. output defaults.

### 15.2 Config Precedence

Precedence **SHOULD** be:

1. command-line flags;
1. environment variables;
1. repository config file;
1. built-in defaults.

## 16. Output Formats

The CLI **MUST** support:

1. human-readable text output;
1. JSON output for automation;
1. SARIF output when feasible for code-scanning integration.

### 16.1 JSON Output

JSON output **MUST** include:

1. command metadata;
1. profile;
1. mode (`offline` or `online`);
1. diagnostic list;
1. summary counts by severity;
1. process exit code.

## 17. CI Integration

The CLI is intended to be the repository's primary semantic validator in CI.

### 17.1 Recommended CI Jobs

A repository integrating `arckit` **SHOULD** run at least these jobs:

1. `arckit fmt --check ARCs adoption` or equivalent formatting check;
1. `arckit validate repo .`;
1. `arckit validate arc <changed-arc-files>` for changed ARC files;
1. `arckit validate adoption <changed-adoption-files>` for changed adoption files;
1. `arckit validate transition <arc-file> --to <status>` when a pull request changes
`status`.

### 17.2 Status-Change Detection

The CI integration **SHOULD** detect whether a pull request changes ARC `status`
and, if so, run the appropriate transition validator automatically.

### 17.3 Label and PR Workflow Integration

`arckit` **MAY** expose machine-readable outputs specifically intended to help GitHub
Actions apply minimal labels such as:

- `blocked`
- `decision-needed`

The CLI **MUST NOT** require repository labels for correctness.

## 18. Rule Catalog Requirements

`arckit rules` **MUST** list all known rules.

`arckit explain <rule-id>` **MUST** display:

1. rule identifier;
1. severity default;
1. title;
1. full description;
1. rationale;
1. remediation guidance;
1. whether the rule is auto-fixable.

## 19. Minimal Required Rule Set

A conforming implementation of this specification **MUST** implement rules covering at least:

1. front-matter presence and order;
1. header formatting;
1. required headers;
1. date validity;
1. ARC file naming;
1. adoption file naming;
1. body required sections;
1. relative ARC links;
1. adoption-summary path consistency;
1. sponsor-to-implementation URL consistency;
1. adoption summary schema validity;
1. cross-artifact ARC number consistency;
1. transition validation for `Review`, `Last Call`, and `Final`.

## 20. Non-Machine-Verifiable Checks

The following checks are outside the CLI's authority and **MUST** remain editorial:

1. whether a proposal addresses a real ecosystem need;
1. whether community consensus exists;
1. whether adoption evidence is sufficient in substance, not merely present;
1. whether unresolved design questions are truly resolved;
1. whether a reference implementation is adequate in quality;
1. whether maintenance is meaningfully active.

The CLI **SHOULD** emit informational reminders for such checks when validating transitions.

## 21. Security and Trustworthiness

The CLI **MUST NOT** silently modify semantic content.

Online mode **MUST** clearly identify network-derived validation results and network failures.

The CLI **SHOULD** make it easy for users to distinguish:

1. malformed artifact problems;
1. missing repository evidence problems;
1. advisory warnings about process gaps;
1. transient network failures.

## 22. Conformance

An implementation conforms to this specification if it:

1. provides the required command surface or a compatible equivalent;
1. implements the required validation domains;
1. emits stable rule identifiers and severities;
1. supports offline operation for core validation;
1. supports CI-oriented exit codes and machine-readable output.

## 23. Recommended Initial Backlog

The recommended implementation order is:

1. `fmt`
1. `validate arc`
1. `validate adoption`
1. `validate repo`
1. `validate transition`
1. `init arc`
1. online checks
1. SARIF output

This order prioritizes the checks most useful during authoring and the checks most important for CI enforcement.
