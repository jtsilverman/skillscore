package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jtsilverman/skillscore/internal/analyzer"
	"github.com/jtsilverman/skillscore/internal/github"
	"github.com/jtsilverman/skillscore/internal/report"
)

var (
	jsonOut bool
	quiet   bool
	verbose bool
)

func init() {
	flag.BoolVar(&jsonOut, "json", false, "Output results as JSON")
	flag.BoolVar(&quiet, "quiet", false, "Output only the grade")
	flag.BoolVar(&verbose, "verbose", false, "Show all individual checks")
}

func usage() {
	fmt.Fprintf(os.Stderr, `skillscore - Quality scorer for Claude Code skills

Usage:
  skillscore [flags] <skill-dir>               Score a local skill
  skillscore [flags] github <owner/repo/path>   Score a GitHub skill
  skillscore scan [flags] <dir>                  Score all skills in a directory
  skillscore index [flags]                       Index and score curated skill repos

Flags:
`)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	switch args[0] {
	case "scan":
		runScan(args[1:])
	case "index":
		runIndex(args[1:])
	case "github":
		runGitHub(args[1:])
	default:
		runLocal(args[0])
	}
}

func runLocal(path string) {
	r, err := analyzer.AnalyzeSkill(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	outputReport(r)
}

func runScan(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: skillscore scan <dir>")
		os.Exit(1)
	}

	dir := args[0]
	skills := findSkillDirs(dir)

	if len(skills) == 0 {
		fmt.Fprintf(os.Stderr, "No skills found in %s\n", dir)
		os.Exit(1)
	}

	var reports []*analyzer.SkillReport
	for _, s := range skills {
		r, err := analyzer.AnalyzeSkill(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", s, err)
			continue
		}
		reports = append(reports, r)
	}

	// Sort by score descending
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Overall.Points > reports[j].Overall.Points
	})

	if jsonOut {
		report.RenderJSONMulti(os.Stdout, reports)
		return
	}

	fmt.Printf("  Found %d skills in %s\n\n", len(reports), dir)
	for _, r := range reports {
		if quiet {
			fmt.Printf("  %-3s %3.0f/100  %s\n", r.Overall.Grade, r.Overall.Points, r.Name)
		} else {
			fmt.Println(report.RenderCompact(r))
		}
	}
	fmt.Println()
}

func runIndex(args []string) {
	fmt.Println("index: not yet implemented")
}

func runGitHub(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: skillscore github <owner/repo/path>")
		os.Exit(1)
	}

	spec := args[0]
	fmt.Fprintf(os.Stderr, "Fetching %s from GitHub...\n", spec)

	tmpDir, cleanup, err := github.FetchSkill(spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	r, err := analyzer.AnalyzeSkill(tmpDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing: %v\n", err)
		os.Exit(1)
	}
	// Override the path with the GitHub spec for display
	r.Path = "github:" + spec
	outputReport(r)
}

func outputReport(r *analyzer.SkillReport) {
	if jsonOut {
		report.RenderJSON(os.Stdout, r)
		return
	}
	if quiet {
		fmt.Printf("%s (%.0f/100)\n", r.Overall.Grade, r.Overall.Points)
		return
	}
	fmt.Println(report.RenderFull(r, verbose))
}

// findSkillDirs walks a directory tree and returns paths containing SKILL.md.
func findSkillDirs(root string) []string {
	var dirs []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.EqualFold(info.Name(), "SKILL.md") {
			dirs = append(dirs, filepath.Dir(path))
		}
		return nil
	})
	return dirs
}
