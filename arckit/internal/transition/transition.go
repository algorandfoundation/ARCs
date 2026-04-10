package transition

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/algorandfoundation/ARCs/arckit/internal/adoption"
	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

func Validate(path string, target string) ([]diag.Diagnostic, error) {
	document, diagnostics, err := arc.Load(path)
	if err != nil {
		return diagnostics, err
	}
	repoRoot := arc.FindRepoRoot(filepath.Dir(path))
	diagnostics = append(diagnostics, arc.Validate(document, repoRoot)...)

	require := func(condition bool, message string) {
		if !condition {
			diagnostics = append(diagnostics, diag.NewWithHint("R:019", diag.OriginNative, path, 1, 1, message, "Add the missing evidence or metadata required for this transition."))
		}
	}

	require(len(requiredSections(document, []string{"Abstract", "Motivation", "Specification", "Rationale", "Security Considerations"})) == 0, "required ARC sections are missing")
	require(hasField(document, "implementation-required"), "implementation-required must be explicitly declared")

	switch target {
	case "Review":
		require(document.Status == "Draft", "transition to Review requires current status Draft")
		if document.ImplementationRequired {
			require(document.ImplementationURL != "", "implementation-url is required when implementation-required is true")
			require(document.ImplementationMaintainer != "", "implementation-maintainer is required when implementation-required is true")
			require(hasSection(document, "Reference Implementation"), "Reference Implementation section is required when implementation-required is true")
			require(hasSection(document, "Test Cases"), "Test Cases section is required when implementation-required is true")
		}
	case "Last Call":
		require(document.Status == "Review", "transition to Last Call requires current status Review")
		summary := loadSummary(document, &diagnostics)
		require(summary != nil, "transition to Last Call requires a valid adoption summary")
		require(document.AdoptionSummary == fmt.Sprintf("adoption/arc-%04d.yaml", document.Number), "adoption-summary must point to the matching adoption file")
		if document.ImplementationRequired && summary != nil {
			require(summary.ReferenceImplementation != nil, "reference-implementation status is required in the adoption summary")
			if summary.ReferenceImplementation != nil {
				require(slices.Contains([]string{"wip", "shipped"}, summary.ReferenceImplementation.Status), "reference-implementation.status must be wip or shipped")
			}
		}
	case "Final":
		require(document.Status == "Last Call", "transition to Final requires current status Last Call")
		require(document.LastCallDeadline != "", "last-call-deadline is required for transition to Final")
		summary := loadSummary(document, &diagnostics)
		if summary == nil {
			require(false, "transition to Final requires a valid adoption summary")
			break
		}
		require(summary.HasAnyEvidence(), "transition to Final requires at least one non-empty evidence entry")
		if document.ImplementationRequired {
			require(document.ImplementationURL != "" && document.ImplementationMaintainer != "", "reference implementation metadata is required in the ARC file")
			require(summary.ReferenceImplementation != nil, "reference-implementation status is required in the adoption summary")
			if summary.ReferenceImplementation != nil {
				require(summary.ReferenceImplementation.Status == "shipped", "reference-implementation.status must be shipped")
			}
			require(summary.HasActorEvidence(), "transition to Final requires at least one adoption actor entry with evidence")
		}
	case "Idle":
		require(document.Status == "Final", "transition to Idle requires current status Final")
		require(document.IdleSince != "", "idle-since is required for transition to Idle")
		hasSummary := loadSummary(document, &diagnostics) != nil
		require(hasSummary, "transition to Idle requires a valid adoption summary")
	default:
		diagnostics = append(diagnostics, diag.NewWithHint("R:026", diag.OriginNative, path, 1, 1, fmt.Sprintf("unsupported transition target %q", target), "Use one of Review, Last Call, Final, or Idle."))
		return diagnostics, nil
	}

	for _, reminder := range []string{
		"Confirm consensus and dissent handling in the tracking issue.",
		"Confirm that the adoption evidence is substantively adequate.",
		"Confirm reference implementation quality and maintenance with the editor.",
		"Confirm editor approval before changing ARC status.",
	} {
		diagnostics = append(diagnostics, diag.NewWithHint("R:020", diag.OriginNative, path, 1, 1, reminder, "This reminder is informational and must be handled outside the CLI."))
	}

	diag.SortDiagnostics(diagnostics)
	return diagnostics, nil
}

func loadSummary(document *arc.Document, diagnostics *[]diag.Diagnostic) *adoption.Summary {
	if document.AdoptionSummary == "" {
		return nil
	}
	root := arc.FindRepoRoot(filepath.Dir(document.Path))
	path := filepath.Join(root, filepath.FromSlash(document.AdoptionSummary))
	registry, registryDiagnostics, registryErr := adoption.LoadVettedAdopters(adoption.RegistryPath(root))
	*diagnostics = append(*diagnostics, registryDiagnostics...)
	if registryErr == nil && len(registryDiagnostics) == 0 {
		registryDiagnostics = adoption.ValidateVettedAdopters(registry)
		*diagnostics = append(*diagnostics, registryDiagnostics...)
	}
	if registryErr != nil || len(registryDiagnostics) != 0 {
		registry = nil
	}
	summary, loadDiagnostics, err := adoption.Load(path)
	*diagnostics = append(*diagnostics, loadDiagnostics...)
	if err != nil {
		return nil
	}
	*diagnostics = append(*diagnostics, adoption.Validate(summary, document, registry)...)
	return summary
}

func hasSection(document *arc.Document, section string) bool {
	_, ok := document.Sections[section]
	return ok
}

func requiredSections(document *arc.Document, names []string) []string {
	missing := make([]string, 0)
	for _, name := range names {
		if !hasSection(document, name) {
			missing = append(missing, name)
		}
	}
	return missing
}

func hasField(document *arc.Document, name string) bool {
	_, ok := document.Fields[name]
	return ok
}
