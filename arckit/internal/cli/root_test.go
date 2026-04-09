package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLICommands(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos")
	validDraftRoot := filepath.Join(root, "valid-draft")
	validDraftARC := filepath.Join(validDraftRoot, "ARCs", "arc-0042.md")
	validDraftAdoption := filepath.Join(validDraftRoot, "adoption", "arc-0042.yaml")
	transitionARC := filepath.Join(root, "transition-final", "ARCs", "arc-0044.md")

	tempRoot := t.TempDir()
	fmtPath := filepath.Join(tempRoot, "ARCs", "arc-0007.md")
	if err := os.MkdirAll(filepath.Dir(fmtPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(fmtPath, []byte(`---
title: Example
arc: 7
description: Example
author: Example
discussions-to: https://example.com
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

Text
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	tests := []struct {
		name      string
		args      []string
		wantCode  int
		wantInOut string
	}{
		{name: "rules", args: []string{"rules"}, wantCode: 0, wantInOut: "R:001"},
		{name: "explain", args: []string{"explain", "R:001"}, wantCode: 0, wantInOut: "Missing front matter"},
		{name: "validate-arc", args: []string{"validate", "arc", validDraftARC}, wantCode: 0, wantInOut: "summary:"},
		{name: "validate-adoption", args: []string{"validate", "adoption", validDraftAdoption}, wantCode: 0, wantInOut: "summary:"},
		{name: "validate-links", args: []string{"validate", "links", transitionARC}, wantCode: 0, wantInOut: "summary:"},
		{name: "validate-repo", args: []string{"validate", "repo", validDraftRoot}, wantCode: 0, wantInOut: "summary:"},
		{name: "validate-transition", args: []string{"validate", "transition", transitionARC, "--to", "Final"}, wantCode: 0, wantInOut: "R:020"},
		{name: "fmt", args: []string{"fmt", fmtPath}, wantCode: 0, wantInOut: "summary:"},
		{name: "init", args: []string{"init", "arc", "--root", tempRoot, "--number", "77", "--title", "CLI Init ARC", "--type", "Standards Track", "--category", "Interface", "--sub-category", "Application", "--sponsor", "Foundation"}, wantCode: 0, wantInOut: "arc-0077.md"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			exitCode := ExecuteArgs(test.args, stdout, stderr)
			if exitCode != test.wantCode {
				t.Fatalf("ExecuteArgs() exit code = %d, want %d, stdout=%s stderr=%s", exitCode, test.wantCode, stdout.String(), stderr.String())
			}
			output := stdout.String() + stderr.String()
			if !strings.Contains(output, test.wantInOut) {
				t.Fatalf("expected output to contain %q, got %q", test.wantInOut, output)
			}
		})
	}
}

func TestCLIValidateCommandsHonorConfig(t *testing.T) {
	t.Run("validate-arc", func(t *testing.T) {
		root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "missing-adoption"))
		writeConfig(t, root, `{
  "ignoreArcs": [43]
}`)
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		path := filepath.Join(root, "ARCs", "arc-0043.md")
		exitCode := ExecuteArgs([]string{"validate", "arc", path}, stdout, stderr)
		assertCommandSucceeded(t, exitCode, stdout, stderr)
		assertContains(t, stdout.String(), "summary: 0 error(s), 0 warning(s), 0 info")
	})

	t.Run("validate-adoption", func(t *testing.T) {
		root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
		writeConfig(t, root, `{
  "ignoreArcs": [42]
}`)
		path := filepath.Join(root, "adoption", "arc-0042.yaml")
		if err := os.WriteFile(path, []byte(": invalid"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		exitCode := ExecuteArgs([]string{"validate", "adoption", path}, stdout, stderr)
		assertCommandSucceeded(t, exitCode, stdout, stderr)
		assertContains(t, stdout.String(), "summary: 0 error(s), 0 warning(s), 0 info")
	})

	t.Run("validate-links", func(t *testing.T) {
		root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
		writeConfig(t, root, `{
  "ignoreArcs": [42]
}`)
		path := filepath.Join(root, "ARCs", "arc-0042.md")
		if err := os.WriteFile(path, []byte("not an arc"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		exitCode := ExecuteArgs([]string{"validate", "links", path}, stdout, stderr)
		assertCommandSucceeded(t, exitCode, stdout, stderr)
		assertContains(t, stdout.String(), "summary: 0 error(s), 0 warning(s), 0 info")
	})

	t.Run("validate-repo", func(t *testing.T) {
		root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "missing-adoption"))
		writeConfig(t, root, `{
  "ignoreByArc": {
    "40-45": ["R:012", "R:013"]
  }
}`)
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		exitCode := ExecuteArgs([]string{"validate", "repo", root}, stdout, stderr)
		assertCommandSucceeded(t, exitCode, stdout, stderr)
		assertContains(t, stdout.String(), "summary: 0 error(s), 0 warning(s), 0 info")
	})

	t.Run("validate-transition", func(t *testing.T) {
		root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
		writeConfig(t, root, `{
  "ignoreArcs": [44]
}`)
		path := filepath.Join(root, "ARCs", "arc-0044.md")
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		exitCode := ExecuteArgs([]string{"validate", "transition", path, "--to", "Final"}, stdout, stderr)
		assertCommandSucceeded(t, exitCode, stdout, stderr)
		assertContains(t, stdout.String(), "summary: 0 error(s), 0 warning(s), 0 info")
	})
}

func TestCLIValidateAdoptionRequiresVettedAdoptersRegistry(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	if err := os.Remove(filepath.Join(root, "adoption", "vetted-adopters.yaml")); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	path := filepath.Join(root, "adoption", "arc-0042.yaml")
	exitCode := ExecuteArgs([]string{"validate", "adoption", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "R:022")
}

func TestCLIValidateAdoptionRejectsUnvettedAdopter(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	writeVettedAdopters(t, root, `wallets: []
explorers: []
sdk-libraries: []
infra: []
dapps-protocols: []
`)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	path := filepath.Join(root, "adoption", "arc-0044.yaml")
	exitCode := ExecuteArgs([]string{"validate", "adoption", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "R:023")
}

func TestCLIValidateAdoptionRejectsFinalARCWithoutTrackedAdoption(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	path := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(path, []byte(`arc: 42
title: Example ARC
status: Final
last-reviewed: 2026-04-09
sponsor: Foundation
implementation-required: false
adoption:
  wallets: []
  explorers: []
  sdk-libraries: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "adoption", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "R:025")
}

func TestCLIValidateAdoptionReportsHelpfulActorSchemaError(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	path := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(path, []byte(`arc: 42
title: Example ARC
status: Final
last-reviewed: 2026-04-09
sponsor: Foundation
implementation-required: false
adoption:
  wallets: []
  explorers:
    - wow
  sdk-libraries: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "adoption", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "adoption.explorers[0] must be an actor object")
	if strings.Contains(stdout.String(), "cannot unmarshal") {
		t.Fatalf("expected targeted schema error, got stdout=%s", stdout.String())
	}
}

func TestCLIValidateRepoRejectsFinalARCWithoutTrackedAdoption(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0042.md"), []byte(`---
arc: 42
title: Example ARC
description: Example ARC for testing.
author:
  - Example Author (@example)
discussions-to: https://example.com/discussion
status: Final
type: Standards Track
created: 2026-03-26
sponsor: Foundation
implementation-required: false
adoption-summary: adoption/arc-0042.yaml
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
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "adoption", "arc-0042.yaml"), []byte(`arc: 42
title: Example ARC
status: Final
last-reviewed: 2026-04-09
sponsor: Foundation
implementation-required: false
adoption:
  wallets: []
  explorers: []
  sdk-libraries: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "repo", root}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "R:025")
}

func TestCLIValidateCommandsRejectInvalidConfig(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	writeConfig(t, root, `{
  "ignoreRules": ["R:999"]
}`)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	path := filepath.Join(root, "ARCs", "arc-0042.md")
	exitCode := ExecuteArgs([]string{"validate", "arc", path}, stdout, stderr)
	if exitCode != 2 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 2, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "R:028")
}

func TestCLIValidateArcIgnoreConfigBypassesRepoSuppressions(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ARCs"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "adoption"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeConfig(t, root, `{
  "ignoreRules": ["R:021"]
}`)

	path := filepath.Join(root, "ARCs", "arc-0001.md")
	content := `---
arc: 1
title: Example
description: Example description
author: Example Author, Other Author
discussions-to: https://example.com/discussion
status: Draft
type: Standards Track
created: 2026-04-09
sponsor: Foundation
implementation-required: false
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

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "arc", path}, stdout, stderr)
	assertCommandSucceeded(t, exitCode, stdout, stderr)
	assertContains(t, stdout.String(), "summary: 0 error(s), 0 warning(s), 0 info")

	stdout.Reset()
	stderr.Reset()
	exitCode = ExecuteArgs([]string{"validate", "--ignore-config", "arc", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "R:021")
}

func TestCLIValidateArcEnforceRuleBypassesOnlyMatchingSuppressions(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ARCs"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeConfig(t, root, `{
  "ignoreRules": ["R:021"],
  "ignoreByArc": {
    "0": ["R:008"]
  }
}`)

	path := filepath.Join(root, "ARCs", "arc-0000.md")
	content := `---
arc: 0
title: Example
description: Example description
author: Example Author, Other Author
discussions-to: https://example.com/discussion
status: Draft
type: Meta
created: 2026-04-09
sponsor: Foundation
implementation-required: false
---

## Abstract

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

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "arc", path}, stdout, stderr)
	assertCommandSucceeded(t, exitCode, stdout, stderr)
	assertContains(t, stdout.String(), "summary: 0 error(s), 0 warning(s), 0 info")

	stdout.Reset()
	stderr.Reset()
	exitCode = ExecuteArgs([]string{"validate", "--enforce-rule", "R:021", "arc", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "R:021")
	if strings.Contains(stdout.String(), "R:008") {
		t.Fatalf("expected --enforce-rule to keep unrelated suppressions, got stdout=%s", stdout.String())
	}
}

func TestCLIWriterFailureReturnsInvocationError(t *testing.T) {
	stdout := failingWriter{}
	stderr := &bytes.Buffer{}

	exitCode := ExecuteArgs([]string{"rules"}, stdout, stderr)
	if exitCode != 2 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 2, stderr=%s", exitCode, stderr.String())
	}
	assertContains(t, stderr.String(), "R:026")
	assertContains(t, stderr.String(), "write failed")
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func copyRepoFixture(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, relative)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, info.Mode())
	}); err != nil {
		t.Fatalf("copyRepoFixture() error = %v", err)
	}
	return dst
}

func writeConfig(t *testing.T, root string, content string) {
	t.Helper()
	path := filepath.Join(root, ".arckit.jsonc")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func writeVettedAdopters(t *testing.T, root string, content string) {
	t.Helper()
	path := filepath.Join(root, "adoption", "vetted-adopters.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func assertCommandSucceeded(t *testing.T, exitCode int, stdout *bytes.Buffer, stderr *bytes.Buffer) {
	t.Helper()
	if exitCode != 0 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 0, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
}

func assertContains(t *testing.T, output string, want string) {
	t.Helper()
	if !strings.Contains(output, want) {
		t.Fatalf("expected output to contain %q, got %q", want, output)
	}
}
