package arc

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
)

var (
	arcPathPattern  = regexp.MustCompile(`(^|.*/)ARCs/arc-(\d{4})\.md$`)
	datePattern     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	assetDirPattern = regexp.MustCompile(`(^|.*/)assets/arc-(\d{4})(/.*)?$`)
)

var fieldOrder = []string{
	"arc",
	"title",
	"description",
	"author",
	"discussions-to",
	"status",
	"type",
	"category",
	"sub-category",
	"created",
	"updated",
	"sponsor",
	"implementation-required",
	"implementation-url",
	"implementation-maintainer",
	"adoption-summary",
	"last-call-deadline",
	"idle-since",
	"requires",
	"supersedes",
	"superseded-by",
	"extends",
	"extended-by",
}

var requiredFields = []string{
	"arc",
	"title",
	"description",
	"author",
	"discussions-to",
	"status",
	"type",
	"created",
	"sponsor",
	"implementation-required",
}

var requiredSections = []string{
	"Abstract",
	"Motivation",
	"Specification",
	"Rationale",
	"Security Considerations",
}

var (
	canonicalStringSequenceFields = map[string]struct{}{
		"author":                    {},
		"updated":                   {},
		"implementation-maintainer": {},
	}
	canonicalIntSequenceFields = map[string]struct{}{
		"requires":    {},
		"supersedes":  {},
		"extends":     {},
		"extended-by": {},
	}
)

type Link struct {
	Destination string
	Line        int
}

type Document struct {
	Path                     string
	Raw                      []byte
	FrontMatter              []byte
	Body                     []byte
	BodyStartLine            int
	Fields                   map[string]any
	FieldOrder               []string
	FieldLines               map[string]int
	Sections                 map[string]int
	Links                    []Link
	ExternalLinks            []string
	FilenameNumber           int
	HasFilenameNumber        bool
	Number                   int
	HasNumber                bool
	Title                    string
	Description              string
	Status                   string
	Type                     string
	Category                 string
	SubCategory              string
	Sponsor                  string
	ImplementationRequired   bool
	ImplementationURL        string
	ImplementationMaintainer string
	AdoptionSummary          string
	LastCallDeadline         string
	IdleSince                string
	Requires                 []int
	Supersedes               []int
	SupersededBy             []int
	Extends                  []int
	ExtendedBy               []int
}

func Load(path string) (*Document, []diag.Diagnostic, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, []diag.Diagnostic{
			diag.NewWithHint("R:027", diag.OriginNative, path, 0, 0, err.Error(), "Check the file path and permissions, then retry."),
		}, err
	}

	document := &Document{
		Path:       filepath.Clean(path),
		Raw:        content,
		Fields:     map[string]any{},
		FieldLines: map[string]int{},
		Sections:   map[string]int{},
	}

	diagnostics := make([]diag.Diagnostic, 0)

	matches := arcPathPattern.FindStringSubmatch(filepath.ToSlash(document.Path))
	if len(matches) != 3 {
		diagnostics = append(diagnostics, diag.NewWithHint("R:002", diag.OriginNative, document.Path, 1, 1, "ARC files must live under ARCs/arc-####.md", "Move or rename the file to ARCs/arc-####.md."))
	} else {
		number, _ := strconv.Atoi(matches[2])
		document.FilenameNumber = number
		document.HasFilenameNumber = true
	}

	frontMatter, body, bodyStartLine, parseDiagnostics := splitFrontMatter(document.Path, content)
	diagnostics = append(diagnostics, parseDiagnostics...)
	document.FrontMatter = frontMatter
	document.Body = body
	document.BodyStartLine = bodyStartLine
	if len(frontMatter) == 0 {
		parseMarkdown(document)
		return document, diagnostics, nil
	}
	diagnostics = append(diagnostics, frontMatterBlankLineDiagnostics(document.Path, frontMatter)...)

	root := yaml.Node{}
	if err := yaml.Unmarshal(frontMatter, &root); err != nil {
		diagnostics = append(diagnostics, diag.NewWithHint("R:005", diag.OriginNative, document.Path, 1, 1, err.Error(), "Fix the YAML syntax in the front matter block."))
		parseMarkdown(document)
		return document, diagnostics, nil
	}

	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		diagnostics = append(diagnostics, diag.NewWithHint("R:005", diag.OriginNative, document.Path, 1, 1, "front matter must decode to a mapping", "Use top-level key/value pairs in the front matter block."))
		parseMarkdown(document)
		return document, diagnostics, nil
	}

	mapping := root.Content[0]
	orderIndex := make(map[string]int, len(fieldOrder))
	for index, name := range fieldOrder {
		orderIndex[name] = index
	}

	lastIndex := -1
	for index := 0; index < len(mapping.Content); index += 2 {
		keyNode := mapping.Content[index]
		valueNode := mapping.Content[index+1]
		key := keyNode.Value
		document.FieldOrder = append(document.FieldOrder, key)
		document.FieldLines[key] = keyNode.Line
		if _, ok := orderIndex[key]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:006", diag.OriginNative, document.Path, keyNode.Line, keyNode.Column, fmt.Sprintf("unknown ARC field %q", key), "Remove the unknown field or move the data into the document body."))
			continue
		}
		if orderIndex[key] < lastIndex {
			diagnostics = append(diagnostics, diag.NewWithHint("R:003", diag.OriginNative, document.Path, keyNode.Line, keyNode.Column, fmt.Sprintf("field %q is out of order", key), "Reorder the front matter fields to match the canonical order."))
		} else {
			lastIndex = orderIndex[key]
		}
		var value any
		if err := valueNode.Decode(&value); err != nil {
			diagnostics = append(diagnostics, diag.NewWithHint("R:005", diag.OriginNative, document.Path, valueNode.Line, valueNode.Column, fmt.Sprintf("could not decode %q: %v", key, err), "Fix the YAML value for this field."))
			continue
		}
		document.Fields[key] = value
	}

	parseMarkdown(document)
	return document, diagnostics, nil
}

func Validate(document *Document, repoRoot string) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)

	for _, field := range requiredFields {
		if !hasValue(document.Fields, field) {
			diagnostics = append(diagnostics, diag.NewWithHint("R:004", diag.OriginNative, document.Path, document.FieldLines[field], 1, fmt.Sprintf("missing required field %q", field), "Add the field to the ARC front matter."))
		}
	}

	document.Number = 0
	document.HasNumber = false
	if number, ok := intField(document, "arc"); ok {
		if document.HasFilenameNumber && number != document.FilenameNumber {
			diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["arc"], 1, fmt.Sprintf("front matter arc value %d does not match filename number %d", number, document.FilenameNumber), "Keep the filename and the arc field aligned."))
		}
		document.Number = number
		document.HasNumber = true
	} else if hasField(document.Fields, "arc") {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["arc"], 1, "field \"arc\" must be an integer", "Use a numeric ARC identifier."))
	}

	document.Title = stringField(document, "title")
	document.Description = stringField(document, "description")
	document.Status = stringField(document, "status")
	document.Type = stringField(document, "type")
	document.Category = stringField(document, "category")
	document.SubCategory = stringField(document, "sub-category")
	document.Sponsor = stringField(document, "sponsor")
	document.ImplementationURL = stringField(document, "implementation-url")
	document.ImplementationMaintainer = strings.Join(stringSequenceField(document, "implementation-maintainer", false), ", ")
	document.AdoptionSummary = stringField(document, "adoption-summary")
	document.LastCallDeadline = stringField(document, "last-call-deadline")
	document.IdleSince = stringField(document, "idle-since")
	document.Requires = intSequenceField(document, "requires")
	document.Supersedes = intSequenceField(document, "supersedes")
	document.Extends = intSequenceField(document, "extends")
	document.ExtendedBy = intSequenceField(document, "extended-by")
	if value, ok := intField(document, "superseded-by"); ok {
		document.SupersededBy = []int{value}
	} else if hasField(document.Fields, "superseded-by") {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["superseded-by"], 1, "field \"superseded-by\" must be an integer ARC number", "Use a single numeric ARC identifier."))
	}

	if value, ok := boolField(document, "implementation-required"); ok {
		document.ImplementationRequired = value
	} else if hasField(document.Fields, "implementation-required") {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["implementation-required"], 1, "field \"implementation-required\" must be true or false", "Use a YAML boolean value."))
	}

	for _, name := range []string{"created", "last-call-deadline", "idle-since"} {
		value := stringField(document, name)
		if value == "" {
			continue
		}
		if !datePattern.MatchString(value) {
			diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines[name], 1, fmt.Sprintf("field %q must use YYYY-MM-DD", name), "Use an ISO date in YYYY-MM-DD format."))
		}
	}
	if hasField(document.Fields, "updated") {
		values := stringSequenceField(document, "updated", true)
		for _, value := range values {
			if !datePattern.MatchString(value) {
				diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["updated"], 1, "field \"updated\" must use YYYY-MM-DD", "Use ISO dates in YYYY-MM-DD format inside the YAML sequence."))
				break
			}
		}
	}

	if document.Title != "" && includesARCNumber(document.Title, document.Number) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["title"], 1, "title must not include the ARC number", "Remove the ARC number from the title."))
	}
	if document.Description != "" && includesARCNumber(document.Description, document.Number) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["description"], 1, "description must not include the ARC number", "Remove the ARC number from the description."))
	}

	if document.Status != "" && !IsValidStatus(document.Status) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["status"], 1, fmt.Sprintf("unsupported ARC status %q", document.Status), "Use one of the supported ARC status values."))
	}
	if document.Type != "" && !slices.Contains([]string{"Standards Track", "Meta"}, document.Type) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["type"], 1, fmt.Sprintf("unsupported ARC type %q", document.Type), "Use either \"Standards Track\" or \"Meta\"."))
	}
	if document.Sponsor != "" && !slices.Contains([]string{"Foundation", "Ecosystem"}, document.Sponsor) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["sponsor"], 1, fmt.Sprintf("unsupported sponsor %q", document.Sponsor), "Use either \"Foundation\" or \"Ecosystem\"."))
	}
	if document.AdoptionSummary != "" {
		if strings.HasPrefix(document.AdoptionSummary, "/") || !strings.HasPrefix(filepath.ToSlash(document.AdoptionSummary), "adoption/") {
			diagnostics = append(diagnostics, diag.NewWithHint("R:013", diag.OriginNative, document.Path, document.FieldLines["adoption-summary"], 1, "adoption-summary must be a relative path under adoption/", "Use a path like adoption/arc-0042.yaml."))
		}
	}
	if RequiresAdoptionSummary(document.Status) && document.AdoptionSummary == "" {
		diagnostics = append(diagnostics, diag.NewWithHint("R:012", diag.OriginNative, document.Path, document.FieldLines["status"], 1, fmt.Sprintf("status %q requires an adoption summary", document.Status), "Set adoption-summary to the matching adoption/arc-####.yaml file and add the file."))
	}
	if document.Status == "Last Call" && document.LastCallDeadline == "" {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["status"], 1, "status \"Last Call\" requires last-call-deadline", "Add last-call-deadline in YYYY-MM-DD format."))
	}
	if document.Status == "Idle" && document.IdleSince == "" {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["status"], 1, "status \"Idle\" requires idle-since", "Add idle-since in YYYY-MM-DD format."))
	}
	if document.ImplementationRequired && RequiresImplementationDeclaration(document.Status) {
		if document.ImplementationURL == "" {
			diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["implementation-required"], 1, fmt.Sprintf("status %q with implementation-required true requires implementation-url", document.Status), "Declare the canonical reference implementation repository in ARC front matter."))
		}
		if document.ImplementationMaintainer == "" {
			diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["implementation-required"], 1, fmt.Sprintf("status %q with implementation-required true requires implementation-maintainer", document.Status), "Declare the canonical implementation maintainers in ARC front matter."))
		}
	}

	diagnostics = append(diagnostics, validateCanonicalYAMLFieldShapes(document)...)

	for _, section := range requiredSections {
		if _, ok := document.Sections[section]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:008", diag.OriginNative, document.Path, 1, 1, fmt.Sprintf("missing required section %q", section), "Add the missing level-2 section to the ARC."))
		}
	}

	diagnostics = append(diagnostics, ValidateLinks(document, repoRoot)...)
	return diagnostics
}

func validateCanonicalYAMLFieldShapes(document *Document) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)

	for field := range canonicalStringSequenceFields {
		value, ok := document.Fields[field]
		if !ok {
			continue
		}
		if isCanonicalStringSequence(value, field == "updated") {
			continue
		}
		diagnostics = append(diagnostics, diag.NewWithHint("R:021", diag.OriginNative, document.Path, document.FieldLines[field], 1, fmt.Sprintf("field %q must use a YAML sequence", field), canonicalShapeHint(field)))
	}

	for field := range canonicalIntSequenceFields {
		value, ok := document.Fields[field]
		if !ok {
			continue
		}
		if isCanonicalIntSequence(value) {
			continue
		}
		diagnostics = append(diagnostics, diag.NewWithHint("R:021", diag.OriginNative, document.Path, document.FieldLines[field], 1, fmt.Sprintf("field %q must use a YAML sequence of ARC numbers", field), canonicalShapeHint(field)))
	}

	return diagnostics
}

func canonicalShapeHint(field string) string {
	switch field {
	case "author":
		return "Use a YAML sequence, for example:\nauthor:\n  - Example Author (@example)"
	case "updated":
		return "Use a YAML sequence of dates, for example:\nupdated:\n  - 2026-04-09"
	case "implementation-maintainer":
		return "Use a YAML sequence, for example:\nimplementation-maintainer:\n  - algorandfoundation"
	default:
		return fmt.Sprintf("Use a YAML sequence, for example:\n%s:\n  - 42", field)
	}
}

func isCanonicalStringSequence(value any, allowDates bool) bool {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return false
	}
	for _, item := range items {
		switch typed := item.(type) {
		case string:
			if strings.TrimSpace(typed) == "" {
				return false
			}
		case time.Time:
			if !allowDates {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func isCanonicalIntSequence(value any) bool {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return false
	}
	for _, item := range items {
		switch item.(type) {
		case int, int64, float64:
		default:
			return false
		}
	}
	return true
}

func ValidateLinks(document *Document, repoRoot string) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)
	root := repoRoot
	if root == "" {
		root = FindRepoRoot(filepath.Dir(document.Path))
	}
	root = filepath.Clean(root)

	for _, link := range document.Links {
		destination := strings.TrimSpace(link.Destination)
		if destination == "" || strings.HasPrefix(destination, "#") {
			continue
		}
		parsed, err := url.Parse(destination)
		if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") {
			document.ExternalLinks = append(document.ExternalLinks, destination)
			continue
		}
		if parsed != nil && parsed.Scheme != "" {
			continue
		}
		if strings.HasPrefix(destination, "/") {
			diagnostics = append(diagnostics, diag.NewWithHint("R:009", diag.OriginNative, document.Path, link.Line, 1, fmt.Sprintf("repo-local link %q must be relative, not root-relative", destination), "Use a relative path from the ARC file to the target."))
			continue
		}
		target := destination
		if hash := strings.Index(target, "#"); hash >= 0 {
			target = target[:hash]
		}
		if target == "" {
			continue
		}
		resolved := filepath.Clean(filepath.Join(filepath.Dir(document.Path), filepath.FromSlash(target)))
		if !withinRoot(root, resolved) {
			diagnostics = append(diagnostics, diag.NewWithHint("R:009", diag.OriginNative, document.Path, link.Line, 1, fmt.Sprintf("link %q resolves outside the repository root", destination), "Keep repo-local links inside the repository."))
			continue
		}
		if _, err := os.Stat(resolved); err != nil {
			diagnostics = append(diagnostics, diag.NewWithHint("R:009", diag.OriginNative, document.Path, link.Line, 1, fmt.Sprintf("link target %q does not exist", destination), "Create the target file or update the link."))
			continue
		}
		relative := filepath.ToSlash(strings.TrimPrefix(resolved, root+string(os.PathSeparator)))
		if matched, _ := regexp.MatchString(`(^|/)arc-\d{4}\.md$`, filepath.ToSlash(target)); matched && !strings.HasPrefix(relative, "ARCs/arc-") {
			diagnostics = append(diagnostics, diag.NewWithHint("R:009", diag.OriginNative, document.Path, link.Line, 1, fmt.Sprintf("ARC link %q must resolve under ARCs/", destination), "Target ARC files under ARCs/arc-####.md."))
		}
		if assetDirPattern.MatchString(filepath.ToSlash(relative)) {
			expected := fmt.Sprintf("assets/arc-%04d/", document.Number)
			if !strings.HasPrefix(relative+"/", expected) {
				diagnostics = append(diagnostics, diag.NewWithHint("R:010", diag.OriginNative, document.Path, link.Line, 1, fmt.Sprintf("asset link %q escapes the ARC asset subtree", destination), "Keep ARC asset links inside the matching assets/arc-####/ directory."))
			}
		}
	}

	if document.AdoptionSummary != "" && RequiresAdoptionSummary(document.Status) {
		resolved := filepath.Clean(filepath.Join(root, filepath.FromSlash(document.AdoptionSummary)))
		if !withinRoot(root, resolved) {
			diagnostics = append(diagnostics, diag.NewWithHint("R:013", diag.OriginNative, document.Path, document.FieldLines["adoption-summary"], 1, "adoption-summary resolves outside the repository root", "Use a relative path under adoption/."))
		} else if _, err := os.Stat(resolved); err != nil {
			diagnostics = append(diagnostics, diag.NewWithHint("R:013", diag.OriginNative, document.Path, document.FieldLines["adoption-summary"], 1, "required adoption-summary target does not exist", "Create the adoption summary file or update the path."))
		}
	}

	return diagnostics
}

func FindRepoRoot(start string) string {
	current := filepath.Clean(start)
	for {
		if current == string(filepath.Separator) || current == "" {
			break
		}
		if dirExists(filepath.Join(current, "ARCs")) || pathExists(filepath.Join(current, ".arckit.jsonc")) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return filepath.Clean(start)
}

func RequiresAdoptionSummary(status string) bool {
	switch status {
	case "Last Call", "Final", "Idle", "Deprecated":
		return true
	default:
		return false
	}
}

func RequiresImplementationDeclaration(status string) bool {
	switch status {
	case "Review", "Last Call", "Final", "Idle", "Deprecated":
		return true
	default:
		return false
	}
}

func IsValidStatus(status string) bool {
	switch status {
	case "Draft", "Review", "Last Call", "Final", "Stagnant", "Withdrawn", "Idle", "Deprecated", "Living":
		return true
	default:
		return false
	}
}

func OrderedFields() []string {
	out := make([]string, len(fieldOrder))
	copy(out, fieldOrder)
	return out
}

func splitFrontMatter(path string, content []byte) ([]byte, []byte, int, []diag.Diagnostic) {
	if !bytes.HasPrefix(content, []byte("---")) {
		return nil, content, 1, []diag.Diagnostic{
			diag.NewWithHint("R:001", diag.OriginNative, path, 1, 1, "ARC files must start with a YAML front matter block", "Add a front matter block delimited by --- at the top of the file."),
		}
	}
	lines := bytes.Split(content, []byte("\n"))
	if len(lines) == 0 || strings.TrimSpace(string(lines[0])) != "---" {
		return nil, content, 1, []diag.Diagnostic{
			diag.NewWithHint("R:001", diag.OriginNative, path, 1, 1, "front matter must start with a line containing only ---", "Start the file with --- on its own line."),
		}
	}

	offset := len(lines[0]) + 1
	for index := 1; index < len(lines); index++ {
		line := strings.TrimSpace(string(lines[index]))
		if line == "---" {
			frontMatter := content[len(lines[0])+1 : offset-1]
			body := content[offset+len(lines[index]):]
			body = bytes.TrimPrefix(body, []byte("\n"))
			return frontMatter, body, index + 2, nil
		}
		offset += len(lines[index]) + 1
	}
	return nil, content, 1, []diag.Diagnostic{
		diag.NewWithHint("R:001", diag.OriginNative, path, 1, 1, "front matter is not closed with a second --- delimiter", "Terminate the front matter block with --- on its own line."),
	}
}

func frontMatterBlankLineDiagnostics(path string, frontMatter []byte) []diag.Diagnostic {
	lines := strings.Split(string(frontMatter), "\n")
	diagnostics := make([]diag.Diagnostic, 0)
	for index, line := range lines {
		if strings.TrimSpace(line) != "" {
			continue
		}
		diagnostics = append(diagnostics, diag.NewWithHint("R:024", diag.OriginNative, path, index+2, 1, "front matter must not contain empty lines", "Remove empty lines from the front matter block or run arckit fmt."))
	}
	return diagnostics
}

func parseMarkdown(document *Document) {
	parser := goldmark.New()
	reader := text.NewReader(document.Body)
	tree := parser.Parser().Parse(reader)
	source := document.Body
	bodyStartLine := document.BodyStartLine
	if bodyStartLine == 0 {
		bodyStartLine = 1
	}
	_ = ast.Walk(tree, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *ast.Heading:
			if typed.Level != 2 {
				return ast.WalkContinue, nil
			}
			title := strings.TrimSpace(plainText(source, typed))
			document.Sections[title] = nodeLine(source, typed, bodyStartLine)
		case *ast.Link:
			document.Links = append(document.Links, Link{
				Destination: string(typed.Destination),
				Line:        nodeLine(source, typed, bodyStartLine),
			})
		}
		return ast.WalkContinue, nil
	})
}

func plainText(source []byte, node ast.Node) string {
	builder := strings.Builder{}
	_ = ast.Walk(node, func(current ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := current.(type) {
		case *ast.Text:
			builder.Write(typed.Segment.Value(source))
		case *ast.String:
			builder.Write(typed.Value)
		}
		return ast.WalkContinue, nil
	})
	return builder.String()
}

func nodeLine(source []byte, node ast.Node, baseLine int) int {
	if lines := safeLines(node); lines != nil && lines.Len() > 0 {
		return baseLine + bytes.Count(source[:lines.At(0).Start], []byte("\n"))
	}
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if line := nodeLine(source, child, baseLine); line > 0 {
			return line
		}
	}
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if lines := safeLines(parent); lines != nil && lines.Len() > 0 {
			return baseLine + bytes.Count(source[:lines.At(0).Start], []byte("\n"))
		}
	}
	return baseLine
}

func includesARCNumber(value string, number int) bool {
	lower := strings.ToLower(value)
	candidates := []string{
		fmt.Sprintf("arc-%04d", number),
		fmt.Sprintf("arc %d", number),
		fmt.Sprintf("arc%d", number),
	}
	for _, candidate := range candidates {
		if strings.Contains(lower, candidate) {
			return true
		}
	}
	return false
}

func hasField(values map[string]any, key string) bool {
	_, ok := values[key]
	return ok
}

func hasValue(values map[string]any, key string) bool {
	value, ok := values[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) != ""
	default:
		return true
	}
}

func stringField(document *Document, key string) string {
	value, ok := document.Fields[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case time.Time:
		return typed.Format("2006-01-02")
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func intField(document *Document, key string) (int, bool) {
	value, ok := document.Fields[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	}
	return 0, false
}

func boolField(document *Document, key string) (bool, bool) {
	value, ok := document.Fields[key]
	if !ok {
		return false, false
	}
	typed, ok := value.(bool)
	return typed, ok
}

func intSequenceField(document *Document, key string) []int {
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
			}
		}
		return out
	default:
		return nil
	}
}

func stringSequenceField(document *Document, key string, allowDates bool) []string {
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
				if !allowDates {
					return nil
				}
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

func withinRoot(root, path string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	if path == root {
		return true
	}
	return strings.HasPrefix(path, root+string(os.PathSeparator))
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func safeLines(node ast.Node) *text.Segments {
	defer func() {
		_ = recover()
	}()
	return node.Lines()
}
