package report

import (
	"encoding/json"
	"io"

	"github.com/jtsilverman/skillscore/internal/analyzer"
)

// RenderJSON writes the report as formatted JSON.
func RenderJSON(w io.Writer, r *analyzer.SkillReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// RenderJSONMulti writes multiple reports as a JSON object with a "skills" array.
func RenderJSONMulti(w io.Writer, reports []*analyzer.SkillReport) error {
	wrapper := struct {
		Skills []*analyzer.SkillReport `json:"skills"`
		Count  int                     `json:"count"`
	}{
		Skills: reports,
		Count:  len(reports),
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(wrapper)
}
