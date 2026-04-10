package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

func TestLoadMissingConfig(t *testing.T) {
	root := t.TempDir()
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.IgnoreARC(42) {
		t.Fatalf("IgnoreARC(42) = true, want false")
	}
}

func TestLoadValidJSONCConfig(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{
  // Ignore one ARC completely.
  "ignoreArcs": [0, 42],
  "ignoreRules": ["R:020"],
  /* Ignore these rules for exact and range selectors. */
  "ignoreByArc": {
    "0": ["R:008"],
    "0043": ["R:009", "R:013"],
    "50-60": ["R:011"]
  }
}`)

	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.IgnoreARC(42) {
		t.Fatalf("IgnoreARC(42) = false, want true")
	}
	if !cfg.IgnoreARC(0) {
		t.Fatalf("IgnoreARC(0) = false, want true")
	}
	if !cfg.IgnoreRule("R:020") {
		t.Fatalf("IgnoreRule(R:020) = false, want true")
	}
	if !cfg.IgnoreRuleForARC("R:008", 0) {
		t.Fatalf("IgnoreRuleForARC(R:008, 0) = false, want true")
	}
	if !cfg.IgnoreRuleForARC("R:009", 43) {
		t.Fatalf("IgnoreRuleForARC(R:009, 43) = false, want true")
	}
	if !cfg.IgnoreRuleForARC("R:011", 55) {
		t.Fatalf("IgnoreRuleForARC(R:011, 55) = false, want true")
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "invalid-json",
			content: `{"ignoreArcs":[}`,
			wantErr: "could not parse",
		},
		{
			name:    "unknown-key",
			content: `{"extra":true}`,
			wantErr: "unknown top-level key",
		},
		{
			name:    "reversed-range",
			content: `{"ignoreByArc":{"60-50":["R:011"]}}`,
			wantErr: "greater than end",
		},
		{
			name:    "unknown-rule",
			content: `{"ignoreRules":["R:999"]}`,
			wantErr: "unknown rule",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeConfig(t, root, test.content)
			_, err := Load(root)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Load() error = %v, want substring %q", err, test.wantErr)
			}
		})
	}
}

func TestFilterDiagnostics(t *testing.T) {
	cfg := Config{
		ignoreArcs: map[int]struct{}{
			0:  {},
			42: {},
		},
		ignoreRules: map[string]struct{}{
			"R:020": {},
		},
		ignoreByArc: []arcRuleIgnore{
			{
				start: 0,
				end:   0,
				rules: map[string]struct{}{
					"R:008": {},
				},
			},
			{
				start: 50,
				end:   60,
				rules: map[string]struct{}{
					"R:011": {},
				},
			},
		},
	}

	diagnostics := []diag.Diagnostic{
		{RuleID: "R:008", File: filepath.Join("ARCs", "arc-0000.md")},
		{RuleID: "R:020", File: filepath.Join("ARCs", "arc-0041.md")},
		{RuleID: "R:012", File: filepath.Join("ARCs", "arc-0042.md")},
		{RuleID: "R:011", File: filepath.Join("assets", "arc-0055", "example.txt")},
		{RuleID: "R:009", File: filepath.Join("ARCs", "arc-0061.md")},
		{RuleID: "R:022", File: filepath.Join("adoption", "vetted-adopters.yaml")},
	}

	filtered := cfg.FilterDiagnostics(diagnostics)
	if len(filtered) != 2 {
		t.Fatalf("len(FilterDiagnostics()) = %d, want 2", len(filtered))
	}
	if filtered[0].RuleID != "R:009" {
		t.Fatalf("FilterDiagnostics()[0].RuleID = %q, want R:009", filtered[0].RuleID)
	}
	if filtered[1].RuleID != "R:022" {
		t.Fatalf("FilterDiagnostics()[1].RuleID = %q, want R:022", filtered[1].RuleID)
	}
}

func TestWithRuleEnforced(t *testing.T) {
	cfg := Config{
		ignoreArcs: map[int]struct{}{
			0:  {},
			42: {},
		},
		ignoreRules: map[string]struct{}{
			"R:020": {},
			"R:021": {},
		},
		ignoreByArc: []arcRuleIgnore{
			{
				start: 0,
				end:   0,
				rules: map[string]struct{}{
					"R:008": {},
					"R:021": {},
				},
			},
		},
	}

	enforced := cfg.WithRuleEnforced("R:021")
	if enforced.IgnoreARC(42) != cfg.IgnoreARC(42) {
		t.Fatalf("WithRuleEnforced() should preserve ignoreArcs")
	}
	if enforced.IgnoreRule("R:021") {
		t.Fatalf("WithRuleEnforced() should remove global suppression for R:021")
	}
	if enforced.IgnoreRule("R:020") != cfg.IgnoreRule("R:020") {
		t.Fatalf("WithRuleEnforced() should preserve unrelated global suppressions")
	}
	if enforced.IgnoreRuleForARC("R:021", 0) {
		t.Fatalf("WithRuleEnforced() should remove ARC-specific suppression for R:021")
	}
	if !enforced.IgnoreRuleForARC("R:008", 0) {
		t.Fatalf("WithRuleEnforced() should preserve unrelated ARC-specific suppressions")
	}
	if !cfg.IgnoreRule("R:021") || !cfg.IgnoreRuleForARC("R:021", 0) {
		t.Fatalf("WithRuleEnforced() should not mutate the original config")
	}
}

func writeConfig(t *testing.T, root string, content string) {
	t.Helper()
	path := filepath.Join(root, FileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
