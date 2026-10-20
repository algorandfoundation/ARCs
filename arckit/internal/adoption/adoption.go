package adoption

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
	"gopkg.in/yaml.v3"
)

var adoptionPathPattern = regexp.MustCompile(`(^|.*/)adoption/arc-(\d{4})\.yaml$`)
var reviewDatePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

type Actor struct {
	Name     string `yaml:"name" json:"name"`
	Status   string `yaml:"status" json:"status"`
	Evidence string `yaml:"evidence" json:"evidence"`
	Notes    string `yaml:"notes" json:"notes"`
}

type ReferenceImplementation struct {
	Status string `yaml:"status" json:"status"`
	Notes  string `yaml:"notes" json:"notes"`
}

type AdoptionSection struct {
	Wallets        []Actor `yaml:"wallets" json:"wallets"`
	Explorers      []Actor `yaml:"explorers" json:"explorers"`
	Tooling        []Actor `yaml:"tooling" json:"tooling"`
	Infra          []Actor `yaml:"infra" json:"infra"`
	DappsProtocols []Actor `yaml:"dapps-protocols" json:"dapps-protocols"`
}

type SummarySection struct {
	AdoptionReadiness string   `yaml:"adoption-readiness" json:"adoption-readiness"`
	Blockers          []string `yaml:"blockers" json:"blockers"`
	Notes             string   `yaml:"notes" json:"notes"`
}

type Summary struct {
	Path                        string                   `json:"path"`
	Arc                         int                      `yaml:"arc" json:"arc"`
	Title                       string                   `yaml:"title" json:"title"`
	LastReviewed                string                   `yaml:"last-reviewed" json:"last-reviewed"`
	ReferenceImplementation     *ReferenceImplementation `yaml:"reference-implementation" json:"reference-implementation,omitempty"`
	Adoption                    AdoptionSection          `yaml:"adoption" json:"adoption"`
	Summary                     SummarySection           `yaml:"summary" json:"summary"`
	keys                        map[string]struct{}
	adoptionKeys                map[string]struct{}
	referenceImplementationKeys map[string]struct{}
}

func Load(path string) (*Summary, []diag.Diagnostic, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, []diag.Diagnostic{
			diag.NewWithHint("R:027", diag.OriginNative, path, 0, 0, err.Error(), "Check the file path and permissions, then retry."),
		}, err
	}

	summary := &Summary{Path: filepath.Clean(path), keys: map[string]struct{}{}, adoptionKeys: map[string]struct{}{}, referenceImplementationKeys: map[string]struct{}{}}
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
		if adoptionNode := mappingValue(mapping, "adoption"); adoptionNode != nil && adoptionNode.Kind == yaml.MappingNode {
			summary.adoptionKeys = mappingKeys(adoptionNode)
		}
		if referenceImplementationNode := mappingValue(mapping, "reference-implementation"); referenceImplementationNode != nil && referenceImplementationNode.Kind == yaml.MappingNode {
			summary.referenceImplementationKeys = mappingKeys(referenceImplementationNode)
		}
	}
	if err := yaml.Unmarshal(content, summary); err != nil {
		if schemaDiagnostics := adoptionSchemaDiagnostics(summary.Path, &root); len(schemaDiagnostics) > 0 {
			diagnostics = append(diagnostics, schemaDiagnostics...)
		} else {
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, err.Error(), "Use the required adoption summary schema."))
		}
		return summary, diagnostics, nil
	}

	if hasExpectedNumber && summary.Arc != 0 && expectedNumber != summary.Arc {
		diagnostics = append(diagnostics, diag.NewWithHint("R:017", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("filename ARC number %d does not match arc value %d", expectedNumber, summary.Arc), "Keep the filename and the arc field aligned."))
	}
	return summary, diagnostics, nil
}

func Validate(summary *Summary, document *arc.Document, registry *VettedAdopters) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)

	for _, key := range []string{"arc", "title", "last-reviewed", "adoption", "summary"} {
		if _, ok := summary.keys[key]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:015", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("missing required field %q", key), "Add the missing top-level field to the adoption summary."))
		}
	}

	allowedTopLevelKeys := map[string]struct{}{
		"arc":                      {},
		"title":                    {},
		"last-reviewed":            {},
		"reference-implementation": {},
		"adoption":                 {},
		"summary":                  {},
	}
	for key := range summary.keys {
		if _, ok := allowedTopLevelKeys[key]; ok {
			continue
		}
		switch key {
		case "status":
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, "status is not allowed in adoption summaries", "Derive ARC status from the matching ARC front matter and remove status from this YAML file."))
		case "sponsor":
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, "sponsor is not allowed in adoption summaries", "Derive sponsor from the matching ARC front matter and remove sponsor from this YAML file."))
		case "implementation-required":
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, "implementation-required is not allowed in adoption summaries", "Derive implementation-required from the matching ARC front matter and remove implementation-required from this YAML file."))
		default:
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("unsupported top-level adoption field %q", key), "Keep only the supported top-level adoption summary fields defined in the specification."))
		}
	}
	if summary.LastReviewed != "" && !reviewDatePattern.MatchString(summary.LastReviewed) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, "last-reviewed must use YYYY-MM-DD", "Use an ISO date in YYYY-MM-DD format."))
	}
	if document != nil && document.ImplementationRequired && summary.ReferenceImplementation == nil {
		diagnostics = append(diagnostics, diag.NewWithHint("R:015", diag.OriginNative, summary.Path, 1, 1, "reference-implementation is required when the ARC front matter sets implementation-required to true", "Add the reference-implementation block or update the ARC front matter if a canonical implementation is not required."))
	}
	if document != nil && !document.ImplementationRequired && summary.ReferenceImplementation != nil {
		diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, "reference-implementation is only allowed when the ARC front matter sets implementation-required to true", "Remove the reference-implementation block or update the ARC front matter if the ARC requires a canonical implementation."))
	}
	if summary.ReferenceImplementation != nil {
		if _, ok := summary.referenceImplementationKeys["status"]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:015", diag.OriginNative, summary.Path, 1, 1, "reference-implementation.status is required", "Keep only status and optional notes under reference-implementation, and set status."))
		}
		for key := range summary.referenceImplementationKeys {
			if key == "status" || key == "notes" {
				continue
			}
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("reference-implementation.%s is not allowed in adoption summaries", key), "Declare the canonical implementation URL and maintainers in the ARC front matter, and keep only status with optional notes in the adoption summary."))
		}
		if !IsKnownReferenceImplementationStatus(summary.ReferenceImplementation.Status) {
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("unsupported reference-implementation.status %q", summary.ReferenceImplementation.Status), "Use one of planned, wip, shipped, or archived."))
		}
	}
	if summary.Summary.AdoptionReadiness != "" && !IsKnownReadiness(summary.Summary.AdoptionReadiness) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("unsupported summary.adoption-readiness %q", summary.Summary.AdoptionReadiness), "Use one of low, medium, or high."))
	} else {
		actorCount := summary.ActorCount()
		expectedReadiness := normalizedAdoptionReadinessForActorCount(actorCount)
		switch summary.Summary.AdoptionReadiness {
		case ReadinessLow:
			if actorCount >= 3 {
				diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("summary.adoption-readiness %q is too low for %d adopters across all categories; expected %q", summary.Summary.AdoptionReadiness, actorCount, expectedReadiness), "Raise adoption-readiness to match the tracked adopter count or remove adopter entries that should not be counted yet."))
			}
		case ReadinessMedium:
			if actorCount < 3 {
				diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("summary.adoption-readiness %q requires at least 3 adopters across all categories, found %d", summary.Summary.AdoptionReadiness, actorCount), "Lower adoption-readiness to low or add more adopter entries across the adoption categories."))
			} else if actorCount >= 5 {
				diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("summary.adoption-readiness %q is too low for %d adopters across all categories; expected %q", summary.Summary.AdoptionReadiness, actorCount, expectedReadiness), "Raise adoption-readiness to match the tracked adopter count or remove adopter entries that should not be counted yet."))
			}
		case ReadinessHigh:
			if actorCount < 5 {
				diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("summary.adoption-readiness %q requires at least 5 adopters across all categories, found %d", summary.Summary.AdoptionReadiness, actorCount), "Lower adoption-readiness or add more adopter entries across the adoption categories."))
			}
		}
	}
	for _, key := range CategoryNames() {
		if _, ok := summary.adoptionKeys[key]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:015", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("adoption.%s is required", key), "Define all canonical adoption categories, even when they are empty lists."))
		}
	}
	for key := range summary.adoptionKeys {
		if !IsKnownCategory(key) {
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("unsupported adoption category %q", key), "Use only wallets, explorers, tooling, infra, and dapps-protocols under adoption."))
		}
	}

	for _, group := range summary.ActorCategories() {
		for index, actor := range group.Actors {
			if strings.TrimSpace(actor.Name) == "" {
				diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("%s[%d] is missing name", group.Name, index), "Set a non-empty actor name."))
			} else if !isVettedAdopterName(actor.Name) {
				diagnostics = append(diagnostics, diag.NewWithHint("R:023", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("%s[%d].name must be lower-kebab-case, got %q", group.Name, index, actor.Name), "Use a lower-kebab adopter name from adoption/vetted-adopters.yaml."))
			} else if registry != nil && !registry.Contains(group.Name, actor.Name) {
				diagnostics = append(diagnostics, diag.NewWithHint("R:023", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("%s[%d].name %q is not present in vetted adopters category %q", group.Name, index, actor.Name, group.Name), "Add the adopter to adoption/vetted-adopters.yaml or use an existing vetted adopter name."))
			}
			if !IsKnownActorStatus(actor.Status) {
				diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("%s[%d] has unsupported status %q", group.Name, index, actor.Status), "Use one of planned, in_progress, shipped, declined, or unknown."))
			}
		}
	}

	if document != nil && strings.TrimSpace(document.Status) == "Final" && !summary.HasAnyActors() {
		diagnostics = append(diagnostics, diag.NewWithHint("R:025", diag.OriginNative, summary.Path, 1, 1, "Final ARC adoption summaries must include at least one adopter entry", "Add at least one vetted adopter to an adoption category before marking the ARC Final."))
	}

	if document == nil {
		return diagnostics
	}

	if summary.Arc != 0 && document.HasNumber && summary.Arc != document.Number {
		diagnostics = append(diagnostics, diag.NewWithHint("R:017", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("arc %d does not match ARC file number %d", summary.Arc, document.Number), "Keep the ARC and adoption summary numbers aligned."))
	}
	if strings.TrimSpace(summary.Title) != "" && strings.TrimSpace(document.Title) != "" && strings.TrimSpace(summary.Title) != strings.TrimSpace(document.Title) {
		diagnostics = append(diagnostics, diag.NewWithHint("R:017", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("adoption title %q does not match ARC title %q", summary.Title, document.Title), "Keep the adoption summary title aligned with the matching ARC title."))
	}
	return diagnostics
}

func (summary *Summary) HasAnyEvidence() bool {
	for _, group := range summary.ActorCategories() {
		for _, actor := range group.Actors {
			if strings.TrimSpace(actor.Evidence) != "" {
				return true
			}
		}
	}
	return false
}

func (summary *Summary) HasAnyActors() bool {
	return summary.ActorCount() > 0
}

func (summary *Summary) ActorCount() int {
	count := 0
	for _, group := range summary.ActorCategories() {
		count += len(group.Actors)
	}
	return count
}

func (summary *Summary) NormalizedAdoptionReadiness() string {
	return normalizedAdoptionReadinessForActorCount(summary.ActorCount())
}

func normalizedAdoptionReadinessForActorCount(actorCount int) string {
	switch {
	case actorCount >= 5:
		return ReadinessHigh
	case actorCount >= 3:
		return ReadinessMedium
	default:
		return ReadinessLow
	}
}

func (summary *Summary) KeySet() map[string]struct{} {
	if summary == nil {
		return map[string]struct{}{}
	}
	return summary.keys
}

func (summary *Summary) ReferenceImplementationKeySet() map[string]struct{} {
	if summary == nil {
		return map[string]struct{}{}
	}
	return summary.referenceImplementationKeys
}

func (summary *Summary) AdoptionKeySet() map[string]struct{} {
	if summary == nil {
		return map[string]struct{}{}
	}
	return summary.adoptionKeys
}

func adoptionSchemaDiagnostics(path string, root *yaml.Node) []diag.Diagnostic {
	if root == nil || len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return nil
	}

	document := root.Content[0]
	referenceImplementationNode := mappingValue(document, "reference-implementation")
	if referenceImplementationNode != nil && referenceImplementationNode.Kind != yaml.MappingNode {
		return []diag.Diagnostic{
			diag.NewWithHint("R:016", diag.OriginNative, path, referenceImplementationNode.Line, referenceImplementationNode.Column, fmt.Sprintf("reference-implementation must be a mapping, got %s", yamlNodeType(referenceImplementationNode)), "Define reference-implementation as a mapping with a status field and optional notes field."),
		}
	}
	adoptionNode := mappingValue(document, "adoption")
	if adoptionNode == nil {
		return nil
	}
	if adoptionNode.Kind != yaml.MappingNode {
		return []diag.Diagnostic{
			diag.NewWithHint("R:016", diag.OriginNative, path, adoptionNode.Line, adoptionNode.Column, fmt.Sprintf("adoption must be a mapping of categories, got %s", yamlNodeType(adoptionNode)), "Define adoption as a mapping with wallets, explorers, tooling, infra, and dapps-protocols keys."),
		}
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, category := range CategoryNames() {
		categoryNode := mappingValue(adoptionNode, category)
		if categoryNode == nil {
			continue
		}
		if categoryNode.Kind != yaml.SequenceNode {
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, path, categoryNode.Line, categoryNode.Column, fmt.Sprintf("adoption.%s must be a sequence of actor objects, got %s", category, yamlNodeType(categoryNode)), "Use a YAML list of actor objects for this adoption category."))
			continue
		}
		for index, item := range categoryNode.Content {
			if item.Kind == yaml.MappingNode {
				continue
			}
			diagnostics = append(diagnostics, diag.NewWithHint("R:016", diag.OriginNative, path, item.Line, item.Column, fmt.Sprintf("adoption.%s[%d] must be an actor object with name, status, evidence, and notes fields, got %s", category, index, yamlNodeType(item)), "Replace the list item with an object entry such as \"- name: example-adopter\" plus status, evidence, and notes fields."))
		}
	}
	return diagnostics
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	return nil
}

func mappingKeys(mapping *yaml.Node) map[string]struct{} {
	keys := map[string]struct{}{}
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return keys
	}
	for index := 0; index < len(mapping.Content); index += 2 {
		keys[mapping.Content[index].Value] = struct{}{}
	}
	return keys
}

func yamlNodeType(node *yaml.Node) string {
	if node == nil {
		return "unknown value"
	}
	switch node.Kind {
	case yaml.MappingNode:
		return "mapping"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.AliasNode:
		return "alias"
	case yaml.ScalarNode:
		switch node.ShortTag() {
		case "!!str":
			return fmt.Sprintf("string %q", node.Value)
		case "!!int", "!!float":
			return fmt.Sprintf("number %q", node.Value)
		case "!!bool":
			return fmt.Sprintf("boolean %q", node.Value)
		case "!!null":
			return "null"
		default:
			if node.Value != "" {
				return fmt.Sprintf("scalar %q", node.Value)
			}
			return "scalar"
		}
	default:
		return "unknown value"
	}
}
