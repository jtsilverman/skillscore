package parser

import (
	"testing"
)

func TestAnalyzeMarkdown_Basic(t *testing.T) {
	content := `# Title

Some text with **bold** and *italic*.

## Section One

A paragraph with several words here.

## Section Two

- item one
- item two
- item three

1. first
2. second

` + "```python\nprint('hello')\n```\n"

	a := AnalyzeMarkdown(content)

	if a.HeadingCount != 3 {
		t.Errorf("HeadingCount=%d, want 3", a.HeadingCount)
	}
	if a.CodeBlockCount != 1 {
		t.Errorf("CodeBlockCount=%d, want 1", a.CodeBlockCount)
	}
	if len(a.CodeLanguages) != 1 || a.CodeLanguages[0] != "python" {
		t.Errorf("CodeLanguages=%v, want [python]", a.CodeLanguages)
	}
	if a.OrderedLists != 1 {
		t.Errorf("OrderedLists=%d, want 1", a.OrderedLists)
	}
	if a.UnorderedLists != 1 {
		t.Errorf("UnorderedLists=%d, want 1", a.UnorderedLists)
	}
	if a.WordCount == 0 {
		t.Error("WordCount should be > 0")
	}
	if a.LineCount == 0 {
		t.Error("LineCount should be > 0")
	}
}

func TestAnalyzeMarkdown_Links(t *testing.T) {
	content := `See [reference guide](reference/guide.md) for details.

Also check [examples](examples.md).
`
	a := AnalyzeMarkdown(content)

	if len(a.Links) != 2 {
		t.Fatalf("Links count=%d, want 2", len(a.Links))
	}
	if a.Links[0].Destination != "reference/guide.md" {
		t.Errorf("Link[0].Destination=%q", a.Links[0].Destination)
	}
	if a.Links[1].Destination != "examples.md" {
		t.Errorf("Link[1].Destination=%q", a.Links[1].Destination)
	}
}

func TestAnalyzeMarkdown_MultipleCodeBlocks(t *testing.T) {
	content := "```bash\necho hello\n```\n\nSome text.\n\n```python\nprint('hi')\n```\n\n```\nno language\n```\n"

	a := AnalyzeMarkdown(content)

	if a.CodeBlockCount != 3 {
		t.Errorf("CodeBlockCount=%d, want 3", a.CodeBlockCount)
	}
	if len(a.CodeLanguages) != 2 {
		t.Errorf("CodeLanguages count=%d, want 2", len(a.CodeLanguages))
	}
}

func TestAnalyzeMarkdown_Empty(t *testing.T) {
	a := AnalyzeMarkdown("")
	if a.HeadingCount != 0 {
		t.Errorf("HeadingCount=%d for empty", a.HeadingCount)
	}
	if a.CodeBlockCount != 0 {
		t.Errorf("CodeBlockCount=%d for empty", a.CodeBlockCount)
	}
}

func TestAnalyzeMarkdown_Headings(t *testing.T) {
	content := `# H1

## H2

### H3

#### H4
`
	a := AnalyzeMarkdown(content)

	if a.HeadingCount != 4 {
		t.Errorf("HeadingCount=%d, want 4", a.HeadingCount)
	}
	if len(a.Headings) < 4 {
		t.Fatalf("Headings count=%d, want 4", len(a.Headings))
	}
	if a.Headings[0].Level != 1 {
		t.Errorf("Headings[0].Level=%d, want 1", a.Headings[0].Level)
	}
	if a.Headings[3].Level != 4 {
		t.Errorf("Headings[3].Level=%d, want 4", a.Headings[3].Level)
	}
}
