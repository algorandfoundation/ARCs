package cli

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

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
		return applyAdoptionFix(path)
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
	if unsafeErr := unsafeFrontMatterError(diagnostics); unsafeErr != nil {
		return unsafeErr
	}

	reordered, err := normalizeDocument(document)
	if err != nil {
		return err
	}

	fixed := []byte(reordered)
	if bytes.Equal(content, fixed) {
		return nil
	}
	return os.WriteFile(path, fixed, 0o644)
}

func unsafeFrontMatterError(diagnostics []diag.Diagnostic) error {
	for _, diagnostic := range diagnostics {
		switch diagnostic.RuleID {
		case "R:005":
			return fmt.Errorf("ARC front matter is not valid YAML (%s); pre-commit YAML hooks do not inspect ARC Markdown front matter, so fix the header and rerun fmt", diagnostic.Message)
		case "R:001", "R:006":
			return fmt.Errorf("ARC front matter could not be safely reformatted because of %s", diagnostic.Message)
		}
	}
	return nil
}

func normalizeDocument(document *arc.Document) (string, error) {
	entries, preamble, suffix, err := frontMatterEntries(document)
	if err != nil {
		return "", err
	}
	body := normalizeBody(document)

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
	if len(body) > 0 {
		builder.Write(body)
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
	if len(document.StringSequenceField(entry.key, entry.key == "updated")) == 0 {
		return nil
	}
	return entry.lines
}

func normalizeIntSequenceChunk(document *arc.Document, entry frontMatterEntry) []string {
	if !arc.IsIntSequenceField(entry.key) {
		return nil
	}
	values := document.IntSequenceField(entry.key)
	if len(values) == 0 {
		return nil
	}
	values = sortedUniqueInts(values)
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

func sortedUniqueInts(values []int) []int {
	out := append([]int(nil), values...)
	sort.Ints(out)
	write := 0
	for _, value := range out {
		if write > 0 && out[write-1] == value {
			continue
		}
		out[write] = value
		write++
	}
	return out[:write]
}

type bodySection struct {
	title      string
	orderIndex int
	startLine  int
	endLine    int
}

func normalizeBody(document *arc.Document) []byte {
	if len(document.Body) == 0 {
		return document.Body
	}

	sections, ok := reorderableBodySections(document)
	if !ok || len(sections) <= 1 {
		return document.Body
	}

	alreadyOrdered := true
	for idx := 1; idx < len(sections); idx++ {
		if sections[idx-1].orderIndex > sections[idx].orderIndex {
			alreadyOrdered = false
			break
		}
	}
	if alreadyOrdered {
		return document.Body
	}

	lines := splitLinesPreserveNewlines(document.Body)
	if len(lines) == 0 {
		return document.Body
	}
	if sections[0].startLine < 0 || sections[0].startLine > len(lines) {
		return document.Body
	}

	preamble := append([]string(nil), lines[:sections[0].startLine]...)
	ordered := append([]bodySection(nil), sections...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].orderIndex < ordered[j].orderIndex
	})

	builder := strings.Builder{}
	for _, line := range preamble {
		builder.WriteString(line)
	}
	for _, section := range ordered {
		if section.startLine < 0 || section.endLine < section.startLine || section.endLine > len(lines) {
			return document.Body
		}
		for _, line := range lines[section.startLine:section.endLine] {
			builder.WriteString(line)
		}
	}
	return []byte(builder.String())
}

func reorderableBodySections(document *arc.Document) ([]bodySection, bool) {
	bodyStartLine := document.BodyStartLine
	if bodyStartLine == 0 {
		bodyStartLine = 1
	}

	allowedOrder := map[string]int{}
	for idx, title := range arc.OrderedLevel2Sections() {
		allowedOrder[title] = idx
	}

	sections := make([]bodySection, 0)
	seen := map[string]struct{}{}
	for _, heading := range document.Headings {
		if heading.Level != 2 {
			continue
		}
		orderIndex, ok := allowedOrder[heading.Title]
		if !ok {
			return nil, false
		}
		if _, exists := seen[heading.Title]; exists {
			return nil, false
		}
		seen[heading.Title] = struct{}{}
		startLine := heading.Line - bodyStartLine
		if startLine < 0 {
			return nil, false
		}
		sections = append(sections, bodySection{
			title:      heading.Title,
			orderIndex: orderIndex,
			startLine:  startLine,
		})
	}

	if len(sections) == 0 {
		return nil, false
	}

	lines := splitLinesPreserveNewlines(document.Body)
	for idx := range sections {
		endLine := len(lines)
		if idx+1 < len(sections) {
			endLine = sections[idx+1].startLine
		}
		if endLine < sections[idx].startLine {
			return nil, false
		}
		sections[idx].endLine = endLine
	}
	return sections, true
}

func splitLinesPreserveNewlines(content []byte) []string {
	if len(content) == 0 {
		return nil
	}
	lines := strings.SplitAfter(string(content), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func applyAdoptionFix(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	summary, diagnostics, err := adoption.Load(path)
	if err != nil {
		return err
	}
	if unsafeErr := unsafeAdoptionError(diagnostics); unsafeErr != nil {
		return unsafeErr
	}

	normalized, err := normalizeAdoptionContent(content, summary.NormalizedAdoptionReadiness())
	if err != nil {
		return err
	}
	if bytes.Equal(content, normalized) {
		return nil
	}
	if len(summary.KeySet()) == 0 {
		return fmt.Errorf("adoption summary could not be safely reformatted")
	}
	return os.WriteFile(path, normalized, 0o644)
}

func unsafeAdoptionError(diagnostics []diag.Diagnostic) error {
	for _, diagnostic := range diagnostics {
		switch diagnostic.RuleID {
		case "R:016":
			return fmt.Errorf("adoption summary is not valid YAML or schema-safe for deterministic reordering (%s); fix the file and rerun fmt", diagnostic.Message)
		}
	}
	return nil
}

type mappingChunk struct {
	key           string
	known         bool
	orderIndex    int
	originalIndex int
	rank          float64
	startLine     int
	endLine       int
	valueNode     *yaml.Node
}

func normalizeAdoptionContent(content []byte, normalizedReadiness string) ([]byte, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, err
	}
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return content, nil
	}

	lines := splitLinesPreserveNewlines(content)
	if len(lines) == 0 {
		return content, nil
	}
	out, ok := renderOrderedMapping(lines, root.Content[0], 0, len(lines), "top", normalizedReadiness)
	if !ok {
		return content, nil
	}
	return []byte(strings.Join(out, "")), nil
}

func renderOrderedMapping(lines []string, mapping *yaml.Node, regionStart int, regionEnd int, context string, normalizedReadiness string) ([]string, bool) {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return append([]string(nil), lines[regionStart:regionEnd]...), true
	}
	if len(mapping.Content) == 0 {
		return append([]string(nil), lines[regionStart:regionEnd]...), true
	}

	entries, ok := mappingChunks(mapping, regionEnd, orderForAdoptionContext(context))
	if !ok || len(entries) == 0 {
		return append([]string(nil), lines[regionStart:regionEnd]...), ok
	}

	firstStart := entries[0].startLine
	if firstStart < regionStart || firstStart > regionEnd {
		return append([]string(nil), lines[regionStart:regionEnd]...), false
	}

	builder := make([]string, 0, regionEnd-regionStart)
	builder = append(builder, lines[regionStart:firstStart]...)
	for _, entry := range reorderedMappingChunks(entries) {
		if entry.startLine < regionStart || entry.endLine < entry.startLine || entry.endLine > regionEnd {
			return append([]string(nil), lines[regionStart:regionEnd]...), false
		}
		if entry.valueNode != nil && entry.valueNode.Kind == yaml.MappingNode && entry.valueNode.Line-1 > entry.startLine && entry.valueNode.Line-1 <= entry.endLine {
			nestedStart := entry.valueNode.Line - 1
			nested, ok := renderOrderedMapping(lines, entry.valueNode, nestedStart, entry.endLine, entry.key, normalizedReadiness)
			if !ok {
				return append([]string(nil), lines[regionStart:regionEnd]...), false
			}
			builder = append(builder, lines[entry.startLine:nestedStart]...)
			builder = append(builder, nested...)
			continue
		}
		if context == "summary" && entry.key == "adoption-readiness" {
			builder = append(builder, renderReadinessChunk(lines[entry.startLine:entry.endLine], normalizedReadiness)...)
			continue
		}
		builder = append(builder, lines[entry.startLine:entry.endLine]...)
	}
	return builder, true
}

func renderReadinessChunk(lines []string, readiness string) []string {
	if len(lines) == 0 || strings.TrimSpace(readiness) == "" {
		return lines
	}

	first := lines[0]
	colon := strings.Index(first, ":")
	if colon < 0 {
		return lines
	}

	updated := append([]string(nil), lines...)
	comment := ""
	lineEnding := ""
	if strings.HasSuffix(first, "\r\n") {
		lineEnding = "\r\n"
		first = strings.TrimSuffix(first, "\r\n")
	} else if strings.HasSuffix(first, "\n") {
		lineEnding = "\n"
		first = strings.TrimSuffix(first, "\n")
	}
	if marker := strings.Index(first[colon+1:], "#"); marker >= 0 {
		comment = strings.TrimSpace(first[colon+1+marker:])
	}

	updated[0] = first[:colon+1] + " " + readiness
	if comment != "" {
		updated[0] += " " + comment
	}
	updated[0] += lineEnding
	return updated
}

func mappingChunks(mapping *yaml.Node, regionEnd int, orderLookup map[string]int) ([]mappingChunk, bool) {
	entries := make([]mappingChunk, 0, len(mapping.Content)/2)
	for idx := 0; idx < len(mapping.Content); idx += 2 {
		keyNode := mapping.Content[idx]
		valueNode := mapping.Content[idx+1]
		startLine := keyNode.Line - 1
		endLine := regionEnd
		if idx+2 < len(mapping.Content) {
			endLine = mapping.Content[idx+2].Line - 1
		}
		if startLine < 0 || endLine < startLine {
			return nil, false
		}
		orderIndex, known := orderLookup[keyNode.Value]
		entries = append(entries, mappingChunk{
			key:           keyNode.Value,
			known:         known,
			orderIndex:    orderIndex,
			originalIndex: len(entries),
			startLine:     startLine,
			endLine:       endLine,
			valueNode:     valueNode,
		})
	}
	return entries, true
}

func reorderedMappingChunks(entries []mappingChunk) []mappingChunk {
	assignMappingChunkRanks(entries)
	out := append([]mappingChunk(nil), entries...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].rank == out[j].rank {
			return out[i].originalIndex < out[j].originalIndex
		}
		return out[i].rank < out[j].rank
	})
	return out
}

func assignMappingChunkRanks(entries []mappingChunk) {
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

		prevOrder, hasPrev := nearestKnownMappingOrder(entries, start-1, -1)
		nextOrder, hasNext := nearestKnownMappingOrder(entries, end+1, 1)
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

func nearestKnownMappingOrder(entries []mappingChunk, start int, step int) (int, bool) {
	for idx := start; idx >= 0 && idx < len(entries); idx += step {
		if entries[idx].known {
			return entries[idx].orderIndex, true
		}
	}
	return 0, false
}

func orderForAdoptionContext(context string) map[string]int {
	var ordered []string
	switch context {
	case "top":
		ordered = []string{"arc", "title", "last-reviewed", "reference-implementation", "adoption", "summary"}
	case "reference-implementation":
		ordered = []string{"status", "notes"}
	case "adoption":
		ordered = adoption.CategoryNames()
	case "summary":
		ordered = []string{"adoption-readiness", "blockers", "notes"}
	default:
		return map[string]int{}
	}

	orderLookup := make(map[string]int, len(ordered))
	for idx, key := range ordered {
		orderLookup[key] = idx
	}
	return orderLookup
}
