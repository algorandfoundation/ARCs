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
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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
  - Example Author ()
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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
  - Example Author ()
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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
  - Example Author ()
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1

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
	if strings.Contains(text, "title: Example\n\narc: 1") || strings.Contains(text, "discussions-to: https://github.com/algorandfoundation/ARCs/issues/1\n\nstatus: Draft") {
		t.Fatalf("expected front matter blank lines to be removed, got:\n%s", text)
	}
	if !strings.Contains(text, "implementation-required: false\n---\n\n## Abstract\n") {
		t.Fatalf("expected body spacing to remain intact after front matter cleanup, got:\n%s", text)
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
  - Example Author ()
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

func TestApplyNativeFixPreservesMixedIntSequenceEntries(t *testing.T) {
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
  - Example Author ()
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Meta
created: 2026-03-26
sponsor: Foundation
implementation-required: false
requires:
  - 4
  - bad
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
	if !strings.Contains(text, "requires:\n  - 4\n  - bad\n") {
		t.Fatalf("expected mixed int sequence to be preserved verbatim, got:\n%s", text)
	}
}

func TestApplyNativeFixSortsAndDeduplicatesNumericARCLists(t *testing.T) {
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
  - Example Author ()
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Meta
created: 2026-03-26
sponsor: Foundation
implementation-required: false
requires:
  - 22
  - 4
  - 22
extended-by:
  - 89
  - 62
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

	if err := applyNativeFix(path); err != nil {
		t.Fatalf("applyNativeFix() error = %v", err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	if !strings.Contains(text, "requires:\n  - 4\n  - 22\n") {
		t.Fatalf("expected requires to be sorted and deduplicated, got:\n%s", text)
	}
	if !strings.Contains(text, "extended-by:\n  - 62\n  - 89\n") {
		t.Fatalf("expected extended-by to be sorted, got:\n%s", text)
	}
}

func TestApplyNativeFixReordersCanonicalLevel2Sections(t *testing.T) {
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
  - Example Author ()
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Draft
type: Meta
created: 2026-03-26
sponsor: Foundation
implementation-required: false
---

## Abstract

Abstract text.

## Specification

Specification text.

## Motivation

Motivation text.

## Security Considerations

Security text.

## Rationale

Rationale text.

## Copyright

Copyright and related rights waived via CC0 1.0.
`
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

	motivationIndex := strings.Index(text, "## Motivation")
	specificationIndex := strings.Index(text, "## Specification")
	rationaleIndex := strings.Index(text, "## Rationale")
	securityIndex := strings.Index(text, "## Security Considerations")
	copyrightIndex := strings.Index(text, "## Copyright")
	if !(motivationIndex < specificationIndex && specificationIndex < rationaleIndex && rationaleIndex < securityIndex && securityIndex < copyrightIndex) {
		t.Fatalf("expected canonical level-2 section order, got:\n%s", text)
	}
}

func TestApplyNativeFixKeepsImplementationMaintainerYAMLSafe(t *testing.T) {
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
  - Example Author ()
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Final
type: Meta
created: 2026-03-26
sponsor: Foundation
implementation-required: true
implementation-url: https://github.com/example/arc1
implementation-maintainer:
  - "@example"
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

Text`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := applyNativeFix(path); err != nil {
		t.Fatalf("first applyNativeFix() error = %v", err)
	}
	if err := applyNativeFix(path); err != nil {
		t.Fatalf("second applyNativeFix() error = %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	if !strings.Contains(text, "implementation-maintainer:\n  - \"@example\"\n") {
		t.Fatalf("expected implementation-maintainer quote style to be preserved, got:\n%s", text)
	}
}

func TestApplyNativeFixReportsInvalidFrontMatterYAMLClearly(t *testing.T) {
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
  - Example Author ()
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Final
type: Meta
created: 2026-03-26
sponsor: Foundation
implementation-required: true
implementation-url: https://github.com/example/arc1
implementation-maintainer:
  - @example
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

Text`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := applyNativeFix(path)
	if err == nil {
		t.Fatal("expected applyNativeFix() to reject invalid front matter YAML")
	}
	if !strings.Contains(err.Error(), "pre-commit YAML hooks do not inspect ARC Markdown front matter") {
		t.Fatalf("expected fmt error to explain ARC/front-matter hook boundary, got: %v", err)
	}
}

func TestApplyAdoptionFixReordersCanonicalAdoptionKeys(t *testing.T) {
	root := t.TempDir()
	adoptionDir := filepath.Join(root, "adoption")
	if err := os.MkdirAll(adoptionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(adoptionDir, "arc-0001.yaml")
	content := `summary:
  notes: ""
  blockers: []
  adoption-readiness: low
adoption:
  tooling: []
  wallets: []
  dapps-protocols: []
  explorers: []
  infra: []
reference-implementation:
  notes: ""
  status: shipped
title: Example ARC
arc: 1
last-reviewed: 2026-04-09
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := applyAdoptionFix(path); err != nil {
		t.Fatalf("applyAdoptionFix() error = %v", err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	if !strings.HasPrefix(text, "arc: 1\ntitle: Example ARC\nlast-reviewed: 2026-04-09\nreference-implementation:\n  status: shipped\n  notes: \"\"\nadoption:\n  wallets: []\n  explorers: []\n  tooling: []\n  infra: []\n  dapps-protocols: []\nsummary:\n  adoption-readiness: low\n  blockers: []\n  notes: \"\"\n") {
		t.Fatalf("expected canonical adoption key ordering, got:\n%s", text)
	}
}
