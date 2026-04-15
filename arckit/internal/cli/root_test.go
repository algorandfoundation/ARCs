package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/testutil"
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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

func TestCLIInitARCRejectsInvalidCategoryMetadata(t *testing.T) {
	tempRoot := t.TempDir()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	arcPath := filepath.Join(tempRoot, "ARCs", "arc-0077.md")

	exitCode := ExecuteArgs([]string{
		"--format", "json",
		"init", "arc",
		"--root", tempRoot,
		"--number", "77",
		"--title", "CLI Init ARC",
		"--type", "Standards Track",
		"--sub-category", "General",
		"--sponsor", "Foundation",
	}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(arcPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected %s not to be created, stat err=%v", arcPath, err)
	}
	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "\"rule_id\": \"R:030\"") || !strings.Contains(output, "sub-category requires category") {
		t.Fatalf("expected R:030 output, got %q", output)
	}
}

func TestCLISummaryRepoWritesDefaultOutput(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "missing-adoption"))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"summary", "repo", root}, stdout, stderr)
	assertCommandSucceeded(t, exitCode, stdout, stderr)

	outputPath := filepath.Join(root, "arc-summary.md")
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected %s to exist, stat err=%v", outputPath, err)
	}
	assertContains(t, stdout.String(), outputPath)

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", outputPath, err)
	}
	assertContains(t, string(content), "# ARC State Summary")
	assertContains(t, string(content), "## Validation Snapshot")
	assertContains(t, string(content), "## Adoption Watch")
}

func TestCLISummaryRepoWritesCustomOutput(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	relativeOut := filepath.Join("editor", "state.md")
	exitCode := ExecuteArgs([]string{"summary", "repo", "--out", relativeOut, root}, stdout, stderr)
	assertCommandSucceeded(t, exitCode, stdout, stderr)

	outputPath := filepath.Join(root, relativeOut)
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected %s to exist, stat err=%v", outputPath, err)
	}
	assertContains(t, stdout.String(), outputPath)

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", outputPath, err)
	}
	assertContains(t, string(content), "## Relationship Watch")
}

func TestCLISummaryRepoRejectsJSONOutput(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"--format", "json", "summary", "repo", root}, stdout, stderr)
	if exitCode != 2 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 2, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), `"rule_id": "R:026"`)
	assertContains(t, stdout.String(), "summary repo only supports text output in v1")
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
tooling: []
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
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0042.md"), []byte(`---
arc: 42
title: Example ARC
description: Example ARC for testing.
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	path := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(path, []byte(`arc: 42
title: Example ARC
last-reviewed: 2026-04-09
adoption:
  wallets: []
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

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "adoption", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "R:025")
}

func TestCLIValidateAdoptionRejectsLegacyReferenceImplementationIdentityFields(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	path := filepath.Join(root, "adoption", "arc-0044.yaml")
	if err := os.WriteFile(path, []byte(`arc: 44
title: Transition Ready ARC
last-reviewed: 2026-03-26
reference-implementation:
  repository: https://github.com/example/arc-0044
  maintainers:
    - "@maintainer"
  status: shipped
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

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "adoption", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "reference-implementation.repository is not allowed")
}

func TestCLIValidateAdoptionAllowsReferenceImplementationWithoutNotes(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "transition-final"))
	path := filepath.Join(root, "adoption", "arc-0044.yaml")
	if err := os.WriteFile(path, []byte(`arc: 44
title: Transition Ready ARC
last-reviewed: 2026-03-26
reference-implementation:
  status: shipped
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

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "adoption", path}, stdout, stderr)
	assertCommandSucceeded(t, exitCode, stdout, stderr)
	assertContains(t, stdout.String(), "summary: 0 error(s), 0 warning(s), 0 info")
}

func TestCLIValidateAdoptionRejectsAdoptionReadinessWithoutEnoughAdopters(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	writeVettedAdopters(t, root, `wallets:
  - example-wallet
explorers:
  - example-explorer
tooling: []
infra: []
dapps-protocols: []
`)
	path := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(path, []byte(`arc: 42
title: Example ARC
last-reviewed: 2026-04-09
adoption:
  wallets:
    - name: example-wallet
      status: shipped
      evidence: https://example.com/wallet-proof
      notes: ""
  explorers:
    - name: example-explorer
      status: shipped
      evidence: https://example.com/explorer-proof
      notes: ""
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: medium
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
	assertContains(t, stdout.String(), `summary.adoption-readiness "medium" requires at least 3 adopters`)
}

func TestCLIFmtCanNormalizeAdoptionReadiness(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "adoption"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(path, []byte(`arc: 42
title: Example ARC
last-reviewed: 2026-04-09
adoption:
  wallets:
    - name: wallet-one
      status: shipped
      evidence: https://example.com/wallet-one
      notes: ""
    - name: wallet-two
      status: shipped
      evidence: https://example.com/wallet-two
      notes: ""
  explorers:
    - name: explorer-one
      status: shipped
      evidence: https://example.com/explorer-one
      notes: ""
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

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"fmt", path}, stdout, stderr)
	assertCommandSucceeded(t, exitCode, stdout, stderr)

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	assertContains(t, string(updated), "adoption-readiness: medium")
}

func TestCLIValidateAdoptionReportsHelpfulActorSchemaError(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	path := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(path, []byte(`arc: 42
title: Example ARC
last-reviewed: 2026-04-09
adoption:
  wallets: []
  explorers:
    - wow
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

func TestCLIValidateAdoptionRejectsLegacyARCMetadataFields(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	path := filepath.Join(root, "adoption", "arc-0042.yaml")
	if err := os.WriteFile(path, []byte(`arc: 42
title: Example ARC
status: Draft
last-reviewed: 2026-04-09
sponsor: Foundation
implementation-required: false
adoption:
  wallets: []
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

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "adoption", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "status is not allowed in adoption summaries")
	assertContains(t, stdout.String(), "sponsor is not allowed in adoption summaries")
	assertContains(t, stdout.String(), "implementation-required is not allowed in adoption summaries")
}

func TestCLIValidateArcRequiresImplementationDeclarationForReviewAndLater(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ARCs"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(root, "ARCs", "arc-0001.md")
	if err := os.WriteFile(path, []byte(`---
arc: 1
title: Example
description: Example description
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Review
type: Standards Track
created: 2026-04-09
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
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "arc", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "requires implementation-url")
	assertContains(t, stdout.String(), "requires implementation-maintainer")
}

func TestCLIValidateArcRejectsNonCanonicalImplementationURL(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ARCs"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(root, "ARCs", "arc-0044.md")
	if err := os.WriteFile(path, []byte(`---
arc: 44
title: Example
description: Example description
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
status: Review
type: Standards Track
created: 2026-04-09
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
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := ExecuteArgs([]string{"validate", "arc", path}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("ExecuteArgs() exit code = %d, want 1, stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	assertContains(t, stdout.String(), "R:029")
	assertContains(t, stdout.String(), "https://github.com/algorandfoundation/arc44")
}

func TestCLIValidateRepoRejectsFinalARCWithoutTrackedAdoption(t *testing.T) {
	root := copyRepoFixture(t, filepath.Join("..", "..", "testdata", "repos", "valid-draft"))
	if err := os.WriteFile(filepath.Join(root, "ARCs", "arc-0042.md"), []byte(`---
arc: 42
title: Example ARC
description: Example ARC for testing.
author:
  - Example Author (@example)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "adoption", "arc-0042.yaml"), []byte(`arc: 42
title: Example ARC
last-reviewed: 2026-04-09
adoption:
  wallets: []
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
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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
discussions-to: https://github.com/algorandfoundation/ARCs/issues/1
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

## Copyright

Copyright and related rights waived via CC0 1.0.
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

var copyRepoFixture = testutil.CopyDir

func writeConfig(t *testing.T, root string, content string) {
	t.Helper()
	testutil.WriteTrimmedFile(t, filepath.Join(root, ".arckit.jsonc"), content)
}

func writeVettedAdopters(t *testing.T, root string, content string) {
	t.Helper()
	testutil.WriteTrimmedFile(t, filepath.Join(root, "adoption", "vetted-adopters.yaml"), content)
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
