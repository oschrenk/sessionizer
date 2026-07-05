package model

// Entry represents a searchable project or directory
type Entry struct {
	Label string
	Path  string
	// Layout names a layout resolved from configDir/layouts/<name>.yml
	Layout string
	// LayoutPath is a direct path to a layout file (env and ~ expanded)
	LayoutPath string
}
