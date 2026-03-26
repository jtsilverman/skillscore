package parser

import (
	"strings"
	"testing"
)

func TestParseFrontmatter_Valid(t *testing.T) {
	input := `---
name: deploy-app
description: Deploy the application to production. Use when the user asks to deploy.
disable-model-invocation: true
allowed-tools: Bash(npm *)
---

# Deploy

1. Run tests
2. Build
3. Push
`
	result, err := ParseSkill(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFrontmatter {
		t.Fatal("expected HasFrontmatter=true")
	}
	if result.ParseError != nil {
		t.Fatalf("unexpected parse error: %v", result.ParseError)
	}
	fm := result.Frontmatter
	if fm.Name != "deploy-app" {
		t.Errorf("name=%q, want %q", fm.Name, "deploy-app")
	}
	if fm.Description == "" {
		t.Error("description should not be empty")
	}
	if !fm.DisableModelInvocation {
		t.Error("disable-model-invocation should be true")
	}
	if fm.AllowedTools != "Bash(npm *)" {
		t.Errorf("allowed-tools=%q, want %q", fm.AllowedTools, "Bash(npm *)")
	}
	if !strings.Contains(result.Body, "# Deploy") {
		t.Error("body should contain markdown content")
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	input := `# Just a markdown file

Some content here.
`
	result, err := ParseSkill(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasFrontmatter {
		t.Error("expected HasFrontmatter=false")
	}
	if result.Frontmatter != nil {
		t.Error("expected nil Frontmatter")
	}
	if !strings.Contains(result.Body, "Just a markdown file") {
		t.Error("body should contain all content")
	}
}

func TestParseFrontmatter_EmptyFrontmatter(t *testing.T) {
	input := `---
---

Body content.
`
	result, err := ParseSkill(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFrontmatter {
		t.Error("expected HasFrontmatter=true")
	}
	if result.Frontmatter == nil {
		t.Fatal("expected non-nil Frontmatter")
	}
	if result.Frontmatter.Name != "" {
		t.Errorf("name should be empty, got %q", result.Frontmatter.Name)
	}
}

func TestParseFrontmatter_Malformed(t *testing.T) {
	input := `---
name: [invalid yaml
  this is broken
---

Body.
`
	result, err := ParseSkill(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFrontmatter {
		t.Error("expected HasFrontmatter=true")
	}
	if result.ParseError == nil {
		t.Error("expected parse error for malformed YAML")
	}
}

func TestParseFrontmatter_Unclosed(t *testing.T) {
	input := `---
name: broken
description: never closed
`
	result, err := ParseSkill(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFrontmatter {
		t.Error("expected HasFrontmatter=true")
	}
	if result.ParseError == nil {
		t.Error("expected parse error for unclosed frontmatter")
	}
}

func TestParseFrontmatter_EmptyInput(t *testing.T) {
	result, err := ParseSkill(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasFrontmatter {
		t.Error("expected HasFrontmatter=false for empty input")
	}
}

func TestParseFrontmatter_AllFields(t *testing.T) {
	input := `---
name: my-skill
description: Does things
argument-hint: "[filename]"
disable-model-invocation: true
user-invocable: false
allowed-tools: Read, Grep
model: sonnet
effort: high
context: fork
agent: Explore
shell: bash
---

Content.
`
	result, err := ParseSkill(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fm := result.Frontmatter
	if fm.Name != "my-skill" {
		t.Errorf("name=%q", fm.Name)
	}
	if fm.ArgumentHint != "[filename]" {
		t.Errorf("argument-hint=%q", fm.ArgumentHint)
	}
	if fm.UserInvocable == nil || *fm.UserInvocable != false {
		t.Error("user-invocable should be false")
	}
	if fm.Model != "sonnet" {
		t.Errorf("model=%q", fm.Model)
	}
	if fm.Effort != "high" {
		t.Errorf("effort=%q", fm.Effort)
	}
	if fm.Context != "fork" {
		t.Errorf("context=%q", fm.Context)
	}
	if fm.Agent != "Explore" {
		t.Errorf("agent=%q", fm.Agent)
	}
}
