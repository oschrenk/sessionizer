package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveLayoutPath(t *testing.T) {
	// sessionDir holds an optional local .sessionizer.yml
	sessionDir := t.TempDir()
	localLayout := filepath.Join(sessionDir, layoutFileName)

	// configDir holds named layouts under layouts/
	configDir := t.TempDir()
	layoutsDir := filepath.Join(configDir, "layouts")
	if err := os.MkdirAll(layoutsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	namedLayout := filepath.Join(layoutsDir, "work.yml")
	if err := os.WriteFile(namedLayout, []byte("windows: []"), 0o644); err != nil {
		t.Fatal(err)
	}

	// a direct layout file somewhere else entirely
	directLayout := filepath.Join(t.TempDir(), "default.yml")
	if err := os.WriteFile(directLayout, []byte("windows: []"), 0o644); err != nil {
		t.Fatal(err)
	}

	writeLocal := func(t *testing.T) {
		if err := os.WriteFile(localLayout, []byte("windows: []"), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(localLayout) })
	}

	tests := []struct {
		name       string
		setup      func(t *testing.T)
		layout     string
		layoutPath string
		want       string
	}{
		{
			name: "none configured returns empty",
			want: "",
		},
		{
			name:       "direct layoutPath is used when it exists",
			layoutPath: directLayout,
			want:       directLayout,
		},
		{
			name:       "missing layoutPath falls through to none",
			layoutPath: filepath.Join(t.TempDir(), "missing.yml"),
			want:       "",
		},
		{
			name:   "named layout is resolved from configDir/layouts",
			layout: "work",
			want:   namedLayout,
		},
		{
			name:       "local .sessionizer.yml wins over layoutPath",
			setup:      writeLocal,
			layoutPath: directLayout,
			want:       localLayout,
		},
		{
			name:   "local .sessionizer.yml wins over named layout",
			setup:  writeLocal,
			layout: "work",
			want:   localLayout,
		},
		{
			name:       "layoutPath wins over named layout",
			layout:     "work",
			layoutPath: directLayout,
			want:       directLayout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}
			got := resolveLayoutPath(sessionDir, tt.layout, tt.layoutPath, configDir)
			if got != tt.want {
				t.Errorf("resolveLayoutPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
