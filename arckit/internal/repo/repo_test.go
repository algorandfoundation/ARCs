package repo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/config"
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
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "missing-adoption"))
	writeConfig(t, root, `{
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
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "missing-adoption"))
	writeConfig(t, root, `{
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
	writeVettedAdopters(t, root, `wallets: []
explorers: []
sdk-libraries: []
infra: []
dapps-protocols: []
`)

	arc1 := `---
arc: 1
title: First
description: First ARC
author:
  - Example Author
discussions-to: https://example.com/discussion
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
`
	arc2 := `---
arc: 2
title: Second
description: Second ARC
author:
  - Example Author
discussions-to: https://example.com/discussion
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
	writeVettedAdopters(t, root, `wallets: []
explorers: []
sdk-libraries: []
infra: []
dapps-protocols: []
`)

	content := `---
arc: 0
title: ARC Zero
description: Repository process document.
author:
  - Example Author
discussions-to: https://example.com/discussion
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

func TestValidateRepoRequiresVettedAdoptersRegistry(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
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
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	writeVettedAdopters(t, root, `wallets: []
explorers: []
sdk-libraries: []
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
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0042.md"), []byte(`---
arc: 42
title: Example ARC
description: Example ARC for testing.
author:
  - Example Author (@example)
discussions-to: https://example.com/discussion
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
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "adoption", "arc-0042.yaml"), []byte(`arc: 42
title: Example ARC
status: Final
last-reviewed: 2026-04-09
sponsor: Foundation
implementation-required: false
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
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0042.md"), []byte(`---
arc: 42
title: Example ARC
description: Example ARC for testing.
author:
  - Example Author (@example)
discussions-to: https://example.com/discussion
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
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "adoption", "arc-0042.yaml"), []byte(`arc: 42
title: Example ARC
status: Final
last-reviewed: 2026-04-09
sponsor: Foundation
implementation-required: false
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
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	writeConfig(t, root, `{
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

func copyRepoFixture(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, relative)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, info.Mode())
	}); err != nil {
		t.Fatalf("copyRepoFixture() error = %v", err)
	}
	return dst
}

func writeVettedAdopters(t *testing.T, root string, content string) {
	t.Helper()
	path := filepath.Join(root, "adoption", "vetted-adopters.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func writeConfig(t *testing.T, root string, content string) {
	t.Helper()
	path := filepath.Join(root, config.FileName)
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
