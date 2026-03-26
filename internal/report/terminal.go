package report

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jtsilverman/skillscore/internal/analyzer"
)

var (
	gradeStyleA = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#22c55e"))
	gradeStyleB = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#86efac"))
	gradeStyleC = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#facc15"))
	gradeStyleD = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f97316"))
	gradeStyleF = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ef4444"))

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#e2e8f0"))
	dimLabelStyle = lipgloss.NewStyle().Width(14).Foreground(lipgloss.Color("#94a3b8"))
	passStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e"))
	failStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))
	suggHighStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))
	suggMedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#f97316"))
	suggLowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))
)

// RenderFull renders a complete terminal report for a single skill.
func RenderFull(r *analyzer.SkillReport, verbose bool) string {
	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render(fmt.Sprintf("  %s", r.Name)))
	b.WriteString("  ")
	b.WriteString(gradeStyle(r.Overall.Grade).Render(
		fmt.Sprintf("%s (%.0f/100)", r.Overall.Grade, r.Overall.Points)))
	b.WriteString("\n\n")

	// Dimension bars
	dims := []struct {
		label string
		score analyzer.DimensionScore
	}{
		{"Structure", r.Structure},
		{"Description", r.Description},
		{"Content", r.Content},
		{"Engineering", r.Engineering},
		{"Packaging", r.Packaging},
	}

	for _, d := range dims {
		b.WriteString(renderDimensionBar(d.label, d.score))
		b.WriteString("\n")
	}

	// Verbose: show individual checks
	if verbose {
		b.WriteString("\n")
		for _, d := range dims {
			b.WriteString(titleStyle.Render(fmt.Sprintf("  %s Checks:", d.label)))
			b.WriteString("\n")
			for _, c := range d.score.Checks {
				icon := passStyle.Render("  ✓")
				if !c.Passed {
					icon = failStyle.Render("  ✗")
				}
				b.WriteString(fmt.Sprintf("%s %s\n", icon, c.Detail))
			}
			b.WriteString("\n")
		}
	}

	// Top suggestions (max 5)
	if len(r.Suggestions) > 0 {
		b.WriteString(titleStyle.Render("  Suggestions:"))
		b.WriteString("\n")
		limit := 5
		if len(r.Suggestions) < limit {
			limit = len(r.Suggestions)
		}
		for i := 0; i < limit; i++ {
			s := r.Suggestions[i]
			style := suggLowStyle
			prefix := "  ○"
			switch s.Priority {
			case "high":
				style = suggHighStyle
				prefix = "  ●"
			case "medium":
				style = suggMedStyle
				prefix = "  ◐"
			}
			b.WriteString(style.Render(fmt.Sprintf("%s %s", prefix, s.Message)))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// RenderCompact renders a single-line summary for scan mode.
func RenderCompact(r *analyzer.SkillReport) string {
	grade := gradeStyle(r.Overall.Grade).Render(
		fmt.Sprintf("%-3s", r.Overall.Grade))
	score := fmt.Sprintf("%3.0f", r.Overall.Points)
	name := r.Name
	if len(name) > 30 {
		name = name[:27] + "..."
	}
	return fmt.Sprintf("  %s  %s/100  %-30s  S:%2.0f D:%2.0f C:%2.0f E:%2.0f P:%2.0f",
		grade, score, name,
		r.Structure.Points, r.Description.Points,
		r.Content.Points, r.Engineering.Points, r.Packaging.Points)
}

func renderDimensionBar(label string, d analyzer.DimensionScore) string {
	bar := renderBar(d.Points, 20)
	return fmt.Sprintf("  %s %s %s",
		dimLabelStyle.Render(label),
		bar,
		gradeStyle(d.Grade).Render(fmt.Sprintf("%3.0f", d.Points)))
}

func renderBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	style := gradeStyleA
	switch {
	case pct < 60:
		style = gradeStyleF
	case pct < 70:
		style = gradeStyleD
	case pct < 80:
		style = gradeStyleC
	case pct < 90:
		style = gradeStyleB
	}

	return style.Render(strings.Repeat("█", filled)) + strings.Repeat("░", empty)
}

func gradeStyle(grade string) lipgloss.Style {
	switch {
	case strings.HasPrefix(grade, "A"):
		return gradeStyleA
	case strings.HasPrefix(grade, "B"):
		return gradeStyleB
	case strings.HasPrefix(grade, "C"):
		return gradeStyleC
	case strings.HasPrefix(grade, "D"):
		return gradeStyleD
	default:
		return gradeStyleF
	}
}
