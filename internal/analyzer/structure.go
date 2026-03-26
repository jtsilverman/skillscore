package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jtsilverman/skillscore/internal/parser"
)

var (
	validNameRe   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	reservedWords = []string{"anthropic", "claude"}
	backslashRe   = regexp.MustCompile(`\\[a-zA-Z]`)
)

// ScoreStructure evaluates the structural quality of a skill.
func ScoreStructure(dir string, result *parser.ParseResult) DimensionScore {
	var checks []Check

	// Check 1: SKILL.md exists
	skillPath := filepath.Join(dir, "SKILL.md")
	_, err := os.Stat(skillPath)
	checks = append(checks, Check{
		Name:   "skill_md_exists",
		Passed: err == nil,
		Weight: 25,
		Detail: checkDetail(err == nil, "SKILL.md exists at expected path", "SKILL.md not found"),
	})

	// Check 2: Has valid frontmatter
	hasFM := result.HasFrontmatter && result.ParseError == nil
	detail := "Valid YAML frontmatter present"
	if !result.HasFrontmatter {
		detail = "No YAML frontmatter found (missing --- delimiters)"
	} else if result.ParseError != nil {
		detail = fmt.Sprintf("Frontmatter parse error: %v", result.ParseError)
	}
	checks = append(checks, Check{
		Name:   "valid_frontmatter",
		Passed: hasFM,
		Weight: 25,
		Detail: detail,
	})

	// Check 3: Name field valid
	nameValid := false
	nameDetail := "No name field"
	if result.Frontmatter != nil && result.Frontmatter.Name != "" {
		name := result.Frontmatter.Name
		switch {
		case len(name) > 64:
			nameDetail = fmt.Sprintf("Name too long (%d chars, max 64)", len(name))
		case !validNameRe.MatchString(name):
			nameDetail = "Name must be lowercase letters, numbers, and hyphens only"
		case containsReserved(name):
			nameDetail = fmt.Sprintf("Name contains reserved word (found in %q)", name)
		default:
			nameValid = true
			nameDetail = fmt.Sprintf("Name %q is valid", name)
		}
	}
	checks = append(checks, Check{
		Name:   "name_valid",
		Passed: nameValid,
		Weight: 20,
		Detail: nameDetail,
	})

	// Check 4: Description present
	hasDesc := result.Frontmatter != nil && strings.TrimSpace(result.Frontmatter.Description) != ""
	checks = append(checks, Check{
		Name:   "description_present",
		Passed: hasDesc,
		Weight: 20,
		Detail: checkDetail(hasDesc, "Description field is present", "Description field is missing or empty"),
	})

	// Check 5: No Windows-style backslash paths
	fullContent := result.Body
	if result.Frontmatter != nil {
		fullContent = result.Frontmatter.Description + "\n" + result.Body
	}
	hasBackslash := backslashRe.MatchString(fullContent)
	checks = append(checks, Check{
		Name:   "no_backslash_paths",
		Passed: !hasBackslash,
		Weight: 10,
		Detail: checkDetail(!hasBackslash, "No Windows-style backslash paths found", "Windows-style backslash paths detected; use forward slashes"),
	})

	return computeDimension(checks)
}

func containsReserved(name string) bool {
	lower := strings.ToLower(name)
	for _, word := range reservedWords {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

func checkDetail(passed bool, passMsg, failMsg string) string {
	if passed {
		return passMsg
	}
	return failMsg
}

func computeDimension(checks []Check) DimensionScore {
	totalWeight := 0
	earned := 0
	for _, c := range checks {
		totalWeight += c.Weight
		if c.Passed {
			earned += c.Weight
		}
	}
	points := 0.0
	if totalWeight > 0 {
		points = float64(earned) / float64(totalWeight) * 100
	}
	return DimensionScore{
		Score: Score{
			Points: points,
			Grade:  PointsToGrade(points),
		},
		Checks: checks,
	}
}
