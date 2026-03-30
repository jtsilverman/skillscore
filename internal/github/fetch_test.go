package github

import (
	"os"
	"testing"
)

func TestTreeDiscover(t *testing.T) {
	// Skip if no network access or rate limited
	if os.Getenv("SKIP_NETWORK_TESTS") != "" {
		t.Skip("SKIP_NETWORK_TESTS set")
	}

	// Test against anthropics/skills which has skills in a "skills" subdirectory
	paths, err := TreeDiscover("anthropics", "skills")
	if err != nil {
		t.Fatalf("TreeDiscover failed: %v", err)
	}

	if len(paths) == 0 {
		t.Fatal("Expected at least one skill path, got 0")
	}

	// Verify paths look reasonable (should be under "skills/" prefix)
	foundSkillsPrefix := false
	for _, p := range paths {
		if len(p) > 0 && p[:min(7, len(p))] == "skills/" {
			foundSkillsPrefix = true
			break
		}
	}
	if !foundSkillsPrefix {
		t.Errorf("Expected paths under skills/ prefix, got: %v", paths[:min(5, len(paths))])
	}

	t.Logf("Found %d skill paths in anthropics/skills", len(paths))
	for _, p := range paths[:min(5, len(paths))] {
		t.Logf("  %s", p)
	}
}

func TestTreeDiscoverLargeRepo(t *testing.T) {
	if os.Getenv("SKIP_NETWORK_TESTS") != "" {
		t.Skip("SKIP_NETWORK_TESTS set")
	}

	// Test against antigravity which has 1340+ skills (exceeds Contents API 1000 cap)
	paths, err := TreeDiscover("sickn33", "antigravity-awesome-skills")
	if err != nil {
		t.Fatalf("TreeDiscover failed: %v", err)
	}

	// Should find more than 1000 (Contents API limit)
	if len(paths) < 1000 {
		t.Errorf("Expected 1000+ skills in antigravity repo, got %d", len(paths))
	}

	t.Logf("Found %d skill paths in antigravity repo", len(paths))
}

func TestTreeDiscoverNotFound(t *testing.T) {
	if os.Getenv("SKIP_NETWORK_TESTS") != "" {
		t.Skip("SKIP_NETWORK_TESTS set")
	}

	_, err := TreeDiscover("nonexistent-owner-12345", "nonexistent-repo-12345")
	if err == nil {
		t.Fatal("Expected error for nonexistent repo, got nil")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
