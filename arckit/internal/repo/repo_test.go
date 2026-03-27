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

func writeConfig(t *testing.T, root string, content string) {
	t.Helper()
	path := filepath.Join(root, config.FileName)
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
