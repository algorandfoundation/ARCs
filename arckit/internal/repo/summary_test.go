package repo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/algorandfoundation/ARCs/arckit/internal/config"
)

var fixedSummaryNow = time.Date(2026, 4, 10, 9, 30, 0, 0, time.UTC)

func TestBuildSummaryAggregatesEditorState(t *testing.T) {
	root := writeSummaryFixtureRepo(t)

	state, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	summary := BuildSummary(state, diagnostics, fixedSummaryNow)

	if summary.ValidationSummary.Errors != 1 || summary.ValidationSummary.Warnings != 0 || summary.ValidationSummary.Info != 0 {
		t.Fatalf("ValidationSummary = %+v, want 1 error and no warnings/info", summary.ValidationSummary)
	}
	if summary.TotalARCs != 6 {
		t.Fatalf("TotalARCs = %d, want 6", summary.TotalARCs)
	}
	if summary.ARCsWithAssets != 1 {
		t.Fatalf("ARCsWithAssets = %d, want 1", summary.ARCsWithAssets)
	}
	if summary.AdoptionSummaryCoverage != 5 {
		t.Fatalf("AdoptionSummaryCoverage = %d, want 5", summary.AdoptionSummaryCoverage)
	}
	if summary.TotalAdoptionFiles != 5 {
		t.Fatalf("TotalAdoptionFiles = %d, want 5", summary.TotalAdoptionFiles)
	}
	if summary.TotalAdopterEntries != 5 {
		t.Fatalf("TotalAdopterEntries = %d, want 5", summary.TotalAdopterEntries)
	}
	if summary.ARCsWithTrackedAdopters != 4 {
		t.Fatalf("ARCsWithTrackedAdopters = %d, want 4", summary.ARCsWithTrackedAdopters)
	}

	assertSummaryCount(t, summary.StatusCounts, "Draft", 1)
	assertSummaryCount(t, summary.StatusCounts, "Final", 3)
	assertSummaryCount(t, summary.StatusCounts, "Idle", 1)
	assertSummaryCount(t, summary.StatusCounts, "Last Call", 1)
	assertSummaryCount(t, summary.TypeCounts, "Meta", 1)
	assertSummaryCount(t, summary.TypeCounts, "Standards Track", 5)
	assertSummaryCount(t, summary.AdoptionReadinessCounts, "low", 5)
	assertSummaryCount(t, summary.ReferenceImplementationStatusCounts, "shipped", 1)
	assertSummaryCount(t, summary.ReferenceImplementationStatusCounts, "wip", 1)
	assertSummaryCount(t, summary.AdopterEntriesByCategory, "wallets", 2)
	assertSummaryCount(t, summary.AdopterEntriesByCategory, "explorers", 1)
	assertSummaryCount(t, summary.AdopterEntriesByCategory, "tooling", 2)
	assertSummaryCount(t, summary.AdopterEntriesByStatus, "shipped", 4)
	assertSummaryCount(t, summary.AdopterEntriesByStatus, "in_progress", 1)

	if len(summary.OverdueLastCall) != 1 {
		t.Fatalf("len(OverdueLastCall) = %d, want 1", len(summary.OverdueLastCall))
	}
	if row := summary.OverdueLastCall[0]; row.ARC != 44 || row.Days != 9 {
		t.Fatalf("OverdueLastCall[0] = %+v, want ARC 44 overdue by 9 days", row)
	}
	if len(summary.UpcomingLastCall) != 0 {
		t.Fatalf("UpcomingLastCall = %+v, want empty", summary.UpcomingLastCall)
	}
	if len(summary.IdleARCs) != 1 {
		t.Fatalf("len(IdleARCs) = %d, want 1", len(summary.IdleARCs))
	}
	if row := summary.IdleARCs[0]; row.ARC != 47 || row.Days != 68 {
		t.Fatalf("IdleARCs[0] = %+v, want ARC 47 idle for 68 days", row)
	}
	if len(summary.ImplementationRequiredNotShipped) != 1 || summary.ImplementationRequiredNotShipped[0].ARC != 47 {
		t.Fatalf("ImplementationRequiredNotShipped = %+v, want only ARC 47", summary.ImplementationRequiredNotShipped)
	}

	if len(summary.FinalZeroAdopters) != 1 || summary.FinalZeroAdopters[0].ARC != 45 {
		t.Fatalf("FinalZeroAdopters = %+v, want only ARC 45", summary.FinalZeroAdopters)
	}
	if len(summary.FinalLowAdopters) != 2 {
		t.Fatalf("len(FinalLowAdopters) = %d, want 2", len(summary.FinalLowAdopters))
	}
	if summary.FinalLowAdopters[0].ARC != 48 || summary.FinalLowAdopters[1].ARC != 46 {
		t.Fatalf("FinalLowAdopters = %+v, want ARC 48 before ARC 46", summary.FinalLowAdopters)
	}

	if len(summary.StaleAdoptionReviews) != 3 {
		t.Fatalf("len(StaleAdoptionReviews) = %d, want 3", len(summary.StaleAdoptionReviews))
	}
	if summary.StaleAdoptionReviews[0].ARC != 44 || summary.StaleAdoptionReviews[1].ARC != 45 || summary.StaleAdoptionReviews[2].ARC != 47 {
		t.Fatalf("StaleAdoptionReviews = %+v, want ARCs 44, 45, 47 in order", summary.StaleAdoptionReviews)
	}

	if len(summary.TopAdoptersByCoverage) < 2 {
		t.Fatalf("TopAdoptersByCoverage = %+v, want at least 2 rows", summary.TopAdoptersByCoverage)
	}
	if summary.TopAdoptersByCoverage[0].Name != "tool-one" || summary.TopAdoptersByCoverage[0].Count != 2 {
		t.Fatalf("TopAdoptersByCoverage[0] = %+v, want tool-one covering 2 ARCs", summary.TopAdoptersByCoverage[0])
	}
	if summary.TopAdoptersByCoverage[1].Name != "wallet-one" || summary.TopAdoptersByCoverage[1].Count != 2 {
		t.Fatalf("TopAdoptersByCoverage[1] = %+v, want wallet-one covering 2 ARCs", summary.TopAdoptersByCoverage[1])
	}

	if len(summary.TopARCsByAdopters) < 4 {
		t.Fatalf("TopARCsByAdopters = %+v, want at least 4 rows", summary.TopARCsByAdopters)
	}
	if summary.TopARCsByAdopters[0].ARC != 46 || summary.TopARCsByAdopters[0].Count != 2 {
		t.Fatalf("TopARCsByAdopters[0] = %+v, want ARC 46 with 2 adopters", summary.TopARCsByAdopters[0])
	}
	if len(summary.TopRequiresTargets) != 2 {
		t.Fatalf("TopRequiresTargets = %+v, want 2 rows", summary.TopRequiresTargets)
	}
	if summary.TopRequiresTargets[0].ARC != 44 || summary.TopRequiresTargets[0].Count != 2 {
		t.Fatalf("TopRequiresTargets[0] = %+v, want ARC 44 referenced twice", summary.TopRequiresTargets[0])
	}
	if len(summary.TopExtendsTargets) != 1 || summary.TopExtendsTargets[0].ARC != 46 || summary.TopExtendsTargets[0].Count != 2 {
		t.Fatalf("TopExtendsTargets = %+v, want ARC 46 referenced twice", summary.TopExtendsTargets)
	}
	if len(summary.SupersessionRows) != 2 {
		t.Fatalf("SupersessionRows = %+v, want 2 rows", summary.SupersessionRows)
	}
}

func TestBuildSummaryMarkdownGolden(t *testing.T) {
	root := writeSummaryFixtureRepo(t)

	state, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	summary := BuildSummary(state, diagnostics, fixedSummaryNow)
	summary.Root = "/repo"

	got := summary.Markdown()
	wantPath := filepath.Join("testdata", "summary.golden.md")
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", wantPath, err)
	}
	if got != string(want) {
		t.Fatalf("Markdown() mismatch\n--- got ---\n%s\n--- want ---\n%s", got, string(want))
	}
}

func TestBuildSummaryUsesValidatedStateWithIgnoredARCs(t *testing.T) {
	root := writeSummaryFixtureRepo(t)
	writeConfig(t, root, `{
  "ignoreArcs": [47]
}`)

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	state, diagnostics, err := Validate(root, cfg)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	summary := BuildSummary(state, diagnostics, fixedSummaryNow)
	if summary.TotalARCs != 5 {
		t.Fatalf("TotalARCs = %d, want 5 after ignoring ARC 47", summary.TotalARCs)
	}
	if len(summary.IdleARCs) != 0 {
		t.Fatalf("IdleARCs = %+v, want empty after ignoring ARC 47", summary.IdleARCs)
	}
	if len(summary.ImplementationRequiredNotShipped) != 0 {
		t.Fatalf("ImplementationRequiredNotShipped = %+v, want empty after ignoring ARC 47", summary.ImplementationRequiredNotShipped)
	}
	if summary.TotalAdoptionFiles != 4 {
		t.Fatalf("TotalAdoptionFiles = %d, want 4 after ignoring ARC 47", summary.TotalAdoptionFiles)
	}
}

func TestBuildSummaryNormalizesInvalidActorStatuses(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"ARCs", "adoption"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
	}

	writeVettedAdopters(t, root, `wallets:
  - wallet-one
explorers:
  - explorer-one
tooling: []
infra: []
dapps-protocols: []
`)
	writeSummaryARC(t, root, 42, summaryARCOptions{
		Title:           "Invalid Status ARC",
		Status:          "Final",
		Type:            "Standards Track",
		AdoptionSummary: true,
	})
	writeSummaryAdoption(t, root, 42, `arc: 42
title: Invalid Status ARC
last-reviewed: 2026-04-05
adoption:
  wallets:
    - name: wallet-one
      evidence: https://example.com/wallet-one
      notes: ""
  explorers:
    - name: explorer-one
      status: maybe
      evidence: https://example.com/explorer-one
      notes: ""
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`)

	state, diagnostics, err := Validate(root, config.Config{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	summary := BuildSummary(state, diagnostics, fixedSummaryNow)

	assertSummaryCount(t, summary.AdopterEntriesByStatus, "missing", 1)
	assertSummaryCount(t, summary.AdopterEntriesByStatus, "unknown", 1)
	for _, row := range summary.AdopterEntriesByStatus {
		if row.Label == "" {
			t.Fatalf("AdopterEntriesByStatus contains blank label row: %+v", summary.AdopterEntriesByStatus)
		}
	}
}

func TestWriteTableEscapesMarkdownCells(t *testing.T) {
	var out bytes.Buffer

	writeTable(&out, []string{"Title", "Count"}, func(writeRow func(...string)) {
		writeRow("Name | Alias\nSecond Line", "1")
	})

	got := out.String()
	if !strings.Contains(got, `Name \| Alias<br>Second Line`) {
		t.Fatalf("writeTable() output = %q, want escaped markdown cell", got)
	}
}

func assertSummaryCount(t *testing.T, rows []SummaryCount, label string, want int) {
	t.Helper()
	for _, row := range rows {
		if row.Label == label {
			if row.Count != want {
				t.Fatalf("count for %q = %d, want %d", label, row.Count, want)
			}
			return
		}
	}
	t.Fatalf("count for %q not found in %+v", label, rows)
}

func writeSummaryFixtureRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	for _, dir := range []string{"ARCs", "adoption", "assets"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
	}

	writeVettedAdopters(t, root, `wallets:
  - wallet-one
explorers:
  - explorer-one
tooling:
  - tool-one
infra: []
dapps-protocols: []
`)

	writeSummaryARC(t, root, 44, summaryARCOptions{
		Title:                    "Last Call ARC",
		Status:                   "Last Call",
		Type:                     "Standards Track",
		ImplementationRequired:   true,
		AdoptionSummary:          true,
		LastCallDeadline:         "2026-04-01",
		ImplementationMaintainer: "@maintainer",
	})
	writeSummaryAdoption(t, root, 44, `arc: 44
title: Last Call ARC
last-reviewed: 2026-03-01
reference-implementation:
  status: shipped
  notes: ""
adoption:
  wallets:
    - name: wallet-one
      status: shipped
      evidence: https://example.com/wallet-one
      notes: ""
  explorers: []
  tooling: []
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`)

	writeSummaryARC(t, root, 45, summaryARCOptions{
		Title:           "Final ARC With Placeholder Adoption",
		Status:          "Final",
		Type:            "Standards Track",
		AdoptionSummary: true,
		SupersededBy:    46,
	})
	writeSummaryAdoption(t, root, 45, `arc: 45
title: Final ARC With Placeholder Adoption
last-reviewed: 2026-03-01
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
`)

	writeSummaryARC(t, root, 46, summaryARCOptions{
		Title:           "Final ARC With Two Adopters",
		Status:          "Final",
		Type:            "Standards Track",
		AdoptionSummary: true,
		Requires:        []int{44},
		ExtendedBy:      []int{47, 49},
		Supersedes:      []int{45},
		HasAssets:       true,
	})
	writeSummaryAdoption(t, root, 46, `arc: 46
title: Final ARC With Two Adopters
last-reviewed: 2026-04-05
adoption:
  wallets:
    - name: wallet-one
      status: shipped
      evidence: https://example.com/wallet-one-final
      notes: ""
  explorers: []
  tooling:
    - name: tool-one
      status: in_progress
      evidence: https://example.com/tool-one-final
      notes: ""
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`)

	writeSummaryARC(t, root, 47, summaryARCOptions{
		Title:                    "Idle ARC With WIP Implementation",
		Status:                   "Idle",
		Type:                     "Standards Track",
		ImplementationRequired:   true,
		AdoptionSummary:          true,
		IdleSince:                "2026-02-01",
		Requires:                 []int{44},
		Extends:                  []int{46},
		ImplementationMaintainer: "@maintainer",
	})
	writeSummaryAdoption(t, root, 47, `arc: 47
title: Idle ARC With WIP Implementation
last-reviewed: 2026-03-05
reference-implementation:
  status: wip
  notes: ""
adoption:
  wallets: []
  explorers: []
  tooling:
    - name: tool-one
      status: shipped
      evidence: https://example.com/tool-one-idle
      notes: ""
  infra: []
  dapps-protocols: []
summary:
  adoption-readiness: low
  blockers: []
  notes: ""
`)

	writeSummaryARC(t, root, 48, summaryARCOptions{
		Title:           "Final ARC With One Adopter",
		Status:          "Final",
		Type:            "Standards Track",
		AdoptionSummary: true,
	})
	writeSummaryAdoption(t, root, 48, `arc: 48
title: Final ARC With One Adopter
last-reviewed: 2026-04-05
adoption:
  wallets: []
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
`)

	writeSummaryARC(t, root, 49, summaryARCOptions{
		Title:    "Draft Meta ARC",
		Status:   "Draft",
		Type:     "Meta",
		Requires: []int{46},
		Extends:  []int{46},
	})

	return root
}

type summaryARCOptions struct {
	Title                    string
	Status                   string
	Type                     string
	ImplementationRequired   bool
	ImplementationMaintainer string
	AdoptionSummary          bool
	LastCallDeadline         string
	IdleSince                string
	Requires                 []int
	Extends                  []int
	ExtendedBy               []int
	Supersedes               []int
	SupersededBy             int
	HasAssets                bool
}

func writeSummaryARC(t *testing.T, root string, number int, options summaryARCOptions) {
	t.Helper()

	var content strings.Builder
	content.WriteString("---\n")
	content.WriteString("arc: " + formatInt(number) + "\n")
	content.WriteString("title: " + options.Title + "\n")
	content.WriteString("description: Fixture ARC for summary tests.\n")
	content.WriteString("author:\n  - Example Author (@example)\n")
	content.WriteString("discussions-to: https://example.com/discussion\n")
	content.WriteString("status: " + options.Status + "\n")
	content.WriteString("type: " + options.Type + "\n")
	content.WriteString("created: 2026-01-01\n")
	content.WriteString("sponsor: Foundation\n")
	if options.ImplementationRequired {
		content.WriteString("implementation-required: true\n")
		content.WriteString("implementation-url: https://github.com/algorandfoundation/arc" + formatInt(number) + "\n")
		content.WriteString("implementation-maintainer:\n  - \"" + options.ImplementationMaintainer + "\"\n")
	} else {
		content.WriteString("implementation-required: false\n")
	}
	if options.AdoptionSummary {
		content.WriteString("adoption-summary: adoption/arc-" + formatPaddedInt(number) + ".yaml\n")
	}
	if options.LastCallDeadline != "" {
		content.WriteString("last-call-deadline: " + options.LastCallDeadline + "\n")
	}
	if options.IdleSince != "" {
		content.WriteString("idle-since: " + options.IdleSince + "\n")
	}
	if len(options.Requires) > 0 {
		content.WriteString("requires:\n")
		for _, value := range options.Requires {
			content.WriteString("  - " + formatInt(value) + "\n")
		}
	}
	if len(options.Supersedes) > 0 {
		content.WriteString("supersedes:\n")
		for _, value := range options.Supersedes {
			content.WriteString("  - " + formatInt(value) + "\n")
		}
	}
	if options.SupersededBy != 0 {
		content.WriteString("superseded-by: " + formatInt(options.SupersededBy) + "\n")
	}
	if len(options.Extends) > 0 {
		content.WriteString("extends:\n")
		for _, value := range options.Extends {
			content.WriteString("  - " + formatInt(value) + "\n")
		}
	}
	if len(options.ExtendedBy) > 0 {
		content.WriteString("extended-by:\n")
		for _, value := range options.ExtendedBy {
			content.WriteString("  - " + formatInt(value) + "\n")
		}
	}
	content.WriteString("---\n\n")
	content.WriteString("## Abstract\n\nText\n\n")
	content.WriteString("## Motivation\n\nText\n\n")
	content.WriteString("## Specification\n\nText\n\n")
	content.WriteString("## Rationale\n\nText\n\n")
	content.WriteString("## Security Considerations\n\nText\n")

	path := filepath.Join(root, "ARCs", "arc-"+formatPaddedInt(number)+".md")
	if err := os.WriteFile(path, []byte(content.String()), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}

	if options.HasAssets {
		assetDir := filepath.Join(root, "assets", "arc-"+formatPaddedInt(number))
		if err := os.MkdirAll(assetDir, 0o755); err != nil {
			t.Fatalf("MkdirAll(%s) error = %v", assetDir, err)
		}
	}
}

func writeSummaryAdoption(t *testing.T, root string, number int, content string) {
	t.Helper()
	path := filepath.Join(root, "adoption", "arc-"+formatPaddedInt(number)+".yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func formatInt(value int) string {
	return fmt.Sprintf("%d", value)
}

func formatPaddedInt(value int) string {
	return fmt.Sprintf("%04d", value)
}
