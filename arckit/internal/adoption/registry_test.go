package adoption

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndValidateVettedAdopters(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "valid-draft")
	registry, diagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", diagnostics)
	}
	if diagnostics := ValidateVettedAdopters(registry); len(diagnostics) != 0 {
		t.Fatalf("ValidateVettedAdopters() diagnostics = %v", diagnostics)
	}
}

func TestValidateVettedAdoptersRejectsInvalidContent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, VettedAdoptersFileName)
	content := `wallets:
  - example-wallet
  - ExampleWallet
  - example-wallet
explorers: []
infra: []
dapps-protocols: []
extra: []
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	registry, diagnostics, err := LoadVettedAdopters(path)
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("LoadVettedAdopters() diagnostics = %v", diagnostics)
	}

	diagnostics = ValidateVettedAdopters(registry)
	if len(diagnostics) == 0 {
		t.Fatalf("expected validation diagnostics, got none")
	}

	want := []string{
		`missing required vetted adopter category "sdk-libraries"`,
		`unsupported vetted adopter category "extra"`,
		`wallets[1] must be lower-kebab-case`,
		`duplicate vetted adopter "example-wallet" in wallets`,
	}
	for _, substring := range want {
		found := false
		for _, diagnostic := range diagnostics {
			if diagnostic.RuleID == "R:022" && strings.Contains(diagnostic.Message, substring) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected diagnostic containing %q, got %+v", substring, diagnostics)
		}
	}
}

func TestLoadVettedAdoptersRejectsInvalidSchema(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, VettedAdoptersFileName)
	content := `wallets: []
explorers: {}
sdk-libraries: []
infra: []
dapps-protocols: []
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, diagnostics, err := LoadVettedAdopters(path)
	if err != nil {
		t.Fatalf("LoadVettedAdopters() error = %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatalf("expected schema diagnostics, got none")
	}
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:022" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected R:022 diagnostics, got %+v", diagnostics)
	}
}
