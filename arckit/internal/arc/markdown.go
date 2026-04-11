package arc

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func parseMarkdown(document *Document) {
	parser := goldmark.New()
	reader := text.NewReader(document.Body)
	tree := parser.Parser().Parse(reader)
	source := document.Body
	bodyStartLine := document.BodyStartLine
	if bodyStartLine == 0 {
		bodyStartLine = 1
	}
	_ = ast.Walk(tree, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *ast.Heading:
			if typed.Level != 2 {
				return ast.WalkContinue, nil
			}
			title := strings.TrimSpace(plainText(source, typed))
			document.Sections[title] = nodeLine(source, typed, bodyStartLine)
		case *ast.Link:
			document.Links = append(document.Links, Link{
				Destination: string(typed.Destination),
				Line:        nodeLine(source, typed, bodyStartLine),
			})
		}
		return ast.WalkContinue, nil
	})
}

func plainText(source []byte, node ast.Node) string {
	builder := strings.Builder{}
	_ = ast.Walk(node, func(current ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := current.(type) {
		case *ast.Text:
			builder.Write(typed.Segment.Value(source))
		case *ast.String:
			builder.Write(typed.Value)
		}
		return ast.WalkContinue, nil
	})
	return builder.String()
}

func nodeLine(source []byte, node ast.Node, baseLine int) int {
	if lines := safeLines(node); lines != nil && lines.Len() > 0 {
		return baseLine + bytes.Count(source[:lines.At(0).Start], []byte("\n"))
	}
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if line := nodeLine(source, child, baseLine); line > 0 {
			return line
		}
	}
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if lines := safeLines(parent); lines != nil && lines.Len() > 0 {
			return baseLine + bytes.Count(source[:lines.At(0).Start], []byte("\n"))
		}
	}
	return baseLine
}

func safeLines(node ast.Node) *text.Segments {
	defer func() {
		_ = recover()
	}()
	return node.Lines()
}
