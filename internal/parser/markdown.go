package parser

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// MarkdownAnalysis holds structural metrics extracted from markdown content.
type MarkdownAnalysis struct {
	LineCount      int
	WordCount      int
	HeadingCount   int
	Headings       []Heading
	CodeBlockCount int
	CodeLanguages  []string
	OrderedLists   int
	UnorderedLists int
	Links          []Link
}

// Heading is a parsed heading with its level and text.
type Heading struct {
	Level int
	Text  string
}

// Link is a parsed link reference.
type Link struct {
	Destination string
	Text        string
}

// AnalyzeMarkdown parses a markdown string and extracts structural metrics.
func AnalyzeMarkdown(content string) *MarkdownAnalysis {
	a := &MarkdownAnalysis{}

	// Line and word counts
	lines := strings.Split(content, "\n")
	a.LineCount = len(lines)
	for _, line := range lines {
		words := strings.Fields(line)
		a.WordCount += len(words)
	}

	// Parse AST
	source := []byte(content)
	md := goldmark.New()
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	// Walk AST
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.Heading:
			a.HeadingCount++
			headingText := string(n.Text(source))
			a.Headings = append(a.Headings, Heading{
				Level: n.Level,
				Text:  headingText,
			})

		case *ast.FencedCodeBlock:
			a.CodeBlockCount++
			lang := string(n.Language(source))
			if lang != "" {
				a.CodeLanguages = append(a.CodeLanguages, lang)
			}

		case *ast.List:
			if n.IsOrdered() {
				a.OrderedLists++
			} else {
				a.UnorderedLists++
			}

		case *ast.Link:
			dest := string(n.Destination)
			linkText := string(n.Text(source))
			a.Links = append(a.Links, Link{
				Destination: dest,
				Text:        linkText,
			})
		}

		return ast.WalkContinue, nil
	})

	return a
}
