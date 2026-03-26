package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/jtsilverman/skillscore/internal/analyzer"
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
  skillscore [flags] <skill-dir>          Score a local skill
  skillscore [flags] github <owner/repo/path>  Score a GitHub skill
  skillscore scan [flags] <dir>           Score all skills in a directory
  skillscore index [flags]                Index and score curated skill repos

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
	report, err := analyzer.AnalyzeSkill(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(report)
		return
	}

	if quiet {
		fmt.Printf("%s (%0.f/100)\n", report.Overall.Grade, report.Overall.Points)
		return
	}

	// Full terminal report (will be implemented in Task 3.1)
	fmt.Printf("Skill: %s\nGrade: %s (%0.f/100)\n", report.Name, report.Overall.Grade, report.Overall.Points)
}

func runScan(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: skillscore scan <dir>")
		os.Exit(1)
	}
	fmt.Println("scan: not yet implemented")
}

func runIndex(args []string) {
	fmt.Println("index: not yet implemented")
}

func runGitHub(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: skillscore github <owner/repo/path>")
		os.Exit(1)
	}
	fmt.Println("github: not yet implemented")
}
