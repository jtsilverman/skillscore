package analyzer

import (
	"testing"

	"github.com/jtsilverman/skillscore/internal/parser"
)

func TestEngineering_GoodSkill(t *testing.T) {
	result, err := parser.ParseSkillDir("../../testdata/good-skill")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	score := ScoreEngineering("../../testdata/good-skill", result)

	if score.Points < 70 {
		t.Errorf("good skill engineering score=%0.f, want >= 70", score.Points)
	}

	checkMap := checkResultMap(score.Checks)
	if !checkMap["has_error_handling"] {
		t.Error("good skill scripts have try/except")
	}
	if !checkMap["forward_slash_paths"] {
		t.Error("good skill scripts use forward slashes")
	}
}

func TestEngineering_BadSkill(t *testing.T) {
	result, err := parser.ParseSkillDir("../../testdata/bad-skill")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	score := ScoreEngineering("../../testdata/bad-skill", result)

	// Bad skill has no scripts, should get full marks
	if score.Points != 100 {
		t.Errorf("bad skill (no scripts) engineering score=%0.f, want 100", score.Points)
	}
}

func TestEngineering_NoScripts(t *testing.T) {
	result := &parser.ParseResult{Body: "# Content"}
	score := ScoreEngineering("/nonexistent-empty-dir", result)

	if score.Points != 100 {
		t.Errorf("no-scripts score=%0.f, want 100", score.Points)
	}
}
