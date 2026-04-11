package repo

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

func (summary RepoSummary) Markdown() string {
	var out bytes.Buffer

	writeLine := func(format string, args ...any) {
		_, _ = fmt.Fprintf(&out, format+"\n", args...)
	}
	writeBlank := func() {
		_, _ = out.WriteString("\n")
	}

	writeLine("# ARC State Summary")
	writeBlank()

	writeLine("## Validation Snapshot")
	writeBlank()
	writeLine("- Generated: `%s`", summary.GeneratedAt.Format(time.RFC3339))
	writeLine("- Repository: `%s`", summary.Root)
	writeLine("- Validation summary: `%d error(s), %d warning(s), %d info`", summary.ValidationSummary.Errors, summary.ValidationSummary.Warnings, summary.ValidationSummary.Info)
	writeBlank()

	writeLine("## State Overview")
	writeBlank()
	writeLine("- Total ARCs: `%d`", summary.TotalARCs)
	writeLine("- ARCs with assets: `%d`", summary.ARCsWithAssets)
	writeLine("- Adoption summary coverage: `%d/%d`", summary.AdoptionSummaryCoverage, summary.TotalARCs)
	writeLine("- Total adoption files: `%d`", summary.TotalAdoptionFiles)
	writeLine("- Total adopter entries: `%d`", summary.TotalAdopterEntries)
	writeLine("- ARCs with at least one adopter: `%d`", summary.ARCsWithTrackedAdopters)
	writeBlank()

	writeLine("### Counts by Status")
	writeBlank()
	writeCountTable(&out, "Status", summary.StatusCounts)
	writeBlank()
	writeLine("### Counts by Type")
	writeBlank()
	writeCountTable(&out, "Type", summary.TypeCounts)
	writeBlank()
	writeLine("### Adoption Readiness Distribution")
	writeBlank()
	writeCountTable(&out, "Adoption Readiness", summary.AdoptionReadinessCounts)
	writeBlank()
	writeLine("### Reference Implementation Status Counts")
	writeBlank()
	writeCountTable(&out, "Reference Implementation Status", summary.ReferenceImplementationStatusCounts)
	writeBlank()

	writeLine("## Transition Watch")
	writeBlank()
	writeLine("### Overdue Last Call")
	writeBlank()
	writeLastCallTable(&out, summary.OverdueLastCall, "Days Overdue")
	writeBlank()
	writeLine("### Upcoming Last Call (next %d days)", upcomingLastCallWindowDays)
	writeBlank()
	writeLastCallTable(&out, summary.UpcomingLastCall, "Days Remaining")
	writeBlank()
	writeLine("### Idle ARCs")
	writeBlank()
	writeIdleTable(&out, summary.IdleARCs)
	writeBlank()
	writeLine("### Implementation-Required ARCs Not Shipped")
	writeBlank()
	writeImplementationTable(&out, summary.ImplementationRequiredNotShipped)
	writeBlank()

	writeLine("## Adoption Watch")
	writeBlank()
	writeLine("### Final ARCs With Zero Adopters")
	writeBlank()
	writeAdoptionActionTable(&out, summary.FinalZeroAdopters)
	writeBlank()
	writeLine("### Final ARCs With 1-2 Adopters")
	writeBlank()
	writeAdoptionActionTable(&out, summary.FinalLowAdopters)
	writeBlank()
	writeLine("### Stale Adoption Reviews (>%d days)", staleAdoptionThresholdDays)
	writeBlank()
	writeStaleAdoptionTable(&out, summary.StaleAdoptionReviews)
	writeBlank()
	writeLine("### Adoption Totals")
	writeBlank()
	writeLine("#### Adopter Entries by Category")
	writeBlank()
	writeCountTable(&out, "Category", summary.AdopterEntriesByCategory)
	writeBlank()
	writeLine("#### Adopter Entries by Actor Status")
	writeBlank()
	writeCountTable(&out, "Actor Status", summary.AdopterEntriesByStatus)
	writeBlank()
	writeLine("#### Top Adopters by Distinct ARC Coverage")
	writeBlank()
	writeNamedCountTable(&out, "Adopter", summary.TopAdoptersByCoverage)
	writeBlank()
	writeLine("#### Top ARCs by Adopter Count")
	writeBlank()
	writeARCCountTable(&out, summary.TopARCsByAdopters)
	writeBlank()

	writeLine("## Relationship Watch")
	writeBlank()
	writeLine("### Top ARCs Most Referenced by `requires`")
	writeBlank()
	writeReferencedARCTable(&out, summary.TopRequiresTargets, "Requires References")
	writeBlank()
	writeLine("### Top ARCs Most Referenced by `extends`")
	writeBlank()
	writeReferencedARCTable(&out, summary.TopExtendsTargets, "Extends References")
	writeBlank()
	writeLine("### Non-Empty `supersedes` / `superseded-by` Pairs")
	writeBlank()
	writeSupersessionTable(&out, summary.SupersessionRows)
	writeBlank()

	writeLine("## Data Notes")
	writeBlank()
	writeLine("- This report is local and offline-only.")
	writeLine("- It uses ARC front matter, adoption YAML, vetted adopters, asset directories, and ARC relationship fields.")
	writeLine("- State age is only known where explicit dates exist: `last-call-deadline`, `idle-since`, and `last-reviewed`.")
	writeLine("- Missing `sponsor`, missing `implementation-required`, and sparse `updated` usage are migration-state and are not treated as backlog in this report.")

	return out.String()
}

func writeCountTable(out *bytes.Buffer, label string, rows []SummaryCount) {
	if len(rows) == 0 {
		_, _ = out.WriteString("None\n")
		return
	}
	writeTable(out, []string{label, "Count"}, func(writeRow func(...string)) {
		for _, row := range rows {
			writeRow(row.Label, fmt.Sprintf("%d", row.Count))
		}
	})
}

func writeLastCallTable(out *bytes.Buffer, rows []LastCallSummaryRow, daysLabel string) {
	if len(rows) == 0 {
		_, _ = out.WriteString("None\n")
		return
	}
	writeTable(out, []string{"ARC", "Title", "Deadline", daysLabel, "Editor Action"}, func(writeRow func(...string)) {
		for _, row := range rows {
			writeRow(fmt.Sprintf("%d", row.ARC), row.Title, row.Date, fmt.Sprintf("%d", row.Days), row.Action)
		}
	})
}

func writeIdleTable(out *bytes.Buffer, rows []IdleSummaryRow) {
	if len(rows) == 0 {
		_, _ = out.WriteString("None\n")
		return
	}
	writeTable(out, []string{"ARC", "Title", "Idle Since", "Days Idle", "Editor Action"}, func(writeRow func(...string)) {
		for _, row := range rows {
			writeRow(fmt.Sprintf("%d", row.ARC), row.Title, row.Date, fmt.Sprintf("%d", row.Days), row.Action)
		}
	})
}

func writeImplementationTable(out *bytes.Buffer, rows []ImplementationSummaryRow) {
	if len(rows) == 0 {
		_, _ = out.WriteString("None\n")
		return
	}
	writeTable(out, []string{"ARC", "Title", "Status", "Ref Impl Status", "Editor Action"}, func(writeRow func(...string)) {
		for _, row := range rows {
			writeRow(fmt.Sprintf("%d", row.ARC), row.Title, row.Status, row.ReferenceImplementation, row.Action)
		}
	})
}

func writeAdoptionActionTable(out *bytes.Buffer, rows []AdoptionSummaryRow) {
	if len(rows) == 0 {
		_, _ = out.WriteString("None\n")
		return
	}
	writeTable(out, []string{"ARC", "Title", "Adoption Readiness", "Last Reviewed", "Editor Action"}, func(writeRow func(...string)) {
		for _, row := range rows {
			writeRow(fmt.Sprintf("%d", row.ARC), row.Title, row.AdoptionReadiness, row.LastReviewed, row.Action)
		}
	})
}

func writeStaleAdoptionTable(out *bytes.Buffer, rows []StaleAdoptionSummaryRow) {
	if len(rows) == 0 {
		_, _ = out.WriteString("None\n")
		return
	}
	writeTable(out, []string{"ARC", "Title", "Status", "Last Reviewed", "Age (days)", "Editor Action"}, func(writeRow func(...string)) {
		for _, row := range rows {
			writeRow(fmt.Sprintf("%d", row.ARC), row.Title, row.Status, row.LastReviewed, fmt.Sprintf("%d", row.AgeDays), row.Action)
		}
	})
}

func writeNamedCountTable(out *bytes.Buffer, label string, rows []NamedCount) {
	if len(rows) == 0 {
		_, _ = out.WriteString("None\n")
		return
	}
	writeTable(out, []string{label, "Distinct ARCs"}, func(writeRow func(...string)) {
		for _, row := range rows {
			writeRow(row.Name, fmt.Sprintf("%d", row.Count))
		}
	})
}

func writeARCCountTable(out *bytes.Buffer, rows []ARCCount) {
	if len(rows) == 0 {
		_, _ = out.WriteString("None\n")
		return
	}
	writeTable(out, []string{"ARC", "Title", "Status", "Adopters"}, func(writeRow func(...string)) {
		for _, row := range rows {
			writeRow(fmt.Sprintf("%d", row.ARC), row.Title, row.Status, fmt.Sprintf("%d", row.Count))
		}
	})
}

func writeReferencedARCTable(out *bytes.Buffer, rows []ReferencedARCCount, countLabel string) {
	if len(rows) == 0 {
		_, _ = out.WriteString("None\n")
		return
	}
	writeTable(out, []string{"ARC", "Title", countLabel}, func(writeRow func(...string)) {
		for _, row := range rows {
			writeRow(fmt.Sprintf("%d", row.ARC), row.Title, fmt.Sprintf("%d", row.Count))
		}
	})
}

func writeSupersessionTable(out *bytes.Buffer, rows []SupersessionSummaryRow) {
	if len(rows) == 0 {
		_, _ = out.WriteString("None\n")
		return
	}
	writeTable(out, []string{"ARC", "Title", "Supersedes", "Superseded By"}, func(writeRow func(...string)) {
		for _, row := range rows {
			writeRow(fmt.Sprintf("%d", row.ARC), row.Title, row.Supersedes, row.SupersededBy)
		}
	})
}

func writeTable(out *bytes.Buffer, headers []string, writeRows func(func(...string))) {
	_, _ = out.WriteString("| " + strings.Join(sanitizeMarkdownTableCells(headers), " | ") + " |\n")
	separators := make([]string, 0, len(headers))
	for range headers {
		separators = append(separators, "---")
	}
	_, _ = out.WriteString("| " + strings.Join(separators, " | ") + " |\n")
	writeRows(func(values ...string) {
		_, _ = out.WriteString("| " + strings.Join(sanitizeMarkdownTableCells(values), " | ") + " |\n")
	})
}

func sanitizeMarkdownTableCells(values []string) []string {
	sanitized := make([]string, 0, len(values))
	for _, value := range values {
		sanitized = append(sanitized, sanitizeMarkdownTableCell(value))
	}
	return sanitized
}

func sanitizeMarkdownTableCell(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, "\n", "<br>")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}
