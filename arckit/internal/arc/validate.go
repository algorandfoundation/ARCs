package arc

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

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
	diagnostics = append(diagnostics, ValidateCategoryMetadata(document.Path, document.FieldLines["category"], document.FieldLines["sub-category"], document.Category, document.SubCategory)...)
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
	if document.ImplementationRequired && document.ImplementationURL != "" {
		if expectedURL, ok := ExpectedImplementationURL(document); ok && strings.TrimSpace(document.ImplementationURL) != expectedURL {
			diagnostics = append(diagnostics, diag.NewWithHint("R:029", diag.OriginNative, document.Path, document.FieldLines["implementation-url"], 1, fmt.Sprintf("implementation-url must be %s when sponsor is %q and arc is %d", expectedURL, document.Sponsor, document.Number), fmt.Sprintf("Update implementation-url to %s.", expectedURL)))
		}
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
		relative, err := relativeToRoot(root, resolved)
		if err != nil {
			diagnostics = append(diagnostics, diag.NewWithHint("R:009", diag.OriginNative, document.Path, link.Line, 1, fmt.Sprintf("link %q resolves outside the repository root", destination), "Keep repo-local links inside the repository."))
			continue
		}
		relative = filepath.ToSlash(relative)
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

func ExpectedImplementationURL(document *Document) (string, bool) {
	if document == nil || !document.HasNumber {
		return "", false
	}
	switch document.Sponsor {
	case "Foundation":
		return fmt.Sprintf("https://github.com/algorandfoundation/arc%d", document.Number), true
	case "Ecosystem":
		return fmt.Sprintf("https://github.com/algorandecosystem/arc%d", document.Number), true
	default:
		return "", false
	}
}

func ValidateCategoryMetadata(path string, categoryLine int, subCategoryLine int, category string, subCategory string) []diag.Diagnostic {
	category = strings.TrimSpace(category)
	subCategory = strings.TrimSpace(subCategory)
	diagnostics := make([]diag.Diagnostic, 0)

	if categoryLine <= 0 {
		categoryLine = 1
	}
	if subCategoryLine <= 0 {
		subCategoryLine = 1
	}

	if category != "" && !IsValidCategory(category) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:030", diag.OriginNative, path, categoryLine, 1, fmt.Sprintf("unsupported category %q", category), fmt.Sprintf("Use one of %s.", quotedList(validCategories))))
	}
	if subCategory != "" && !IsValidSubCategory(subCategory) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:030", diag.OriginNative, path, subCategoryLine, 1, fmt.Sprintf("unsupported sub-category %q", subCategory), fmt.Sprintf("Use one of %s.", quotedList(validSubCategories))))
	}
	if subCategory != "" && category == "" {
		diagnostics = append(diagnostics, diag.NewWithHint("R:030", diag.OriginNative, path, subCategoryLine, 1, "sub-category requires category", fmt.Sprintf("Add category with one of %s, or remove sub-category.", quotedList(validCategories))))
	}

	return diagnostics
}

func quotedList(values []string) string {
	if len(values) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}
	if len(quoted) == 1 {
		return quoted[0]
	}
	if len(quoted) == 2 {
		return quoted[0] + " or " + quoted[1]
	}
	return strings.Join(quoted[:len(quoted)-1], ", ") + ", or " + quoted[len(quoted)-1]
}

func IsValidStatus(status string) bool {
	switch status {
	case "Draft", "Review", "Last Call", "Final", "Stagnant", "Withdrawn", "Idle", "Deprecated", "Living":
		return true
	default:
		return false
	}
}
