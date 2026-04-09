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
	if len(created) != 3 {
		t.Fatalf("expected 3 created paths, got %d", len(created))
	}
	for _, path := range []string{
		filepath.Join(root, "ARCs", "arc-0099.md"),
		filepath.Join(root, "adoption", "arc-0099.yaml"),
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
