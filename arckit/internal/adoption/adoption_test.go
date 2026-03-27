package adoption

import (
	"path/filepath"
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
	for _, diagnostic := range Validate(summary, document) {
		if diagnostic.Severity == diag.SeverityError {
			t.Fatalf("unexpected validation error: %+v", diagnostic)
		}
	}
}
