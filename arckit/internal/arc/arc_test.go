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

func TestValidateRequiresImplementationDeclarationForReviewAndLater(t *testing.T) {
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
author:
  - Example Author
discussions-to: https://example.com/discussion
status: Review
type: Standards Track
created: 2026-04-08
sponsor: Foundation
implementation-required: true
adoption-summary: adoption/arc-0001.yaml
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
	if len(diagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %v", diagnostics)
	}

	validationDiagnostics := Validate(document, root)
	missingURL := false
	missingMaintainer := false
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:007" && strings.Contains(diagnostic.Message, "requires implementation-url") {
			missingURL = true
		}
		if diagnostic.RuleID == "R:007" && strings.Contains(diagnostic.Message, "requires implementation-maintainer") {
			missingMaintainer = true
		}
	}
	if !missingURL || !missingMaintainer {
		t.Fatalf("expected implementation declaration diagnostics, got %+v", validationDiagnostics)
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
author:
  - Example Author
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
author:
  - Example Author
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

func TestLoadRejectsBlankLinesInFrontMatter(t *testing.T) {
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
author:
  - Example Author
discussions-to: https://example.com/discussion
status: Draft
type: Meta
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
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleID == "R:024" {
			if diagnostic.Line != 4 {
				t.Fatalf("expected R:024 at line 4, got %+v", diagnostic)
			}
			return
		}
	}
	t.Fatalf("expected R:024 diagnostic, got %+v", diagnostics)
}

func TestValidateLinksReportsFileLineNumbers(t *testing.T) {
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
author:
  - Example Author
discussions-to: https://example.com/discussion
status: Draft
type: Meta
created: 2026-04-08
sponsor: Foundation
implementation-required: false
---

## Abstract

See [broken](./missing.md).

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
	if len(diagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %+v", diagnostics)
	}

	validationDiagnostics := Validate(document, root)
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:009" {
			if diagnostic.Line != 17 {
				t.Fatalf("expected R:009 at line 17, got %+v", diagnostic)
			}
			return
		}
	}
	t.Fatalf("expected R:009 diagnostic, got %+v", validationDiagnostics)
}

func TestValidateRejectsLegacyScalarListFields(t *testing.T) {
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
author: Example Author, Other Author
discussions-to: https://example.com/discussion
status: Draft
type: Standards Track
created: 2026-04-08
updated: 2026-04-09, 2026-04-10
sponsor: Foundation
implementation-required: false
implementation-maintainer: algorandfoundation, algorandecosystem
requires: 4, 22
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
	if len(diagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %+v", diagnostics)
	}

	validationDiagnostics := Validate(document, root)
	count := 0
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:021" {
			count++
		}
	}
	if count != 4 {
		t.Fatalf("expected 4 R:021 diagnostics, got %+v", validationDiagnostics)
	}
	if len(document.Requires) != 0 {
		t.Fatalf("expected legacy scalar requires to be ignored by relationship logic, got %+v", document.Requires)
	}
	if document.ImplementationMaintainer != "" {
		t.Fatalf("expected legacy scalar implementation-maintainer to be ignored, got %q", document.ImplementationMaintainer)
	}
}

func TestValidateRejectsNonScalarSupersededBy(t *testing.T) {
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
author:
  - Example Author
discussions-to: https://example.com/discussion
status: Draft
type: Standards Track
created: 2026-04-08
sponsor: Foundation
implementation-required: false
superseded-by:
  - 2
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
	if len(diagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %+v", diagnostics)
	}

	validationDiagnostics := Validate(document, root)
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:007" && strings.Contains(diagnostic.Message, "superseded-by") {
			return
		}
	}
	t.Fatalf("expected R:007 for non-scalar superseded-by, got %+v", validationDiagnostics)
}

func TestValidateDetectsARCZeroFilenameMismatch(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	path := filepath.Join(arcDir, "arc-0000.md")
	content := `---
arc: 1
title: Example
description: Example description
author:
  - Example Author
discussions-to: https://example.com/discussion
status: Draft
type: Meta
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
	if len(diagnostics) != 0 {
		t.Fatalf("Load() diagnostics = %+v", diagnostics)
	}

	validationDiagnostics := Validate(document, root)
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:007" && strings.Contains(diagnostic.Message, "filename number 0") {
			return
		}
	}
	t.Fatalf("expected R:007 ARC/file mismatch for arc-0000.md, got %+v", validationDiagnostics)
}

func TestFindRepoRootUsesConfigFileMarker(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "docs", "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	configPath := filepath.Join(root, ".arckit.jsonc")
	if err := os.WriteFile(configPath, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", configPath, err)
	}

	found := FindRepoRoot(nested)
	if found != root {
		t.Fatalf("FindRepoRoot(%q) = %q, want %q", nested, found, root)
	}
}
