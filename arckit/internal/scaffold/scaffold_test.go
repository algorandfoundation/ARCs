package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitARCCreatesExpectedFiles(t *testing.T) {
	root := t.TempDir()
	created, diagnostics, err := InitARC(InitOptions{
		Root:                   root,
		Number:                 99,
		Title:                  "Scaffolded ARC",
		Type:                   "Meta",
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
}
