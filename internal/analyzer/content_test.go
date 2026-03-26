package analyzer

import (
	"testing"

	"github.com/jtsilverman/skillscore/internal/parser"
)

func TestContent_GoodSkill(t *testing.T) {
	result, err := parser.ParseSkillDir("../../testdata/good-skill")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	md := parser.AnalyzeMarkdown(result.Body)
	score := ScoreContent("../../testdata/good-skill", result, md)

	if score.Points < 80 {
		t.Errorf("good skill content score=%0.f, want >= 80", score.Points)
	}
}

func TestContent_BadSkill(t *testing.T) {
	result, err := parser.ParseSkillDir("../../testdata/bad-skill")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	md := parser.AnalyzeMarkdown(result.Body)
	score := ScoreContent("../../testdata/bad-skill", result, md)

	if score.Points >= 50 {
		t.Errorf("bad skill content score=%0.f, want < 50", score.Points)
	}

	checkMap := checkResultMap(score.Checks)
	if checkMap["has_code_examples"] {
		t.Error("bad skill should lack code examples")
	}
	if checkMap["has_workflow_steps"] {
		t.Error("bad skill should lack workflow steps")
	}
	if checkMap["has_structure"] {
		t.Error("bad skill should lack heading structure")
	}
}

func TestContent_BodyTooLong(t *testing.T) {
	longBody := ""
	for i := 0; i < 600; i++ {
		longBody += "This is line content.\n"
	}
	md := parser.AnalyzeMarkdown(longBody)
	result := &parser.ParseResult{Body: longBody}
	score := ScoreContent("/tmp", result, md)

	checkMap := checkResultMap(score.Checks)
	if checkMap["body_under_500_lines"] {
		t.Error("should flag body over 500 lines")
	}
}

func TestContent_MinimalContent(t *testing.T) {
	md := parser.AnalyzeMarkdown("do stuff")
	result := &parser.ParseResult{Body: "do stuff"}
	score := ScoreContent("/tmp", result, md)

	checkMap := checkResultMap(score.Checks)
	if checkMap["has_substantial_content"] {
		t.Error("should flag sparse content")
	}
}
