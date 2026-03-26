# SkillScore

## Overview

A Go CLI that scores Claude Code skills against Anthropic's official best practices. Point it at a local skill directory or a GitHub repo and get a detailed quality report with actionable feedback. The twist: while existing registries (60k+ skills) sort by GitHub stars, nobody actually analyzes whether skills follow the published authoring guidelines. SkillScore turns Anthropic's qualitative checklist into a quantitative scoring algorithm.

## Scope

**Timebox:** 2-3 days

**Building:**
- Go CLI (`skillscore`) that analyzes local skill directories
- Quality scoring engine with 5 dimensions (structure, description, content, engineering, packaging)
- GitHub mode: fetch and score any `owner/repo` or `owner/repo/path/to/skill`
- `skillscore index` command: crawl curated GitHub skill repos, score each, generate `scored-index.json`
- Static HTML/JS web directory (GitHub Pages): browse, search, filter, sort skills by quality score
- GitHub Actions cron to re-index daily
- Human-readable terminal report with color-coded grades (A-F)
- JSON output mode for piping/automation
- `skillscore scan <dir>` to batch-score all skills in a directory tree

**Not building:**
- Skill installation/package management (commoditized by existing tools)
- User accounts, ratings, or social features
- npm-style popularity/maintenance scoring (requires download tracking infra)
- Bubbletea TUI (stretch goal if time permits)

**Ship target:** GitHub release with prebuilt binaries (goreleaser), `go install` support, GitHub Pages directory

## Stack

- **Language:** Go 1.24 (stdlib-heavy, minimal dependencies)
- **TUI:** Charm stack (bubbletea, lipgloss, bubbles)
- **YAML parsing:** `gopkg.in/yaml.v3`
- **Markdown parsing:** `github.com/yuin/goldmark` (AST analysis)
- **GitHub API:** `net/http` + GitHub REST API (no SDK needed)
- **Why Go:** Fast single binary, cross-platform, good for CLI tools. Adds real Go to Jake's portfolio (browser-bench shows as HTML on GitHub). Bubbletea TUI is a visual differentiator for demos.

## Architecture

```
skillscore/
├── cmd/
│   └── skillscore/
│       └── main.go            # CLI entrypoint, flag parsing
├── internal/
│   ├── analyzer/
│   │   ├── analyzer.go        # Orchestrates all scoring dimensions
│   │   ├── structure.go       # Frontmatter + file structure checks
│   │   ├── description.go     # Description quality heuristics
│   │   ├── content.go         # Markdown body analysis
│   │   ├── engineering.go     # Script/code quality checks
│   │   └── packaging.go       # File organization, size, refs
│   ├── parser/
│   │   ├── frontmatter.go     # YAML frontmatter extraction
│   │   └── markdown.go        # Goldmark AST helpers
│   ├── github/
│   │   └── fetch.go           # Fetch skill from GitHub repo
│   ├── report/
│   │   ├── terminal.go        # Color-coded terminal output
│   │   └── json.go            # JSON output formatter
│   └── indexer/
│       ├── indexer.go         # Crawl curated repos, score all skills
│       └── sources.go         # Curated list of skill repo URLs
├── web/
│   ├── index.html             # Static directory page
│   ├── app.js                 # Client-side search/filter/sort
│   └── style.css              # Styling
├── testdata/
│   ├── good-skill/            # High-scoring test skill
│   ├── bad-skill/             # Low-scoring test skill
│   └── edge-cases/            # Partial, malformed, etc.
├── .github/
│   └── workflows/
│       ├── release.yml        # Goreleaser on tag push
│       └── index.yml          # Daily cron: index → score → deploy Pages
├── go.mod
├── go.sum
├── README.md
└── .goreleaser.yml
```

### Data Model

```go
type SkillReport struct {
    Path        string          `json:"path"`
    Name        string          `json:"name"`
    Overall     Score           `json:"overall"`     // Weighted composite
    Structure   DimensionScore  `json:"structure"`
    Description DimensionScore  `json:"description"`
    Content     DimensionScore  `json:"content"`
    Engineering DimensionScore  `json:"engineering"`
    Packaging   DimensionScore  `json:"packaging"`
    Suggestions []Suggestion    `json:"suggestions"`
}

type Score struct {
    Points float64 `json:"points"`  // 0-100
    Grade  string  `json:"grade"`   // A, B, C, D, F
}

type DimensionScore struct {
    Score
    Checks []Check `json:"checks"`
}

type Check struct {
    Name    string `json:"name"`
    Passed  bool   `json:"passed"`
    Weight  int    `json:"weight"`  // Relative importance within dimension
    Detail  string `json:"detail"`  // Why it passed/failed
}

type Suggestion struct {
    Priority string `json:"priority"` // "high", "medium", "low"
    Message  string `json:"message"`
}
```

### Scoring Algorithm

Five dimensions, weighted by importance:

| Dimension | Weight | What it measures |
|-----------|--------|-----------------|
| Structure (25%) | Frontmatter validity, file organization | Does it follow the SKILL.md spec? |
| Description (25%) | Trigger specificity, third-person, keyword richness | Will Claude discover it? |
| Content (25%) | Conciseness, examples, workflows, progressive disclosure | Will Claude execute it well? |
| Engineering (15%) | Script quality, error handling, deps listed | Are bundled scripts reliable? |
| Packaging (10%) | README, reasonable size, no bloat, reference depth | Is it well-organized? |

#### Structure Checks (25%)
- Has valid YAML frontmatter (required `---` delimiters)
- `name` field present and valid (lowercase, hyphens, <=64 chars, no reserved words)
- `description` field present and non-empty
- SKILL.md exists as entrypoint
- No Windows-style backslash paths

#### Description Checks (25%)
- Includes action verb (what it does)
- Includes trigger context ("Use when...", "Use for...", or similar)
- Written in third person (no "I can", "You can")
- Under 1024 characters
- Specific keywords (not vague like "helps with stuff")
- No reserved words ("anthropic", "claude" in name)

#### Content Checks (25%)
- SKILL.md body under 500 lines
- Has at least one code example or template
- Has workflow steps (numbered or bulleted sequence)
- No deeply nested references (>1 level from SKILL.md)
- Consistent terminology (flag repeated synonyms)
- Token efficiency: ratio of unique information to total tokens
- Has validation/feedback loop for multi-step workflows

#### Engineering Checks (15%)
- Scripts have error handling (try/except, error returns)
- No magic numbers without comments
- Dependencies explicitly listed
- Scripts use forward-slash paths
- Execution intent clear ("Run X" vs "See X")
- (Only scored if skill contains scripts; otherwise full marks)

#### Packaging Checks (10%)
- Total skill directory under 50KB (excluding binary assets)
- Reference files have descriptive names (not `doc1.md`)
- Supporting files organized in subdirectories
- No unnecessary files (`.DS_Store`, `node_modules`, etc.)
- Reference files linked from SKILL.md

### CLI Interface

```bash
# Score a single local skill
skillscore ./my-skill/
skillscore ~/.claude/skills/deploy/

# Score a GitHub skill
skillscore github anthropics/skills/skills/skill-creator

# Score all skills in a directory
skillscore scan ~/.claude/skills/

# JSON output
skillscore --json ./my-skill/

# Quiet mode (just the grade)
skillscore --quiet ./my-skill/
# Output: B+ (78/100)
```

## Task List

## Phase 1: Foundation

### Task 1.1: Initialize Go module and project structure
**Files:** `go.mod` (create), `cmd/skillscore/main.go` (create), `internal/analyzer/analyzer.go` (create)
**Do:** Initialize Go module as `github.com/jtsilverman/skillscore`. Create directory structure. Stub out main.go with cobra-less flag parsing (stdlib `flag` package). Create analyzer.go with `AnalyzeSkill(path string) (*SkillReport, error)` signature and data types.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go build ./cmd/skillscore/ && echo "ok"`

### Task 1.2: YAML frontmatter parser
**Files:** `internal/parser/frontmatter.go` (create), `internal/parser/frontmatter_test.go` (create), `testdata/good-skill/SKILL.md` (create), `testdata/bad-skill/SKILL.md` (create)
**Do:** Parse YAML frontmatter between `---` delimiters. Extract name, description, and all optional fields. Create test skills: good-skill has proper frontmatter + body, bad-skill has missing/invalid fields. Handle edge cases: no frontmatter, empty frontmatter, malformed YAML.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/parser/ -v`

### Task 1.3: Markdown AST parser
**Files:** `internal/parser/markdown.go` (create), `internal/parser/markdown_test.go` (create)
**Do:** Use goldmark to parse markdown body (after frontmatter). Extract: heading structure, code blocks (count + languages), list items (ordered/unordered), link references to other files, line count, word count. Expose as a `MarkdownAnalysis` struct.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/parser/ -v`

## Phase 2: Scoring Engine

### Task 2.1: Structure scorer
**Files:** `internal/analyzer/structure.go` (create), `internal/analyzer/structure_test.go` (create)
**Do:** Implement all structure checks: valid frontmatter, name field validation (lowercase, hyphens, <=64 chars, no reserved words "anthropic"/"claude"), description present, SKILL.md exists, no backslash paths. Each check returns a Check struct with pass/fail, weight, and detail message. Score is weighted sum normalized to 0-100.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/analyzer/ -run TestStructure -v`

### Task 2.2: Description scorer
**Files:** `internal/analyzer/description.go` (create), `internal/analyzer/description_test.go` (create)
**Do:** Implement description checks: has action verb (maintain a verb list), has trigger context (regex for "Use when", "Use for", "Trigger when" etc.), third-person check (flag "I can", "You can", "I will"), under 1024 chars, specificity check (flag vague terms like "helps with", "does stuff", "general purpose"), no reserved words in name.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/analyzer/ -run TestDescription -v`

### Task 2.3: Content scorer
**Files:** `internal/analyzer/content.go` (create), `internal/analyzer/content_test.go` (create)
**Do:** Implement content checks: body under 500 lines, has code examples (check code block count > 0), has workflow steps (detect ordered lists or "Step N" patterns), reference depth check (parse links, check if linked files also link to other files), terminology consistency (extract key terms, flag when >2 synonyms used for same concept using a basic synonym map).
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/analyzer/ -run TestContent -v`

### Task 2.4: Engineering scorer
**Files:** `internal/analyzer/engineering.go` (create), `internal/analyzer/engineering_test.go` (create)
**Do:** Implement engineering checks: scan for script files (*.py, *.sh, *.js, *.ts) in skill directory. If none found, return full marks. If found: check for error handling patterns (try/except, if err != nil, try/catch), check for magic numbers (bare numeric literals without adjacent comments), check deps listed in SKILL.md, check forward-slash paths, check execution intent ("Run" vs "See" before script references).
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/analyzer/ -run TestEngineering -v`

### Task 2.5: Packaging scorer
**Files:** `internal/analyzer/packaging.go` (create), `internal/analyzer/packaging_test.go` (create)
**Do:** Implement packaging checks: total directory size under 50KB (excluding common binary patterns), descriptive file names (flag generic names like doc1.md, file.md, temp.md), subdirectory organization (scripts/, reference/, examples/), no junk files (.DS_Store, node_modules, __pycache__), reference files linked from SKILL.md.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/analyzer/ -run TestPackaging -v`

### Task 2.6: Score aggregation and grading
**Files:** `internal/analyzer/analyzer.go` (modify)
**Do:** Wire all 5 scorers into `AnalyzeSkill`. Apply dimension weights (structure 25%, description 25%, content 25%, engineering 15%, packaging 10%). Compute overall score 0-100. Map to grades: A (90-100), B (80-89), C (70-79), D (60-69), F (<60). Add +/- modifiers (e.g., B+ for 85-89). Generate prioritized suggestions from failed checks.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/analyzer/ -run TestAnalyzeSkill -v` (test against good-skill and bad-skill testdata)

## Phase 3: Output and GitHub

### Task 3.1: Terminal report renderer
**Files:** `internal/report/terminal.go` (create)
**Do:** Color-coded terminal output using lipgloss. Show: skill name, overall grade (large, colored), dimension breakdown (5 bars with scores), top suggestions. Green for A-B, yellow for C, red for D-F. Compact mode for `scan` (one line per skill). Verbose mode shows all individual checks.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go build ./cmd/skillscore/ && ./skillscore testdata/good-skill/ && ./skillscore testdata/bad-skill/` (visual check: good-skill scores high, bad-skill scores low)

### Task 3.2: JSON output and scan mode
**Files:** `internal/report/json.go` (create), `cmd/skillscore/main.go` (modify)
**Do:** JSON output with `--json` flag. Scan mode: `skillscore scan <dir>` walks directory tree finding SKILL.md files, scores each, outputs sorted table. Wire up all CLI flags: `--json`, `--quiet`, `--verbose`, `scan` subcommand.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go build ./cmd/skillscore/ && ./skillscore --json testdata/good-skill/ | jq .overall` (returns valid JSON with score)

### Task 3.3: GitHub fetch mode
**Files:** `internal/github/fetch.go` (create), `internal/github/fetch_test.go` (create)
**Do:** Parse `github owner/repo/path` argument. Use GitHub REST API to fetch SKILL.md and directory listing (no auth required for public repos). Download to temp directory, run analyzer, clean up. Handle: repo not found, SKILL.md not found, rate limiting (warn user). Support `github anthropics/skills/skills/skill-creator` format.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/github/ -v && go build ./cmd/skillscore/ && ./skillscore github anthropics/skills/skills/skill-creator`

## Phase 4: Web Directory and Indexer

### Task 4.1: Indexer: crawl and score skill repos
**Files:** `internal/indexer/sources.go` (create), `internal/indexer/indexer.go` (create), `internal/indexer/indexer_test.go` (create), `cmd/skillscore/main.go` (modify)
**Do:** Curated list of GitHub skill sources: anthropics/skills, travisvn/awesome-claude-skills, top repos from search. `skillscore index` command: for each source, use GitHub API to discover SKILL.md files recursively, fetch and score each, write `scored-index.json` with all results (skill name, repo, path, scores, grade, description). Support `GITHUB_TOKEN` env var for rate limits. Add `--output` flag for index file path.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go build ./cmd/skillscore/ && ./skillscore index --output testdata/test-index.json && cat testdata/test-index.json | jq '.skills | length'` (returns > 0 skills indexed)

### Task 4.2: Static web directory
**Files:** `web/index.html` (create), `web/app.js` (create), `web/style.css` (create)
**Do:** Single-page static site that loads `scored-index.json`. Features: search by name/description, filter by grade (A-F), sort by score/name/repo, click to expand dimension breakdown. Clean dark theme. Each skill card shows: name, grade badge, repo link, description, 5 dimension mini-bars. Mobile responsive. No build step (vanilla HTML/JS/CSS).
**Validate:** Open `web/index.html` in browser with a test `scored-index.json`. Search, filter, and sort all work. Skills display with grades and dimension breakdowns.

### Task 4.3: GitHub Actions for daily index + Pages deploy
**Files:** `.github/workflows/index.yml` (create)
**Do:** GitHub Actions workflow: runs daily at 6 AM UTC, checks out repo, builds skillscore, runs `skillscore index --output web/scored-index.json`, commits and deploys to GitHub Pages. Uses GITHUB_TOKEN for API access. Manual trigger option for on-demand reindex.
**Validate:** Workflow YAML is valid (act or manual review). Pages deployment config correct.

### Task 4.4: End-to-end integration tests
**Files:** `cmd/skillscore/main_test.go` (create), `testdata/edge-cases/` (create various edge case skills)
**Do:** Test full CLI pipeline: (1) score good-skill, verify A/B grade, (2) score bad-skill, verify D/F grade, (3) JSON output parses correctly, (4) scan mode finds all skills in testdata, (5) quiet mode outputs just grade, (6) edge cases: empty SKILL.md, no frontmatter, only frontmatter, huge file, deeply nested refs. (7) GitHub mode against a known public skill.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./... -v -count=1`

## Phase 5: Ship

### Task 5.1: README, goreleaser, deploy
**Files:** `README.md` (create), `.goreleaser.yml` (create), `.github/workflows/release.yml` (create)
**Do:** README with: problem statement ("60k+ skills, no quality signal"), live directory link, demo GIF placeholder, install instructions (go install + binary download), usage examples, scoring breakdown table, "what I learned" section. Goreleaser config for cross-platform binaries. GitHub Actions release workflow on tag push.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go build ./cmd/skillscore/ && ./skillscore --help` (shows usage) `&& cat README.md | head -5` (README exists)

### Task 5.2: Create repo, push, verify
**Files:** n/a (git operations)
**Do:** Create GitHub repo `jtsilverman/skillscore`. Push all code. Enable GitHub Pages. Verify README renders, directory loads. Tag v0.1.0. Verify CI runs.
**Validate:** `gh repo view jtsilverman/skillscore --json name,description` (repo exists)

## The One Hard Thing

**The quality scoring algorithm.** Turning Anthropic's qualitative best-practices guide into a quantitative scorer that produces useful, actionable grades.

Why it's hard:
- Many checks are subjective (is a description "specific enough"?). Need to find heuristics that correlate with actual quality without being gameable or producing false positives.
- Description quality analysis requires NLP-adjacent heuristics (verb detection, vague term identification, person detection) without importing a full NLP library.
- The synonym consistency check needs a curated map of common term clusters in the skill domain.
- Weighting matters a lot: bad weights produce scores that don't match human intuition. Need to calibrate against known good (Anthropic official) and known bad skills.

Proposed approach:
- Start with the official Anthropic checklist as ground truth (21 checks across their quality guide).
- Build each check as a standalone function with clear pass/fail logic.
- Calibrate weights by scoring Anthropic's own skills (should all get A) and deliberately bad test skills (should get D/F).
- Use word lists and regex patterns for NLP-adjacent checks (verb lists, vague term lists, person pronouns).

Fallback:
- If NLP heuristics produce too many false positives, simplify to structural-only checks (frontmatter, line count, file organization) which are objective and still valuable. Reframe as "structural quality" rather than "overall quality."

## Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Scoring algorithm feels arbitrary or gameable | High | Calibrate against Anthropic's own skills. Document every check with rationale. Allow users to see individual check results, not just the grade. |
| GitHub API rate limiting (60 req/hr unauthenticated) | Medium | Indexer requires GITHUB_TOKEN env var (5000 req/hr). CLI warns when approaching unauthenticated limit. |
| Indexer takes too long for many repos | Medium | Start with curated list of ~5 repos. Parallelize fetches with goroutines. Cache scored results between runs. |
| Skill format evolves after spec freeze | Low | Scoring checks are modular. New checks can be added without restructuring. |
