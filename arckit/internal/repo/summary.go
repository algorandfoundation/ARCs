package repo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/algorandfoundation/ARCs/arckit/internal/adoption"
	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

const (
	upcomingLastCallWindowDays = 14
	staleAdoptionThresholdDays = 30
	dateLayout                 = "2006-01-02"
)

var (
	statusOrder = []string{"Draft", "Review", "Last Call", "Final", "Stagnant", "Withdrawn", "Idle", "Deprecated", "Living"}
	typeOrder   = []string{"Standards Track", "Meta"}
)

type RepoSummary struct {
	Root                                string
	GeneratedAt                         time.Time
	ValidationSummary                   diag.Summary
	TotalARCs                           int
	StatusCounts                        []SummaryCount
	TypeCounts                          []SummaryCount
	ARCsWithAssets                      int
	AdoptionSummaryCoverage             int
	TotalAdoptionFiles                  int
	TotalAdopterEntries                 int
	ARCsWithTrackedAdopters             int
	AdoptionReadinessCounts             []SummaryCount
	ReferenceImplementationStatusCounts []SummaryCount
	OverdueLastCall                     []LastCallSummaryRow
	UpcomingLastCall                    []LastCallSummaryRow
	IdleARCs                            []IdleSummaryRow
	ImplementationRequiredNotShipped    []ImplementationSummaryRow
	FinalZeroAdopters                   []AdoptionSummaryRow
	FinalLowAdopters                    []AdoptionSummaryRow
	StaleAdoptionReviews                []StaleAdoptionSummaryRow
	AdopterEntriesByCategory            []SummaryCount
	AdopterEntriesByStatus              []SummaryCount
	TopAdoptersByCoverage               []NamedCount
	TopARCsByAdopters                   []ARCCount
	TopRequiresTargets                  []ReferencedARCCount
	TopExtendsTargets                   []ReferencedARCCount
	SupersessionRows                    []SupersessionSummaryRow
}

type SummaryCount struct {
	Label string
	Count int
}

type NamedCount struct {
	Name  string
	Count int
}

type ARCCount struct {
	ARC    int
	Title  string
	Status string
	Count  int
}

type ReferencedARCCount struct {
	ARC   int
	Title string
	Count int
}

type LastCallSummaryRow struct {
	ARC    int
	Title  string
	Date   string
	Days   int
	Action string
}

type IdleSummaryRow struct {
	ARC    int
	Title  string
	Date   string
	Days   int
	Action string
}

type ImplementationSummaryRow struct {
	ARC                     int
	Title                   string
	Status                  string
	ReferenceImplementation string
	Action                  string
}

type AdoptionSummaryRow struct {
	ARC               int
	Title             string
	AdoptionReadiness string
	LastReviewed      string
	AdopterCount      int
	Action            string
}

type StaleAdoptionSummaryRow struct {
	ARC          int
	Title        string
	Status       string
	LastReviewed string
	AgeDays      int
	Action       string
}

type SupersessionSummaryRow struct {
	ARC          int
	Title        string
	Supersedes   string
	SupersededBy string
}

func BuildSummary(state State, diagnostics []diag.Diagnostic, now time.Time) RepoSummary {
	summary := RepoSummary{
		Root:               filepath.Clean(state.Root),
		GeneratedAt:        now.UTC(),
		ValidationSummary:  diag.Summarize(diagnostics),
		TotalARCs:          len(state.ARCs),
		TotalAdoptionFiles: len(state.Adoptions),
	}

	statusCounts := map[string]int{}
	typeCounts := map[string]int{}
	readinessCounts := map[string]int{}
	refImplCounts := map[string]int{}
	adopterCategoryCounts := map[string]int{
		"wallets":         0,
		"explorers":       0,
		"tooling":         0,
		"infra":           0,
		"dapps-protocols": 0,
	}
	adopterStatusCounts := map[string]int{
		"planned":     0,
		"in_progress": 0,
		"shipped":     0,
		"declined":    0,
		"unknown":     0,
	}
	adopterCoverage := map[string]map[int]struct{}{}
	topARCsByAdopters := make([]ARCCount, 0)
	requiresTargets := map[int]int{}
	extendsTargets := map[int]int{}

	currentDate := normalizeDay(now)
	for number, document := range state.ARCs {
		statusCounts[document.Status]++
		typeCounts[document.Type]++
		if strings.TrimSpace(document.AdoptionSummary) != "" {
			summary.AdoptionSummaryCoverage++
		}
		if assetDirExists(state.Root, number) {
			summary.ARCsWithAssets++
		}

		if document.Status == "Last Call" {
			if deadline, ok := parseDate(document.LastCallDeadline); ok {
				days := int(currentDate.Sub(deadline).Hours() / 24)
				if days > 0 {
					summary.OverdueLastCall = append(summary.OverdueLastCall, LastCallSummaryRow{
						ARC:    number,
						Title:  document.Title,
						Date:   document.LastCallDeadline,
						Days:   days,
						Action: "Decide Final, extend Last Call, or move back to Review",
					})
				} else if days >= -upcomingLastCallWindowDays {
					summary.UpcomingLastCall = append(summary.UpcomingLastCall, LastCallSummaryRow{
						ARC:    number,
						Title:  document.Title,
						Date:   document.LastCallDeadline,
						Days:   -days,
						Action: "Prepare review decision before deadline",
					})
				}
			}
		}

		if document.Status == "Idle" {
			if idleSince, ok := parseDate(document.IdleSince); ok {
				summary.IdleARCs = append(summary.IdleARCs, IdleSummaryRow{
					ARC:    number,
					Title:  document.Title,
					Date:   document.IdleSince,
					Days:   int(currentDate.Sub(idleSince).Hours() / 24),
					Action: "Confirm Idle should remain or restart editorial follow-up",
				})
			}
		}

		for _, target := range document.Requires {
			requiresTargets[target]++
		}
		for _, target := range document.Extends {
			extendsTargets[target]++
		}
		if len(document.Supersedes) > 0 || len(document.SupersededBy) > 0 {
			summary.SupersessionRows = append(summary.SupersessionRows, SupersessionSummaryRow{
				ARC:          number,
				Title:        document.Title,
				Supersedes:   joinARCNumbers(document.Supersedes),
				SupersededBy: joinARCNumbers(document.SupersededBy),
			})
		}

		adoptionSummary := state.Adoptions[number]
		refImplStatus := "missing"
		adopterCount := 0
		adoptionReadiness := "missing"
		lastReviewed := "missing"

		if adoptionSummary != nil {
			readiness := strings.TrimSpace(adoptionSummary.Summary.AdoptionReadiness)
			if readiness != "" {
				readinessCounts[readiness]++
				adoptionReadiness = readiness
			}
			lastReviewedValue := strings.TrimSpace(adoptionSummary.LastReviewed)
			if lastReviewedValue != "" {
				lastReviewed = lastReviewedValue
			}
			if adoptionSummary.ReferenceImplementation != nil {
				status := strings.TrimSpace(adoptionSummary.ReferenceImplementation.Status)
				if status != "" {
					refImplStatus = status
					refImplCounts[status]++
				}
			}

			actorGroups := adoptionGroups(adoptionSummary)
			for _, group := range []struct {
				key    string
				actors []adoption.Actor
			}{
				{key: "wallets", actors: actorGroups["wallets"]},
				{key: "explorers", actors: actorGroups["explorers"]},
				{key: "tooling", actors: actorGroups["tooling"]},
				{key: "infra", actors: actorGroups["infra"]},
				{key: "dapps-protocols", actors: actorGroups["dapps-protocols"]},
			} {
				adopterCategoryCounts[group.key] += len(group.actors)
				adopterCount += len(group.actors)
				for _, actor := range group.actors {
					adopterStatusCounts[actor.Status]++
					if strings.TrimSpace(actor.Name) == "" {
						continue
					}
					coverage, ok := adopterCoverage[actor.Name]
					if !ok {
						coverage = map[int]struct{}{}
						adopterCoverage[actor.Name] = coverage
					}
					coverage[number] = struct{}{}
				}
			}
		}

		summary.TotalAdopterEntries += adopterCount
		if adopterCount > 0 {
			summary.ARCsWithTrackedAdopters++
			topARCsByAdopters = append(topARCsByAdopters, ARCCount{
				ARC:    number,
				Title:  document.Title,
				Status: document.Status,
				Count:  adopterCount,
			})
		}

		if document.ImplementationRequired && refImplStatus != "shipped" {
			summary.ImplementationRequiredNotShipped = append(summary.ImplementationRequiredNotShipped, ImplementationSummaryRow{
				ARC:                     number,
				Title:                   document.Title,
				Status:                  document.Status,
				ReferenceImplementation: refImplStatus,
				Action:                  "Check canonical implementation readiness before transition",
			})
		}

		if document.Status == "Final" {
			row := AdoptionSummaryRow{
				ARC:               number,
				Title:             document.Title,
				AdoptionReadiness: adoptionReadiness,
				LastReviewed:      lastReviewed,
			}
			switch {
			case adopterCount == 0:
				row.Action = "Backfill at least one tracked adopter or confirm historical exception"
				summary.FinalZeroAdopters = append(summary.FinalZeroAdopters, row)
			case adopterCount <= 2:
				row.AdopterCount = adopterCount
				row.Action = "Check whether additional adoption evidence exists"
				summary.FinalLowAdopters = append(summary.FinalLowAdopters, row)
			}
		}

		if adoptionSummary != nil {
			if reviewedAt, ok := parseDate(adoptionSummary.LastReviewed); ok {
				ageDays := int(currentDate.Sub(reviewedAt).Hours() / 24)
				if ageDays > staleAdoptionThresholdDays {
					summary.StaleAdoptionReviews = append(summary.StaleAdoptionReviews, StaleAdoptionSummaryRow{
						ARC:          number,
						Title:        document.Title,
						Status:       document.Status,
						LastReviewed: adoptionSummary.LastReviewed,
						AgeDays:      ageDays,
						Action:       "Refresh adoption summary and evidence",
					})
				}
			}
		}
	}

	summary.StatusCounts = orderedCounts(statusCounts, statusOrder)
	summary.TypeCounts = orderedCounts(typeCounts, typeOrder)
	summary.AdoptionReadinessCounts = orderedCounts(readinessCounts, []string{"low", "medium", "high"})
	summary.ReferenceImplementationStatusCounts = orderedCounts(refImplCounts, []string{"planned", "wip", "shipped", "archived"})
	summary.AdopterEntriesByCategory = orderedCounts(adopterCategoryCounts, []string{"wallets", "explorers", "tooling", "infra", "dapps-protocols"})
	summary.AdopterEntriesByStatus = orderedCounts(adopterStatusCounts, []string{"planned", "in_progress", "shipped", "declined", "unknown"})
	summary.TopAdoptersByCoverage = topAdopterCoverage(adopterCoverage)
	summary.TopARCsByAdopters = topARCs(topARCsByAdopters)
	summary.TopRequiresTargets = topReferencedARCs(requiresTargets, state.ARCs)
	summary.TopExtendsTargets = topReferencedARCs(extendsTargets, state.ARCs)

	sort.Slice(summary.OverdueLastCall, func(i, j int) bool {
		if summary.OverdueLastCall[i].Days != summary.OverdueLastCall[j].Days {
			return summary.OverdueLastCall[i].Days > summary.OverdueLastCall[j].Days
		}
		return summary.OverdueLastCall[i].ARC < summary.OverdueLastCall[j].ARC
	})
	sort.Slice(summary.UpcomingLastCall, func(i, j int) bool {
		if summary.UpcomingLastCall[i].Days != summary.UpcomingLastCall[j].Days {
			return summary.UpcomingLastCall[i].Days < summary.UpcomingLastCall[j].Days
		}
		return summary.UpcomingLastCall[i].ARC < summary.UpcomingLastCall[j].ARC
	})
	sort.Slice(summary.IdleARCs, func(i, j int) bool {
		if summary.IdleARCs[i].Days != summary.IdleARCs[j].Days {
			return summary.IdleARCs[i].Days > summary.IdleARCs[j].Days
		}
		return summary.IdleARCs[i].ARC < summary.IdleARCs[j].ARC
	})
	sort.Slice(summary.ImplementationRequiredNotShipped, func(i, j int) bool {
		return summary.ImplementationRequiredNotShipped[i].ARC < summary.ImplementationRequiredNotShipped[j].ARC
	})
	sort.Slice(summary.FinalZeroAdopters, func(i, j int) bool {
		return summary.FinalZeroAdopters[i].ARC < summary.FinalZeroAdopters[j].ARC
	})
	sort.Slice(summary.FinalLowAdopters, func(i, j int) bool {
		if summary.FinalLowAdopters[i].AdopterCount != summary.FinalLowAdopters[j].AdopterCount {
			return summary.FinalLowAdopters[i].AdopterCount < summary.FinalLowAdopters[j].AdopterCount
		}
		return summary.FinalLowAdopters[i].ARC < summary.FinalLowAdopters[j].ARC
	})
	sort.Slice(summary.StaleAdoptionReviews, func(i, j int) bool {
		if summary.StaleAdoptionReviews[i].AgeDays != summary.StaleAdoptionReviews[j].AgeDays {
			return summary.StaleAdoptionReviews[i].AgeDays > summary.StaleAdoptionReviews[j].AgeDays
		}
		return summary.StaleAdoptionReviews[i].ARC < summary.StaleAdoptionReviews[j].ARC
	})
	sort.Slice(summary.SupersessionRows, func(i, j int) bool {
		return summary.SupersessionRows[i].ARC < summary.SupersessionRows[j].ARC
	})

	return summary
}

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
	writeCountTable(&out, "Status", summary.StatusCounts)
	writeBlank()
	writeLine("### Counts by Type")
	writeCountTable(&out, "Type", summary.TypeCounts)
	writeBlank()
	writeLine("### Adoption Readiness Distribution")
	writeCountTable(&out, "Adoption Readiness", summary.AdoptionReadinessCounts)
	writeBlank()
	writeLine("### Reference Implementation Status Counts")
	writeCountTable(&out, "Reference Implementation Status", summary.ReferenceImplementationStatusCounts)
	writeBlank()

	writeLine("## Transition Watch")
	writeBlank()
	writeLine("### Overdue Last Call")
	writeLastCallTable(&out, summary.OverdueLastCall, "Days Overdue")
	writeBlank()
	writeLine("### Upcoming Last Call (next %d days)", upcomingLastCallWindowDays)
	writeLastCallTable(&out, summary.UpcomingLastCall, "Days Remaining")
	writeBlank()
	writeLine("### Idle ARCs")
	writeIdleTable(&out, summary.IdleARCs)
	writeBlank()
	writeLine("### Implementation-Required ARCs Not Shipped")
	writeImplementationTable(&out, summary.ImplementationRequiredNotShipped)
	writeBlank()

	writeLine("## Adoption Watch")
	writeBlank()
	writeLine("### Final ARCs With Zero Adopters")
	writeAdoptionActionTable(&out, summary.FinalZeroAdopters)
	writeBlank()
	writeLine("### Final ARCs With 1-2 Adopters")
	writeAdoptionActionTable(&out, summary.FinalLowAdopters)
	writeBlank()
	writeLine("### Stale Adoption Reviews (>%d days)", staleAdoptionThresholdDays)
	writeStaleAdoptionTable(&out, summary.StaleAdoptionReviews)
	writeBlank()
	writeLine("### Adoption Totals")
	writeBlank()
	writeLine("#### Adopter Entries by Category")
	writeCountTable(&out, "Category", summary.AdopterEntriesByCategory)
	writeBlank()
	writeLine("#### Adopter Entries by Actor Status")
	writeCountTable(&out, "Actor Status", summary.AdopterEntriesByStatus)
	writeBlank()
	writeLine("#### Top Adopters by Distinct ARC Coverage")
	writeNamedCountTable(&out, "Adopter", summary.TopAdoptersByCoverage)
	writeBlank()
	writeLine("#### Top ARCs by Adopter Count")
	writeARCCountTable(&out, summary.TopARCsByAdopters)
	writeBlank()

	writeLine("## Relationship Watch")
	writeBlank()
	writeLine("### Top ARCs Most Referenced by `requires`")
	writeReferencedARCTable(&out, summary.TopRequiresTargets, "Requires References")
	writeBlank()
	writeLine("### Top ARCs Most Referenced by `extends`")
	writeReferencedARCTable(&out, summary.TopExtendsTargets, "Extends References")
	writeBlank()
	writeLine("### Non-Empty `supersedes` / `superseded-by` Pairs")
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

func orderedCounts(counts map[string]int, preferred []string) []SummaryCount {
	ordered := make([]SummaryCount, 0, len(counts))
	seen := map[string]struct{}{}
	for _, label := range preferred {
		if count, ok := counts[label]; ok {
			ordered = append(ordered, SummaryCount{Label: label, Count: count})
			seen[label] = struct{}{}
		}
	}
	remaining := make([]string, 0, len(counts))
	for label := range counts {
		if _, ok := seen[label]; ok {
			continue
		}
		remaining = append(remaining, label)
	}
	sort.Strings(remaining)
	for _, label := range remaining {
		ordered = append(ordered, SummaryCount{Label: label, Count: counts[label]})
	}
	return ordered
}

func topAdopterCoverage(coverage map[string]map[int]struct{}) []NamedCount {
	rows := make([]NamedCount, 0, len(coverage))
	for name, arcs := range coverage {
		rows = append(rows, NamedCount{Name: name, Count: len(arcs)})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].Name < rows[j].Name
	})
	if len(rows) > 10 {
		rows = rows[:10]
	}
	return rows
}

func topARCs(rows []ARCCount) []ARCCount {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].ARC < rows[j].ARC
	})
	if len(rows) > 10 {
		rows = rows[:10]
	}
	return rows
}

func topReferencedARCs(counts map[int]int, documents map[int]*arc.Document) []ReferencedARCCount {
	rows := make([]ReferencedARCCount, 0, len(counts))
	for number, count := range counts {
		title := ""
		if document, ok := documents[number]; ok && document != nil {
			title = document.Title
		}
		rows = append(rows, ReferencedARCCount{ARC: number, Title: title, Count: count})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].ARC < rows[j].ARC
	})
	if len(rows) > 10 {
		rows = rows[:10]
	}
	return rows
}

func adoptionGroups(summary *adoption.Summary) map[string][]adoption.Actor {
	if summary == nil {
		return map[string][]adoption.Actor{}
	}
	return map[string][]adoption.Actor{
		"wallets":         summary.Adoption.Wallets,
		"explorers":       summary.Adoption.Explorers,
		"tooling":         summary.Adoption.Tooling,
		"infra":           summary.Adoption.Infra,
		"dapps-protocols": summary.Adoption.DappsProtocols,
	}
}

func assetDirExists(root string, number int) bool {
	info, err := os.Stat(filepath.Join(root, "assets", fmt.Sprintf("arc-%04d", number)))
	return err == nil && info.IsDir()
}

func parseDate(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func normalizeDay(now time.Time) time.Time {
	return time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func joinARCNumbers(numbers []int) string {
	if len(numbers) == 0 {
		return "None"
	}
	parts := make([]string, 0, len(numbers))
	for _, number := range numbers {
		parts = append(parts, fmt.Sprintf("%d", number))
	}
	return strings.Join(parts, ", ")
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
	_, _ = out.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	separators := make([]string, 0, len(headers))
	for range headers {
		separators = append(separators, "---")
	}
	_, _ = out.WriteString("| " + strings.Join(separators, " | ") + " |\n")
	writeRows(func(values ...string) {
		_, _ = out.WriteString("| " + strings.Join(values, " | ") + " |\n")
	})
}
