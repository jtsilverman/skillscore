package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jtsilverman/skillscore/internal/analyzer"
)

func makeReport(name string, score float64, grade string) *analyzer.SkillReport {
	return &analyzer.SkillReport{
		Path: "/tmp/test",
		Name: name,
		Desc: "A test skill",
		Overall: analyzer.Score{
			Points: score,
			Grade:  grade,
		},
		Structure:   analyzer.DimensionScore{Score: analyzer.Score{Points: 80, Grade: "B+"}},
		Description: analyzer.DimensionScore{Score: analyzer.Score{Points: 70, Grade: "C+"}},
		Content:     analyzer.DimensionScore{Score: analyzer.Score{Points: 60, Grade: "D+"}},
		Engineering: analyzer.DimensionScore{Score: analyzer.Score{Points: 90, Grade: "A-"}},
		Packaging:   analyzer.DimensionScore{Score: analyzer.Score{Points: 50, Grade: "D"}},
		Suggestions: []analyzer.Suggestion{
			{Priority: "high", Message: "Add error handling"},
			{Priority: "low", Message: "Improve docs"},
		},
	}
}

func TestRenderCompact_NameTruncation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string // what the output should contain
	}{
		{
			name:     "short name unchanged",
			input:    "my-skill",
			contains: "my-skill",
		},
		{
			name:     "exactly 30 chars unchanged",
			input:    "abcdefghijklmnopqrstuvwxyz1234",
			contains: "abcdefghijklmnopqrstuvwxyz1234",
		},
		{
			name:     "over 30 chars truncated with ellipsis",
			input:    "this-is-a-very-long-skill-name-that-exceeds-thirty",
			contains: "this-is-a-very-long-skill-n...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := makeReport(tt.input, 85, "B+")
			out := RenderCompact(r)
			if !strings.Contains(out, tt.contains) {
				t.Errorf("RenderCompact() output missing %q\ngot: %s", tt.contains, out)
			}
		})
	}
}

func TestRenderCompact_TruncatedNameLength(t *testing.T) {
	// A name longer than 30 chars should become 27 chars + "..."
	longName := "abcdefghijklmnopqrstuvwxyz12345678" // 34 chars
	r := makeReport(longName, 85, "B+")
	out := RenderCompact(r)

	truncated := longName[:27] + "..."
	if !strings.Contains(out, truncated) {
		t.Errorf("expected truncated name %q in output, got: %s", truncated, out)
	}
	// The original full name should NOT appear
	if strings.Contains(out, longName) {
		t.Errorf("full name %q should not appear in compact output", longName)
	}
}

func TestRenderCompact_ContainsDimensionScores(t *testing.T) {
	r := makeReport("test-skill", 85, "B+")
	out := RenderCompact(r)
	// Should contain dimension labels
	for _, prefix := range []string{"S:", "D:", "C:", "E:", "P:"} {
		if !strings.Contains(out, prefix) {
			t.Errorf("RenderCompact() missing dimension prefix %q in: %s", prefix, out)
		}
	}
}

func TestRenderJSON_ValidJSON(t *testing.T) {
	r := makeReport("json-skill", 75, "C+")
	var buf bytes.Buffer
	err := RenderJSON(&buf, r)
	if err != nil {
		t.Fatalf("RenderJSON() error: %v", err)
	}

	// Must be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("RenderJSON() produced invalid JSON: %v\noutput: %s", err, buf.String())
	}

	// Check key fields are present
	if parsed["name"] != "json-skill" {
		t.Errorf("expected name 'json-skill', got %v", parsed["name"])
	}
}

func TestRenderJSONMulti_ValidJSON(t *testing.T) {
	reports := []*analyzer.SkillReport{
		makeReport("skill-one", 80, "B+"),
		makeReport("skill-two", 60, "D+"),
	}

	var buf bytes.Buffer
	err := RenderJSONMulti(&buf, reports)
	if err != nil {
		t.Fatalf("RenderJSONMulti() error: %v", err)
	}

	// Must be valid JSON
	var parsed struct {
		Skills []json.RawMessage `json:"skills"`
		Count  int               `json:"count"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("RenderJSONMulti() produced invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if parsed.Count != 2 {
		t.Errorf("count=%d, want 2", parsed.Count)
	}
	if len(parsed.Skills) != 2 {
		t.Errorf("skills array length=%d, want 2", len(parsed.Skills))
	}
}

func TestRenderJSONMulti_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := RenderJSONMulti(&buf, nil)
	if err != nil {
		t.Fatalf("RenderJSONMulti() error: %v", err)
	}

	var parsed struct {
		Skills []json.RawMessage `json:"skills"`
		Count  int               `json:"count"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("RenderJSONMulti() produced invalid JSON for empty input: %v", err)
	}
	if parsed.Count != 0 {
		t.Errorf("count=%d, want 0", parsed.Count)
	}
}

func TestRenderFull_NoPanic(t *testing.T) {
	r := makeReport("full-test-skill", 85, "B+")
	// Add some checks to dimensions for verbose mode
	r.Structure.Checks = []analyzer.Check{
		{Name: "has-skill-md", Passed: true, Weight: 3, Detail: "SKILL.md found"},
		{Name: "has-readme", Passed: false, Weight: 1, Detail: "README.md missing"},
	}

	// Non-verbose
	out := RenderFull(r, false)
	if out == "" {
		t.Error("RenderFull(verbose=false) returned empty string")
	}
	if !strings.Contains(out, "full-test-skill") {
		t.Error("RenderFull() output missing skill name")
	}

	// Verbose
	outVerbose := RenderFull(r, true)
	if outVerbose == "" {
		t.Error("RenderFull(verbose=true) returned empty string")
	}
	// Verbose should be longer (includes check details)
	if len(outVerbose) <= len(out) {
		t.Error("verbose output should be longer than non-verbose")
	}
}

func TestRenderFull_WithSuggestions(t *testing.T) {
	r := makeReport("suggestion-skill", 70, "C+")
	out := RenderFull(r, false)
	if !strings.Contains(out, "Suggestions") {
		t.Error("RenderFull() should show suggestions section")
	}
}

func TestRenderFull_NoSuggestions(t *testing.T) {
	r := makeReport("perfect-skill", 100, "A+")
	r.Suggestions = nil
	out := RenderFull(r, false)
	if strings.Contains(out, "Suggestions") {
		t.Error("RenderFull() should not show suggestions section when there are none")
	}
}
