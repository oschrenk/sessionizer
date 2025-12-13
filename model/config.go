package model

// Config holds the application configuration
type Config struct {
	DefaultName     string
	DefaultPath     string
	SearchDirs      []string
	SearchEntries   []string
	Ignore          []string
	RooterPatterns  []string
}
