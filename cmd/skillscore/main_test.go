package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binary string

func TestMain(m *testing.M) {
	// Build the binary once
	dir, _ := os.MkdirTemp("", "skillscore-test-*")
	binary = filepath.Join(dir, "skillscore")
	cmd := exec.Command("go", "build", "-o", binary, "./")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("build failed: " + string(out))
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func run(args ...string) (string, error) {
	cmd := exec.Command(binary, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestE2E_GoodSkillGrade(t *testing.T) {
	out, err := run("--quiet", "../../testdata/good-skill/")
	if err != nil {
		t.Fatalf("error: %v\n%s", err, out)
	}
	out = strings.TrimSpace(out)
	// Should be A or B range
	if !strings.HasPrefix(out, "A") && !strings.HasPrefix(out, "B") {
		t.Errorf("good skill grade=%q, want A or B range", out)
	}
}

func TestE2E_BadSkillGrade(t *testing.T) {
	out, err := run("--quiet", "../../testdata/bad-skill/")
	if err != nil {
		t.Fatalf("error: %v\n%s", err, out)
	}
	out = strings.TrimSpace(out)
	// Should be D or F range
	if !strings.HasPrefix(out, "D") && !strings.HasPrefix(out, "F") {
		t.Errorf("bad skill grade=%q, want D or F range", out)
	}
}

func TestE2E_JSONOutput(t *testing.T) {
	out, err := run("--json", "../../testdata/good-skill/")
	if err != nil {
		t.Fatalf("error: %v\n%s", err, out)
	}

	var report map[string]interface{}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, out)
	}

	overall, ok := report["overall"].(map[string]interface{})
	if !ok {
		t.Fatal("missing 'overall' in JSON")
	}
	if _, ok := overall["points"]; !ok {
		t.Error("missing 'points' in overall")
	}
	if _, ok := overall["grade"]; !ok {
		t.Error("missing 'grade' in overall")
	}
}

func TestE2E_ScanMode(t *testing.T) {
	out, err := run("scan", "../../testdata/")
	if err != nil {
		t.Fatalf("error: %v\n%s", err, out)
	}
	// Should find at least the good-skill and bad-skill
	if !strings.Contains(out, "deploy-app") {
		t.Error("scan should find deploy-app (good-skill)")
	}
	if !strings.Contains(out, "HELPER_Tool") {
		t.Error("scan should find HELPER_Tool (bad-skill)")
	}
}

func TestE2E_ScanJSON(t *testing.T) {
	out, err := run("--json", "scan", "../../testdata/")
	if err != nil {
		t.Fatalf("error: %v\n%s", err, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	skills, ok := result["skills"].([]interface{})
	if !ok {
		t.Fatal("missing 'skills' array")
	}
	if len(skills) < 2 {
		t.Errorf("expected at least 2 skills, got %d", len(skills))
	}
}

func TestE2E_EmptySkill(t *testing.T) {
	out, err := run("--quiet", "../../testdata/edge-cases/empty-skill/")
	if err != nil {
		t.Fatalf("error: %v\n%s", err, out)
	}
	// Should not crash, should give a low grade
	out = strings.TrimSpace(out)
	if out == "" {
		t.Error("expected some output for empty skill")
	}
}

func TestE2E_NoFrontmatter(t *testing.T) {
	out, err := run("--quiet", "../../testdata/edge-cases/no-frontmatter/")
	if err != nil {
		t.Fatalf("error: %v\n%s", err, out)
	}
	out = strings.TrimSpace(out)
	// Should get a low grade due to missing frontmatter
	if strings.HasPrefix(out, "A") {
		t.Errorf("no-frontmatter skill should not get an A, got %q", out)
	}
}

func TestE2E_OnlyFrontmatter(t *testing.T) {
	out, err := run("--quiet", "../../testdata/edge-cases/only-frontmatter/")
	if err != nil {
		t.Fatalf("error: %v\n%s", err, out)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		t.Error("expected output for only-frontmatter skill")
	}
}

func TestE2E_VerboseMode(t *testing.T) {
	out, err := run("--verbose", "../../testdata/good-skill/")
	if err != nil {
		t.Fatalf("error: %v\n%s", err, out)
	}
	// Verbose should include check details
	if !strings.Contains(out, "Checks:") {
		t.Error("verbose mode should show check details")
	}
}

func TestE2E_Help(t *testing.T) {
	out, _ := run() // No args
	if !strings.Contains(out, "Usage:") {
		t.Error("no-args should show usage")
	}
}
