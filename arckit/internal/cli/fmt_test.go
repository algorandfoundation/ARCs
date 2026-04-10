package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/config"
)

func TestApplyNativeFixReordersFrontMatterWithoutNormalizingBodyWhitespace(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(arcDir, "arc-0001.md")
	content := `---
title: Example
arc: 1
description: Example description
author: Example Author, Another Author
discussions-to: https://example.com/discussion
status: Draft
type: Meta
created: 2026-03-26
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

Text`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := applyNativeFix(path); err != nil {
		t.Fatalf("applyNativeFix() error = %v", err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	if !strings.Contains(text, "arc: 1\ntitle: Example\n") {
		t.Fatalf("expected reordered front matter, got:\n%s", text)
	}
	if !strings.Contains(text, "author: Example Author, Another Author\n") {
		t.Fatalf("expected non-canonical author scalar to be preserved, got:\n%s", text)
	}
	if !strings.Contains(text, "created: 2026-03-26\n") {
		t.Fatalf("expected created to remain YYYY-MM-DD, got:\n%s", text)
	}
	if !strings.Contains(text, "## Abstract\n\nText\n\n## Motivation\n") {
		t.Fatalf("expected body whitespace to be preserved, got:\n%s", text)
	}
	if !strings.HasSuffix(text, "Text") {
		t.Fatalf("expected final newline policy to remain unchanged, got:\n%s", text)
	}
}

func TestApplyNativeFixKeepsDateFieldsAsDateOnly(t *testing.T) {
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
status: Last Call
type: Meta
created: 2026-03-26
last-call-deadline: 2026-04-01
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

Text`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := applyNativeFix(path); err != nil {
		t.Fatalf("applyNativeFix() error = %v", err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	if strings.Contains(text, "T00:00:00Z") {
		t.Fatalf("expected date-only fields, got:\n%s", text)
	}
	if !strings.Contains(text, "created: 2026-03-26\n") || !strings.Contains(text, "last-call-deadline: 2026-04-01\n") {
		t.Fatalf("expected date-only fields, got:\n%s", text)
	}
}

func TestApplyNativeFixPreservesExtraBlankLinesAfterFrontMatter(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(arcDir, "arc-0001.md")
	content := `---
title: Example
arc: 1
description: Example description
author:
  - Example Author
discussions-to: https://example.com/discussion
status: Draft
type: Meta
created: 2026-03-26
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

Text`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := applyNativeFix(path); err != nil {
		t.Fatalf("applyNativeFix() error = %v", err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	if !strings.Contains(text, "implementation-required: false\n---\n\n\n## Abstract\n") {
		t.Fatalf("expected extra blank lines after front matter to be preserved, got:\n%s", text)
	}
}

func TestApplyNativeFixRemovesBlankLinesInsideFrontMatter(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(arcDir, "arc-0001.md")
	content := `---
title: Example

arc: 1
description: Example description
author:
  - Example Author
discussions-to: https://example.com/discussion

status: Draft
type: Meta
created: 2026-03-26
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

Text`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := applyNativeFix(path); err != nil {
		t.Fatalf("applyNativeFix() error = %v", err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	if strings.Contains(text, "title: Example\n\narc: 1") || strings.Contains(text, "discussions-to: https://example.com/discussion\n\nstatus: Draft") {
		t.Fatalf("expected front matter blank lines to be removed, got:\n%s", text)
	}
	if !strings.Contains(text, "implementation-required: false\n---\n\n## Abstract\n") {
		t.Fatalf("expected body spacing to remain intact after front matter cleanup, got:\n%s", text)
	}
}

func TestApplyAdoptionFixRejectsUnknownAdoptionCategory(t *testing.T) {
	root := t.TempDir()
	adoptionDir := filepath.Join(root, "adoption")
	if err := os.MkdirAll(adoptionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(adoptionDir, "arc-0001.yaml")
	content := `arc: 1
title: Example
last-reviewed: 2026-04-10
adoption:
  wallets: []
  explorers: []
  tooling: []
  infra: []
  dapps-protocols: []
  typo-category: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := applyAdoptionFixWithConfig(path)
	if err == nil {
		t.Fatalf("expected applyAdoptionFixWithConfig() to reject unknown adoption category")
	}
	if !strings.Contains(err.Error(), "could not be safely reformatted") {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if string(updated) != content {
		t.Fatalf("expected content to remain unchanged, got:\n%s", string(updated))
	}
}

func TestApplyAdoptionFixRejectsMissingCanonicalAdoptionCategory(t *testing.T) {
	root := t.TempDir()
	adoptionDir := filepath.Join(root, "adoption")
	if err := os.MkdirAll(adoptionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(adoptionDir, "arc-0001.yaml")
	content := `arc: 1
title: Example
last-reviewed: 2026-04-10
adoption:
  wallets: []
  explorers: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := applyAdoptionFixWithConfig(path)
	if err == nil {
		t.Fatalf("expected applyAdoptionFixWithConfig() to reject missing canonical adoption category")
	}
	if !strings.Contains(err.Error(), "could not be safely reformatted") {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if string(updated) != content {
		t.Fatalf("expected content to remain unchanged, got:\n%s", string(updated))
	}
}

func TestApplyNativeFixWithConfigHonorsIgnoredRulesAndPreservesUnknownFieldText(t *testing.T) {
	root := t.TempDir()
	arcDir := filepath.Join(root, "ARCs")
	if err := os.MkdirAll(arcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	configPath := filepath.Join(root, config.FileName)
	if err := os.WriteFile(configPath, []byte("{\"ignoreRules\":[\"R:006\"]}"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", configPath, err)
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	path := filepath.Join(arcDir, "arc-0001.md")
	content := `---
arc: 1
title: Example
description: Example description
author:
  - Example Author
discussions-to: https://example.com/discussion
status: Idle
idle-since: 2026-04-01
type: Standards Track
custom-field: keep me
requires: 4, 22
created: 2026-03-26
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

Text`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := applyNativeFixWithConfig(path, cfg); err != nil {
		t.Fatalf("applyNativeFixWithConfig() error = %v", err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	if !strings.Contains(text, "status: Idle\ntype: Standards Track\ncreated: 2026-03-26\nsponsor: Foundation\ncustom-field: keep me\n") {
		t.Fatalf("expected created/sponsor/custom-field ordering with unknown field preserved, got:\n%s", text)
	}
	if !strings.Contains(text, "requires: 4, 22\n") {
		t.Fatalf("expected non-canonical requires scalar to be preserved, got:\n%s", text)
	}
	if !strings.Contains(text, "implementation-required: false\nidle-since: 2026-04-01\nrequires: 4, 22\n") {
		t.Fatalf("expected idle-since before requires, got:\n%s", text)
	}
}
