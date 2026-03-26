package analyzer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jtsilverman/skillscore/internal/parser"
)

var (
	scriptExts = []string{".py", ".sh", ".js", ".ts", ".rb"}

	// Error handling patterns by language
	errorPatterns = []*regexp.Regexp{
		regexp.MustCompile(`try\s*:`),                      // Python try:
		regexp.MustCompile(`except\s+`),                    // Python except
		regexp.MustCompile(`try\s*\{`),                     // JS/TS try {
		regexp.MustCompile(`catch\s*\(`),                   // JS/TS catch(
		regexp.MustCompile(`if\s+err\s*!=\s*nil`),          // Go
		regexp.MustCompile(`rescue\b`),                     // Ruby
		regexp.MustCompile(`set\s+-e`),                     // Bash set -e
		regexp.MustCompile(`\|\|\s*(exit|return|echo|die)`), // Bash || exit
	}

	// Magic number: bare integer/float literal on its own line or in assignment
	// (not in comments, strings, or common safe patterns like 0, 1, range)
	magicNumberRe = regexp.MustCompile(`(?m)^\s*[A-Z_]+\s*=\s*\d{2,}`)

	// Backslash path in scripts
	scriptBackslashRe = regexp.MustCompile(`\\[a-zA-Z]`)

	// Dependency install patterns
	depPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)pip install`),
		regexp.MustCompile(`(?i)npm install`),
		regexp.MustCompile(`(?i)requirements?\.txt`),
		regexp.MustCompile(`(?i)package\.json`),
		regexp.MustCompile(`(?i)import\s+\w+`),
	}

	// Execution intent patterns
	execIntentRe = regexp.MustCompile(`(?i)(run\s+|execute\s+|` + "`" + `[^` + "`" + `]+` + "`" + `)`)
)

// ScoreEngineering evaluates bundled script quality in a skill.
func ScoreEngineering(dir string, result *parser.ParseResult) DimensionScore {
	scripts := findScripts(dir)

	// No scripts = full marks (not applicable)
	if len(scripts) == 0 {
		return DimensionScore{
			Score: Score{Points: 100, Grade: "A+"},
			Checks: []Check{{
				Name:   "no_scripts",
				Passed: true,
				Weight: 100,
				Detail: "No bundled scripts; engineering checks not applicable",
			}},
		}
	}

	var checks []Check

	// Check 1: Scripts have error handling
	hasErrorHandling := false
	for _, s := range scripts {
		content, err := os.ReadFile(s)
		if err != nil {
			continue
		}
		for _, pat := range errorPatterns {
			if pat.Match(content) {
				hasErrorHandling = true
				break
			}
		}
		if hasErrorHandling {
			break
		}
	}
	checks = append(checks, Check{
		Name:   "has_error_handling",
		Passed: hasErrorHandling,
		Weight: 30,
		Detail: checkDetail(hasErrorHandling,
			"Scripts include error handling",
			"Scripts lack error handling (try/except, catch, set -e)"),
	})

	// Check 2: No magic numbers (constants without comments)
	hasMagic := false
	for _, s := range scripts {
		content, err := os.ReadFile(s)
		if err != nil {
			continue
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
				continue
			}
			if magicNumberRe.MatchString(line) {
				// Check if the line itself or the previous line has a comment
				hasComment := strings.Contains(line, "#") || strings.Contains(line, "//")
				if !hasComment && i > 0 {
					prevTrimmed := strings.TrimSpace(lines[i-1])
					hasComment = strings.HasPrefix(prevTrimmed, "#") || strings.HasPrefix(prevTrimmed, "//")
				}
				if !hasComment {
					hasMagic = true
					break
				}
			}
		}
		if hasMagic {
			break
		}
	}
	checks = append(checks, Check{
		Name:   "no_magic_numbers",
		Passed: !hasMagic,
		Weight: 15,
		Detail: checkDetail(!hasMagic,
			"Constants are documented",
			"Magic numbers detected; add comments explaining constant values"),
	})

	// Check 3: Dependencies mentioned in SKILL.md
	hasDepMention := false
	body := result.Body
	for _, pat := range depPatterns {
		if pat.MatchString(body) {
			hasDepMention = true
			break
		}
	}
	checks = append(checks, Check{
		Name:   "deps_documented",
		Passed: hasDepMention,
		Weight: 20,
		Detail: checkDetail(hasDepMention,
			"Dependencies are documented in SKILL.md",
			"Scripts may need dependencies; document required packages in SKILL.md"),
	})

	// Check 4: Forward-slash paths in scripts
	hasBackslash := false
	for _, s := range scripts {
		content, err := os.ReadFile(s)
		if err != nil {
			continue
		}
		if scriptBackslashRe.Match(content) {
			hasBackslash = true
			break
		}
	}
	checks = append(checks, Check{
		Name:   "forward_slash_paths",
		Passed: !hasBackslash,
		Weight: 15,
		Detail: checkDetail(!hasBackslash,
			"Scripts use forward-slash paths",
			"Windows-style backslash paths found in scripts; use forward slashes"),
	})

	// Check 5: Execution intent clear in SKILL.md
	hasExecIntent := execIntentRe.MatchString(body)
	checks = append(checks, Check{
		Name:   "execution_intent_clear",
		Passed: hasExecIntent,
		Weight: 20,
		Detail: checkDetail(hasExecIntent,
			"SKILL.md clearly indicates how to run scripts",
			"Add clear execution instructions (e.g., 'Run scripts/deploy.py')"),
	})

	return computeDimension(checks)
}

func findScripts(dir string) []string {
	var scripts []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		for _, se := range scriptExts {
			if ext == se {
				scripts = append(scripts, path)
				break
			}
		}
		return nil
	})
	return scripts
}
