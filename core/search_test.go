package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oschrenk/sessionizer/model"
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

func TestBuildEntriesDefaultLayout(t *testing.T) {
	// a search directory holding one project, marked by a .git rooter
	searchDir := t.TempDir()
	projectDir := filepath.Join(searchDir, "walked")
	if err := os.MkdirAll(filepath.Join(projectDir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	base := func() model.Config {
		return model.Config{
			StartupName:    "startup",
			StartupPath:    t.TempDir(),
			DefaultLayout:  "global",
			SearchDirs:     []string{searchDir},
			RooterPatterns: []string{".git"},
			SearchEntries: []model.SearchEntry{
				{Path: "/tmp/bare", Name: "bare"},
				{Path: "/tmp/own", Name: "own", Layout: "entry-own"},
			},
		}
	}

	// layoutOf finds the entry by label and reports its resolved layout keys
	layoutOf := func(t *testing.T, entries []model.Entry, label string) model.Entry {
		t.Helper()
		for _, e := range entries {
			if e.Label == label {
				return e
			}
		}
		t.Fatalf("no entry labelled %q in %v", label, entries)
		return model.Entry{}
	}

	t.Run("default layout reaches walked, manual and startup entries", func(t *testing.T) {
		entries, err := BuildEntries(base())
		if err != nil {
			t.Fatal(err)
		}

		for _, label := range []string{"startup", "walked", "bare"} {
			if got := layoutOf(t, entries, label).Layout; got != "global" {
				t.Errorf("%s: Layout = %q, want %q", label, got, "global")
			}
		}
	})

	t.Run("entry layout wins over default layout", func(t *testing.T) {
		entries, err := BuildEntries(base())
		if err != nil {
			t.Fatal(err)
		}

		if got := layoutOf(t, entries, "own").Layout; got != "entry-own" {
			t.Errorf("Layout = %q, want %q", got, "entry-own")
		}
	})

	t.Run("startup layout wins over default layout", func(t *testing.T) {
		config := base()
		config.StartupLayout = "startup-own"

		entries, err := BuildEntries(config)
		if err != nil {
			t.Fatal(err)
		}

		if got := layoutOf(t, entries, "startup").Layout; got != "startup-own" {
			t.Errorf("Layout = %q, want %q", got, "startup-own")
		}
	})

	t.Run("startup layout path wins over default layout", func(t *testing.T) {
		config := base()
		config.StartupLayoutPath = "/tmp/startup.yml"

		entries, err := BuildEntries(config)
		if err != nil {
			t.Fatal(err)
		}

		// the default must not be seeded alongside a layout path, or it would
		// resolve as a fallback the startup section never asked for
		startup := layoutOf(t, entries, "startup")
		if startup.Layout != "" {
			t.Errorf("Layout = %q, want empty", startup.Layout)
		}
		if startup.LayoutPath != "/tmp/startup.yml" {
			t.Errorf("LayoutPath = %q, want %q", startup.LayoutPath, "/tmp/startup.yml")
		}
	})

	t.Run("no default layout leaves entries untouched", func(t *testing.T) {
		config := base()
		config.DefaultLayout = ""

		entries, err := BuildEntries(config)
		if err != nil {
			t.Fatal(err)
		}

		for _, label := range []string{"startup", "walked", "bare"} {
			if got := layoutOf(t, entries, label).Layout; got != "" {
				t.Errorf("%s: Layout = %q, want empty", label, got)
			}
		}
	})

	t.Run("no startup name drops the startup entry", func(t *testing.T) {
		config := base()
		config.StartupName = ""

		entries, err := BuildEntries(config)
		if err != nil {
			t.Fatal(err)
		}

		for _, e := range entries {
			if e.Label == "startup" {
				t.Fatal("startup entry offered without a startup name")
			}
		}
	})
}

func TestBuildEntriesDeduplicates(t *testing.T) {
	searchDir := t.TempDir()
	project := filepath.Join(searchDir, "ops", "dotfiles")
	if err := os.MkdirAll(filepath.Join(project, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	base := model.Config{
		SearchDirs:     []string{searchDir},
		RooterPatterns: []string{".git"},
	}

	t.Run("startup wins over the walked directory", func(t *testing.T) {
		config := base
		config.StartupName = "dotfiles"
		config.StartupPath = project

		entries, err := BuildEntries(config)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d: %v", len(entries), entries)
		}
		if entries[0].Label != "dotfiles" {
			t.Errorf("expected label %q, got %q", "dotfiles", entries[0].Label)
		}
	})

	t.Run("a manual entry wins over the walked directory", func(t *testing.T) {
		config := base
		config.SearchEntries = []model.SearchEntry{{Path: project, Name: "config/dotfiles"}}

		entries, err := BuildEntries(config)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d: %v", len(entries), entries)
		}
		if entries[0].Label != "config/dotfiles" {
			t.Errorf("expected label %q, got %q", "config/dotfiles", entries[0].Label)
		}
	})

	t.Run("startup wins over a manual entry", func(t *testing.T) {
		config := base
		config.SearchDirs = nil
		config.StartupName = "dotfiles"
		config.StartupPath = project
		config.SearchEntries = []model.SearchEntry{{Path: project, Name: "config/dotfiles"}}

		entries, err := BuildEntries(config)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d: %v", len(entries), entries)
		}
		if entries[0].Label != "dotfiles" {
			t.Errorf("expected label %q, got %q", "dotfiles", entries[0].Label)
		}
	})

	t.Run("a trailing slash is still the same path", func(t *testing.T) {
		config := base
		config.StartupName = "dotfiles"
		config.StartupPath = project + "/"

		entries, err := BuildEntries(config)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d: %v", len(entries), entries)
		}
	})

	t.Run("distinct paths are all kept", func(t *testing.T) {
		otherDir := t.TempDir()
		other := filepath.Join(otherDir, "tools", "sessionizer")
		if err := os.MkdirAll(filepath.Join(other, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
		config := base
		config.SearchDirs = []string{searchDir, otherDir}
		config.SearchEntries = []model.SearchEntry{{Path: t.TempDir()}}

		entries, err := BuildEntries(config)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 3 {
			t.Fatalf("expected 3 entries, got %d: %v", len(entries), entries)
		}
	})
}
