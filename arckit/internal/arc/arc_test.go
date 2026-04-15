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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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

func TestValidateRequiresCanonicalImplementationURL(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	path := filepath.Join(arcDir, "arc-0044.md")
	content := `---
arc: 44
title: Example
description: Example description
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Standards Track
created: 2026-04-08
sponsor: Foundation
implementation-required: true
implementation-url: https://github.com/example/arc-0044
implementation-maintainer:
  - algorandfoundation
adoption-summary: adoption/arc-0044.yaml
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
	found := false
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:029" && strings.Contains(diagnostic.Message, "https://github.com/algorandfoundation/arc44") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected canonical implementation-url diagnostic, got %+v", validationDiagnostics)
	}
}

func TestValidateAcceptsCanonicalImplementationURLForEcosystemSponsor(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	path := filepath.Join(arcDir, "arc-0044.md")
	content := `---
arc: 44
title: Example
description: Example description
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Review
type: Standards Track
created: 2026-04-08
sponsor: Ecosystem
implementation-required: true
implementation-url: https://github.com/algorandecosystem/arc44
implementation-maintainer:
  - algorandecosystem
adoption-summary: adoption/arc-0044.yaml
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:029" {
			t.Fatalf("unexpected canonical implementation-url diagnostic: %+v", diagnostic)
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
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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

func TestValidateRejectsUnsupportedCategory(t *testing.T) {
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Standards Track
category: ARC
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
	found := false
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:030" && strings.Contains(diagnostic.Message, `unsupported category "ARC"`) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected R:030 unsupported category diagnostic, got %+v", validationDiagnostics)
	}
}

func TestValidateRejectsUnsupportedSubCategory(t *testing.T) {
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Standards Track
category: Interface
sub-category: Asa
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
	found := false
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:030" && strings.Contains(diagnostic.Message, `unsupported sub-category "Asa"`) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected R:030 unsupported sub-category diagnostic, got %+v", validationDiagnostics)
	}
}

func TestValidateRejectsSubCategoryWithoutCategory(t *testing.T) {
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Meta
sub-category: General
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
	found := false
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:030" && strings.Contains(diagnostic.Message, "sub-category requires category") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected R:030 sub-category requires category diagnostic, got %+v", validationDiagnostics)
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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

func TestValidateIgnoresEmailAutolinksInBodyText(t *testing.T) {
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Meta
created: 2026-04-08
sponsor: Foundation
implementation-required: false
---

## Abstract

Random J. User <address@dom.ain> (@github-handle)

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
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
			t.Fatalf("expected email autolink to be ignored, got %+v", validationDiagnostics)
		}
	}
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
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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

func TestValidateRejectsLegacyMetadataRequirements(t *testing.T) {
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
last-call-deadline: 2026-04-30
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
	foundDiscussion := false
	foundAuthor := false
	foundConditional := false
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:032" && strings.Contains(diagnostic.Message, "discussions-to") {
			foundDiscussion = true
		}
		if diagnostic.RuleID == "R:032" && strings.Contains(diagnostic.Message, "GitHub handle") {
			foundAuthor = true
		}
		if diagnostic.RuleID == "R:033" && strings.Contains(diagnostic.Message, "last-call-deadline") {
			foundConditional = true
		}
	}
	if !foundDiscussion || !foundAuthor || !foundConditional {
		t.Fatalf("expected discussions-to, author, and conditional field diagnostics, got %+v", validationDiagnostics)
	}
}

func TestValidateRejectsLegacyTitleDescriptionAndSortedListRules(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	path := filepath.Join(arcDir, "arc-0001.md")
	content := `---
arc: 1
title: A
description: This standard depends on arc 3.
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Standards Track
created: 2026-04-08
sponsor: Foundation
implementation-required: false
requires:
  - 5
  - 4
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
	foundLength := false
	foundStandard := false
	foundARCSpelling := false
	foundRequires := false
	foundSorted := false
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:031" && strings.Contains(diagnostic.Message, "between 2 and 44") {
			foundLength = true
		}
		if diagnostic.RuleID == "R:031" && strings.Contains(diagnostic.Message, "standard-like wording") {
			foundStandard = true
		}
		if diagnostic.RuleID == "R:031" && strings.Contains(diagnostic.Message, "arc 3") {
			foundARCSpelling = true
		}
		if diagnostic.RuleID == "R:031" && strings.Contains(diagnostic.Message, "must also appear in requires") {
			foundRequires = true
		}
		if diagnostic.RuleID == "R:034" {
			foundSorted = true
		}
	}
	if !foundLength || !foundStandard || !foundARCSpelling || !foundRequires || !foundSorted {
		t.Fatalf("expected title/description/list diagnostics, got %+v", validationDiagnostics)
	}
}

func TestValidateRejectsLegacyBodySectionAndLinkRules(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	targetPath := filepath.Join(arcDir, "arc-0002.md")
	targetContent := `---
arc: 2
title: Target ARC
description: Example target ARC.
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/2
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

## Copyright

Copyright and related rights waived via CC0 1.0.
`
	if err := os.WriteFile(targetPath, []byte(targetContent), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	path := filepath.Join(arcDir, "arc-0001.md")
	content := `---
arc: 1
title: Example
description: Example description
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Meta
created: 2026-04-08
sponsor: Foundation
implementation-required: false
---

## Abstract

Text with arc-2 before the link and [ARC-2](./arc-0002.md) after it.

## Motivation

See [absolute repo](https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0002.md).

## Security Considerations

Text

## Rationale

Text

## Extra

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
	foundMissingCopyright := false
	foundExtraSection := false
	foundOutOfOrder := false
	foundBodyCase := false
	foundFirstMention := false
	foundAbsolute := false
	for _, diagnostic := range validationDiagnostics {
		if diagnostic.RuleID == "R:035" && strings.Contains(diagnostic.Message, "missing required section") {
			foundMissingCopyright = true
		}
		if diagnostic.RuleID == "R:035" && strings.Contains(diagnostic.Message, "unsupported level-2 section") {
			foundExtraSection = true
		}
		if diagnostic.RuleID == "R:035" && strings.Contains(diagnostic.Message, "out of order") {
			foundOutOfOrder = true
		}
		if diagnostic.RuleID == "R:036" && strings.Contains(diagnostic.Message, "arc-2") {
			foundBodyCase = true
		}
		if diagnostic.RuleID == "R:036" && strings.Contains(diagnostic.Message, "first body mention") {
			foundFirstMention = true
		}
		if diagnostic.RuleID == "R:037" {
			foundAbsolute = true
		}
	}
	if !foundMissingCopyright || !foundExtraSection || !foundOutOfOrder || !foundBodyCase || !foundFirstMention || !foundAbsolute {
		t.Fatalf("expected section/body/link diagnostics, got %+v", validationDiagnostics)
	}
}

func TestValidateAllowsExternalHTMLAnchorLinks(t *testing.T) {
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Meta
created: 2026-04-08
sponsor: Foundation
implementation-required: false
---

## Abstract

See <a href="https://example.com/spec">external spec</a>.

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
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
		if diagnostic.RuleID == "R:037" {
			t.Fatalf("expected external HTML anchor to be allowed, got %+v", validationDiagnostics)
		}
	}
}

func TestValidateRejectsRepositoryAbsoluteHTMLAnchorLinks(t *testing.T) {
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
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Meta
created: 2026-04-08
sponsor: Foundation
implementation-required: false
---

## Abstract

See <a href="https://github.com/algorandfoundation/ARCs/blob/main/ARCs/arc-0002.md">ARC-2</a>.

## Motivation

Text

## Specification

Text

## Rationale

Text

## Security Considerations

Text

## Copyright

Copyright and related rights waived via CC0 1.0.
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
		if diagnostic.RuleID == "R:037" {
			return
		}
	}
	t.Fatalf("expected repository absolute HTML anchor to trigger R:037, got %+v", validationDiagnostics)
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
