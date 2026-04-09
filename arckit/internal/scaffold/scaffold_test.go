package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitARCCreatesExpectedFiles(t *testing.T) {
	root := t.TempDir()
	created, diagnostics, err := InitARC(InitOptions{
		Root:                   root,
		Number:                 99,
		Title:                  "Scaffolded ARC",
		Type:                   "Standards Track",
		Category:               "Interface",
		SubCategory:            "Application",
		Sponsor:                "Foundation",
		ImplementationRequired: true,
	})
	if err != nil {
		t.Fatalf("InitARC() error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("InitARC() diagnostics = %v", diagnostics)
	}
	if len(created) != 4 {
		t.Fatalf("expected 4 created paths, got %d", len(created))
	}
	for _, path := range []string{
		filepath.Join(root, "ARCs", "arc-0099.md"),
		filepath.Join(root, "adoption", "arc-0099.yaml"),
		filepath.Join(root, "adoption", "vetted-adopters.yaml"),
		filepath.Join(root, "assets", "arc-0099"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}

	arcContent, err := os.ReadFile(filepath.Join(root, "ARCs", "arc-0099.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(arcContent)
	if !strings.Contains(text, "author:\n  - TBD\ndiscussions-to:\nstatus: Draft\ntype: Standards Track\ncategory: Interface\nsub-category: Application\ncreated: ") {
		t.Fatalf("expected scaffolded ARC to include category fields, got:\n%s", text)
	}
}

func TestInitARCPreservesExistingVettedAdoptersRegistry(t *testing.T) {
	root := t.TempDir()
	registryPath := filepath.Join(root, "adoption", "vetted-adopters.yaml")
	if err := os.MkdirAll(filepath.Dir(registryPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	original := "wallets:\n  - existing-wallet\nexplorers: []\nsdk-libraries: []\ninfra: []\ndapps-protocols: []\n"
	if err := os.WriteFile(registryPath, []byte(original), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	created, diagnostics, err := InitARC(InitOptions{
		Root:                   root,
		Number:                 100,
		Title:                  "Scaffolded ARC",
		Type:                   "Standards Track",
		Sponsor:                "Foundation",
		ImplementationRequired: false,
	})
	if err != nil {
		t.Fatalf("InitARC() error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("InitARC() diagnostics = %v", diagnostics)
	}
	if len(created) != 3 {
		t.Fatalf("expected 3 created paths when registry already exists, got %d", len(created))
	}

	content, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != original {
		t.Fatalf("expected existing vetted adopters registry to be preserved, got:\n%s", string(content))
	}
}
