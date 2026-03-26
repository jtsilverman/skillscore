package analyzer

import (
	"regexp"
	"strings"

	"github.com/jtsilverman/skillscore/internal/parser"
)

var (
	// Action verbs that indicate what a skill does
	actionVerbs = []string{
		"deploy", "build", "test", "analyze", "generate", "create", "extract",
		"process", "monitor", "review", "scan", "validate", "format", "convert",
		"migrate", "optimize", "debug", "configure", "install", "manage",
		"run", "execute", "transform", "parse", "compile", "check", "fix",
		"update", "sync", "merge", "split", "filter", "search", "index",
		"score", "grade", "report", "export", "import", "fetch", "push",
		"pull", "publish", "release", "clean", "lint", "refactor",
	}

	// Trigger phrases that indicate when to use a skill
	triggerRe = regexp.MustCompile(`(?i)(use when|use for|trigger when|use this when|invoke when|run when|activate when|use if)`)

	// First/second person patterns that should be avoided
	personRe = regexp.MustCompile(`(?i)\b(i can|i will|i help|you can|you should|we can|my |your )\b`)

	// Vague terms that indicate low specificity
	vagueTerms = []string{
		"helps with stuff", "does stuff", "general purpose", "helps with things",
		"various tasks", "many things", "different things", "useful tool",
		"helper", "utility", "misc", "generic",
	}
)

// ScoreDescription evaluates the quality of the skill's description field.
func ScoreDescription(result *parser.ParseResult) DimensionScore {
	var checks []Check

	desc := ""
	if result.Frontmatter != nil {
		desc = result.Frontmatter.Description
	}

	// Check 1: Has action verb
	hasVerb := false
	if desc != "" {
		lower := strings.ToLower(desc)
		for _, verb := range actionVerbs {
			// Match verb at word boundaries (starts of words)
			if strings.Contains(lower, verb) {
				hasVerb = true
				break
			}
		}
	}
	checks = append(checks, Check{
		Name:   "has_action_verb",
		Passed: hasVerb,
		Weight: 20,
		Detail: checkDetail(hasVerb, "Description contains an action verb", "Description lacks an action verb describing what the skill does"),
	})

	// Check 2: Has trigger context
	hasTrigger := triggerRe.MatchString(desc)
	checks = append(checks, Check{
		Name:   "has_trigger_context",
		Passed: hasTrigger,
		Weight: 25,
		Detail: checkDetail(hasTrigger, "Description includes trigger context (e.g., 'Use when...')", "Add trigger context like 'Use when...' to help Claude discover this skill"),
	})

	// Check 3: Written in third person
	hasFirstSecond := personRe.MatchString(desc)
	checks = append(checks, Check{
		Name:   "third_person",
		Passed: !hasFirstSecond,
		Weight: 15,
		Detail: checkDetail(!hasFirstSecond, "Description uses third person (recommended)", "Description uses first/second person; rewrite in third person for better discovery"),
	})

	// Check 4: Under 1024 characters
	underLimit := len(desc) <= 1024
	checks = append(checks, Check{
		Name:   "under_char_limit",
		Passed: underLimit,
		Weight: 10,
		Detail: checkDetail(underLimit, "Description within 1024 char limit", "Description exceeds 1024 character limit"),
	})

	// Check 5: Specificity (not vague)
	isVague := false
	if desc != "" {
		lower := strings.ToLower(desc)
		for _, term := range vagueTerms {
			if strings.Contains(lower, term) {
				isVague = true
				break
			}
		}
		// Also flag very short descriptions as likely vague
		if len(desc) < 20 {
			isVague = true
		}
	} else {
		isVague = true
	}
	checks = append(checks, Check{
		Name:   "specific_description",
		Passed: !isVague,
		Weight: 20,
		Detail: checkDetail(!isVague, "Description is specific and meaningful", "Description is too vague or short; be specific about what the skill does"),
	})

	// Check 6: Non-empty description
	hasDesc := strings.TrimSpace(desc) != ""
	checks = append(checks, Check{
		Name:   "description_non_empty",
		Passed: hasDesc,
		Weight: 10,
		Detail: checkDetail(hasDesc, "Description is present", "Description is missing; Claude needs this for skill discovery"),
	})

	return computeDimension(checks)
}
