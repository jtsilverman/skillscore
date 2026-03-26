package indexer

// Source represents a GitHub repository to crawl for skills.
type Source struct {
	Owner string
	Repo  string
	Path  string // Optional subdirectory to scan (empty = root)
	Label string // Human-readable label
}

// DefaultSources is the curated list of skill repositories to index.
var DefaultSources = []Source{
	{Owner: "anthropics", Repo: "skills", Path: "skills", Label: "Anthropic Official"},
	{Owner: "travisvn", Repo: "awesome-claude-skills", Path: "", Label: "Awesome Claude Skills"},
	{Owner: "slavingia", Repo: "skills", Path: "", Label: "Slavingia Skills"},
}
