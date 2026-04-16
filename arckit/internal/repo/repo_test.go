package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/config"
	"github.com/algorandfoundation/ARCs/arckit/internal/testutil"
)

func TestValidateRepoMissingRequiredAdoption(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "missing-adoption")
	_, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:013" || diagnostic.RuleID == "R:012" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected missing adoption diagnostic, got %+v", diagnostics)
	}
}

func TestValidateRepoIgnoresConfiguredARCFootprint(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "missing-adoption"))
	testutil.WriteTrimmedFile(t, filepath.Join(root, config.FileName), `{
  "ignoreArcs": [43]
}`)
	if err := os.MkdirAll(filepath.Join(root, "assets", "arc-0043"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "assets", "arc-0043", "example.txt"), []byte("example"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatalf("Validate() diagnostics = %v, want non-empty without config", diagnostics)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	_, diagnostics, err = Validate(root, cfg)
	if err != nil {
		t.Fatalf("Validate() with config error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("Validate() with config diagnostics = %+v, want none", diagnostics)
	}
}

func TestValidateRepoIgnoresRuleOnConfiguredRange(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "missing-adoption"))
	testutil.WriteTrimmedFile(t, filepath.Join(root, config.FileName), `{
  "ignoreByArc": {
    "40-45": ["R:012", "R:013"]
  }
}`)

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	_, diagnostics, err := Validate(root, cfg)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:012" || diagnostic.RuleID == "R:013" {
			t.Fatalf("unexpected missing-adoption diagnostic with config: %+v", diagnostics)
		}
	}
}

func TestValidateRepoDoesNotDeriveRelationshipsFromLegacyScalarLists(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ARCs"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "adoption"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	testutil.WriteTrimmedFile(t, filepath.Join(root, "adoption", "vetted-adopters.yaml"), `wallets: []
explorers: []
tooling: []
infra: []
dapps-protocols: []
`)

	arc1 := `---
arc: 1
title: First
description: First ARC
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Standards Track
created: 2026-04-09
sponsor: Foundation
implementation-required: false
extends: 2
---

## Abstract

Text

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
`
	arc2 := `---
arc: 2
title: Second
description: Second ARC
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Standards Track
created: 2026-04-09
sponsor: Foundation
implementation-required: false
---

## Abstract

Text

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
`
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0001.md"), []byte(arc1), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0002.md"), []byte(arc2), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:011" {
			t.Fatalf("expected legacy scalar relationship field not to feed R:011, got %+v", diagnostics)
		}
	}
}

func TestValidateRepoIncludesARCZeroInState(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ARCs"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "adoption"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	testutil.WriteTrimmedFile(t, filepath.Join(root, "adoption", "vetted-adopters.yaml"), `wallets: []
explorers: []
tooling: []
infra: []
dapps-protocols: []
`)

	content := `---
arc: 0
title: ARC Zero
description: Repository process document.
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Living
type: Meta
created: 2026-04-09
sponsor: Foundation
implementation-required: false
---

## Abstract

Text

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
`
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0000.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	state, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("Validate() diagnostics = %+v", diagnostics)
	}
	if _, ok := state.ARCs[0]; !ok {
		t.Fatalf("expected ARC 0 to be included in repo state, got %+v", state.ARCs)
	}
}

func TestValidateRepoDoesNotReportOrphanedAdoptionWhenMatchingARCFileHasInvalidFrontMatter(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ARCs"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "adoption"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	testutil.WriteTrimmedFile(t, filepath.Join(root, "adoption", "vetted-adopters.yaml"), `wallets: []
explorers: []
tooling: []
infra: []
dapps-protocols: []
`)

	testutil.WriteTrimmedFile(t, filepath.Join(root, "ARCs", "arc-0089.md"), `---
arc: 89
title: Example ARC
description: Example description
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Last Call
type: Standards Track
created: 2026-04-09
sponsor: Foundation
implementation-required: true
implementation-url: https://github.com/example/arc89
implementation-maintainer:
  - [@example
adoption-summary: adoption/arc-0089.yaml
last-call-deadline: 2026-05-01
---

## Abstract

Text

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text
`)

	testutil.WriteTrimmedFile(t, filepath.Join(root, "adoption", "arc-0089.yaml"), `arc: 89
title: Example ARC
last-reviewed: 2026-04-09
reference-implementation:
  status: shipped
adoption:
  wallets: []
  explorers: []
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`)

	_, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	var sawInvalidFrontMatter bool
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:005" {
			sawInvalidFrontMatter = true
		}
		if diagnostic.RuleID == "R:018" && diagnostic.Message == "orphaned adoption summary for ARC 89" {
			t.Fatalf("expected no orphaned adoption false positive, got %+v", diagnostics)
		}
	}
	if !sawInvalidFrontMatter {
		t.Fatalf("expected invalid front matter diagnostic, got %+v", diagnostics)
	}
}

func TestValidateRepoRejectsMaturityRegression(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ARCs"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "adoption"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	testutil.WriteTrimmedFile(t, filepath.Join(root, "adoption", "vetted-adopters.yaml"), `wallets: []
explorers: []
tooling: []
infra: []
dapps-protocols: []
`)

	source := `---
arc: 1
title: Source ARC
description: Example source ARC.
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Review
type: Standards Track
created: 2026-04-10
sponsor: Foundation
implementation-required: false
requires:
  - 2
---

## Abstract

Text

## Motivation

Text

## Specification

See [ARC-2](arc-0002.md).

## Rationale

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
`
	target := `---
arc: 2
title: Target ARC
description: Example target ARC.
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/2
status: Draft
type: Standards Track
created: 2026-04-10
sponsor: Foundation
implementation-required: false
---

## Abstract

Text

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
`
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0001.md"), []byte(source), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0002.md"), []byte(target), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	foundRequires := false
	foundBodyLink := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID != "R:038" {
			continue
		}
		if diagnostic.File == filepath.Join(root, "ARCs", "arc-0001.md") && diagnostic.Line == 12 && diagnostic.Message == "requires target ARC-2 must be at least as mature as ARC-1" {
			foundRequires = true
		}
		if diagnostic.File == filepath.Join(root, "ARCs", "arc-0001.md") && diagnostic.Message == "ARC body link to ARC-2 must not point to a less mature ARC" {
			foundBodyLink = true
		}
	}
	if !foundRequires || !foundBodyLink {
		t.Fatalf("expected requires and body-link maturity diagnostics, got %+v", diagnostics)
	}
}

func TestValidateRepoRequiresVettedAdoptersRegistry(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	if err := os.Remove(filepath.Join(root, "adoption", "vetted-adopters.yaml")); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	_, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:022" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected missing vetted adopters diagnostic, got %+v", diagnostics)
	}
}

func TestValidateRepoRejectsUnvettedAdopter(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	testutil.WriteTrimmedFile(t, filepath.Join(root, "adoption", "vetted-adopters.yaml"), `wallets: []
explorers: []
tooling: []
infra: []
dapps-protocols: []
`)

	_, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:023" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected invalid adopter reference diagnostic, got %+v", diagnostics)
	}
}

func TestValidateRepoRejectsFinalARCWithoutTrackedAdoption(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0042.md"), []byte(`---
arc: 42
title: Example ARC
description: Example ARC for testing.
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Final
type: Standards Track
created: 2026-03-26
sponsor: Foundation
implementation-required: false
adoption-summary: adoption/arc-0042.yaml
---

## Abstract

Text

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "adoption", "arc-0042.yaml"), []byte(`arc: 42
title: Example ARC
last-reviewed: 2026-04-09
adoption:
  wallets: []
  explorers: []
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:025" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected Final ARC empty adoption diagnostic, got %+v", diagnostics)
	}
}

func TestValidateRepoCanSuppressFinalARCWithoutTrackedAdoptionByARC(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0042.md"), []byte(`---
arc: 42
title: Example ARC
description: Example ARC for testing.
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Final
type: Standards Track
created: 2026-03-26
sponsor: Foundation
implementation-required: false
adoption-summary: adoption/arc-0042.yaml
---

## Abstract

Text

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "adoption", "arc-0042.yaml"), []byte(`arc: 42
title: Example ARC
last-reviewed: 2026-04-09
adoption:
  wallets: []
  explorers: []
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	testutil.WriteTrimmedFile(t, filepath.Join(root, config.FileName), `{
  "ignoreByArc": {
    "42": ["R:025"]
  }
}`)

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	_, diagnostics, err := Validate(root, cfg)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:025" {
			t.Fatalf("expected R:025 to be suppressed by ARC config, got %+v", diagnostics)
		}
	}
}
