package model

// SearchEntry represents a manual entry with a path and optional custom name
type SearchEntry struct {
	Path   string
	Name   string
	Layout string
}

// Config holds the application configuration
type Config struct {
	StartupName string
	StartupPath string
	// StartupLayoutPath is a direct path to a layout file for the startup entry
	StartupLayoutPath string
	SearchDirs        []string
	SearchEntries     []SearchEntry
	Ignore            []string
	RooterPatterns    []string
}
