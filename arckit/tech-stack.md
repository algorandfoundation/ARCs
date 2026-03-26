# arckit Tech Stack

## 1. Purpose

This document defines the implementation stack for `arckit` as a native-first Go CLI
inside this repository.

The goals are:

1. keep the first full Go project simple enough to build and maintain confidently;
1. keep required user dependencies low;
1. keep CI and release automation explicit and easy to reason about;
1. pin the toolchain and all automation dependencies deliberately for security.

## 2. Fixed Product Decisions

These choices are fixed for v1:

1. `arckit` is a single Go binary.
1. Core validation must work with just Go and the repository contents.
1. Optional backend tooling may add Markdown style checks and online link checks, but
   must never be required for core offline validation.
1. The CLI remains broad, but internal variability is kept small:
   - no `.arckit.yaml`
   - no user-local config
   - no profile system
   - no per-backend mode matrix
   - no SARIF in v1
1. `arckit validate repo .` is the canonical CI validation command.

## 3. Go Toolchain Policy

### 3.1 Go Version

Use one exact Go patch release for development and CI.

As of March 26, 2026, the default pinned version is:

- `go1.26.1`

Implementation policy:

1. declare the exact version in `arckit/go.mod`;
1. declare the same preferred toolchain with a `toolchain` line;
1. pin CI to the same exact patch release;
1. do not use floating `stable`, major-only, or major-minor-only Go version selectors.

The initial `go.mod` policy should be:

```go
go 1.26.1
toolchain go1.26.1
```

### 3.2 Module Layout

Use one Go module rooted at `arckit/`.

Do not:

1. add a repository-root `go.mod`;
1. add a `go.work` file;
1. split the tool into multiple modules in v1.

This repository is a docs-first monorepo. The Go module should cover only the tool.

## 4. Direct Dependencies

Keep direct Go dependencies intentionally small.

### 4.1 Required Direct Dependencies

Use exactly these third-party building blocks unless a concrete implementation problem
proves otherwise:

1. `cobra` for the command tree, help output, and subcommand parsing;
1. `goldmark` for Markdown AST parsing;
1. `yaml.v3` for adoption summary YAML parsing.

Everything else should prefer the standard library.

### 4.2 Standard Library Responsibilities

Use the standard library for:

1. JSON rendering;
1. file walking and path validation;
1. process execution for backend tools;
1. testing;
1. fixture loading;
1. backend version probing;
1. conservative file rewriting in `fmt`;
1. command exit handling and logging.

### 4.3 Explicitly Rejected for v1

Do not use:

1. Viper;
1. a Docker SDK;
1. generic schema-validation frameworks;
1. assertion libraries;
1. `golangci-lint`;
1. Mage or Task;
1. GoReleaser;
1. plugin systems or pluggable rule engines.

The aim is a tool that is easy to understand with ordinary Go knowledge.

## 5. Internal Layout

Keep the package layout compact and behavior-oriented:

```text
arckit/
├── cmd/arckit/
├── internal/cli/
├── internal/arc/
├── internal/adoption/
├── internal/repo/
├── internal/transition/
├── internal/backend/
├── internal/diag/
├── internal/scaffold/
├── testdata/
├── go.mod
└── README.md
```

Package responsibilities:

### 5.1 `cli`

Owns:

1. Cobra command definitions;
1. global flag parsing;
1. command dispatch;
1. text and JSON output wiring.

### 5.2 `arc`

Owns:

1. front matter parsing;
1. section discovery;
1. ARC-only validation;
1. local ARC link and asset rules.

### 5.3 `adoption`

Owns:

1. adoption YAML decoding;
1. enum validation;
1. adoption summary consistency checks.

### 5.4 `repo`

Owns:

1. repository file discovery;
1. ARC to adoption mapping;
1. asset tree checks;
1. cross-file relationship reciprocity.

### 5.5 `transition`

Owns:

1. machine-verifiable transition checks;
1. manual-check reminder diagnostics.

### 5.6 `backend`

Owns:

1. backend detection on `PATH`;
1. Docker fallback invocation;
1. version discovery;
1. backend-unavailable diagnostics.

### 5.7 `diag`

Owns:

1. rule metadata;
1. severities;
1. file positions;
1. text and JSON-friendly diagnostic structures.

### 5.8 `scaffold`

Owns:

1. `init arc` templates;
1. deterministic ARC and adoption stub generation.

## 6. Backend Strategy

### 6.1 Supported Backends

Only these external tools are supported in v1:

1. `markdownlint-cli2`
1. `lychee`

### 6.2 Resolution Policy

Backend resolution is fixed:

1. try the system binary first;
1. otherwise try a pinned Docker image digest;
1. otherwise emit a backend-unavailable diagnostic and continue native validation.

There is no repo config or user config for backends in v1.

### 6.3 When Backends Are Used

Use `markdownlint-cli2` only for:

1. generic Markdown linting during validation;
1. generic Markdown autofix during `fmt`.

Use `lychee` only for:

1. external link reachability in `--online` mode.

Do not move repo-semantic rules into external tools. Relative ARC links, adoption-path
rules, asset-tree rules, and transition rules stay inside `arckit`.

## 7. Developer Workflow

### 7.1 Required Local Tooling

For core development and usage, require only:

1. Go `1.26.1`

Optional local tooling:

1. `markdownlint-cli2`
1. `lychee`
1. Docker

Users without those optional tools must still be able to run native validation successfully.

### 7.2 Default Commands

Prefer plain Go commands and a thin `Makefile` if a wrapper is useful.

The default developer command set should be:

```make
fmt:
	cd arckit && gofmt -w -s ./...

vet:
	cd arckit && go vet ./...

test:
	cd arckit && go test ./...

build:
	cd arckit && go build ./cmd/arckit
```

Do not introduce extra task runners in v1.

## 8. Testing Strategy

Use the standard library test stack only.

### 8.1 Test Types

Use:

1. unit tests for parsers, validators, and backend resolution;
1. fixture tests for repo-like validation scenarios;
1. golden tests for text and JSON output, `rules`, `explain`, and `init arc`;
1. CLI integration tests for each public command.

### 8.2 Test Data

Keep repo-like fixtures under `arckit/testdata/`.

Required fixture coverage:

1. valid ARC without adoption file where one is not yet required;
1. ARC missing an adoption file when one is required;
1. transition to `Review`, `Last Call`, `Final`, and `Idle`;
1. backend present through system binary;
1. backend present through Docker;
1. backend unavailable with native validation still succeeding;
1. online link failures reported separately from semantic failures.

## 9. CI/CD Workflows

Keep CI explicit and small. v1 needs three workflows.

### 9.1 PR Tool Workflow

Run when `arckit/**` or relevant workflow files change.

Steps:

1. checkout;
1. set up pinned Go `1.26.1`;
1. run `gofmt -s` check;
1. run `go vet ./...`;
1. run `go test ./...`;
1. run `go build ./cmd/arckit`.

### 9.2 PR Repo Validation Workflow

Run when `ARCs/**`, `adoption/**`, `templates/**`, or `arckit/**` changes.

Steps:

1. checkout;
1. set up pinned Go `1.26.1`;
1. build `arckit`;
1. run `arckit validate repo .`.

This is the canonical required validation gate.

### 9.3 Scheduled or Manual Online Workflow

Run on a schedule and by manual dispatch.

Steps:

1. checkout;
1. set up pinned Go `1.26.1`;
1. make the pinned backend tools available;
1. build `arckit`;
1. run `arckit validate repo . --online`.

This workflow may be separately triaged if external network instability causes failures.

## 10. Release Strategy

Start simple and subdirectory-aware.

### 10.1 Version Tags

Release from tags shaped like:

```text
arckit/v0.1.0
arckit/v0.2.0
```

This keeps releases aligned with the `arckit/` Go module.

### 10.2 Release Workflow

The release workflow should:

1. trigger on `arckit/v*` tags;
1. build binaries for:
   - Linux amd64
   - Linux arm64
   - macOS amd64
   - macOS arm64
   - Windows amd64
   - Windows arm64
1. archive the binaries;
1. generate SHA256 checksums;
1. publish artifacts to GitHub Releases.

Do not add signing, provenance, Homebrew automation, or GoReleaser in v1.

## 11. Security and Pinning Policy

### 11.1 Toolchain and Module Pinning

Pin:

1. the Go patch version in `go.mod`;
1. the preferred Go toolchain in `go.mod`;
1. all direct module versions in `go.mod`;
1. all module checksums in `go.sum`.

Updates must be explicit and reviewed.

### 11.2 CI Pinning

Pin:

1. every GitHub Action to a full commit SHA;
1. the Go patch version in every workflow;
1. backend package versions exactly when installed directly;
1. backend Docker images by digest, not mutable tags.

Do not use floating `latest`, `stable`, or major-only tags in CI.

### 11.3 Workflow Permissions

Default workflow permissions should be read-only:

```yaml
permissions:
  contents: read
```

Widen permissions only per job when a specific release or upload step requires it.

## 12. Deferred for Later

These are explicitly deferred until the first real version is working:

1. SARIF output;
1. multi-version required CI matrices;
1. signing and provenance;
1. Homebrew or package-manager distribution;
1. config files;
1. broader linter suites;
1. deeper automation around pull requests or labels.

## 13. Reference Inputs

These sources informed the versioning and workflow defaults in this document:

1. [Go 1.26 release notes](https://go.dev/doc/go1.26)
1. [Go downloads page](https://go.dev/dl/)
1. [Go toolchain selection](https://go.dev/doc/toolchain)
1. [Go module source and subdirectory tags](https://go.dev/doc/modules/managing-source)
1. [GitHub Actions secure use reference](https://docs.github.com/en/actions/reference/security/secure-use)
