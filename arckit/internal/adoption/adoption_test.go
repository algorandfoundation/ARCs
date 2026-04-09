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
