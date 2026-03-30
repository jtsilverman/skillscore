package indexer

import "strings"

// Categories and their keyword signals. Order matters: first match wins.
// Keywords are matched against lowercase name + description.
var categoryRules = []struct {
	Name     string
	Keywords []string
}{
	{"Security", []string{"security", "vulnerability", "exploit", "pentest", "audit", "cve", "owasp", "threat", "malware", "firewall", "encrypt", "auth", "authentication", "authorization", "csrf", "xss", "injection", "sandbox", "hardening"}},
	{"Testing", []string{"test", "testing", "spec", "tdd", "bdd", "coverage", "assertion", "mock", "fixture", "e2e", "integration test", "unit test", "pytest", "jest", "vitest", "playwright", "cypress"}},
	{"DevOps", []string{"deploy", "ci/cd", "cicd", "pipeline", "docker", "kubernetes", "k8s", "terraform", "ansible", "nginx", "aws", "gcp", "azure", "cloud", "infrastructure", "helm", "vercel", "railway", "netlify", "heroku", "monitoring", "observability", "logging", "prometheus", "grafana"}},
	{"Data", []string{"data", "database", "sql", "postgres", "mysql", "mongodb", "redis", "elasticsearch", "analytics", "etl", "pipeline", "csv", "json", "parquet", "pandas", "dataframe", "bigquery", "warehouse", "migration", "schema"}},
	{"AI/ML", []string{"ai", "machine learning", "ml", "llm", "gpt", "claude", "openai", "anthropic", "embedding", "vector", "rag", "fine-tune", "neural", "model", "inference", "prompt", "agent", "chatbot", "nlp", "transformer", "diffusion", "training"}},
	{"Frontend", []string{"frontend", "react", "vue", "svelte", "angular", "next.js", "nextjs", "nuxt", "css", "tailwind", "html", "component", "ui", "ux", "design system", "responsive", "animation", "dom", "browser", "web app", "spa"}},
	{"Backend", []string{"backend", "api", "rest", "graphql", "grpc", "server", "endpoint", "middleware", "route", "express", "fastapi", "flask", "django", "spring", "microservice", "webhook", "socket", "websocket"}},
	{"Mobile", []string{"mobile", "ios", "android", "react native", "flutter", "swift", "kotlin", "expo", "app store", "google play"}},
	{"CLI", []string{"cli", "command line", "terminal", "shell", "bash", "zsh", "script", "automation", "cron", "task runner"}},
	{"Documentation", []string{"documentation", "docs", "readme", "changelog", "jsdoc", "docstring", "wiki", "guide", "tutorial", "comment"}},
	{"Content", []string{"content", "writing", "blog", "article", "social media", "twitter", "post", "newsletter", "seo", "copywriting", "marketing", "brand"}},
	{"Code Quality", []string{"lint", "format", "refactor", "review", "style", "convention", "prettier", "eslint", "rubocop", "clean code", "solid", "pattern", "best practice", "code smell", "complexity", "technical debt"}},
	{"Git", []string{"git", "github", "gitlab", "bitbucket", "commit", "branch", "merge", "pull request", "pr", "rebase", "version control"}},
	{"Finance", []string{"finance", "trading", "stock", "crypto", "bitcoin", "market", "portfolio", "investment", "payment", "stripe", "billing", "invoice"}},
	{"Game Dev", []string{"game", "unity", "unreal", "godot", "sprite", "physics", "multiplayer", "rpg", "platformer", "game dev"}},
	{"Research", []string{"research", "scraping", "crawl", "fetch", "browse", "search", "wikipedia", "arxiv", "paper", "survey", "analysis"}},
}

// categorize assigns a category to a skill based on its name and description.
func categorize(name, description string) string {
	text := strings.ToLower(name + " " + description)

	for _, rule := range categoryRules {
		for _, kw := range rule.Keywords {
			if strings.Contains(text, kw) {
				return rule.Name
			}
		}
	}

	return "Other"
}
