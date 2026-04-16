package repo

import (
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
	adopterCategoryCounts := map[string]int{}
	for _, category := range adoption.CategoryNames() {
		adopterCategoryCounts[category] = 0
	}
	adopterStatusCounts := map[string]int{}
	for _, status := range adoption.ActorStatuses() {
		adopterStatusCounts[status] = 0
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

			for _, category := range adoption.CategoryNames() {
				actors := adoptionSummary.Actors(category)
				adopterCategoryCounts[category] += len(actors)
				adopterCount += len(actors)
				for _, actor := range actors {
					adopterStatusCounts[normalizeActorStatus(actor.Status)]++
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

		if document.ImplementationRequired && refImplStatus != adoption.ReferenceImplementationStatusShipped {
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
	summary.AdoptionReadinessCounts = orderedCounts(readinessCounts, adoption.ReadinessLevels())
	summary.ReferenceImplementationStatusCounts = orderedCounts(refImplCounts, adoption.ReferenceImplementationStatuses())
	summary.AdopterEntriesByCategory = orderedCounts(adopterCategoryCounts, adoption.CategoryNames())
	summary.AdopterEntriesByStatus = orderedCounts(adopterStatusCounts, append(adoption.ActorStatuses(), "missing"))
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

func normalizeActorStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "":
		return "missing"
	case adoption.ActorStatusPlanned, adoption.ActorStatusInProgress, adoption.ActorStatusShipped, adoption.ActorStatusDeclined, adoption.ActorStatusUnknown:
		return strings.TrimSpace(value)
	default:
		return adoption.ActorStatusUnknown
	}
}
