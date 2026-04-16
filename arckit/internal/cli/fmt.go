package cli

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/config"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

func applyNativeFix(path string) error {
	return applyNativeFixWithConfig(path, config.Config{})
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
	if !arc.IsStringSequenceField(entry.key) {
		return nil
	}
	values := document.StringSequenceField(entry.key, entry.key == "updated")
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
	if !arc.IsIntSequenceField(entry.key) {
		return nil
	}
	values := document.IntSequenceField(entry.key)
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
	if !arc.IsScalarDateField(entry.key) {
		return entry.lines
	}
	value, ok := document.Fields[entry.key]
	if !ok {
		return entry.lines
	}
	normalized, ok := arc.NormalizeScalarDateValue(value)
	if !ok || len(entry.lines) != 1 {
		return entry.lines
	}
	line := strings.TrimSpace(entry.lines[0])
	if !strings.HasPrefix(line, entry.key+":") {
		return entry.lines
	}
	return []string{entry.key + ": " + normalized}
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
