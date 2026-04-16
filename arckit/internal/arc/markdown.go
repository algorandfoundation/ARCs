package arc

import (
	"bytes"
	"strconv"
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
	linkDepth := 0
	codeSpanDepth := 0
	codeBlockDepth := 0
	_ = ast.Walk(tree, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		switch typed := node.(type) {
		case *ast.Link:
			if entering {
				linkDepth++
				document.Links = append(document.Links, Link{
					Destination: string(typed.Destination),
					Line:        nodeLine(source, typed, bodyStartLine),
				})
			} else if linkDepth > 0 {
				linkDepth--
			}
		case *ast.AutoLink:
			if entering {
				if typed.AutoLinkType != ast.AutoLinkURL {
					return ast.WalkContinue, nil
				}
				document.Links = append(document.Links, Link{
					Destination: string(typed.URL(source)),
					Line:        nodeLine(source, typed, bodyStartLine),
				})
			}
		case *ast.CodeSpan:
			if entering {
				codeSpanDepth++
			} else if codeSpanDepth > 0 {
				codeSpanDepth--
			}
		case *ast.FencedCodeBlock:
			if entering {
				codeBlockDepth++
			} else if codeBlockDepth > 0 {
				codeBlockDepth--
			}
		case *ast.CodeBlock:
			if entering {
				codeBlockDepth++
			} else if codeBlockDepth > 0 {
				codeBlockDepth--
			}
		case *ast.Heading:
			if !entering {
				return ast.WalkContinue, nil
			}
			line := nodeLine(source, typed, bodyStartLine)
			document.Headings = append(document.Headings, Heading{
				Level: typed.Level,
				Title: strings.TrimSpace(plainText(source, typed)),
				Line:  line,
			})
			if typed.Level != 2 {
				return ast.WalkContinue, nil
			}
			title := strings.TrimSpace(plainText(source, typed))
			document.Sections[title] = line
		case *ast.Text:
			if entering && codeSpanDepth == 0 && codeBlockDepth == 0 {
				appendARCReferencesFromText(string(typed.Segment.Value(source)), nodeLine(source, typed, bodyStartLine), linkDepth > 0, &document.ARCReferences)
			}
		case *ast.String:
			if entering && codeSpanDepth == 0 && codeBlockDepth == 0 {
				appendARCReferencesFromText(string(typed.Value), nodeLine(source, typed, bodyStartLine), linkDepth > 0, &document.ARCReferences)
			}
		}
		return ast.WalkContinue, nil
	})
}

func appendARCReferencesFromText(value string, baseLine int, inLink bool, out *[]ARCReference) {
	if value == "" {
		return
	}
	indexes := arcRefPattern.FindAllStringSubmatchIndex(value, -1)
	for _, index := range indexes {
		if len(index) < 4 {
			continue
		}
		raw := value[index[0]:index[1]]
		number, err := strconv.Atoi(value[index[2]:index[3]])
		if err != nil {
			continue
		}
		line := baseLine + strings.Count(value[:index[0]], "\n")
		*out = append(*out, ARCReference{
			Raw:       raw,
			Number:    number,
			Line:      line,
			InLink:    inLink,
			Canonical: raw == "ARC-"+strconv.Itoa(number),
		})
	}
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
