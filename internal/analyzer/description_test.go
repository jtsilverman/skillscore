package analyzer

import (
	"testing"

	"github.com/jtsilverman/skillscore/internal/parser"
)

func makeResult(name, desc string) *parser.ParseResult {
	return &parser.ParseResult{
		HasFrontmatter: true,
		Frontmatter: &parser.Frontmatter{
			Name:        name,
			Description: desc,
		},
		Body: "# Content",
	}
}

func TestDescription_Good(t *testing.T) {
	result := makeResult("deploy-app",
		"Deploys the application to production with zero-downtime rolling updates. Use when the user asks to deploy, push to production, or ship a release.")

	score := ScoreDescription(result)
	if score.Points < 80 {
		t.Errorf("good description score=%0.f, want >= 80", score.Points)
	}

	checkMap := checkResultMap(score.Checks)
	if !checkMap["has_action_verb"] {
		t.Error("should have action verb (deploy)")
	}
	if !checkMap["has_trigger_context"] {
		t.Error("should have trigger context (Use when)")
	}
	if !checkMap["third_person"] {
		t.Error("should be third person")
	}
}

func TestDescription_Bad(t *testing.T) {
	result := makeResult("helper", "helps with stuff")

	score := ScoreDescription(result)
	if score.Points >= 50 {
		t.Errorf("bad description score=%0.f, want < 50", score.Points)
	}

	checkMap := checkResultMap(score.Checks)
	if checkMap["specific_description"] {
		t.Error("should flag vague description")
	}
	if checkMap["has_trigger_context"] {
		t.Error("should flag missing trigger context")
	}
}

func TestDescription_FirstPerson(t *testing.T) {
	result := makeResult("my-skill", "I can help you process Excel files and generate reports")

	score := ScoreDescription(result)
	checkMap := checkResultMap(score.Checks)
	if checkMap["third_person"] {
		t.Error("should flag first person ('I can')")
	}
}

func TestDescription_SecondPerson(t *testing.T) {
	result := makeResult("my-skill", "You can use this to process Excel files")

	score := ScoreDescription(result)
	checkMap := checkResultMap(score.Checks)
	if checkMap["third_person"] {
		t.Error("should flag second person ('You can')")
	}
}

func TestDescription_Empty(t *testing.T) {
	result := makeResult("my-skill", "")

	score := ScoreDescription(result)
	if score.Points >= 30 {
		t.Errorf("empty description score=%0.f, want < 30", score.Points)
	}
}

func TestDescription_TooLong(t *testing.T) {
	longDesc := ""
	for i := 0; i < 250; i++ {
		longDesc += "words "
	}
	result := makeResult("my-skill", longDesc)

	score := ScoreDescription(result)
	checkMap := checkResultMap(score.Checks)
	if checkMap["under_char_limit"] {
		t.Error("should flag description over 1024 chars")
	}
}

func checkResultMap(checks []Check) map[string]bool {
	m := make(map[string]bool)
	for _, c := range checks {
		m[c.Name] = c.Passed
	}
	return m
}
