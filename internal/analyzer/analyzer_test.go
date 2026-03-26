package analyzer

import (
	"testing"
)

func TestAnalyzeSkill_GoodSkill(t *testing.T) {
	report, err := AnalyzeSkill("../../testdata/good-skill")
	if err != nil {
		t.Fatalf("analyze error: %v", err)
	}

	if report.Name != "deploy-app" {
		t.Errorf("name=%q, want %q", report.Name, "deploy-app")
	}
	if report.Overall.Points < 75 {
		t.Errorf("overall score=%0.f, want >= 75", report.Overall.Points)
	}
	if report.Overall.Grade == "F" || report.Overall.Grade == "D" || report.Overall.Grade == "D-" {
		t.Errorf("grade=%q, want at least C", report.Overall.Grade)
	}

	// Check that individual dimensions are scored
	if report.Structure.Points == 0 && report.Description.Points == 0 {
		t.Error("dimension scores should not all be zero")
	}

	t.Logf("Good skill: %s (%0.f/100)", report.Overall.Grade, report.Overall.Points)
	t.Logf("  Structure:   %0.f", report.Structure.Points)
	t.Logf("  Description: %0.f", report.Description.Points)
	t.Logf("  Content:     %0.f", report.Content.Points)
	t.Logf("  Engineering: %0.f", report.Engineering.Points)
	t.Logf("  Packaging:   %0.f", report.Packaging.Points)
}

func TestAnalyzeSkill_BadSkill(t *testing.T) {
	report, err := AnalyzeSkill("../../testdata/bad-skill")
	if err != nil {
		t.Fatalf("analyze error: %v", err)
	}

	if report.Overall.Points >= 60 {
		t.Errorf("overall score=%0.f, want < 60", report.Overall.Points)
	}

	// Should have suggestions
	if len(report.Suggestions) == 0 {
		t.Error("bad skill should have improvement suggestions")
	}

	t.Logf("Bad skill: %s (%0.f/100)", report.Overall.Grade, report.Overall.Points)
	t.Logf("  Structure:   %0.f", report.Structure.Points)
	t.Logf("  Description: %0.f", report.Description.Points)
	t.Logf("  Content:     %0.f", report.Content.Points)
	t.Logf("  Engineering: %0.f", report.Engineering.Points)
	t.Logf("  Packaging:   %0.f", report.Packaging.Points)
	t.Logf("  Suggestions: %d", len(report.Suggestions))
}

func TestAnalyzeSkill_NonexistentDir(t *testing.T) {
	_, err := AnalyzeSkill("/nonexistent/skill/path")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestPointsToGrade(t *testing.T) {
	tests := []struct {
		points float64
		want   string
	}{
		{100, "A+"},
		{95, "A"},
		{91, "A-"},
		{88, "B+"},
		{85, "B"},
		{80, "B-"},
		{78, "C+"},
		{75, "C"},
		{70, "C-"},
		{68, "D+"},
		{65, "D"},
		{60, "D-"},
		{55, "F"},
		{0, "F"},
	}
	for _, tt := range tests {
		got := PointsToGrade(tt.points)
		if got != tt.want {
			t.Errorf("PointsToGrade(%0.f)=%q, want %q", tt.points, got, tt.want)
		}
	}
}
