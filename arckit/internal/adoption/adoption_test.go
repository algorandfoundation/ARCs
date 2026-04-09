package adoption

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

func TestValidateMatchingAdoptionSummary(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "valid-draft")
	arcPath := filepath.Join(root, "ARCs", "arc-0042.md")
	adoptionPath := filepath.Join(root, "adoption", "arc-0042.yaml")

	document, diagnostics, err := arc.Load(arcPath)
	if err != nil {
		t.Fatalf("arc.Load() error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("arc.Load() diagnostics = %v", diagnostics)
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
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	registryPath := RegistryPath(root)
	if err := os.WriteFile(registryPath, []byte(`wallets: []
explorers:
  - example-wallet
sdk-libraries: []
infra: []
dapps-protocols: []
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	arcPath := filepath.Join(root, "ARCs", "arc-0044.md")
	adoptionPath := filepath.Join(root, "adoption", "arc-0044.yaml")

	document, diagnostics, err := arc.Load(arcPath)
	if err != nil {
		t.Fatalf("arc.Load() error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("arc.Load() diagnostics = %v", diagnostics)
	}

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

	diagnostics = Validate(summary, document, registry)
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

func TestValidateFinalAdoptionRequiresAtLeastOneAdopter(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	adoptionPath := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 42
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

	diagnostics := Validate(summary, nil, registry)
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

func TestValidateFinalAdoptionAllowsTrackedAdopter(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	adoptionPath := filepath.Join(root, "adoption", "arc-0044.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 44
title: Transition Ready ARC
status: Final
last-reviewed: 2026-03-26
sponsor: Foundation
implementation-required: true
reference-implementation:
  repository: https://github.com/example/arc-0044
  maintainers:
    - "@maintainer"
  status: testable
  notes: ""
adoption:
  wallets:
    - name: example-wallet
      status: shipped
      evidence: https://example.com/wallet-proof
      notes: ""
  explorers: []
  sdk-libraries: []
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

	for _, diagnostic := range Validate(summary, nil, registry) {
		if diagnostic.RuleID == "R:025" {
			t.Fatalf("unexpected Final ARC empty adoption diagnostic: %+v", diagnostic)
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
status: Final
last-reviewed: 2026-04-09
sponsor: ""
implementation-required: false
adoption:
  wallets: []
  explorers:
    - wow
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
