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

type Link struct {
	Destination string
	Line        int
}

type Document struct {
	Path                     string
	Raw                      []byte
	Body                     []byte
	Fields                   map[string]any
	FieldOrder               []string
	FieldLines               map[string]int
	Sections                 map[string]int
	Links                    []Link
	ExternalLinks            []string
	Number                   int
	Title                    string
	Description              string
	Status                   string
	Type                     string
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
		document.Number = number
	}

	frontMatter, body, parseDiagnostics := splitFrontMatter(document.Path, content)
	diagnostics = append(diagnostics, parseDiagnostics...)
	document.Body = body
	if len(frontMatter) == 0 {
		parseMarkdown(document)
		return document, diagnostics, nil
	}

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

	if number, ok := intField(document, "arc"); ok {
		if document.Number != 0 && number != document.Number {
			diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["arc"], 1, fmt.Sprintf("front matter arc value %d does not match filename number %d", number, document.Number), "Keep the filename and the arc field aligned."))
		}
		document.Number = number
	} else if hasField(document.Fields, "arc") {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["arc"], 1, "field \"arc\" must be an integer", "Use a numeric ARC identifier."))
	}

	document.Title = stringField(document, "title")
	document.Description = stringField(document, "description")
	document.Status = stringField(document, "status")
	document.Type = stringField(document, "type")
	document.Sponsor = stringField(document, "sponsor")
	document.ImplementationURL = stringField(document, "implementation-url")
	document.ImplementationMaintainer = stringField(document, "implementation-maintainer")
	document.AdoptionSummary = stringField(document, "adoption-summary")
	document.LastCallDeadline = stringField(document, "last-call-deadline")
	document.IdleSince = stringField(document, "idle-since")
	document.Requires = intListField(document, "requires")
	document.Supersedes = intListField(document, "supersedes")
	document.SupersededBy = intListField(document, "superseded-by")
	document.Extends = intListField(document, "extends")
	document.ExtendedBy = intListField(document, "extended-by")

	if value, ok := boolField(document, "implementation-required"); ok {
		document.ImplementationRequired = value
	} else if hasField(document.Fields, "implementation-required") {
		diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines["implementation-required"], 1, "field \"implementation-required\" must be true or false", "Use a YAML boolean value."))
	}

	for _, name := range []string{"created", "updated", "last-call-deadline", "idle-since"} {
		value := stringField(document, name)
		if value == "" {
			continue
		}
		if !datePattern.MatchString(value) {
			diagnostics = append(diagnostics, diag.NewWithHint("R:007", diag.OriginNative, document.Path, document.FieldLines[name], 1, fmt.Sprintf("field %q must use YYYY-MM-DD", name), "Use an ISO date in YYYY-MM-DD format."))
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

	for _, section := range requiredSections {
		if _, ok := document.Sections[section]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:008", diag.OriginNative, document.Path, 1, 1, fmt.Sprintf("missing required section %q", section), "Add the missing level-2 section to the ARC."))
		}
	}

	diagnostics = append(diagnostics, ValidateLinks(document, repoRoot)...)
	return diagnostics
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
		if current == "." || current == string(filepath.Separator) || current == "" {
			break
		}
		if exists(filepath.Join(current, "ARCs")) && exists(filepath.Join(current, "adoption")) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "."
}

func RequiresAdoptionSummary(status string) bool {
	switch status {
	case "Last Call", "Final", "Idle", "Deprecated":
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

func splitFrontMatter(path string, content []byte) ([]byte, []byte, []diag.Diagnostic) {
	if !bytes.HasPrefix(content, []byte("---")) {
		return nil, content, []diag.Diagnostic{
			diag.NewWithHint("R:001", diag.OriginNative, path, 1, 1, "ARC files must start with a YAML front matter block", "Add a front matter block delimited by --- at the top of the file."),
		}
	}
	lines := bytes.Split(content, []byte("\n"))
	if len(lines) == 0 || strings.TrimSpace(string(lines[0])) != "---" {
		return nil, content, []diag.Diagnostic{
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
			return frontMatter, body, nil
		}
		offset += len(lines[index]) + 1
	}
	return nil, content, []diag.Diagnostic{
		diag.NewWithHint("R:001", diag.OriginNative, path, 1, 1, "front matter is not closed with a second --- delimiter", "Terminate the front matter block with --- on its own line."),
	}
}

func parseMarkdown(document *Document) {
	parser := goldmark.New()
	reader := text.NewReader(document.Body)
	tree := parser.Parser().Parse(reader)
	source := document.Body
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
			document.Sections[title] = nodeLine(source, typed)
		case *ast.Link:
			document.Links = append(document.Links, Link{
				Destination: string(typed.Destination),
				Line:        nodeLine(source, typed),
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

func nodeLine(source []byte, node ast.Node) int {
	if lines := safeLines(node); lines != nil && lines.Len() > 0 {
		return bytes.Count(source[:lines.At(0).Start], []byte("\n")) + 1
	}
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if line := nodeLine(source, child); line > 0 {
			return line
		}
	}
	return 1
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

func intListField(document *Document, key string) []int {
	value, ok := document.Fields[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case int:
		return []int{typed}
	case int64:
		return []int{int(typed)}
	case float64:
		return []int{int(typed)}
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

func withinRoot(root, path string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	if path == root {
		return true
	}
	return strings.HasPrefix(path, root+string(os.PathSeparator))
}

func exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func safeLines(node ast.Node) *text.Segments {
	defer func() {
		_ = recover()
	}()
	return node.Lines()
}
