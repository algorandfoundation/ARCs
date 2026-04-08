package arc

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadAcceptsCategoryField(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	path := filepath.Join(arcDir, "arc-0001.md")
	content := `---
arc: 1
title: Example
description: Example description
author: Example Author
discussions-to: https://example.com/discussion
status: Draft
type: Standards Track
category: Interface
sub-category: Application
created: 2026-04-08
sponsor: Foundation
implementation-required: false
---

## Abstract

Text

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	document, diagnostics, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:006" {
			t.Fatalf("expected category to be recognized, got diagnostics = %+v", diagnostics)
		}
	}

	validationDiagnostics := Validate(document, root)
	if document.Category != "Interface" {
		t.Fatalf("expected category to be captured, got %q", document.Category)
	}
	if document.SubCategory != "Application" {
		t.Fatalf("expected sub-category to be captured, got %q", document.SubCategory)
	}
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.Severity == diag.SeverityError {
			t.Fatalf("unexpected validation error: %+v", diagnostic)
		}
	}
}

func TestValidateStillRejectsUnknownFields(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	path := filepath.Join(arcDir, "arc-0001.md")
	content := `---
arc: 1
title: Example
description: Example description
author: Example Author
discussions-to: https://example.com/discussion
status: Draft
type: Standards Track
category: Interface
custom-field: keep me
created: 2026-04-08
sponsor: Foundation
implementation-required: false
---

## Abstract

Text

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, diagnostics, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	foundUnknown := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:006" && strings.Contains(diagnostic.Message, "custom-field") {
			foundUnknown = true
		}
	}
	if !foundUnknown {
		t.Fatalf("expected R:006 for custom-field, got diagnostics = %+v", diagnostics)
	}
}
