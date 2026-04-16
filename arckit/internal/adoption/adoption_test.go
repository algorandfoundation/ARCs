package adoption

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
	"github.com/algorandfoundation/ARCs/arckit/internal/testutil"
)

func TestValidateMatchingAdoptionSummary(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "valid-draft")
	arcPath := filepath.Join(root, "ARCs", "arc-0042.md")
	adoptionPath := filepath.Join(root, "adoption", "arc-0042.yaml")

	document := loadValidatedDocument(t, root, arcPath)

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}
	registry, registryDiagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}
	if registryValidation := ValidateVettedAdopters(registry); len(registryValidation) != 0 {
		t.Fatalf("ValidateVettedAdopters() diagnostics = %v", registryValidation)
	}
	for _, diagnostic := range Validate(summary, document, registry) {
		if diagnostic.Severity == diag.SeverityError {
			t.Fatalf("unexpected validation error: %+v", diagnostic)
		}
	}
}

func TestValidateAdoptionRejectsActorOutsideMatchingRegistryCategory(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	registryPath := RegistryPath(root)
	if err := os.WriteFile(registryPath, []byte(`wallets: []
explorers:
  - example-wallet
tooling: []
infra: []
dapps-protocols: []
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	arcPath := filepath.Join(root, "ARCs", "arc-0044.md")
	adoptionPath := filepath.Join(root, "adoption", "arc-0044.yaml")

	document := loadValidatedDocument(t, root, arcPath)

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}

	registry, registryDiagnostics, err := LoadVettedAdopters(registryPath)
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}

	diagnostics := Validate(summary, document, registry)
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:023" && strings.Contains(diagnostic.Message, `is not present in vetted adopters category "wallets"`) {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected vetted adopter category diagnostic, got %+v", diagnostics)
	}
}

func TestValidateAdoptionRejectsTitleMismatchWithARC(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	arcPath := filepath.Join(root, "ARCs", "arc-0042.md")
	adoptionPath := filepath.Join(root, "adoption", "arc-0042.yaml")

	document := loadValidatedDocument(t, root, arcPath)
	if err := os.WriteFile(adoptionPath, []byte(`arc: 42
title: Different ARC Title
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

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}
	registry, registryDiagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}

	diagnostics := Validate(summary, document, registry)
	if !containsDiagnosticMessage(diagnostics, `adoption title "Different ARC Title" does not match ARC title "Example ARC"`) {
		t.Fatalf("expected title mismatch diagnostic, got %+v", diagnostics)
	}
}

func TestValidateFinalAdoptionRequiresAtLeastOneAdopter(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	arcPath := filepath.Join(root, "ARCs", "arc-0042.md")
	if err := os.WriteFile(arcPath, []byte(`---
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
	adoptionPath := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 42
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

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}

	registry, registryDiagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}

	document := loadValidatedDocument(t, root, arcPath)
	diagnostics := Validate(summary, document, registry)
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:025" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Final ARC empty adoption diagnostic, got %+v", diagnostics)
	}
}

func TestValidateAdoptionReadinessRequiresMinimumAdopterCounts(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	document := loadValidatedDocument(t, root, filepath.Join(root, "ARCs", "arc-0042.md"))
	adoptionPath := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 42
title: Example ARC
last-reviewed: 2026-04-09
adoption:
  wallets:
    - name: example-wallet
      status: shipped
      evidence: https://example.com/wallet-proof
      notes: ""
  explorers:
    - name: example-explorer
      status: shipped
      evidence: https://example.com/explorer-proof
      notes: ""
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: medium
  blockers: []
  notes: ""
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(RegistryPath(root), []byte(`wallets:
  - example-wallet
explorers:
  - example-explorer
tooling: []
infra: []
dapps-protocols: []
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}
	registry, registryDiagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}

	diagnostics := Validate(summary, document, registry)
	if !containsDiagnosticMessage(diagnostics, `summary.adoption-readiness "medium" requires at least 3 adopters`) {
		t.Fatalf("expected medium readiness threshold diagnostic, got %+v", diagnostics)
	}
}

func TestValidateHighAdoptionReadinessAllowsFiveAdoptersAcrossCategories(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	document := loadValidatedDocument(t, root, filepath.Join(root, "ARCs", "arc-0042.md"))
	adoptionPath := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 42
title: Example ARC
last-reviewed: 2026-04-09
adoption:
  wallets:
    - name: wallet-one
      status: shipped
      evidence: https://example.com/wallet-one
      notes: ""
    - name: wallet-two
      status: shipped
      evidence: https://example.com/wallet-two
      notes: ""
  explorers:
    - name: explorer-one
      status: shipped
      evidence: https://example.com/explorer-one
      notes: ""
  tooling:
    - name: sdk-one
      status: shipped
      evidence: https://example.com/sdk-one
      notes: ""
  infra:
    - name: infra-one
      status: shipped
      evidence: https://example.com/infra-one
      notes: ""
  dapps-protocols: []
summary:
  adoption-readiness: high
  blockers: []
  notes: ""
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(RegistryPath(root), []byte(`wallets:
  - wallet-one
  - wallet-two
explorers:
  - explorer-one
tooling:
  - sdk-one
infra:
  - infra-one
dapps-protocols: []
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}
	registry, registryDiagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}

	for _, diagnostic := range Validate(summary, document, registry) {
		if diagnostic.RuleID == "R:016" && strings.Contains(diagnostic.Message, "summary.adoption-readiness") {
			t.Fatalf("unexpected adoption-readiness diagnostic: %+v", diagnostic)
		}
	}
}

func TestValidateFinalAdoptionAllowsTrackedAdopter(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	arcPath := filepath.Join(root, "ARCs", "arc-0044.md")
	if err := os.WriteFile(arcPath, []byte(`---
arc: 44
title: Transition Ready ARC
description: ARC used to validate Final transitions.
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Final
type: Standards Track
created: 2026-03-26
sponsor: Foundation
implementation-required: true
implementation-url: https://github.com/algorandfoundation/arc44
implementation-maintainer:
  - "@maintainer"
adoption-summary: adoption/arc-0044.yaml
---

## Abstract

Text

## Motivation

Text

## Specification

Text

## Rationale

Text

## Test Cases

Text

## Reference Implementation

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	adoptionPath := filepath.Join(root, "adoption", "arc-0044.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 44
title: Transition Ready ARC
last-reviewed: 2026-03-26
reference-implementation:
  status: shipped
  notes: ""
adoption:
  wallets:
    - name: example-wallet
      status: shipped
      evidence: https://example.com/wallet-proof
      notes: ""
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

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}

	registry, registryDiagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}

	document := loadValidatedDocument(t, root, arcPath)
	for _, diagnostic := range Validate(summary, document, registry) {
		if diagnostic.RuleID == "R:025" {
			t.Fatalf("unexpected Final ARC empty adoption diagnostic: %+v", diagnostic)
		}
	}
}

func TestValidateRejectsLegacyReferenceImplementationIdentityFields(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	document := loadValidatedDocument(t, root, filepath.Join(root, "ARCs", "arc-0044.md"))
	adoptionPath := filepath.Join(root, "adoption", "arc-0044.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 44
title: Transition Ready ARC
last-reviewed: 2026-03-26
reference-implementation:
  repository: https://github.com/example/arc-0044
  maintainers:
    - "@maintainer"
  status: shipped
  notes: ""
adoption:
  wallets:
    - name: example-wallet
      status: shipped
      evidence: https://example.com/wallet-proof
      notes: ""
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

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}

	registry, registryDiagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}

	diagnostics := Validate(summary, document, registry)
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:016" && strings.Contains(diagnostic.Message, "reference-implementation.repository is not allowed") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected legacy reference-implementation identity diagnostic, got %+v", diagnostics)
	}
}

func TestValidateAllowsReferenceImplementationWithoutNotes(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	document := loadValidatedDocument(t, root, filepath.Join(root, "ARCs", "arc-0044.md"))
	adoptionPath := filepath.Join(root, "adoption", "arc-0044.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 44
title: Transition Ready ARC
last-reviewed: 2026-03-26
reference-implementation:
  status: shipped
adoption:
  wallets:
    - name: example-wallet
      status: shipped
      evidence: https://example.com/wallet-proof
      notes: ""
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

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}

	registry, registryDiagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}

	for _, diagnostic := range Validate(summary, document, registry) {
		if diagnostic.RuleID == "R:015" && strings.Contains(diagnostic.Message, "reference-implementation.notes") {
			t.Fatalf("unexpected notes-required diagnostic: %+v", diagnostic)
		}
		if diagnostic.Severity == diag.SeverityError {
			t.Fatalf("unexpected validation error: %+v", diagnostic)
		}
	}
}

func TestLoadRejectsScalarActorEntryWithHelpfulMessage(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "adoption", "arc-0062.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(`arc: 62
title: ASA Circulating Supply
last-reviewed: 2026-04-09
adoption:
  wallets: []
  explorers:
    - wow
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

	_, diagnostics, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:016" && strings.Contains(diagnostic.Message, `adoption.explorers[0] must be an actor object`) {
			found = true
			if strings.Contains(diagnostic.Message, "cannot unmarshal") {
				t.Fatalf("expected targeted schema message, got %+v", diagnostic)
			}
		}
	}
	if !found {
		t.Fatalf("expected targeted actor schema diagnostic, got %+v", diagnostics)
	}
}

func TestValidateRejectsLegacyARCMetadataFields(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	document := loadValidatedDocument(t, root, filepath.Join(root, "ARCs", "arc-0042.md"))
	adoptionPath := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 42
title: Example ARC
status: Draft
last-reviewed: 2026-04-09
sponsor: Foundation
implementation-required: false
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

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}

	registry, registryDiagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}

	diagnostics := Validate(summary, document, registry)
	for _, field := range []string{"status is not allowed", "sponsor is not allowed", "implementation-required is not allowed"} {
		if !containsDiagnosticMessage(diagnostics, field) {
			t.Fatalf("expected diagnostic containing %q, got %+v", field, diagnostics)
		}
	}
}

func TestValidateRequiresReferenceImplementationFromARCFrontMatter(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	document := loadValidatedDocument(t, root, filepath.Join(root, "ARCs", "arc-0044.md"))
	adoptionPath := filepath.Join(root, "adoption", "arc-0044.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 44
title: Transition Ready ARC
last-reviewed: 2026-03-26
adoption:
  wallets:
    - name: example-wallet
      status: shipped
      evidence: https://example.com/wallet-proof
      notes: ""
  explorers: []
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: medium
  blockers: []
  notes: ""
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	summary, loadDiagnostics, err := Load(adoptionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loadDiagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", loadDiagnostics)
	}

	registry, registryDiagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(registryDiagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", registryDiagnostics)
	}

	diagnostics := Validate(summary, document, registry)
	if !containsDiagnosticMessage(diagnostics, "reference-implementation is required when the ARC front matter sets implementation-required to true") {
		t.Fatalf("expected ARC-derived reference-implementation requirement, got %+v", diagnostics)
	}
}

func loadValidatedDocument(t *testing.T, root string, path string) *arc.Document {
	t.Helper()
	document, diagnostics, err := arc.Load(path)
	if err != nil {
		t.Fatalf("arc.Load() error = %v", err)
	}
	diagnostics = append(diagnostics, arc.Validate(document, root)...)
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == diag.SeverityError {
			t.Fatalf("unexpected ARC validation error: %+v", diagnostic)
		}
	}
	return document
}

func containsDiagnosticMessage(diagnostics []diag.Diagnostic, fragment string) bool {
	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic.Message, fragment) {
			return true
		}
	}
	return false
}
