package analyzer

import (
	"testing"

	"github.com/jtsilverman/skillscore/internal/parser"
)

func TestPackaging_GoodSkill(t *testing.T) {
	result, err := parser.ParseSkillDir("../../testdata/good-skill")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	md := parser.AnalyzeMarkdown(result.Body)
	score := ScorePackaging("../../testdata/good-skill", md)

	if score.Points < 80 {
		t.Errorf("good skill packaging score=%0.f, want >= 80", score.Points)
	}
}

func TestPackaging_BadSkill(t *testing.T) {
	result, err := parser.ParseSkillDir("../../testdata/bad-skill")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	md := parser.AnalyzeMarkdown(result.Body)
	score := ScorePackaging("../../testdata/bad-skill", md)

	// Bad skill is minimal, but should still score ok on packaging (no junk, small)
	// It just has SKILL.md and nothing else
	if score.Points < 60 {
		t.Errorf("bad skill packaging score=%0.f, want >= 60 (minimal but clean)", score.Points)
	}
}
