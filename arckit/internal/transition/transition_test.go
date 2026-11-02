package transition

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
	"github.com/algorandfoundation/ARCs/arckit/internal/testutil"
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

func TestValidateFinalTransitionRequiresShippedReferenceImplementationStatus(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	adoptionPath := filepath.Join(root, "adoption", "arc-0044.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 44
title: Transition Ready ARC
reference-implementation:
  status: wip
  notes: ""
adoption:
  wallets:
    - name: example-wallet
      status: shipped
      evidence: https://example.com/wallet-proof
      notes: ""
  explorers: []
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  blockers: []
  notes: ""
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	path := filepath.Join(root, "ARCs", "arc-0044.md")
	diagnostics, err := Validate(path, "Final")
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:019" && diagnostic.Message == "reference-implementation.status must be shipped" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected shipped reference implementation transition diagnostic, got %+v", diagnostics)
	}
}

func TestValidateTransitionRequiresVettedAdoptersRegistry(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
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

func TestValidateFinalTransitionReportsMissingEvidenceOnce(t *testing.T) {
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	adoptionPath := filepath.Join(root, "adoption", "arc-0044.yaml")
	if err := os.WriteFile(adoptionPath, []byte(`arc: 44
title: Transition Ready ARC
reference-implementation:
  status: shipped
  notes: ""
adoption:
  wallets:
    - name: example-wallet
      status: shipped
      evidence: ""
      notes: ""
  explorers: []
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  blockers: []
  notes: ""
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	path := filepath.Join(root, "ARCs", "arc-0044.md")
	diagnostics, err := Validate(path, "Final")
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	count := 0
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:019" && strings.Contains(diagnostic.Message, "evidence") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected one evidence-related transition diagnostic, got %d: %+v", count, diagnostics)
	}
}

func applyFinalStatus(t *testing.T) string {
	t.Helper()
	root := testutil.CopyDir(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	path := filepath.Join(root, "ARCs", "arc-0044.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	updated := strings.Replace(string(content), "status: Last Call\n", "status: Final\n", 1)
	updated = strings.Replace(updated, "last-call-deadline: 2026-04-01\n", "", 1)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func TestValidateAppliedFinalTransition(t *testing.T) {
	diagnostics, err := ValidateApplied(applyFinalStatus(t), "Last Call", "Final")
	if err != nil {
		t.Fatalf("ValidateApplied() error = %v", err)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == diag.SeverityError {
			t.Fatalf("unexpected error diagnostic: %+v", diagnostic)
		}
	}
}

func TestValidateAppliedTransitionRejectsSkippedGate(t *testing.T) {
	diagnostics, err := ValidateApplied(applyFinalStatus(t), "Draft", "Final")
	if err != nil {
		t.Fatalf("ValidateApplied() error = %v", err)
	}
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:019" && diagnostic.Message == "transition to Final requires current status Last Call" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected skipped gate diagnostic, got %+v", diagnostics)
	}
}

func TestValidateAppliedTransitionRequiresTargetStatus(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "repos", "transition-final", "ARCs", "arc-0044.md")
	diagnostics, err := ValidateApplied(path, "Last Call", "Final")
	if err != nil {
		t.Fatalf("ValidateApplied() error = %v", err)
	}
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:019" && diagnostic.Message == "status must be Final after the transition" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected target status diagnostic, got %+v", diagnostics)
	}
}
