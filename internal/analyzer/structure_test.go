package analyzer

import (
	"testing"

	"github.com/jtsilverman/skillscore/internal/parser"
)

func TestStructure_GoodSkill(t *testing.T) {
	result, err := parser.ParseSkillDir("../../testdata/good-skill")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	score := ScoreStructure("../../testdata/good-skill", result)

	if score.Points < 90 {
		t.Errorf("good skill structure score=%0.f, want >= 90", score.Points)
	}
	for _, c := range score.Checks {
		if !c.Passed {
			t.Errorf("check %q failed: %s", c.Name, c.Detail)
		}
	}
}

func TestStructure_BadSkill(t *testing.T) {
	result, err := parser.ParseSkillDir("../../testdata/bad-skill")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	score := ScoreStructure("../../testdata/bad-skill", result)

	// Bad skill has uppercase name + reserved word issue + backslash paths
	if score.Points >= 90 {
		t.Errorf("bad skill structure score=%0.f, want < 90", score.Points)
	}

	// Check that specific checks failed
	checkResults := make(map[string]bool)
	for _, c := range score.Checks {
		checkResults[c.Name] = c.Passed
	}

	if checkResults["name_valid"] {
		t.Error("name_valid should fail for HELPER_Tool (uppercase + underscore)")
	}
	if checkResults["no_backslash_paths"] {
		t.Error("no_backslash_paths should fail (C:\\Users path in content)")
	}
}

func TestStructure_NoFrontmatter(t *testing.T) {
	result := &parser.ParseResult{
		HasFrontmatter: false,
		Body:           "# Just content",
	}
	score := ScoreStructure("/nonexistent", result)

	if score.Points > 30 {
		t.Errorf("no-frontmatter score=%0.f, want <= 30", score.Points)
	}
}

func TestStructure_ReservedNameWords(t *testing.T) {
	result := &parser.ParseResult{
		HasFrontmatter: true,
		Frontmatter: &parser.Frontmatter{
			Name:        "my-claude-helper",
			Description: "Does things",
		},
		Body: "content",
	}
	score := ScoreStructure("/tmp", result)

	checkResults := make(map[string]bool)
	for _, c := range score.Checks {
		checkResults[c.Name] = c.Passed
	}
	if checkResults["name_valid"] {
		t.Error("name containing 'claude' should fail reserved word check")
	}
}
