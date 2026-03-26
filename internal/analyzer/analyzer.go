package analyzer

import (
	"fmt"
	"path/filepath"

	"github.com/jtsilverman/skillscore/internal/parser"
)

// SkillReport is the complete analysis of a single skill.
type SkillReport struct {
	Path        string         `json:"path"`
	Name        string         `json:"name"`
	Desc        string         `json:"description"`
	Overall     Score          `json:"overall"`
	Structure   DimensionScore `json:"structure"`
	Description DimensionScore `json:"description"`
	Content     DimensionScore `json:"content"`
	Engineering DimensionScore `json:"engineering"`
	Packaging   DimensionScore `json:"packaging"`
	Suggestions []Suggestion   `json:"suggestions"`
}

// Score holds a numeric score and letter grade.
type Score struct {
	Points float64 `json:"points"` // 0-100
	Grade  string  `json:"grade"`  // A+, A, A-, B+, ..., F
}

// DimensionScore is a score with its individual checks.
type DimensionScore struct {
	Score
	Checks []Check `json:"checks"`
}

// Check is a single pass/fail quality check.
type Check struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Weight int    `json:"weight"` // Relative importance within dimension
	Detail string `json:"detail"` // Why it passed/failed
}

// Suggestion is an actionable improvement recommendation.
type Suggestion struct {
	Priority string `json:"priority"` // "high", "medium", "low"
	Message  string `json:"message"`
}

// Dimension weights (must sum to 1.0).
const (
	WeightStructure   = 0.25
	WeightDescription = 0.25
	WeightContent     = 0.25
	WeightEngineering = 0.15
	WeightPackaging   = 0.10
)

// PointsToGrade converts a 0-100 score to a letter grade with +/- modifiers.
func PointsToGrade(points float64) string {
	switch {
	case points >= 97:
		return "A+"
	case points >= 93:
		return "A"
	case points >= 90:
		return "A-"
	case points >= 87:
		return "B+"
	case points >= 83:
		return "B"
	case points >= 80:
		return "B-"
	case points >= 77:
		return "C+"
	case points >= 73:
		return "C"
	case points >= 70:
		return "C-"
	case points >= 67:
		return "D+"
	case points >= 63:
		return "D"
	case points >= 60:
		return "D-"
	default:
		return "F"
	}
}

// AnalyzeSkill scores a skill at the given directory path.
func AnalyzeSkill(path string) (*SkillReport, error) {
	result, err := parser.ParseSkillDir(path)
	if err != nil {
		return nil, fmt.Errorf("parse skill at %s: %w", path, err)
	}

	md := parser.AnalyzeMarkdown(result.Body)

	// Score each dimension
	structure := ScoreStructure(path, result)
	description := ScoreDescription(result)
	content := ScoreContent(path, result, md)
	engineering := ScoreEngineering(path, result)
	packaging := ScorePackaging(path, md)

	// Compute weighted overall score
	overall := structure.Points*WeightStructure +
		description.Points*WeightDescription +
		content.Points*WeightContent +
		engineering.Points*WeightEngineering +
		packaging.Points*WeightPackaging

	// Extract skill name and description
	name := filepath.Base(path)
	desc := ""
	if result.Frontmatter != nil {
		if result.Frontmatter.Name != "" {
			name = result.Frontmatter.Name
		}
		desc = result.Frontmatter.Description
	}

	// Generate suggestions from failed checks
	suggestions := generateSuggestions(structure, description, content, engineering, packaging)

	return &SkillReport{
		Path:        path,
		Name:        name,
		Desc:        desc,
		Overall:     Score{Points: overall, Grade: PointsToGrade(overall)},
		Structure:   structure,
		Description: description,
		Content:     content,
		Engineering: engineering,
		Packaging:   packaging,
		Suggestions: suggestions,
	}, nil
}

func generateSuggestions(dims ...DimensionScore) []Suggestion {
	var suggestions []Suggestion
	for _, dim := range dims {
		for _, c := range dim.Checks {
			if c.Passed {
				continue
			}
			priority := "low"
			if c.Weight >= 25 {
				priority = "high"
			} else if c.Weight >= 15 {
				priority = "medium"
			}
			suggestions = append(suggestions, Suggestion{
				Priority: priority,
				Message:  c.Detail,
			})
		}
	}
	return suggestions
}
