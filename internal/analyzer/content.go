package analyzer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jtsilverman/skillscore/internal/parser"
)

var (
	stepPattern = regexp.MustCompile(`(?i)(step\s+\d|phase\s+\d)`)
)

// ScoreContent evaluates the markdown body quality of a skill.
func ScoreContent(dir string, result *parser.ParseResult, md *parser.MarkdownAnalysis) DimensionScore {
	var checks []Check

	// Check 1: Body under 500 lines
	under500 := md.LineCount <= 500
	checks = append(checks, Check{
		Name:   "body_under_500_lines",
		Passed: under500,
		Weight: 15,
		Detail: checkDetail(under500,
			"SKILL.md body is concise (under 500 lines)",
			"SKILL.md body exceeds 500 lines; split into reference files"),
	})

	// Check 2: Has code examples
	hasCode := md.CodeBlockCount > 0
	checks = append(checks, Check{
		Name:   "has_code_examples",
		Passed: hasCode,
		Weight: 20,
		Detail: checkDetail(hasCode,
			"Contains code examples or templates",
			"No code examples found; add examples to guide Claude"),
	})

	// Check 3: Has workflow steps (ordered lists or "Step N" patterns)
	hasWorkflow := md.OrderedLists > 0 || stepPattern.MatchString(result.Body)
	checks = append(checks, Check{
		Name:   "has_workflow_steps",
		Passed: hasWorkflow,
		Weight: 20,
		Detail: checkDetail(hasWorkflow,
			"Contains structured workflow steps",
			"No workflow steps found; add numbered steps for complex procedures"),
	})

	// Check 4: Reference depth (links to files that also link out = too deep)
	depthOk := checkReferenceDepth(dir, md)
	checks = append(checks, Check{
		Name:   "reference_depth_ok",
		Passed: depthOk,
		Weight: 15,
		Detail: checkDetail(depthOk,
			"Reference files are at most one level deep",
			"Nested references detected (A->B->C); keep references one level from SKILL.md"),
	})

	// Check 5: Has headings (structured content)
	hasStructure := md.HeadingCount >= 2
	checks = append(checks, Check{
		Name:   "has_structure",
		Passed: hasStructure,
		Weight: 15,
		Detail: checkDetail(hasStructure,
			"Content is well-structured with headings",
			"Add section headings to organize the skill content"),
	})

	// Check 6: Non-trivial content (not just a one-liner)
	hasContent := md.WordCount >= 30
	checks = append(checks, Check{
		Name:   "has_substantial_content",
		Passed: hasContent,
		Weight: 15,
		Detail: checkDetail(hasContent,
			"Skill has substantial instructions",
			"Skill content is too sparse; add enough detail for Claude to execute effectively"),
	})

	return computeDimension(checks)
}

// checkReferenceDepth looks at linked .md files from SKILL.md and checks
// if any of those files also contain links to other .md files (depth > 1).
func checkReferenceDepth(dir string, md *parser.MarkdownAnalysis) bool {
	for _, link := range md.Links {
		if !strings.HasSuffix(link.Destination, ".md") {
			continue
		}
		// Check if the linked file exists and contains further .md links
		linkedPath := filepath.Join(dir, link.Destination)
		content, err := os.ReadFile(linkedPath)
		if err != nil {
			continue // File doesn't exist locally, skip
		}
		linkedMD := parser.AnalyzeMarkdown(string(content))
		for _, subLink := range linkedMD.Links {
			if strings.HasSuffix(subLink.Destination, ".md") {
				return false // Depth > 1 detected
			}
		}
	}
	return true
}
