package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter holds the parsed YAML frontmatter from a SKILL.md file.
type Frontmatter struct {
	Name                   string `yaml:"name"`
	Description            string `yaml:"description"`
	ArgumentHint           string `yaml:"argument-hint"`
	DisableModelInvocation bool   `yaml:"disable-model-invocation"`
	UserInvocable          *bool  `yaml:"user-invocable"`
	AllowedTools           string `yaml:"allowed-tools"`
	Model                  string `yaml:"model"`
	Effort                 string `yaml:"effort"`
	Context                string `yaml:"context"`
	Agent                  string `yaml:"agent"`
	Shell                  string `yaml:"shell"`
}

// ParseResult holds both the frontmatter and the remaining markdown body.
type ParseResult struct {
	Frontmatter    *Frontmatter
	Body           string
	HasFrontmatter bool
	ParseError     error // Non-nil if frontmatter exists but is malformed
}

// ParseSkillFile reads a SKILL.md file and extracts the frontmatter and body.
func ParseSkillFile(path string) (*ParseResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	return ParseSkill(f)
}

// ParseSkillDir finds and parses the SKILL.md file in a skill directory.
func ParseSkillDir(dir string) (*ParseResult, error) {
	skillPath := filepath.Join(dir, "SKILL.md")
	return ParseSkillFile(skillPath)
}

// ParseSkill parses SKILL.md content from a reader.
func ParseSkill(r io.Reader) (*ParseResult, error) {
	scanner := bufio.NewScanner(r)
	result := &ParseResult{}

	// Check for opening ---
	if !scanner.Scan() {
		return result, nil
	}
	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "---" {
		// No frontmatter; entire content is body
		var body strings.Builder
		body.WriteString(scanner.Text())
		body.WriteString("\n")
		for scanner.Scan() {
			body.WriteString(scanner.Text())
			body.WriteString("\n")
		}
		result.Body = body.String()
		return result, nil
	}

	// Read frontmatter until closing ---
	var fmBuilder strings.Builder
	foundClose := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundClose = true
			break
		}
		fmBuilder.WriteString(line)
		fmBuilder.WriteString("\n")
	}

	if !foundClose {
		// Opened frontmatter but never closed it
		result.HasFrontmatter = true
		result.ParseError = fmt.Errorf("frontmatter opened but never closed with ---")
		result.Body = fmBuilder.String()
		return result, nil
	}

	result.HasFrontmatter = true

	// Parse YAML
	fm := &Frontmatter{}
	if err := yaml.Unmarshal([]byte(fmBuilder.String()), fm); err != nil {
		result.ParseError = fmt.Errorf("invalid YAML in frontmatter: %w", err)
	} else {
		result.Frontmatter = fm
	}

	// Read remaining body
	var body strings.Builder
	for scanner.Scan() {
		body.WriteString(scanner.Text())
		body.WriteString("\n")
	}
	result.Body = body.String()

	return result, nil
}
