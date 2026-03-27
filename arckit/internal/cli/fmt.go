package cli

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/algorandfoundation/ARCs/arckit/internal/arc"
	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
	"gopkg.in/yaml.v3"
)

func applyNativeFix(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	document, diagnostics, err := arc.Load(path)
	if err != nil {
		return err
	}
	if hasUnsafeFrontMatterDiagnostics(diagnostics) {
		return fmt.Errorf("front matter could not be safely reformatted")
	}

	reordered, err := reorderFrontMatter(document)
	if err != nil {
		return err
	}

	fixed := []byte(reordered)
	if bytes.Equal(content, fixed) {
		return nil
	}
	return os.WriteFile(path, fixed, 0o644)
}

func hasUnsafeFrontMatterDiagnostics(diagnostics []diag.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		switch diagnostic.RuleID {
		case "R:001", "R:005", "R:006":
			return true
		}
	}
	return false
}

func reorderFrontMatter(document *arc.Document) (string, error) {
	builder := strings.Builder{}
	builder.WriteString("---\n")
	order := arc.OrderedFields()
	for _, key := range order {
		value, ok := document.Fields[key]
		if !ok {
			continue
		}
		builder.WriteString(key)
		builder.WriteString(":")
		encoded, err := yaml.Marshal(value)
		if err != nil {
			return "", err
		}
		text := strings.TrimRight(string(encoded), "\n")
		if strings.Contains(text, "\n") {
			builder.WriteString("\n")
			for _, line := range strings.Split(text, "\n") {
				builder.WriteString("  ")
				builder.WriteString(line)
				builder.WriteString("\n")
			}
			continue
		}
		builder.WriteString(" ")
		builder.WriteString(text)
		builder.WriteString("\n")
	}
	builder.WriteString("---\n\n")
	builder.Write(document.Body)
	return builder.String(), nil
}
