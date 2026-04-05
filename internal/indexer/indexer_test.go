package indexer

import "testing"

func TestDedup(t *testing.T) {
	tests := []struct {
		name    string
		entries []IndexEntry
		want    int    // expected result length
		check   func([]IndexEntry) string // optional: return error message if invalid
	}{
		{
			name:    "empty input",
			entries: nil,
			want:    0,
		},
		{
			name: "no duplicates passthrough",
			entries: []IndexEntry{
				{Name: "skill-a", Repo: "owner/repo1", Score: 80},
				{Name: "skill-b", Repo: "owner/repo1", Score: 70},
			},
			want: 2,
		},
		{
			name: "same name same repo keeps higher score",
			entries: []IndexEntry{
				{Name: "skill-a", Repo: "owner/repo1", Score: 60},
				{Name: "skill-a", Repo: "owner/repo1", Score: 90},
			},
			want: 1,
			check: func(result []IndexEntry) string {
				if result[0].Score != 90 {
					return "expected score 90, got different"
				}
				return ""
			},
		},
		{
			name: "same name different repo keeps both",
			entries: []IndexEntry{
				{Name: "skill-a", Repo: "owner/repo1", Score: 60},
				{Name: "skill-a", Repo: "owner/repo2", Score: 90},
			},
			want: 2,
		},
		{
			name: "case insensitive name dedup",
			entries: []IndexEntry{
				{Name: "Skill-A", Repo: "owner/repo1", Score: 60},
				{Name: "skill-a", Repo: "owner/repo1", Score: 90},
			},
			want: 1,
			check: func(result []IndexEntry) string {
				if result[0].Score != 90 {
					return "expected higher score (90) to win"
				}
				return ""
			},
		},
		{
			name: "whitespace trimming in name",
			entries: []IndexEntry{
				{Name: "  skill-a  ", Repo: "owner/repo1", Score: 50},
				{Name: "skill-a", Repo: "owner/repo1", Score: 80},
			},
			want: 1,
			check: func(result []IndexEntry) string {
				if result[0].Score != 80 {
					return "expected higher score (80) to win"
				}
				return ""
			},
		},
		{
			name: "multiple duplicates three entries same key",
			entries: []IndexEntry{
				{Name: "skill-a", Repo: "owner/repo1", Score: 50},
				{Name: "skill-a", Repo: "owner/repo1", Score: 90},
				{Name: "skill-a", Repo: "owner/repo1", Score: 70},
			},
			want: 1,
			check: func(result []IndexEntry) string {
				if result[0].Score != 90 {
					return "expected highest score (90) to win"
				}
				return ""
			},
		},
		{
			name: "mixed duplicates and unique",
			entries: []IndexEntry{
				{Name: "skill-a", Repo: "owner/repo1", Score: 50},
				{Name: "skill-b", Repo: "owner/repo1", Score: 60},
				{Name: "skill-a", Repo: "owner/repo1", Score: 80},
				{Name: "skill-c", Repo: "owner/repo2", Score: 70},
			},
			want: 3,
			check: func(result []IndexEntry) string {
				for _, e := range result {
					if e.Name == "skill-a" && e.Score != 80 {
						return "skill-a should have score 80"
					}
				}
				return ""
			},
		},
		{
			name: "lower score duplicate is discarded",
			entries: []IndexEntry{
				{Name: "skill-a", Repo: "owner/repo1", Score: 90},
				{Name: "skill-a", Repo: "owner/repo1", Score: 60},
			},
			want: 1,
			check: func(result []IndexEntry) string {
				if result[0].Score != 90 {
					return "expected first (higher) score to remain"
				}
				return ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dedup(tt.entries)
			if len(result) != tt.want {
				t.Errorf("dedup() returned %d entries, want %d", len(result), tt.want)
			}
			if tt.check != nil && len(result) == tt.want {
				if msg := tt.check(result); msg != "" {
					t.Errorf("dedup() check failed: %s", msg)
				}
			}
		})
	}
}
