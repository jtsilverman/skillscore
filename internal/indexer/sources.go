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
	{Owner: "affaan-m", Repo: "everything-claude-code", Path: "skills", Label: "Everything Claude Code"},
	{Owner: "sickn33", Repo: "antigravity-awesome-skills", Path: "skills", Label: "Antigravity Skills"},
	{Owner: "ComposioHQ", Repo: "awesome-claude-skills", Path: "", Label: "Composio Skills"},
	{Owner: "alirezarezvani", Repo: "claude-skills", Path: "", Label: "Rezvani Skills"},
	{Owner: "Jeffallan", Repo: "claude-skills", Path: "skills", Label: "Jeffallan Skills"},
	{Owner: "Donchitos", Repo: "Claude-Code-Game-Studios", Path: ".claude/skills", Label: "Game Studios"},
	{Owner: "Orchestra-Research", Repo: "AI-Research-SKILLs", Path: "", Label: "AI Research Skills"},
}
