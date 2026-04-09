package adoption

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
	"gopkg.in/yaml.v3"
)

var adoptionPathPattern = regexp.MustCompile(`(^|.*/)adoption/arc-(\d{4})\.yaml$`)

type Actor struct {
	Name     string `yaml:"name" json:"name"`
	Status   string `yaml:"status" json:"status"`
	Evidence string `yaml:"evidence" json:"evidence"`
	Notes    string `yaml:"notes" json:"notes"`
}

type ReferenceImplementation struct {
	Repository  string   `yaml:"repository" json:"repository"`
	Maintainers []string `yaml:"maintainers" json:"maintainers"`
	Status      string   `yaml:"status" json:"status"`
	Notes       string   `yaml:"notes" json:"notes"`
}

type AdoptionSection struct {
	Wallets        []Actor `yaml:"wallets" json:"wallets"`
	Explorers      []Actor `yaml:"explorers" json:"explorers"`
	SDKLibraries   []Actor `yaml:"sdk-libraries" json:"sdk-libraries"`
	Infra          []Actor `yaml:"infra" json:"infra"`
	DappsProtocols []Actor `yaml:"dapps-protocols" json:"dapps-protocols"`
}

type SummarySection struct {
	AdoptionReadiness string   `yaml:"adoption-readiness" json:"adoption-readiness"`
	Blockers          []string `yaml:"blockers" json:"blockers"`
	Notes             string   `yaml:"notes" json:"notes"`
}

type Summary struct {
	Path                    string                   `json:"path"`
	Arc                     int                      `yaml:"arc" json:"arc"`
	Title                   string                   `yaml:"title" json:"title"`
	Status                  string                   `yaml:"status" json:"status"`
	LastReviewed            string                   `yaml:"last-reviewed" json:"last-reviewed"`
	Sponsor                 string                   `yaml:"sponsor" json:"sponsor"`
	ImplementationRequired  bool                     `yaml:"implementation-required" json:"implementation-required"`
	ReferenceImplementation *ReferenceImplementation `yaml:"reference-implementation" json:"reference-implementation,omitempty"`
	Adoption                AdoptionSection          `yaml:"adoption" json:"adoption"`
	Summary                 SummarySection           `yaml:"summary" json:"summary"`
	keys                    map[string]struct{}
}

func Load(path string) (*Summary, []diag.Diagnostic, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, []diag.Diagnostic{
			diag.NewWithHint("R:027", diag.OriginNative, path, 0, 0, err.Error(), "Check the file path and permissions, then retry."),
		}, err
	}

	summary := &Summary{Path: filepath.Clean(path), keys: map[string]struct{}{}}
	diagnostics := make([]diag.Diagnostic, 0)

	matches := adoptionPathPattern.FindStringSubmatch(filepath.ToSlash(summary.Path))
	expectedNumber := 0
	hasExpectedNumber := false
	if len(matches) != 3 {
		diagnostics = append(diagnostics, diag.NewWithHint("R:014", diag.OriginNative, summary.Path, 1, 1, "adoption summaries must live under adoption/arc-####.yaml", "Move or rename the file to adoption/arc-####.yaml."))
	} else {
		expectedNumber, _ = strconv.Atoi(matches[2])
		hasExpectedNumber = true
	}

	root := yaml.Node{}
	if err := yaml.Unmarshal(content, &root); err != nil {
		diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, err.Error(), "Fix the YAML syntax in the adoption summary."))
		return summary, diagnostics, nil
	}
	if len(root.Content) > 0 && root.Content[0].Kind == yaml.MappingNode {
		mapping := root.Content[0]
		for index := 0; index < len(mapping.Content); index += 2 {
			summary.keys[mapping.Content[index].Value] = struct{}{}
		}
	}
	if err := yaml.Unmarshal(content, summary); err != nil {
		diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, err.Error(), "Use the required adoption summary schema."))
		return summary, diagnostics, nil
	}

	if hasExpectedNumber && summary.Arc != 0 && expectedNumber != summary.Arc {
		diagnostics = append(diagnostics, diag.NewWithHint("R:017", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("filename ARC number %d does not match arc value %d", expectedNumber, summary.Arc), "Keep the filename and the arc field aligned."))
	}
	return summary, diagnostics, nil
}

func Validate(summary *Summary, document *arc.Document) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)

	for _, key := range []string{"arc", "title", "status", "last-reviewed", "sponsor", "implementation-required", "adoption", "summary"} {
		if _, ok := summary.keys[key]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:015", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("missing required field %q", key), "Add the missing top-level field to the adoption summary."))
		}
	}
	if summary.ImplementationRequired && summary.ReferenceImplementation == nil {
		diagnostics = append(diagnostics, diag.NewWithHint("R:015", diag.OriginNative, summary.Path, 1, 1, "reference-implementation is required when implementation-required is true", "Add the reference-implementation block."))
	}
	if summary.Status != "" && !arc.IsValidStatus(summary.Status) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("unsupported ARC status %q", summary.Status), "Use one of the supported ARC status values."))
	}
	if summary.LastReviewed != "" && !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(summary.LastReviewed) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, "last-reviewed must use YYYY-MM-DD", "Use an ISO date in YYYY-MM-DD format."))
	}
	if summary.Sponsor != "" && !slices.Contains([]string{"Foundation", "Ecosystem"}, summary.Sponsor) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("unsupported sponsor %q", summary.Sponsor), "Use either \"Foundation\" or \"Ecosystem\"."))
	}
	if summary.ReferenceImplementation != nil {
		if !slices.Contains([]string{"planned", "in_progress", "testable", "shipped", "archived"}, summary.ReferenceImplementation.Status) {
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("unsupported reference-implementation.status %q", summary.ReferenceImplementation.Status), "Use one of planned, in_progress, testable, shipped, or archived."))
		}
	}
	if summary.Summary.AdoptionReadiness != "" && !slices.Contains([]string{"low", "medium", "high"}, summary.Summary.AdoptionReadiness) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("unsupported summary.adoption-readiness %q", summary.Summary.AdoptionReadiness), "Use one of low, medium, or high."))
	}

	for group, actors := range summary.actorGroups() {
		for index, actor := range actors {
			if strings.TrimSpace(actor.Name) == "" {
				diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("%s[%d] is missing name", group, index), "Set a non-empty actor name."))
			}
			if !slices.Contains([]string{"planned", "in_progress", "shipped", "declined", "unknown"}, actor.Status) {
				diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("%s[%d] has unsupported status %q", group, index, actor.Status), "Use one of planned, in_progress, shipped, declined, or unknown."))
			}
		}
	}

	if document == nil {
		return diagnostics
	}

	if summary.Arc != 0 && document.HasNumber && summary.Arc != document.Number {
		diagnostics = append(diagnostics, diag.NewWithHint("R:017", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("arc %d does not match ARC file number %d", summary.Arc, document.Number), "Keep the ARC and adoption summary numbers aligned."))
	}
	if summary.Sponsor != "" && document.Sponsor != "" && summary.Sponsor != document.Sponsor {
		diagnostics = append(diagnostics, diag.NewWithHint("R:017", diag.OriginNative, summary.Path, 1, 1, "sponsor does not match the ARC file", "Update the ARC or adoption summary so the sponsor matches."))
	}
	if summary.ImplementationRequired != document.ImplementationRequired {
		diagnostics = append(diagnostics, diag.NewWithHint("R:017", diag.OriginNative, summary.Path, 1, 1, "implementation-required does not match the ARC file", "Update the ARC or adoption summary so implementation-required matches."))
	}
	if summary.ReferenceImplementation != nil && document.ImplementationURL != "" && summary.ReferenceImplementation.Repository != "" && summary.ReferenceImplementation.Repository != document.ImplementationURL {
		diagnostics = append(diagnostics, diag.NewWithHint("R:017", diag.OriginNative, summary.Path, 1, 1, "reference-implementation.repository does not match implementation-url in the ARC file", "Keep the reference implementation repository consistent across both files."))
	}
	return diagnostics
}

func (summary *Summary) HasAnyEvidence() bool {
	for _, actors := range summary.actorGroups() {
		for _, actor := range actors {
			if strings.TrimSpace(actor.Evidence) != "" {
				return true
			}
		}
	}
	return false
}

func (summary *Summary) HasActorEvidence() bool {
	return summary.HasAnyEvidence()
}

func (summary *Summary) actorGroups() map[string][]Actor {
	return map[string][]Actor{
		"wallets":         summary.Adoption.Wallets,
		"explorers":       summary.Adoption.Explorers,
		"sdk-libraries":   summary.Adoption.SDKLibraries,
		"infra":           summary.Adoption.Infra,
		"dapps-protocols": summary.Adoption.DappsProtocols,
	}
}
