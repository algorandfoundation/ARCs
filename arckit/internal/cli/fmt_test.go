package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
author: Example Author
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
	if !strings.Contains(text, "## Abstract    \n\nText    \n") {
		t.Fatalf("expected body whitespace to be preserved, got:\n%s", text)
	}
	if !strings.HasSuffix(text, "Text") {
		t.Fatalf("expected final newline policy to remain unchanged, got:\n%s", text)
	}
}
