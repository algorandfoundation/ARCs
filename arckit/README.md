# arckit

`arckit` is the ARC repository validator and scaffolding CLI.

## Requirements

- Go `1.26.1`

Generic Markdown/YAML/text hygiene for this repository is handled by the
repository-root `.pre-commit-config.yaml`, not by `arckit`. That includes
advisory Markdown/YAML spelling checks and advisory external link checks.

`arckit fmt` is limited to ARC Markdown files under `ARCs/arc-####.md` and
adoption summaries under `adoption/arc-####.yaml`. It rewrites deterministic
ARC/adoption-specific structure only and does not overlap with the repository
`pre-commit` hooks on generic YAML formatting or linting.

If `fmt` reports invalid ARC front matter YAML, that is still an `arckit`
concern: the repository YAML hooks do not inspect YAML embedded in ARC Markdown
front matter.

Within that scope, `fmt` can sort numeric ARC relationship lists, reorder
canonical ARC level-2 sections, and reorder canonical adoption-summary mapping
keys such as top-level fields plus `reference-implementation`, `adoption`, and
`summary`.

`arckit` owns ARC-specific metadata, section, reference, maturity, and body-link
policy, including rejecting absolute links back into repository content such as
ARCs or assets. External raw HTML anchors are allowed.

## Repo-Local Config

`arckit` auto-discovers an optional repository-root `.arckit.jsonc` file for all
`validate` commands. The file is JSON with comments and supports only repo-local
suppression rules. There is no `--config` flag and no user-local configuration.

Supported keys:

- `ignoreArcs`: ignore an ARC number across its ARC, adoption, and asset footprint
- `ignoreRules`: ignore a rule everywhere
- `ignoreByArc`: ignore rules for exact ARC numbers or inclusive ARC ranges like `50-60`

Migration suppressions are allowed temporarily when the repository is moving
historical ARC files to a new canonical encoding or ARC body/reference policy.

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

Invalid `.arckit.jsonc` content stops validation with exit code `2`.

## Common Commands

```sh
cd arckit
find . -name '*.go' -print0 | xargs -0 gofmt -w -s
go vet ./...
go test ./...
go build ./cmd/arckit
go run ./cmd/arckit fmt ../ARCs/arc-0000.md
go run ./cmd/arckit summary repo ..
```

## Examples

```sh
cd arckit
go run ./cmd/arckit validate repo ..
go run ./cmd/arckit summary repo ..
go run ./cmd/arckit validate arc ../ARCs/arc-0000.md
go run ./cmd/arckit validate arc \
  --enforce-rule R:004 \
  --enforce-rule R:008 \
  --enforce-rule R:025 \
  --enforce-rule R:032 \
  ../ARCs/arc-0000.md
go run ./cmd/arckit validate arc --ignore-config ../ARCs/arc-0000.md
go run ./cmd/arckit validate adoption ../adoption/arc-0042.yaml
go run ./cmd/arckit validate links ../ARCs/arc-0000.md
go run ./cmd/arckit fmt ../ARCs/arc-0000.md
```

`validate adoption` and `validate repo` both require the canonical vetted adopters
registry at `../adoption/vetted-adopters.yaml`. Per-ARC adoption actor names must
be lower-kebab-case identifiers present in the matching registry category.
They also enforce that `summary.adoption-readiness` matches the tracked adopter
count thresholds.

When formatting adoption summaries, `arckit fmt` normalizes
`summary.adoption-readiness` from the tracked adopter count: `low` for fewer than
3 adopters, `medium` for 3-4 adopters, and `high` for 5 or more adopters.

`summary repo` writes a local markdown review artifact at `../arc-summary.md` by
default for ARC Editor workflow use. It is a report generator, not a validation gate.
