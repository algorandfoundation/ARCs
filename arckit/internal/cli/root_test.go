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
		{name: "validate-links", args: []string{"validate", "links", validDraftARC}, wantCode: 0, wantInOut: "summary:"},
		{name: "validate-repo", args: []string{"validate", "repo", validDraftRoot}, wantCode: 0, wantInOut: "summary:"},
		{name: "validate-transition", args: []string{"validate", "transition", transitionARC, "--to", "Final"}, wantCode: 0, wantInOut: "R:020"},
		{name: "fmt", args: []string{"fmt", fmtPath}, wantCode: 0, wantInOut: "summary:"},
		{name: "init", args: []string{"init", "arc", "--root", tempRoot, "--number", "77", "--title", "CLI Init ARC", "--type", "Meta", "--sponsor", "Foundation"}, wantCode: 0, wantInOut: "arc-0077.md"},
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
