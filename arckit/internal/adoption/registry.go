package adoption

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
	"gopkg.in/yaml.v3"
)

const VettedAdoptersFileName = "vetted-adopters.yaml"

var vettedAdopterNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type VettedAdopters struct {
	Path           string   `json:"path"`
	Wallets        []string `yaml:"wallets" json:"wallets"`
	Explorers      []string `yaml:"explorers" json:"explorers"`
	Tooling        []string `yaml:"tooling" json:"tooling"`
	Infra          []string `yaml:"infra" json:"infra"`
	DappsProtocols []string `yaml:"dapps-protocols" json:"dapps-protocols"`
	keys           map[string]struct{}
}

func RegistryPath(root string) string {
	return filepath.Join(filepath.Clean(root), "adoption", VettedAdoptersFileName)
}

func LoadVettedAdopters(path string) (*VettedAdopters, []diag.Diagnostic, error) {
	registry := &VettedAdopters{
		Path: filepath.Clean(path),
		keys: map[string]struct{}{},
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, []diag.Diagnostic{
				diag.NewWithHint("R:022", diag.OriginNative, registry.Path, 1, 1, "required vetted adopters registry does not exist", "Create adoption/vetted-adopters.yaml with the required adopter categories."),
			}, nil
		}
		return nil, []diag.Diagnostic{
			diag.NewWithHint("R:027", diag.OriginNative, registry.Path, 0, 0, err.Error(), "Check the file path and permissions, then retry."),
		}, err
	}

	root := yaml.Node{}
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, []diag.Diagnostic{
			diag.NewWithHint("R:022", diag.OriginNative, registry.Path, 1, 1, err.Error(), "Fix the YAML syntax in adoption/vetted-adopters.yaml."),
		}, nil
	}
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return nil, []diag.Diagnostic{
			diag.NewWithHint("R:022", diag.OriginNative, registry.Path, 1, 1, "vetted adopters registry must decode to one top-level YAML mapping", "Use a mapping with the required adopter categories."),
		}, nil
	}
	mapping := root.Content[0]
	for index := 0; index < len(mapping.Content); index += 2 {
		registry.keys[mapping.Content[index].Value] = struct{}{}
	}

	if err := yaml.Unmarshal(content, registry); err != nil {
		return nil, []diag.Diagnostic{
			diag.NewWithHint("R:022", diag.OriginNative, registry.Path, 1, 1, err.Error(), "Use YAML sequences of lower-kebab adopter names for each required category."),
		}, nil
	}

	return registry, nil, nil
}

func ValidateVettedAdopters(registry *VettedAdopters) []diag.Diagnostic {
	if registry == nil {
		return nil
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, key := range CategoryNames() {
		if _, ok := registry.keys[key]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:022", diag.OriginNative, registry.Path, 1, 1, fmt.Sprintf("missing required vetted adopter category %q", key), "Add the missing adopter category as a YAML sequence."))
		}
	}
	for key := range registry.keys {
		if !IsKnownCategory(key) {
			diagnostics = append(diagnostics, diag.NewWithHint("R:022", diag.OriginNative, registry.Path, 1, 1, fmt.Sprintf("unsupported vetted adopter category %q", key), "Use only wallets, explorers, tooling, infra, and dapps-protocols."))
		}
	}

	for _, group := range registry.Categories() {
		seen := map[string]struct{}{}
		for index, name := range group.Entries {
			if !isVettedAdopterName(name) {
				diagnostics = append(diagnostics, diag.NewWithHint("R:022", diag.OriginNative, registry.Path, 1, 1, fmt.Sprintf("%s[%d] must be lower-kebab-case, got %q", group.Name, index, name), "Use a lower-kebab identifier such as example-wallet."))
				continue
			}
			if _, ok := seen[name]; ok {
				diagnostics = append(diagnostics, diag.NewWithHint("R:022", diag.OriginNative, registry.Path, 1, 1, fmt.Sprintf("duplicate vetted adopter %q in %s", name, group.Name), "Keep each vetted adopter listed only once per category."))
				continue
			}
			seen[name] = struct{}{}
		}
	}

	return diagnostics
}

func (registry *VettedAdopters) Contains(group string, name string) bool {
	if registry == nil {
		return false
	}
	return slices.Contains(registry.Entries(group), name)
}

func isVettedAdopterName(name string) bool {
	return vettedAdopterNamePattern.MatchString(name)
}
