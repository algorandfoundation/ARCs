package arc

import (
	"path/filepath"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

func TestValidateValidDraftARC(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "valid-draft")
	path := filepath.Join(root, "ARCs", "arc-0042.md")

	document, diagnostics, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", diagnostics)
	}
	validationDiagnostics := Validate(document, root)
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.Severity == diag.SeverityError {
			t.Fatalf("unexpected validation error: %+v", diagnostic)
		}
	}
}
