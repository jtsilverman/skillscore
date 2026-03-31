# SkillScore Directory Expansion

## Overview

Expand the SkillScore directory from 3 GitHub sources to 12+, add Jake's own skills as a local source, and add a community submission flow via GitHub Issue templates. The current directory indexes maybe a few dozen skills. After this expansion it should index 1,500+ skills from the Claude Code ecosystem, making it the most comprehensive quality-scored skill directory.

## Scope

### Phase 1 — Core: More Sources + Local Skills

- Switch `discoverSkills` from Contents API to Git Tree API (Contents API caps at 1000 items, antigravity has 1,340 skills)
- Add 9 new GitHub sources to `sources.go` (verified active, contain SKILL.md files)
- Add local source support to the indexer (score skills from a local directory, merge into the same index)
- Add `--local` flag to `skillscore index` to include a local skill directory
- Add Jake's 24 Rock skills as a local source in the GitHub Actions workflow
- Re-run indexer, verify scored-index.json populates with 1,500+ skills
- Deploy to GitHub Pages

### Phase 2 — Full Product: Submission Flow + Scale

- GitHub Issue template for skill submissions (repo URL, skill path, description)
- "Submit Your Skill" button on the web UI linking to the issue template
- Web UI improvements for 1,500+ skills: virtual scrolling or pagination, category tags
- Deduplication: same skill indexed from multiple sources (e.g., a skill in both awesome-lists) should appear once with the best score
- Source badges on skill cards (Official, Community, Local)
- Index metadata: generation timestamp, per-source counts, total indexed vs skipped

### Phase 3 — Stretch

- Category/tag extraction from skill descriptions (auto-classify into dev-tools, content, research, etc.)
- `skillscore submit <owner/repo/path>` CLI command that opens a pre-filled GitHub Issue
- RSS/Atom feed of newly indexed skills
- Skill comparison view (side-by-side scoring of two skills)

### Not building (any phase)

- User accounts, auth, or social features
- Skill installation or package management
- Real-time indexing (webhook-driven). Daily cron is sufficient.
- Backend server. Stays fully static (GitHub Pages + JSON).

### Ship target

GitHub Pages (existing). Update GitHub Actions daily cron to include new sources.

## Stack

No stack changes. Same Go CLI + vanilla HTML/JS/CSS + GitHub Actions. The only new dependency consideration is the Git Tree API endpoint (no library needed, just different URL).

## Architecture

### Changes to existing files

```
internal/indexer/
├── sources.go          (modify — add 9 new sources)
├── indexer.go          (modify — add local source support, Tree API discovery, dedup)
internal/github/
├── fetch.go            (modify — add TreeDiscover function using Git Tree API)
cmd/skillscore/
├── main.go             (modify — add --local flag to index command)
web/
├── index.html          (modify — add Submit button, pagination, source badges)
├── app.js              (modify — pagination, dedup display, category filters)
├── style.css           (modify — source badges, pagination styles)
.github/
├── workflows/index.yml (modify — add local skills step, checkout Rock skills)
├── ISSUE_TEMPLATE/
│   └── submit-skill.yml (create — structured issue template)
```

### New GitHub Sources (verified 2026-03-30)

| # | Repo | Path | Skills | Stars | Label |
|---|------|------|--------|-------|-------|
| 1 | `anthropics/skills` | `skills` | ~17 | 107K | Anthropic Official |
| 2 | `travisvn/awesome-claude-skills` | `` | varies | 10K | Awesome Claude Skills |
| 3 | `slavingia/skills` | `` | varies | 6K | Slavingia Skills |
| 4 | `affaan-m/everything-claude-code` | `skills` | 136 | 119K | Everything Claude Code |
| 5 | `sickn33/antigravity-awesome-skills` | `skills` | 1,340 | 29K | Antigravity Skills |
| 6 | `ComposioHQ/awesome-claude-skills` | `` | ~33 | 49K | Composio Skills |
| 7 | `alirezarezvani/claude-skills` | `` | ~192 | 8K | Rezvani Skills |
| 8 | `Jeffallan/claude-skills` | `` | ~66 | 7.5K | Jeffallan Skills |
| 9 | `VoltAgent/awesome-agent-skills` | `` | ~1,000 | 13K | VoltAgent Skills |
| 10 | `Donchitos/Claude-Code-Game-Studios` | `` | ~36 | 7.5K | Game Studios |
| 11 | `Orchestra-Research/AI-Research-SKILLs` | `` | varies | 6K | AI Research Skills |
| 12 | `hesreallyhim/awesome-claude-code` | `` | varies | 34K | Awesome Claude Code |

Plus local: Jake's 24 skills from `~/Rock/skills/`

### Local Source Data Model

```go
type LocalSource struct {
    Path  string // Absolute path to directory containing skill subdirs
    Label string // e.g., "jtsilverman (local)"
    Repo  string // GitHub repo to link to (optional, for display)
}
```

### Deduplication Strategy

Skills may appear in multiple awesome-lists. Dedup by:
1. Normalize skill name (lowercase, trim)
2. If same name + same repo: keep highest score
3. If same name + different repos: keep both (different implementations)

### Git Tree API Discovery

```go
// TreeDiscover returns all paths containing SKILL.md in a repo
// Uses /git/trees/:sha?recursive=1 (no 1000-item cap)
func TreeDiscover(owner, repo string) ([]string, error)
```

Returns paths like `skills/deploy-app` for each directory containing a SKILL.md.

### Issue Template (submit-skill.yml)

```yaml
name: Submit a Skill
description: Submit a Claude Code skill to be indexed in the directory
labels: ["submission"]
body:
  - type: input
    id: repo
    attributes:
      label: GitHub Repository
      description: "owner/repo format"
      placeholder: "username/my-skill"
    validations:
      required: true
  - type: input
    id: path
    attributes:
      label: Skill Path
      description: "Path to skill directory within the repo (leave empty if root)"
      placeholder: "skills/my-skill"
  - type: textarea
    id: description
    attributes:
      label: Description
      description: "Brief description of what the skill does"
```

## Task List

## Phase 1: Core

### 1A: Git Tree API Discovery

#### Task 1A.1: Add TreeDiscover function
**Files:** `internal/github/fetch.go` (modify), `internal/github/fetch_test.go` (modify)
**Do:** Add `TreeDiscover(owner, repo string) ([]string, error)` function. Uses `GET /repos/{owner}/{repo}/git/trees/main?recursive=1`. Parses response, filters for entries where path ends with `/SKILL.md`, returns parent directory paths. Falls back to `HEAD` if `main` branch 404s. Respects `GITHUB_TOKEN` env var. Handle `truncated: true` response by logging a warning (tree API truncates at 100K entries, which no skill repo should hit).
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/github/ -run TestTreeDiscover -v`

#### Task 1A.2: Switch indexer to Tree API
**Files:** `internal/indexer/indexer.go` (modify)
**Do:** Replace `discoverSkills` implementation. Instead of calling `gh.ListDir` (Contents API, 1000 cap), call `gh.TreeDiscover(src.Owner, src.Repo)` then filter results to those under `src.Path` prefix. This handles repos with 1000+ skills (antigravity has 1,340). Keep the existing `ListDir` approach as fallback if TreeDiscover fails (some repos may not have default branch named `main` or `HEAD`).
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/indexer/ -v`

### 1B: Expand GitHub Sources

#### Task 1B.1: Add new sources to sources.go
**Files:** `internal/indexer/sources.go` (modify)
**Do:** Add 9 new entries to `DefaultSources`. For each, set Owner, Repo, Path (empty string if skills are at root, "skills" if in a subdirectory), and Label. Full list: affaan-m/everything-claude-code (path: "skills"), sickn33/antigravity-awesome-skills (path: "skills"), ComposioHQ/awesome-claude-skills (path: ""), alirezarezvani/claude-skills (path: ""), Jeffallan/claude-skills (path: ""), VoltAgent/awesome-agent-skills (path: ""), Donchitos/Claude-Code-Game-Studios (path: ""), Orchestra-Research/AI-Research-SKILLs (path: ""), hesreallyhim/awesome-claude-code (path: ""). Verify each path by checking the actual repo structure (some may have skills in a subdirectory).
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go build ./cmd/skillscore/ && echo "builds ok"`

### 1C: Local Source Support

#### Task 1C.1: Add local indexing to indexer
**Files:** `internal/indexer/indexer.go` (modify)
**Do:** Add `LocalSource` struct with `Path`, `Label`, and `Repo` fields. Add `RunIndexWithLocal(sources []Source, localSources []LocalSource, outputPath string) error`. For each local source, walk the directory tree finding SKILL.md files (reuse the same `findSkillDirs` pattern from main.go's scan command). Score each skill using `analyzer.AnalyzeSkill`. Set `Repo` field on IndexEntry from LocalSource.Repo (for GitHub links on the web UI). Set `Source` from LocalSource.Label. Merge local entries into the same Index output alongside GitHub entries. Update `RunIndex` to accept an optional `[]LocalSource` parameter (or add a new exported function).
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/indexer/ -run TestLocalSource -v`

#### Task 1C.2: Add --local flag to CLI
**Files:** `cmd/skillscore/main.go` (modify)
**Do:** Add `--local` flag to the `index` subcommand. Accepts a path to a local skills directory. When provided, creates a `LocalSource{Path: localPath, Label: "Local Skills"}` and passes it to the indexer. Support `--local-repo` flag to set the GitHub repo link (e.g., `--local-repo jtsilverman/rock-skills`). If `--local` is not provided, behavior is unchanged (GitHub-only indexing).
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go build ./cmd/skillscore/ && ./skillscore index --local ~/Rock/skills/ --local-repo jtsilverman --output /tmp/test-index.json && python3 -c "import json; d=json.load(open('/tmp/test-index.json')); local=[s for s in d['skills'] if 'Local' in s.get('source','')]; print(f'Local skills: {len(local)}')" && rm /tmp/test-index.json`

### 1D: Build & Deploy

#### Task 1D.1: Run full index and deploy
**Files:** `web/scored-index.json` (create/update), `.github/workflows/index.yml` (modify)
**Do:** Build the binary and run the full index: `./skillscore index --local ~/Rock/skills/ --local-repo jtsilverman --output web/scored-index.json`. Verify the JSON has 1,000+ skills. Update the GitHub Actions workflow to include Jake's skills: add a step that clones/copies the skills directory before running the indexer (or use a separate workflow that runs locally and pushes the JSON). Since Jake's skills are private (not on GitHub), the workflow should work without --local in CI, and Jake can run `skillscore index --local ~/Rock/skills/ --output web/scored-index.json` locally to include his skills. Update workflow to reflect this. Commit and push everything.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && cat web/scored-index.json | python3 -c "import json,sys; d=json.load(sys.stdin); print(f'Skills: {d[\"count\"]}, Sources: {len(d[\"sources\"])}')"` shows 1000+ skills and 12+ sources.

## Phase 2: Full Product

### 2A: Submission Flow

#### Task 2A.1: Create GitHub Issue template
**Files:** `.github/ISSUE_TEMPLATE/submit-skill.yml` (create)
**Do:** Create structured issue template with fields: repo (required, owner/repo format), path (optional, subdirectory), description (optional, what the skill does). Set label "submission". Add a note that submitted skills will be scored and added to the directory in the next daily index run.
**Validate:** File exists and is valid YAML: `cd /Users/rock/Rock/projects/skillscore && python3 -c "import yaml; yaml.safe_load(open('.github/ISSUE_TEMPLATE/submit-skill.yml'))" && echo "valid"`

#### Task 2A.2: Add Submit button to web UI
**Files:** `web/index.html` (modify), `web/style.css` (modify)
**Do:** Add a "Submit Your Skill" button in the header area, next to the stats. Links to `https://github.com/jtsilverman/skillscore/issues/new?template=submit-skill.yml`. Style it as a secondary action (outlined button, not competing with the main content).
**Validate:** Open `web/index.html` in browser. Submit button is visible and links to correct URL.

### 2B: Scale the Web UI

#### Task 2B.1: Pagination for 1,500+ skills
**Files:** `web/app.js` (modify), `web/style.css` (modify)
**Do:** With 1,500+ skills, rendering all at once will be slow. Add pagination: show 50 skills per page, add prev/next buttons and page indicator at the bottom. Preserve search and filter state across pages. When filters change, reset to page 1. Show "Showing X-Y of Z skills" in the stats bar.
**Validate:** Open `web/index.html` with a large scored-index.json. Page loads fast. Pagination controls work. Search resets to page 1.

#### Task 2B.2: Source badges on skill cards
**Files:** `web/app.js` (modify), `web/style.css` (modify)
**Do:** Add a small badge on each skill card showing its source (e.g., "Official", "Antigravity", "Local"). Use distinct colors: gold for Official (Anthropic), blue for curated collections, green for Local. The badge should be subtle and not compete with the grade badge.
**Validate:** Open `web/index.html`. Skills from different sources show different colored badges.

#### Task 2B.3: Deduplication in indexer
**Files:** `internal/indexer/indexer.go` (modify)
**Do:** After collecting all entries, deduplicate. Key: lowercase skill name + repo. If same name appears from same repo via different sources, keep the entry with the highest score and merge source labels. If same name from different repos, keep both. Add `Deduped` count to Index metadata. Log dedup stats to stderr.
**Validate:** `cd /Users/rock/Rock/projects/skillscore && export PATH="$HOME/go-sdk/go/bin:$PATH" && go test ./internal/indexer/ -run TestDedup -v`

### 2C: Ship

#### Task 2C.1: Update README, commit, push, verify
**Files:** `README.md` (modify)
**Do:** Update README to reflect expanded directory: new source count, total skills indexed, mention submission flow. Update the "what I learned" section if needed. Commit all changes, push, verify GitHub Pages deploys with the full index. Tag a new release if appropriate.
**Validate:** `gh repo view jtsilverman/skillscore --json name` succeeds. GitHub Pages site loads with 1,000+ skills.

## Phase 3: Stretch

### 3A: Categories

#### Task 3A.1: Auto-categorize skills by description
**Files:** `internal/indexer/indexer.go` (modify), `web/app.js` (modify), `web/index.html` (modify)
**Do:** Extract category from skill description using keyword matching (dev-tools, content, research, security, automation, travel, finance, etc.). Add `category` field to IndexEntry. Add category filter dropdown to web UI.
**Validate:** Skills in the index have category fields. Web UI filters by category.

#### Task 3A.2: CLI submit command
**Files:** `cmd/skillscore/main.go` (modify)
**Do:** Add `skillscore submit <owner/repo/path>` that opens a pre-filled GitHub Issue URL in the default browser. Score the skill first and include the score in the issue body.
**Validate:** `./skillscore submit anthropics/skills/skills/deploy-app` opens browser with pre-filled issue.

## The One Hard Thing

**Scaling the indexer to 1,500+ skills without hitting GitHub API rate limits.**

Why it's hard:
- The current indexer fetches each skill individually via raw.githubusercontent.com (SKILL.md) + Contents API (supporting files). With 1,340 skills in antigravity alone, that's 2,680+ API calls for one source.
- Unauthenticated rate limit: 60 req/hr. Even with GITHUB_TOKEN: 5,000 req/hr.
- The daily GitHub Actions cron gets a GITHUB_TOKEN automatically, but 12 sources x hundreds of skills = potentially 5,000+ calls per run.

Proposed approach:
- Use Git Tree API for discovery (1 call per repo instead of N calls per directory level)
- For skill fetching, raw.githubusercontent.com doesn't count against API rate limits (it's a CDN). Only the directory listing calls count.
- With Tree API discovery, the bottleneck becomes fetching individual SKILL.md files. But raw.githubusercontent.com is unmetered.
- For supporting files (needed for engineering/packaging scoring), use Contents API but cache aggressively. Most skills are just SKILL.md with no supporting files, so the actual API call count should be manageable.
- Add request counting and early termination if approaching rate limit.

Fallback:
- If rate limits are still an issue, index sources sequentially across multiple cron runs (3-4 sources per day, full cycle every 3 days). The index JSON would merge incrementally.

## Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| GitHub API rate limits with 1,500+ skills | High | Tree API for discovery (1 call/repo), raw.githubusercontent.com for content (unmetered). Add rate limit monitoring. |
| Contents API 1000-item cap truncates large repos | High | Already mitigated: switching to Git Tree API. |
| Some source repos restructure or go offline | Medium | Indexer already handles fetch errors gracefully (skip + continue). Dead sources logged but don't break the run. |
| Jake's skills are private (not on GitHub) | Low | Local source support runs on Jake's machine. CI workflow indexes GitHub-only. Jake runs local index manually or via a separate workflow. |
| Web UI performance with 1,500+ skills | Medium | Pagination (50/page). No framework needed, vanilla JS handles this fine. |
| Duplicate skills across awesome-lists | Medium | Dedup by name+repo. Phase 2 task. |
