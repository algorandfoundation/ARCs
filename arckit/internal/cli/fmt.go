package cli

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

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
		normalized := normalizeFrontMatterValue(key, value)
		if isDateField(key) {
			if text, ok := normalized.(string); ok {
				builder.WriteString(" ")
				builder.WriteString(text)
				builder.WriteString("\n")
				continue
			}
		}
		encoded, err := yaml.Marshal(normalized)
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

func normalizeFrontMatterValue(key string, value any) any {
	if !isDateField(key) {
		return value
	}
	switch typed := value.(type) {
	case time.Time:
		return typed.Format("2006-01-02")
	case string:
		if parsed, err := time.Parse(time.RFC3339, typed); err == nil {
			return parsed.Format("2006-01-02")
		}
	}
	return value
}

func isDateField(key string) bool {
	switch key {
	case "created", "updated", "last-call-deadline", "idle-since":
		return true
	default:
		return false
	}
}
