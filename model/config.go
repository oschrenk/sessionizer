package model

// SearchEntry represents a manual entry with a path and optional custom name
type SearchEntry struct {
	Path   string
	Name   string
	Layout string
}

// Config holds the application configuration
type Config struct {
	DefaultName    string
	DefaultPath    string
	SearchDirs     []string
	SearchEntries  []SearchEntry
	Ignore         []string
	RooterPatterns []string
}
