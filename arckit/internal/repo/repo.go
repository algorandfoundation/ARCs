package repo

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/algorandfoundation/ARCs/arckit/internal/adoption"
	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/config"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

var assetPattern = regexp.MustCompile(`^arc-(\d{4})$`)

type State struct {
	Root      string
	ARCs      map[int]*arc.Document
	Adoptions map[int]*adoption.Summary
}

func Validate(root string, cfg config.Config) (State, []diag.Diagnostic, error) {
	state := State{
		Root:      filepath.Clean(root),
		ARCs:      map[int]*arc.Document{},
		Adoptions: map[int]*adoption.Summary{},
	}
	diagnostics := make([]diag.Diagnostic, 0)
	arcFileNumbers := map[int]struct{}{}

	arcFiles, err := filepath.Glob(filepath.Join(state.Root, "ARCs", "arc-*.md"))
	if err != nil {
		return state, []diag.Diagnostic{
			diag.NewWithHint("R:027", diag.OriginNative, state.Root, 0, 0, err.Error(), "Check the repository root path and try again."),
		}, err
	}
	numberToPaths := map[int][]string{}
	for _, path := range arcFiles {
		if number, ok := config.ARCNumberForPath(path); ok {
			if cfg.IgnoreARC(number) {
				continue
			}
			arcFileNumbers[number] = struct{}{}
		}
		document, loadDiagnostics, loadErr := arc.Load(path)
		diagnostics = append(diagnostics, loadDiagnostics...)
		if loadErr != nil {
			continue
		}
		diagnostics = append(diagnostics, arc.Validate(document, state.Root)...)
		if document.HasNumber {
			numberToPaths[document.Number] = append(numberToPaths[document.Number], document.Path)
			if _, exists := state.ARCs[document.Number]; !exists {
				state.ARCs[document.Number] = document
			}
		}
	}
	for number, paths := range numberToPaths {
		if len(paths) > 1 {
			diagnostics = append(diagnostics, diag.NewWithHint("R:018", diag.OriginNative, paths[0], 1, 1, fmt.Sprintf("duplicate ARC number %d across %s", number, strings.Join(paths, ", ")), "Keep only one ARC file for each ARC number."))
		}
	}

	registry, registryDiagnostics, registryErr := adoption.LoadValidatedRegistry(state.Root)
	diagnostics = append(diagnostics, registryDiagnostics...)
	if registryErr != nil || len(registryDiagnostics) != 0 {
		registry = nil
	}

	adoptionFiles, err := filepath.Glob(filepath.Join(state.Root, "adoption", "arc-*.yaml"))
	if err != nil {
		return state, []diag.Diagnostic{
			diag.NewWithHint("R:027", diag.OriginNative, state.Root, 0, 0, err.Error(), "Check the repository root path and try again."),
		}, err
	}
	adoptionToPaths := map[int][]string{}
	for _, path := range adoptionFiles {
		if number, ok := config.ARCNumberForPath(path); ok && cfg.IgnoreARC(number) {
			continue
		}
		summary, loadDiagnostics, loadErr := adoption.Load(path)
		diagnostics = append(diagnostics, loadDiagnostics...)
		if loadErr != nil {
			continue
		}
		diagnostics = append(diagnostics, adoption.Validate(summary, state.ARCs[summary.Arc], registry)...)
		if summary.Arc != 0 {
			adoptionToPaths[summary.Arc] = append(adoptionToPaths[summary.Arc], summary.Path)
			if _, exists := state.Adoptions[summary.Arc]; !exists {
				state.Adoptions[summary.Arc] = summary
			}
		}
	}
	for number, paths := range adoptionToPaths {
		if len(paths) > 1 {
			diagnostics = append(diagnostics, diag.NewWithHint("R:018", diag.OriginNative, paths[0], 1, 1, fmt.Sprintf("duplicate adoption summary for ARC %d across %s", number, strings.Join(paths, ", ")), "Keep only one adoption summary file per ARC."))
		}
	}

	diagnostics = append(diagnostics, validateAssets(state.Root, arcFileNumbers, cfg)...)
	diagnostics = append(diagnostics, validateMappings(state, arcFileNumbers)...)
	diagnostics = append(diagnostics, validateRelationships(state)...)
	diagnostics = append(diagnostics, validateMaturity(state)...)

	diagnostics = cfg.FilterDiagnostics(diagnostics)
	diag.SortDiagnostics(diagnostics)
	return state, diagnostics, nil
}

func validateAssets(root string, arcFileNumbers map[int]struct{}, cfg config.Config) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)
	entries, err := os.ReadDir(filepath.Join(root, "assets"))
	if err != nil {
		if os.IsNotExist(err) {
			return diagnostics
		}
		return []diag.Diagnostic{
			diag.NewWithHint("R:027", diag.OriginNative, filepath.Join(root, "assets"), 0, 0, err.Error(), "Check the assets directory and permissions."),
		}
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		matches := assetPattern.FindStringSubmatch(entry.Name())
		if len(matches) != 2 {
			continue
		}
		number, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		if cfg.IgnoreARC(number) {
			continue
		}
		if _, ok := arcFileNumbers[number]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:018", diag.OriginNative, filepath.Join(root, "assets", entry.Name()), 1, 1, fmt.Sprintf("orphaned asset directory %q has no matching ARC", entry.Name()), "Remove the orphaned asset tree or add the matching ARC file."))
		}
	}
	return diagnostics
}

func validateMappings(state State, arcFileNumbers map[int]struct{}) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)
	for number, summary := range state.Adoptions {
		if _, ok := arcFileNumbers[number]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:018", diag.OriginNative, summary.Path, 1, 1, fmt.Sprintf("orphaned adoption summary for ARC %d", number), "Remove the orphaned adoption summary or add the matching ARC file."))
		}
	}
	for number, document := range state.ARCs {
		if arc.RequiresAdoptionSummary(document.Status) {
			expected := fmt.Sprintf("adoption/arc-%04d.yaml", number)
			if document.AdoptionSummary == "" {
				diagnostics = append(diagnostics, diag.NewWithHint("R:012", diag.OriginNative, document.Path, document.FieldLines["status"], 1, fmt.Sprintf("ARC %d requires an adoption summary", number), "Set adoption-summary to the matching adoption/arc-####.yaml file."))
				continue
			}
			if filepath.ToSlash(document.AdoptionSummary) != expected {
				diagnostics = append(diagnostics, diag.NewWithHint("R:013", diag.OriginNative, document.Path, document.FieldLines["adoption-summary"], 1, fmt.Sprintf("adoption-summary must point to %s", expected), "Point adoption-summary to the matching adoption file for this ARC."))
			}
			if _, ok := state.Adoptions[number]; !ok {
				diagnostics = append(diagnostics, diag.NewWithHint("R:013", diag.OriginNative, document.Path, document.FieldLines["adoption-summary"], 1, fmt.Sprintf("required adoption summary %s does not exist", expected), "Create the matching adoption summary file."))
			}
		}
	}
	return diagnostics
}

func validateRelationships(state State) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)
	for number, document := range state.ARCs {
		checkReciprocal := func(sourceField string, targets []int, targetField string) {
			for _, target := range targets {
				targetDocument, ok := state.ARCs[target]
				if !ok {
					continue
				}
				var inverse []int
				switch targetField {
				case "superseded-by":
					inverse = targetDocument.SupersededBy
				case "supersedes":
					inverse = targetDocument.Supersedes
				case "extended-by":
					inverse = targetDocument.ExtendedBy
				case "extends":
					inverse = targetDocument.Extends
				}
				if !slices.Contains(inverse, number) {
					diagnostics = append(diagnostics, diag.NewWithHint("R:011", diag.OriginNative, document.Path, document.FieldLines[sourceField], 1, fmt.Sprintf("%s points to ARC %d but %s is missing the reciprocal reference", sourceField, target, targetField), "Keep reciprocal ARC relationship fields aligned when both files exist."))
				}
			}
		}
		checkReciprocal("supersedes", document.Supersedes, "superseded-by")
		checkReciprocal("superseded-by", document.SupersededBy, "supersedes")
		checkReciprocal("extends", document.Extends, "extended-by")
		checkReciprocal("extended-by", document.ExtendedBy, "extends")
	}
	return diagnostics
}

func validateMaturity(state State) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)
	for _, document := range state.ARCs {
		currentTier, ok := maturityTier(document.Status)
		if !ok {
			continue
		}
		for _, target := range document.Requires {
			targetDocument, ok := state.ARCs[target]
			if !ok {
				continue
			}
			targetTier, ok := maturityTier(targetDocument.Status)
			if !ok || targetTier >= currentTier {
				continue
			}
			diagnostics = append(diagnostics, diag.NewWithHint("R:038", diag.OriginNative, document.Path, document.FieldLines["requires"], 1, fmt.Sprintf("requires target ARC-%d must be at least as mature as ARC-%d", target, document.Number), "Update requires to reference ARCs at the same or higher maturity tier, or adjust the current ARC status."))
		}
		for _, link := range document.Links {
			target, ok := linkedARCNumber(document, link.Destination)
			if !ok {
				continue
			}
			targetDocument, ok := state.ARCs[target]
			if !ok {
				continue
			}
			targetTier, ok := maturityTier(targetDocument.Status)
			if !ok || currentTier <= targetTier {
				continue
			}
			diagnostics = append(diagnostics, diag.NewWithHint("R:038", diag.OriginNative, document.Path, link.Line, 1, fmt.Sprintf("ARC body link to ARC-%d must not point to a less mature ARC", target), "Only link directly to ARCs at the same or higher maturity tier from body content."))
		}
	}
	return diagnostics
}

func linkedARCNumber(document *arc.Document, destination string) (int, bool) {
	destination = strings.TrimSpace(destination)
	if destination == "" || strings.HasPrefix(destination, "#") {
		return 0, false
	}
	parsed, err := url.Parse(destination)
	if err == nil && parsed.Scheme != "" {
		return 0, false
	}
	target := destination
	if hash := strings.Index(target, "#"); hash >= 0 {
		target = target[:hash]
	}
	if target == "" || strings.HasPrefix(target, "/") {
		return 0, false
	}
	resolved := filepath.Clean(filepath.Join(filepath.Dir(document.Path), filepath.FromSlash(target)))
	return config.ARCNumberForPath(resolved)
}

func maturityTier(status string) (int, bool) {
	switch strings.TrimSpace(status) {
	case "Draft", "Stagnant":
		return 1, true
	case "Review":
		return 2, true
	case "Last Call":
		return 3, true
	case "Final", "Withdrawn", "Living", "Deprecated", "Idle":
		return 4, true
	default:
		return 0, false
	}
}
