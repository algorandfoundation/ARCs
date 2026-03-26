# arckit Tech Stack

## 1. Purpose

This document recommends the implementation stack for `arckit` as a Go CLI living
in a monorepo that also contains the ARC documents and supporting repository artifacts.

The recommendation is optimized for these constraints:

- one lightweight local CLI for authors and CI;
- minimal required dependencies for contributors;
- optional use of external backends (`markdownlint-cli2` and `lychee`) only when
available via the local system or Docker;
- deterministic repository validation in CI;
- clear separation between ARC-specific validation and generic backend tooling.

## 2. Executive Summary

**Recommended primary stack:**

- **Language:** Go
- **CLI framework:** Cobra
- **Markdown parser:** Goldmark
- **YAML parser:** `goccy/go-yaml`
- **Process execution:** Go standard library (`os/exec`, `context`)
- **Testing:** Go standard library `testing`, table-driven tests, fuzz tests where
useful
- **Linting:** `golangci-lint`
- **CI/CD:** GitHub Actions
- **Release strategy:** GitHub Actions build matrix + GitHub Releases
- **External backends:**
  - `markdownlint-cli2` via system binary or Docker
  - `lychee` via system binary or Docker

**High-level architectural choice:**

`arckit` should be a **single Go binary** that owns:

- ARC-specific validation rules,
- adoption-summary validation,
- cross-artifact repository checks,
- transition/gate checks,
- diagnostics and machine-readable output,
- orchestration of external backends.

It should **not** reimplement Markdown linting or generic HTTP link checking.

## 3. Monorepo Layout

Recommended repository layout:

```text
/
├── ARCs/
├── adoption/
├── assets/
├── docs/
├── templates/
├── .github/
├── README.md
├── arc-site/                 # optional, only if site tooling exists later
└── arckit/
    ├── cmd/
    │   └── arckit/
    │       └── main.go
    ├── internal/
    │   ├── app/
    │   ├── arc/
    │   ├── adoption/
    │   ├── repo/
    │   ├── transition/
    │   ├── backend/
    │   ├── config/
    │   ├── diag/
    │   ├── output/
    │   └── scaffolding/
    ├── testdata/
    │   ├── valid/
    │   ├── invalid/
    │   ├── fixtures/
    │   └── golden/
    ├── scripts/
    ├── go.mod
    ├── go.sum
    ├── .golangci.yml
    ├── .markdownlint-cli2.yaml   # optional if shared with repo root
    ├── .lychee.toml              # optional if shared with repo root
    └── README.md
```

### Recommendation on Go module boundaries

Start with **one Go module** rooted at `arckit/`.

Do **not** create a root-level `go.mod` just because the repository is a monorepo.
The docs/ARCs repository content is not Go code and does not benefit from becoming
part of a Go module.

Use a root-level `go.work` file **only if** you later split `arckit` into multiple
Go modules or add another Go project to the monorepo.

## 4. Go Version Policy

Use the **current stable Go release** as the primary development version.

Recommendation:

- primary development target: current stable Go;
- CI support target: current stable and previous stable minor release;
- minimum supported Go version: define explicitly in `arckit/go.mod` and update
intentionally.

For the first implementation, keep the minimum version reasonably modern to simplify
maintenance and leverage newer standard library improvements.

## 5. Core Implementation Stack

### 5.1 Language: Go

Go `1.26.1`.

### 5.2 CLI Framework: Cobra

Use **Cobra** for:

- subcommands;
- flag handling;
- shell completions;
- built-in help output;
- command tree organization;
- optional docs/manpage generation.

This is a good fit for a command surface like:

```text
arckit fmt
arckit validate arc
arckit validate adoption
arckit validate links
arckit validate repo
arckit validate transition
arckit init arc
arckit tools doctor
arckit rules
arckit explain
```

### 5.3 Markdown Parsing: Goldmark

Use **Goldmark** as the internal Markdown parser for ARC-specific checks.

Why:

- CommonMark-compliant;
- AST-based;
- source-position aware;
- extensible;
- pure Go.

Use Goldmark for:

- front-matter/body boundary handling after front-matter parsing;
- section order validation;
- heading structure checks;
- ARC-to-ARC link detection;
- image/asset link checks;
- extracting references and examples for ARC-specific rules.

Do **not** use Goldmark as a full formatter. Formatting should remain limited and safe.

### 5.4 YAML Parsing: `goccy/go-yaml`

Use **`goccy/go-yaml`** for:

- adoption summary parsing;
- repository config parsing (such as `.arckit.yaml`);
- front-matter decoding after manual field extraction where appropriate.

Why:

- pure Go implementation;
- strong error reporting;
- good support for YAML tooling use-cases;
- better fit than older YAML packages when diagnostics matter.

### 5.5 Validation Strategy

Prefer **typed Go structs plus explicit custom validation** over generic runtime
validation frameworks.

Recommendation:

- decode YAML/front matter into strongly typed structs;
- implement repository rules as explicit validator functions;
- keep rule identifiers, severities, and messages under direct control.

Benefits:

- clearer diagnostics;
- easier mapping to rule IDs such as `R:001`;
- easier control over migration/compatibility profiles;
- less framework magic.

Use generic schema tooling only if you later need external schema publication or
compatibility with non-Go tooling.

## 6. Parsing Strategy

### 6.1 Front Matter

Do **not** rely entirely on a generic front-matter library.

Instead:

1. parse the front-matter block manually;
2. preserve original line ordering and source spans;
3. decode field values into typed structures afterward.

Reason:

`arckit` needs to validate:

- field order;
- duplicate fields;
- whitespace conventions;
- conditional presence;
- filename/number alignment;
- migration-aware header compatibility.

Those checks are easier and more reliable with a small custom front-matter parser.

### 6.2 Body

Use Goldmark for the Markdown body and keep ARC-specific validation separate from
generic Markdown linting.

### 6.3 Formatting

Formatting should be conservative.

Use `arckit` for:

- front-matter normalization;
- safe whitespace cleanup;
- deterministic field ordering;
- repository-specific safe fixes.

Use `markdownlint-cli2` only as a **backend** for generic Markdown autofixes.

## 7. External Backend Strategy

### 7.1 Principle

`arckit` is the required CLI.

`markdownlint-cli2` and `lychee` are **optional backends**, not required user-installed
dependencies.

Users should have three practical paths:

1. use a compatible **system** backend if already installed;
2. use **Docker** as the local fallback;
3. rely on **CI** for full backend-powered reporting if they have neither.

### 7.2 Backend Modes

Recommended backend modes:

- `system`
- `docker`
- `auto`
- `off`

#### `system`

Use a backend executable found on the local `PATH`.

#### `docker`

Run the backend using Docker.

#### `auto`

Prefer `system`; if unavailable, fall back to `docker`; if neither works, emit a
clear diagnostic and continue with core `arckit` checks where possible.

#### `off`

Disable the backend explicitly.

### 7.3 Execution model

Use the standard library:

- `os/exec`
- `exec.CommandContext`
- `context.Context`

Do **not** use a Docker SDK.

Just shell out to:

- the system binary, or
- `docker run ...`

This keeps the implementation small and debuggable.

### 7.4 Markdown backend

Backend recommendation:

- **primary backend:** `markdownlint-cli2`

Use it for:

- Markdown linting;
- Markdown autofix when requested locally;
- check-only mode in validation.

`arckit fmt` should:

1. invoke `markdownlint-cli2 --fix` when the backend is available;
2. apply ARC-specific safe fixes;
3. re-run checks or internal validation as needed.

If the backend is unavailable and the mode is not `off`, `arckit` should emit a 
clear diagnostic such as:

- markdown backend unavailable on system;
- Docker unavailable;
- full Markdown backend checks deferred to CI.

### 7.5 Link backend

Backend recommendation:

- **primary backend:** `lychee`

Use it for:

- external link validation;
- generic local-path and link target validation when helpful;
- scheduled CI link-rot checks.

Keep repository-semantic link rules inside `arckit` itself. For example:

- ARC links must be relative;
- first ARC reference must be linked;
- root-relative links are invalid;
- image links must point to the correct `assets/arc-####/` subtree.

Lychee should handle reachability; `arckit` should handle repository semantics.

## 8. Configuration Strategy

Use a small explicit repository config file, for example:

```yaml
version: 1
profile: strict
backends:
  markdownlint:
    mode: auto
    system_command: markdownlint-cli2
    docker_image: davidanson/markdownlint-cli2:v0.20.0
  lychee:
    mode: auto
    system_command: lychee
    docker_image: lycheeverse/lychee:latest
ci:
  online_checks: true
output:
  default_format: text
```

## 9. Suggested Internal Package Layout

```text
arckit/internal/
├── app/           # command orchestration and service wiring
├── arc/           # ARC front matter/body parsing and ARC-only validation
├── adoption/      # adoption-summary parsing and validation
├── repo/          # repository-wide checks across ARC/adoption/assets
├── transition/    # status-transition checks
├── backend/       # system/docker backends for markdownlint and lychee
├── config/        # .arckit.yaml loading and validation
├── diag/          # rule IDs, severities, positions, renderers
├── output/        # text/json/sarif rendering
└── scaffolding/   # init arc and file generation
```

### Package responsibilities

#### `app`
- command execution flow
- common options
- mode resolution
- backend selection

#### `arc`
- parse front matter
- parse Markdown body
- validate ARC fields and sections
- ARC link rules

#### `adoption`
- parse YAML
- adoption schema validation
- actor/evidence validation

#### `repo`
- file discovery
- ARC/adoption mapping
- reciprocity checks
- asset path checks
- sponsor/repository alignment checks

#### `transition`
- machine-verifiable transition gates
- profile-aware relaxations for migration periods
- manual/editorial reminder diagnostics

#### `backend`
- backend discovery (`system`, `docker`)
- `markdownlint-cli2` adapter
- `lychee` adapter
- version reporting

#### `diag`
- `R:NNN` rule identifiers
- severities
- source locations
- remediation hints
- summary counts

#### `output`
- human-readable text
- JSON
- SARIF (when implemented)

#### `scaffolding`
- `init arc`
- template loading
- deterministic file generation

## 10. Testing Stack

### 10.1 Use the standard library first

Use Go's built-in testing tools as the default:

- `testing`
- subtests
- table-driven tests
- fuzz tests where parsing logic benefits
- benchmarks only where performance becomes meaningful

### 10.2 Test categories

#### Unit tests
For:
- front-matter parser
- field validators
- link rule validators
- adoption-summary validators
- backend resolution logic

#### Golden tests
For:
- text output
- JSON output
- SARIF output
- formatter results
- `arckit explain` output

#### Fixture tests
Use repository-style fixtures under `testdata/`:

- valid ARC docs
- invalid ARC docs
- valid adoption summaries
- invalid adoption summaries
- multi-file repo fixtures

#### Integration tests
For:
- `validate repo`
- `validate transition`
- backend invocation via system tools when available
- backend invocation via Docker in CI on supported runners

### 10.3 Snapshot/golden discipline

Prefer explicit golden files over opaque snapshot tooling.

That keeps outputs reviewable in Git and easier to reason about.

## 11. Linting and Developer Tooling

### 11.1 Go linting

Use **golangci-lint** for the Go codebase.

Recommended baseline linters:

- `govet`
- `staticcheck`
- `errcheck`
- `ineffassign`
- `unused`
- `gocritic`
- `revive`
- `misspell`

Start with a focused ruleset. Avoid turning on every linter at once.

### 11.2 Formatting

Use:

- `gofmt`
- `goimports` (optional but recommended)

### 11.3 Task running

To minimize dependencies, prefer one of these:

- plain `make` as a thin wrapper, or
- small shell scripts in `arckit/scripts/`

Do **not** require Mage, Task, or other extra local tooling for v1.

Examples:

```make
fmt:
	cd arckit && gofmt -w ./...

lint:
	cd arckit && golangci-lint run

test:
	cd arckit && go test ./...

build:
	cd arckit && go build ./cmd/arckit
```

## 12. CI/CD Stack

### 12.1 CI platform

Use **GitHub Actions**.

The repository already lives on GitHub and the CLI is intended to be used in repository CI, so GitHub Actions is the natural default.

### 12.2 Suggested CI jobs

#### `arckit-go-lint`
- checkout
- setup Go
- run `golangci-lint`

#### `arckit-go-test`
- checkout
- setup Go
- run `go test ./...`

#### `arckit-build`
- checkout
- setup Go
- build `./cmd/arckit`
- optional cross-compile smoke checks

#### `repo-validate-offline`
- build `arckit`
- run `arckit validate repo . --offline`

#### `repo-validate-online`
- run on pull request or on a narrower path filter
- enable external backend checks when appropriate
- use Docker or system tools available in CI

#### `link-rot-scheduled`
- scheduled workflow
- full online link validation
- non-blocking or separately triaged if desired

### 12.3 Monorepo-aware paths

Use path filters so Go jobs only run when:

- `arckit/**`
- `.github/workflows/**`
- shared config files relevant to `arckit`

Repository validation jobs should run when ARC docs or adoption files change.

## 13. Release Strategy

### Recommendation: start simple

Because `arckit` lives in a monorepo subdirectory, keep release automation simple
initially.

Recommended approach:

- create tags for the `arckit` module using a subdirectory-aware prefix;
- use GitHub Actions to build binaries for Linux, macOS, and Windows;
- upload artifacts to GitHub Releases;
- optionally generate checksums.

### Tagging convention

Use tags like:

```text
arckit/v0.1.0
arckit/v0.2.0
```

This aligns with Go subdirectory module expectations and keeps releases scoped to `arckit`.

### Why not start with GoReleaser

GoReleaser is excellent, but its monorepo tag-prefix support is tied to its monorepo
features. For a docs + tool monorepo, plain GitHub Actions is simpler to control
at first.

You can revisit GoReleaser later if release complexity grows.

## 14. Recommended Developer UX

### Local workflows

#### Lightweight author
A docs author with no extra local tools should still be able to:

- run `arckit validate arc ...`
- run `arckit validate repo .`
- get ARC-specific diagnostics
- rely on CI for full backend-powered Markdown and link checks if they do not have
system tools or Docker

#### Power user
A contributor with tools installed locally can:

- use `markdownlint-cli2` and `lychee` via `system` mode
- run local autofix and online link checks
- get faster feedback loops

#### Docker user
A contributor without local backends but with Docker can:

- run the exact backend tools through `docker`
- avoid Node/Rust installs entirely

## 15. Observability and Outputs

### Mandatory outputs

Implement these first:

- human-readable text output
- JSON output

### Later output

- SARIF output, once the diagnostic model is stable

Recommendation:

Design diagnostics internally as structured data first, then render them into text/JSON/SARIF.

## 16. Security and Trustworthiness

- never silently rewrite semantic content;
- always distinguish deterministic repository failures from backend/tool availability failures;
- always distinguish semantic validation failures from transient network failures;
- never make Docker or external backends mandatory for core offline validation.

## 17. Implementation Phases

### Phase 1
- Cobra CLI skeleton
- config loading
- diagnostics model
- front-matter parser
- `validate arc`

### Phase 2
- adoption-summary parsing and validation
- `validate adoption`
- `validate repo`

### Phase 3
- `validate transition`
- `init arc`
- JSON output stabilization

### Phase 4
- backend orchestration (`system`, `docker`, `auto`, `off`)
- `markdownlint-cli2` integration
- `lychee` integration
- `tools doctor`

### Phase 5
- safer autofix flows
- SARIF
- deeper monorepo CI integration

## 18. Final Recommendation

For this repository, the best product tradeoff is:

- **Go for the core CLI**
- **Cobra for the command model**
- **Goldmark for Markdown parsing**
- **goccy/go-yaml for YAML**
- **system-or-Docker backends for markdownlint-cli2 and lychee**
- **GitHub Actions for CI and releases**
- **single Go module under `/arckit`**, with `go.work` only when multiple Go modules
actually exist

That stack keeps the contributor footprint small, fits the updated backend strategy,
and stays aligned with a docs-first monorepo where `arckit` is the repository's
semantic validator rather than an all-in-one bundled toolchain.
