package transition

import (
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
