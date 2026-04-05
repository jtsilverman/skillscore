package indexer

import "testing"

func TestCategorize(t *testing.T) {
	tests := []struct {
		name        string
		skillName   string
		description string
		want        string
	}{
		// Direct keyword matches (first-match-wins ordering)
		{"security keyword in name", "security-scanner", "", "Security"},
		{"security keyword in desc", "my-tool", "checks for xss vulnerabilities", "Security"},
		{"testing keyword", "unit-test-helper", "", "Testing"},
		{"testing keyword in desc", "checker", "runs pytest suites", "Testing"},
		{"devops keyword", "docker-deploy", "", "DevOps"},
		{"devops keyword k8s", "k8s-manager", "", "DevOps"},
		{"data keyword", "sql-migrator", "", "Data"},
		{"data keyword in desc", "loader", "loads csv files into postgres", "Data"},
		{"ai/ml keyword", "llm-agent", "", "AI/ML"},
		{"ai/ml keyword in desc", "helper", "uses openai embeddings", "AI/ML"},
		{"frontend keyword", "react-components", "", "Frontend"},
		{"backend keyword", "api-server", "", "Backend"},
		{"backend keyword grpc", "service", "exposes grpc endpoints", "Backend"},
		{"mobile keyword", "flutter-app", "", "Mobile"},
		{"cli keyword", "terminal-tool", "", "CLI"},
		{"documentation keyword", "docs-generator", "", "Documentation"},
		{"content keyword", "blog-writer", "", "Content"},
		{"code quality keyword", "lint-rules", "", "Code Quality"},
		{"git keyword", "github-helper", "", "Git"},
		{"finance keyword", "trading-bot", "", "Finance"},
		{"game dev keyword", "unity-helper", "", "Game Dev"},
		{"research keyword", "arxiv-crawler", "crawls research papers", "Research"},

		// Case insensitivity
		{"uppercase name", "DOCKER-DEPLOY", "", "DevOps"},
		{"mixed case desc", "tool", "Uses Machine Learning models", "AI/ML"},

		// First-match-wins: "test" matches Testing before other categories
		{"first match wins", "test-api-deploy", "integration test for deployment", "Testing"},

		// Edge cases
		{"empty input", "", "", "Other"},
		{"no matching category", "my-cool-thing", "does something neat", "Other"},
		{"whitespace only desc", "gadget", "   ", "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categorize(tt.skillName, tt.description)
			if got != tt.want {
				t.Errorf("categorize(%q, %q) = %q, want %q",
					tt.skillName, tt.description, got, tt.want)
			}
		})
	}
}
