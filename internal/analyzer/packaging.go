package analyzer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jtsilverman/skillscore/internal/parser"
)

var (
	junkFiles = []string{
		".ds_store", "thumbs.db", ".gitkeep",
		"node_modules", "__pycache__", ".venv", "venv",
		".env", ".pyc",
	}

	genericNames = []string{
		"doc1.md", "doc2.md", "file.md", "temp.md", "test.md",
		"data.md", "notes.md", "stuff.md", "misc.md",
	}

	// Max skill directory size: 50KB (excluding binary assets)
	maxDirSize int64 = 50 * 1024
)

// ScorePackaging evaluates the file organization of a skill.
func ScorePackaging(dir string, md *parser.MarkdownAnalysis) DimensionScore {
	var checks []Check

	// Check 1: Total directory size under 50KB
	totalSize := dirSize(dir)
	underSize := totalSize <= maxDirSize
	checks = append(checks, Check{
		Name:   "dir_size_ok",
		Passed: underSize,
		Weight: 20,
		Detail: checkDetail(underSize,
			"Skill directory is within size limits",
			"Skill directory exceeds 50KB; move large files to external references"),
	})

	// Check 2: Descriptive file names
	hasGeneric := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		lower := strings.ToLower(info.Name())
		for _, g := range genericNames {
			if lower == g {
				hasGeneric = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	checks = append(checks, Check{
		Name:   "descriptive_names",
		Passed: !hasGeneric,
		Weight: 20,
		Detail: checkDetail(!hasGeneric,
			"Files have descriptive names",
			"Generic file names detected (e.g., doc1.md); use descriptive names"),
	})

	// Check 3: Supporting files in subdirectories
	hasSubdirs := false
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				hasSubdirs = true
				break
			}
		}
	}
	// Only penalize if there are multiple non-SKILL.md files at root
	rootFiles := 0
	if entries != nil {
		for _, e := range entries {
			if !e.IsDir() && e.Name() != "SKILL.md" {
				rootFiles++
			}
		}
	}
	organized := hasSubdirs || rootFiles <= 1
	checks = append(checks, Check{
		Name:   "organized_structure",
		Passed: organized,
		Weight: 20,
		Detail: checkDetail(organized,
			"Supporting files are organized in subdirectories",
			"Multiple files at root level; organize into subdirectories (scripts/, reference/)"),
	})

	// Check 4: No junk files
	hasJunk := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		lower := strings.ToLower(info.Name())
		for _, j := range junkFiles {
			if lower == j || strings.HasSuffix(lower, ".pyc") {
				hasJunk = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	checks = append(checks, Check{
		Name:   "no_junk_files",
		Passed: !hasJunk,
		Weight: 20,
		Detail: checkDetail(!hasJunk,
			"No unnecessary files (.DS_Store, node_modules, etc.)",
			"Junk files detected; clean up before publishing"),
	})

	// Check 5: Reference files linked from SKILL.md
	linkedOk := checkRefsLinked(dir, md)
	checks = append(checks, Check{
		Name:   "refs_linked",
		Passed: linkedOk,
		Weight: 20,
		Detail: checkDetail(linkedOk,
			"Reference files are linked from SKILL.md",
			"Reference .md files exist but aren't linked from SKILL.md"),
	})

	return computeDimension(checks)
}

func dirSize(dir string) int64 {
	var total int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// Skip common binary extensions
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".png" || ext == ".jpg" || ext == ".gif" || ext == ".ico" || ext == ".woff" || ext == ".woff2" {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

// checkRefsLinked verifies that .md files in the skill directory (other than SKILL.md)
// are linked from SKILL.md.
func checkRefsLinked(dir string, md *parser.MarkdownAnalysis) bool {
	// Collect all linked destinations
	linked := make(map[string]bool)
	for _, l := range md.Links {
		linked[l.Destination] = true
	}

	// Check each .md file in the directory
	var unlinked bool
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".md" {
			return nil
		}
		if info.Name() == "SKILL.md" {
			return nil
		}
		// Get relative path from skill dir
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		if !linked[rel] {
			unlinked = true
			return filepath.SkipAll
		}
		return nil
	})
	return !unlinked
}
