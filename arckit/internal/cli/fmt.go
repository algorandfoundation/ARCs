package cli

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
	"github.com/algorandfoundation/ARCs/arckit/internal/config"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
	"gopkg.in/yaml.v3"
)

func applyNativeFix(path string) error {
	return applyNativeFixWithConfig(path, config.Config{})
}

func applyFixWithConfig(path string, cfg config.Config) error {
	if isAdoptionSummaryPath(path) {
		return applyAdoptionFixWithConfig(path)
	}
	return applyNativeFixWithConfig(path, cfg)
}

func applyNativeFixWithConfig(path string, cfg config.Config) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	document, diagnostics, err := arc.Load(path)
	if err != nil {
		return err
	}
	diagnostics = cfg.FilterDiagnostics(diagnostics)
	if hasUnsafeFrontMatterDiagnostics(diagnostics) {
		return fmt.Errorf("front matter could not be safely reformatted")
	}

	reordered, err := reorderFrontMatter(document)
	if err != nil {
		return err
	}

	fixed := []byte(reordered)
	if bytes.Equal(content, fixed) {
		return nil
	}
	return os.WriteFile(path, fixed, 0o644)
}

func applyAdoptionFixWithConfig(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	summary, diagnostics, err := adoption.Load(path)
	if err != nil {
		return err
	}
	if len(diagnostics) != 0 || hasUnsafeAdoptionSummaryShape(summary) {
		return fmt.Errorf("adoption summary could not be safely reformatted")
	}

	summary.Summary.AdoptionReadiness = summary.NormalizedAdoptionReadiness()

	rendered, err := renderAdoptionSummary(summary)
	if err != nil {
		return err
	}
	if bytes.Equal(content, rendered) {
		return nil
	}
	return os.WriteFile(path, rendered, 0o644)
}

func hasUnsafeAdoptionSummaryShape(summary *adoption.Summary) bool {
	if summary == nil {
		return true
	}
	requiredTopLevelKeys := []string{"arc", "title", "last-reviewed", "adoption", "summary"}
	for _, key := range requiredTopLevelKeys {
		if _, ok := summaryKeys(summary)[key]; !ok {
			return true
		}
	}
	for key := range summaryKeys(summary) {
		switch key {
		case "arc", "title", "last-reviewed", "reference-implementation", "adoption", "summary":
		default:
			return true
		}
	}
	for key := range summaryReferenceImplementationKeys(summary) {
		if key != "status" && key != "notes" {
			return true
		}
	}
	requiredAdoptionKeys := []string{"wallets", "explorers", "tooling", "infra", "dapps-protocols"}
	for _, key := range requiredAdoptionKeys {
		if _, ok := summaryAdoptionKeys(summary)[key]; !ok {
			return true
		}
	}
	for key := range summaryAdoptionKeys(summary) {
		switch key {
		case "wallets", "explorers", "tooling", "infra", "dapps-protocols":
		default:
			return true
		}
	}
	return false
}

func renderAdoptionSummary(summary *adoption.Summary) ([]byte, error) {
	type referenceImplementation struct {
		Status string  `yaml:"status"`
		Notes  *string `yaml:"notes,omitempty"`
	}
	type renderedSummary struct {
		Arc                     int                      `yaml:"arc"`
		Title                   string                   `yaml:"title"`
		LastReviewed            string                   `yaml:"last-reviewed"`
		ReferenceImplementation *referenceImplementation `yaml:"reference-implementation,omitempty"`
		Adoption                adoption.AdoptionSection `yaml:"adoption"`
		Summary                 adoption.SummarySection  `yaml:"summary"`
	}

	var referenceImplementationValue *referenceImplementation
	if summary.ReferenceImplementation != nil {
		referenceImplementationValue = &referenceImplementation{
			Status: summary.ReferenceImplementation.Status,
		}
		if _, ok := summaryReferenceImplementationKeys(summary)["notes"]; ok {
			notes := summary.ReferenceImplementation.Notes
			referenceImplementationValue.Notes = &notes
		}
	}

	rendered, err := yaml.Marshal(renderedSummary{
		Arc:                     summary.Arc,
		Title:                   summary.Title,
		LastReviewed:            summary.LastReviewed,
		ReferenceImplementation: referenceImplementationValue,
		Adoption:                summary.Adoption,
		Summary:                 summary.Summary,
	})
	if err != nil {
		return nil, err
	}
	return rendered, nil
}

func isAdoptionSummaryPath(path string) bool {
	return strings.HasPrefix(filepath.ToSlash(filepath.Clean(path)), "adoption/arc-") || strings.Contains(filepath.ToSlash(filepath.Clean(path)), "/adoption/arc-")
}

func summaryKeys(summary *adoption.Summary) map[string]struct{} {
	if summary == nil {
		return map[string]struct{}{}
	}
	return summary.KeySet()
}

func summaryReferenceImplementationKeys(summary *adoption.Summary) map[string]struct{} {
	if summary == nil {
		return map[string]struct{}{}
	}
	return summary.ReferenceImplementationKeySet()
}

func summaryAdoptionKeys(summary *adoption.Summary) map[string]struct{} {
	if summary == nil {
		return map[string]struct{}{}
	}
	return summary.AdoptionKeySet()
}

func hasUnsafeFrontMatterDiagnostics(diagnostics []diag.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		switch diagnostic.RuleID {
		case "R:001", "R:005", "R:006":
			return true
		}
	}
	return false
}

func reorderFrontMatter(document *arc.Document) (string, error) {
	entries, preamble, suffix, err := frontMatterEntries(document)
	if err != nil {
		return "", err
	}

	builder := strings.Builder{}
	builder.WriteString("---\n")

	for _, line := range preamble {
		if strings.TrimSpace(line) == "" {
			continue
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}

	for _, entry := range reorderedEntries(entries) {
		for _, line := range filteredFrontMatterLines(normalizedEntryLines(document, entry)) {
			builder.WriteString(line)
			builder.WriteString("\n")
		}
	}

	for _, line := range suffix {
		if strings.TrimSpace(line) == "" {
			continue
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}

	builder.WriteString("---\n")
	if len(document.Body) > 0 {
		builder.Write(document.Body)
	}
	return builder.String(), nil
}

func normalizeFrontMatterValue(key string, value any) any {
	if !isScalarDateField(key) {
		return value
	}
	switch typed := value.(type) {
	case time.Time:
		return typed.Format("2006-01-02")
	case string:
		if parsed, err := time.Parse(time.RFC3339, typed); err == nil {
			return parsed.Format("2006-01-02")
		}
	}
	return value
}

func isScalarDateField(key string) bool {
	switch key {
	case "created", "last-call-deadline", "idle-since":
		return true
	default:
		return false
	}
}

func isStringSequenceField(key string) bool {
	switch key {
	case "author", "updated", "implementation-maintainer":
		return true
	default:
		return false
	}
}

func isIntSequenceField(key string) bool {
	switch key {
	case "requires", "supersedes", "extends", "extended-by":
		return true
	default:
		return false
	}
}

type frontMatterEntry struct {
	key           string
	lines         []string
	known         bool
	orderIndex    int
	originalIndex int
	rank          float64
}

func frontMatterEntries(document *arc.Document) ([]frontMatterEntry, []string, []string, error) {
	if len(document.FrontMatter) == 0 {
		return nil, nil, nil, nil
	}
	lines := strings.Split(string(document.FrontMatter), "\n")
	if len(document.FieldOrder) == 0 {
		return nil, lines, nil, nil
	}

	preambleEnd := document.FieldLines[document.FieldOrder[0]] - 1
	if preambleEnd < 0 {
		preambleEnd = 0
	}
	if preambleEnd > len(lines) {
		preambleEnd = len(lines)
	}
	preamble := append([]string(nil), lines[:preambleEnd]...)

	orderLookup := map[string]int{}
	for idx, key := range arc.OrderedFields() {
		orderLookup[key] = idx
	}

	entries := make([]frontMatterEntry, 0, len(document.FieldOrder))
	lastEnd := preambleEnd
	for idx, key := range document.FieldOrder {
		start := document.FieldLines[key] - 1
		if start < 0 || start > len(lines) {
			return nil, nil, nil, fmt.Errorf("front matter field %q has invalid line metadata", key)
		}
		end := len(lines)
		if idx+1 < len(document.FieldOrder) {
			end = document.FieldLines[document.FieldOrder[idx+1]] - 1
		}
		if end < start || end > len(lines) {
			return nil, nil, nil, fmt.Errorf("front matter field %q has invalid line range", key)
		}
		orderIndex, known := orderLookup[key]
		entries = append(entries, frontMatterEntry{
			key:           key,
			lines:         append([]string(nil), lines[start:end]...),
			known:         known,
			orderIndex:    orderIndex,
			originalIndex: idx,
		})
		lastEnd = end
	}

	suffix := append([]string(nil), lines[lastEnd:]...)
	return entries, preamble, suffix, nil
}

func reorderedEntries(entries []frontMatterEntry) []frontMatterEntry {
	if len(entries) == 0 {
		return entries
	}
	assignEntryRanks(entries)
	out := append([]frontMatterEntry(nil), entries...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].rank == out[j].rank {
			return out[i].originalIndex < out[j].originalIndex
		}
		return out[i].rank < out[j].rank
	})
	return out
}

func assignEntryRanks(entries []frontMatterEntry) {
	for idx := range entries {
		if entries[idx].known {
			entries[idx].rank = float64(entries[idx].orderIndex * 1000)
		}
	}

	for start := 0; start < len(entries); {
		if entries[start].known {
			start++
			continue
		}
		end := start
		for end+1 < len(entries) && !entries[end+1].known {
			end++
		}

		prevOrder, hasPrev := nearestKnownOrder(entries, start-1, -1)
		nextOrder, hasNext := nearestKnownOrder(entries, end+1, 1)
		count := end - start + 1

		switch {
		case hasPrev && hasNext:
			low := float64(prevOrder * 1000)
			high := float64(nextOrder * 1000)
			if low > high {
				low, high = high, low
			}
			step := (high - low) / float64(count+1)
			for i := 0; i < count; i++ {
				entries[start+i].rank = low + step*float64(i+1)
			}
		case hasPrev:
			base := float64(prevOrder * 1000)
			for i := 0; i < count; i++ {
				entries[start+i].rank = base + 500 + float64(i)
			}
		case hasNext:
			base := float64(nextOrder * 1000)
			for i := 0; i < count; i++ {
				entries[start+i].rank = base - 500 + float64(i)
			}
		default:
			for i := 0; i < count; i++ {
				entries[start+i].rank = float64(i)
			}
		}
		start = end + 1
	}
}

func nearestKnownOrder(entries []frontMatterEntry, start int, step int) (int, bool) {
	for idx := start; idx >= 0 && idx < len(entries); idx += step {
		if entries[idx].known {
			return entries[idx].orderIndex, true
		}
	}
	return 0, false
}

func normalizedEntryLines(document *arc.Document, entry frontMatterEntry) []string {
	if lines := normalizeStringSequenceChunk(document, entry); len(lines) > 0 {
		return filteredFrontMatterLines(lines)
	}
	if lines := normalizeIntSequenceChunk(document, entry); len(lines) > 0 {
		return filteredFrontMatterLines(lines)
	}
	return filteredFrontMatterLines(normalizeDateChunk(document, entry))
}

func normalizeStringSequenceChunk(document *arc.Document, entry frontMatterEntry) []string {
	if !isStringSequenceField(entry.key) {
		return nil
	}
	values := stringListField(document, entry.key)
	if len(values) == 0 {
		return nil
	}
	lines := []string{entry.key + ":"}
	for _, value := range values {
		lines = append(lines, "  - "+value)
	}
	return lines
}

func normalizeIntSequenceChunk(document *arc.Document, entry frontMatterEntry) []string {
	if !isIntSequenceField(entry.key) {
		return nil
	}
	values := intListField(document, entry.key)
	if len(values) == 0 {
		return nil
	}
	lines := []string{entry.key + ":"}
	for _, value := range values {
		lines = append(lines, fmt.Sprintf("  - %d", value))
	}
	return lines
}

func normalizeDateChunk(document *arc.Document, entry frontMatterEntry) []string {
	if !isScalarDateField(entry.key) {
		return entry.lines
	}
	value, ok := document.Fields[entry.key]
	if !ok {
		return entry.lines
	}
	normalized, ok := normalizeFrontMatterValue(entry.key, value).(string)
	if !ok || len(entry.lines) != 1 {
		return entry.lines
	}
	line := strings.TrimSpace(entry.lines[0])
	if !strings.HasPrefix(line, entry.key+":") {
		return entry.lines
	}
	return []string{entry.key + ": " + normalized}
}

func stringListField(document *arc.Document, key string) []string {
	value, ok := document.Fields[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			switch value := item.(type) {
			case string:
				trimmed := strings.TrimSpace(value)
				if trimmed == "" {
					return nil
				}
				out = append(out, trimmed)
			case time.Time:
				out = append(out, value.Format("2006-01-02"))
			default:
				return nil
			}
		}
		return out
	default:
		return nil
	}
}

func intListField(document *arc.Document, key string) []int {
	value, ok := document.Fields[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case []any:
		out := make([]int, 0, len(typed))
		for _, item := range typed {
			switch value := item.(type) {
			case int:
				out = append(out, value)
			case int64:
				out = append(out, int(value))
			case float64:
				out = append(out, int(value))
			default:
				return nil
			}
		}
		return out
	default:
		return nil
	}
}

func filteredFrontMatterLines(lines []string) []string {
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		filtered = append(filtered, line)
	}
	return filtered
}
