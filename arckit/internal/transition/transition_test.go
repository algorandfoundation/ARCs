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
last-reviewed: 2026-03-26
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
  adoption-readiness: low
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
last-reviewed: 2026-03-26
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
  adoption-readiness: low
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
