package arc

import (
	"regexp"
	"slices"
	"strings"
)

var (
	arcPathPattern  = regexp.MustCompile(`(^|.*/)ARCs/arc-(\d{4})\.md$`)
	datePattern     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	assetDirPattern = regexp.MustCompile(`(^|.*/)assets/arc-(\d{4})(/.*)?$`)
)

var validCategories = []string{"Interface", "Data", "Cryptography", "Protocol", "Governance"}

var validSubCategories = []string{"General", "ASA", "Application", "LSig", "Event", "Library", "Identity", "Explorer", "Wallet"}

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

func OrderedFields() []string {
	out := make([]string, len(fieldOrder))
	copy(out, fieldOrder)
	return out
}

func IsValidCategory(category string) bool {
	return slices.Contains(validCategories, strings.TrimSpace(category))
}

func IsValidSubCategory(subCategory string) bool {
	return slices.Contains(validSubCategories, strings.TrimSpace(subCategory))
}
