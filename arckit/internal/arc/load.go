package arc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
	"gopkg.in/yaml.v3"
)

func Load(path string) (*Document, []diag.Diagnostic, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, []diag.Diagnostic{
			diag.NewWithHint("R:027", diag.OriginNative, path, 0, 0, err.Error(), "Check the file path and permissions, then retry."),
		}, err
	}

	document := &Document{
		Path:       filepath.Clean(path),
		Raw:        content,
		Fields:     map[string]any{},
		FieldLines: map[string]int{},
		Sections:   map[string]int{},
	}

	diagnostics := make([]diag.Diagnostic, 0)

	matches := arcPathPattern.FindStringSubmatch(filepath.ToSlash(document.Path))
	if len(matches) != 3 {
		diagnostics = append(diagnostics, diag.NewWithHint("R:002", diag.OriginNative, document.Path, 1, 1, "ARC files must live under ARCs/arc-####.md", "Move or rename the file to ARCs/arc-####.md."))
	} else {
		number, _ := strconv.Atoi(matches[2])
		document.FilenameNumber = number
		document.HasFilenameNumber = true
	}

	frontMatter, body, bodyStartLine, parseDiagnostics := splitFrontMatter(document.Path, content)
	diagnostics = append(diagnostics, parseDiagnostics...)
	document.FrontMatter = frontMatter
	document.Body = body
	document.BodyStartLine = bodyStartLine
	if len(frontMatter) == 0 {
		parseMarkdown(document)
		return document, diagnostics, nil
	}
	diagnostics = append(diagnostics, frontMatterBlankLineDiagnostics(document.Path, frontMatter)...)

	root := yaml.Node{}
	if err := yaml.Unmarshal(frontMatter, &root); err != nil {
		diagnostics = append(diagnostics, diag.NewWithHint("R:005", diag.OriginNative, document.Path, 1, 1, err.Error(), "Fix the YAML syntax in the front matter block."))
		parseMarkdown(document)
		return document, diagnostics, nil
	}

	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		diagnostics = append(diagnostics, diag.NewWithHint("R:005", diag.OriginNative, document.Path, 1, 1, "front matter must decode to a mapping", "Use top-level key/value pairs in the front matter block."))
		parseMarkdown(document)
		return document, diagnostics, nil
	}

	mapping := root.Content[0]
	orderIndex := make(map[string]int, len(fieldOrder))
	for index, name := range fieldOrder {
		orderIndex[name] = index
	}

	lastIndex := -1
	for index := 0; index < len(mapping.Content); index += 2 {
		keyNode := mapping.Content[index]
		valueNode := mapping.Content[index+1]
		key := keyNode.Value
		document.FieldOrder = append(document.FieldOrder, key)
		document.FieldLines[key] = keyNode.Line
		if _, ok := orderIndex[key]; !ok {
			diagnostics = append(diagnostics, diag.NewWithHint("R:006", diag.OriginNative, document.Path, keyNode.Line, keyNode.Column, fmt.Sprintf("unknown ARC field %q", key), "Remove the unknown field or move the data into the document body."))
			continue
		}
		if orderIndex[key] < lastIndex {
			diagnostics = append(diagnostics, diag.NewWithHint("R:003", diag.OriginNative, document.Path, keyNode.Line, keyNode.Column, fmt.Sprintf("field %q is out of order", key), "Reorder the front matter fields to match the canonical order."))
		} else {
			lastIndex = orderIndex[key]
		}
		var value any
		if err := valueNode.Decode(&value); err != nil {
			diagnostics = append(diagnostics, diag.NewWithHint("R:005", diag.OriginNative, document.Path, valueNode.Line, valueNode.Column, fmt.Sprintf("could not decode %q: %v", key, err), "Fix the YAML value for this field."))
			continue
		}
		document.Fields[key] = value
	}

	parseMarkdown(document)
	return document, diagnostics, nil
}

func splitFrontMatter(path string, content []byte) ([]byte, []byte, int, []diag.Diagnostic) {
	if !bytes.HasPrefix(content, []byte("---")) {
		return nil, content, 1, []diag.Diagnostic{
			diag.NewWithHint("R:001", diag.OriginNative, path, 1, 1, "ARC files must start with a YAML front matter block", "Add a front matter block delimited by --- at the top of the file."),
		}
	}
	lines := bytes.Split(content, []byte("\n"))
	if len(lines) == 0 || strings.TrimSpace(string(lines[0])) != "---" {
		return nil, content, 1, []diag.Diagnostic{
			diag.NewWithHint("R:001", diag.OriginNative, path, 1, 1, "front matter must start with a line containing only ---", "Start the file with --- on its own line."),
		}
	}

	offset := len(lines[0]) + 1
	for index := 1; index < len(lines); index++ {
		line := strings.TrimSpace(string(lines[index]))
		if line == "---" {
			frontMatter := content[len(lines[0])+1 : offset-1]
			body := content[offset+len(lines[index]):]
			body = bytes.TrimPrefix(body, []byte("\n"))
			return frontMatter, body, index + 2, nil
		}
		offset += len(lines[index]) + 1
	}
	return nil, content, 1, []diag.Diagnostic{
		diag.NewWithHint("R:001", diag.OriginNative, path, 1, 1, "front matter is not closed with a second --- delimiter", "Terminate the front matter block with --- on its own line."),
	}
}

func frontMatterBlankLineDiagnostics(path string, frontMatter []byte) []diag.Diagnostic {
	lines := strings.Split(string(frontMatter), "\n")
	diagnostics := make([]diag.Diagnostic, 0)
	for index, line := range lines {
		if strings.TrimSpace(line) != "" {
			continue
		}
		diagnostics = append(diagnostics, diag.NewWithHint("R:024", diag.OriginNative, path, index+2, 1, "front matter must not contain empty lines", "Remove empty lines from the front matter block or run arckit fmt."))
	}
	return diagnostics
}
