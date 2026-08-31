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
	// StartupLayout names a layout for the startup entry, resolved from
	// configDir/layouts/<name>.yml
	StartupLayout string
	// StartupLayoutPath is a direct path to a layout file for the startup entry
	StartupLayoutPath string
	// DefaultLayout names a layout applied to every entry that carries none of
	// its own, resolved from configDir/layouts/<name>.yml
	DefaultLayout  string
	SearchDirs     []string
	SearchEntries  []SearchEntry
	Ignore         []string
	RooterPatterns []string
}
