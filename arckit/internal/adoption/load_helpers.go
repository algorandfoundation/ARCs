package adoption

import (
	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

func LoadValidatedRegistry(root string) (*VettedAdopters, []diag.Diagnostic, error) {
	registry, diagnostics, err := LoadVettedAdopters(RegistryPath(root))
	if err != nil {
		return nil, diagnostics, err
	}
	if len(diagnostics) != 0 {
		return nil, diagnostics, nil
	}

	validationDiagnostics := ValidateVettedAdopters(registry)
	diagnostics = append(diagnostics, validationDiagnostics...)
	if len(validationDiagnostics) != 0 {
		return nil, diagnostics, nil
	}

	return registry, diagnostics, nil
}

func LoadValidatedSummary(path string, document *arc.Document, registry *VettedAdopters) (*Summary, []diag.Diagnostic, error) {
	summary, diagnostics, err := Load(path)
	if err != nil {
		return nil, diagnostics, err
	}
	diagnostics = append(diagnostics, Validate(summary, document, registry)...)
	return summary, diagnostics, nil
}
