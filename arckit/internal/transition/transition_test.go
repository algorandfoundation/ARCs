package transition

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

func TestValidateFinalTransition(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "repos", "transition-final", "ARCs", "arc-0044.md")
	diagnostics, err := Validate(path, "Final")
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == diag.SeverityError {
			t.Fatalf("unexpected error diagnostic: %+v", diagnostic)
		}
	}
}

func TestValidateIdleTransition(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "repos", "transition-idle", "ARCs", "arc-0045.md")
	diagnostics, err := Validate(path, "Idle")
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == diag.SeverityError {
			t.Fatalf("unexpected error diagnostic: %+v", diagnostic)
		}
	}
}

func TestValidateTransitionRequiresVettedAdoptersRegistry(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	if err := os.Remove(filepath.Join(root, "adoption", "vetted-adopters.yaml")); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	path := filepath.Join(root, "ARCs", "arc-0044.md")
	diagnostics, err := Validate(path, "Final")
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
