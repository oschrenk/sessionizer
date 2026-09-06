package core

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/oschrenk/sessionizer/internal/tmux"
	"github.com/oschrenk/sessionizer/internal/tmuxp"
	"github.com/oschrenk/sessionizer/model"
)

const layoutFileName = ".sessionizer.yml"

// EntriesFromDir finds all project directories within a given directory
// that match the rooter patterns, ignoring specified directories
func EntriesFromDir(dir string, ignore []string, rooterPatterns []string) ([]model.Entry, error) {
	projects := []model.Entry{}

	filepath.WalkDir(dir, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if file.IsDir() {
			// check if any rooter pattern exists in this directory
			hasRooterPattern := false
			for _, pattern := range rooterPatterns {
				if _, err := os.Stat(filepath.Join(path, pattern)); err == nil {
					hasRooterPattern = true
					break
				}
			}

			if !hasRooterPattern {
				// ignore directories
				if slices.Contains(ignore, file.Name()) {
					return filepath.SkipDir
				}
				// no rooter pattern found, continue search
				return nil
			} else {
				label := strings.ReplaceAll(path, dir+"/", "")
				projects = append(projects, model.Entry{Label: label, Path: path})

				// don't extends search breadth
				// that stops from build directories or sub-Projects
				// from being picked up
				return filepath.SkipDir
			}
		}

		return nil
	})

	return projects, nil
}

// EntryFromSearchEntry creates an entry from a SearchEntry.
// If the entry has a custom name, it is used as the label.
// Otherwise, the label is derived from the base of the path.
func EntryFromSearchEntry(se model.SearchEntry) model.Entry {
	label := se.Name
	if label == "" {
		label = filepath.Base(se.Path)
	}
	return model.Entry{Label: label, Path: se.Path, Layout: se.Layout}
}

// applyDefaultLayout seeds the global default layout onto an entry that names
// no layout of its own. A local .sessionizer.yml still wins, in resolveLayoutPath.
func applyDefaultLayout(entry model.Entry, defaultLayout string) model.Entry {
	if entry.Layout == "" && entry.LayoutPath == "" {
		entry.Layout = defaultLayout
	}
	return entry
}

// StartupEntry builds the entry for the startup session, with the global
// default layout applied when the startup section names none itself.
func StartupEntry(config model.Config) model.Entry {
	return applyDefaultLayout(model.Entry{
		Label:      config.StartupName,
		Path:       config.StartupPath,
		Layout:     config.StartupLayout,
		LayoutPath: config.StartupLayoutPath,
	}, config.DefaultLayout)
}

// entry sources, ordered from most to least explicit. A path reachable from
// more than one source keeps the entry of the lowest-numbered one.
const (
	sourceStartup = iota
	sourceManual
	sourceWalked
)

// sourcedEntry pairs an entry with the source that produced it, so duplicates
// can be resolved before the source is forgotten.
type sourcedEntry struct {
	entry  model.Entry
	source int
}

// dedupe collapses entries that name the same path. The entry from the most
// explicit source wins, and it takes the position of the first mention of that
// path, so the order of the list does not depend on which source won.
func dedupe(sourced []sourcedEntry) []model.Entry {
	kept := []sourcedEntry{}
	positions := map[string]int{}

	for _, candidate := range sourced {
		key := filepath.Clean(candidate.entry.Path)
		position, seen := positions[key]
		if !seen {
			positions[key] = len(kept)
			kept = append(kept, candidate)
			continue
		}
		if candidate.source < kept[position].source {
			kept[position] = candidate
		}
	}

	entries := make([]model.Entry, len(kept))
	for i, k := range kept {
		entries[i] = k.entry
	}
	return entries
}

// BuildEntries creates a list of all searchable entries based on configuration
func BuildEntries(config model.Config) ([]model.Entry, error) {
	sourced := []sourcedEntry{}
	// TODO this should not allow a session name with `.` or `:`
	// only offer the startup entry when a startup name is configured
	if strings.TrimSpace(config.StartupName) != "" {
		sourced = append(sourced, sourcedEntry{StartupEntry(config), sourceStartup})
	}

	// search through directories
	for _, searchDir := range config.SearchDirs {
		dirProjects, err := EntriesFromDir(searchDir, config.Ignore, config.RooterPatterns)
		if err != nil {
			return nil, err
		}
		for _, dirProject := range dirProjects {
			sourced = append(sourced, sourcedEntry{dirProject, sourceWalked})
		}
	}

	// add specific entries
	for _, entryPath := range config.SearchEntries {
		sourced = append(sourced, sourcedEntry{EntryFromSearchEntry(entryPath), sourceManual})
	}

	// one path, one entry: the startup section and a manual entry both outrank a
	// directory found by walking, which is where the same project shows up twice
	allProjects := dedupe(sourced)

	// the global default is the last level of the layout chain: it reaches every
	// entry that named neither a layout nor a layout path of its own
	for i := range allProjects {
		allProjects[i] = applyDefaultLayout(allProjects[i], config.DefaultLayout)
	}

	return allProjects, nil
}

// resolveLayoutPath returns the path to the layout file to apply, or "" if none.
// Precedence: local .sessionizer.yml > direct layoutPath > named layout from
// configDir/layouts/ > none. The named layout may be the entry's own or the
// global default seeded by applyDefaultLayout; both arrive here the same way.
func resolveLayoutPath(sessionPath string, layout string, layoutPath string, configDir string) string {
	localPath := filepath.Join(sessionPath, layoutFileName)
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	if layoutPath != "" {
		directPath := tmuxp.ExpandPath(layoutPath)
		if _, err := os.Stat(directPath); err == nil {
			return directPath
		}
	}
	if layout != "" {
		namedPath := filepath.Join(configDir, "layouts", layout+".yml")
		if _, err := os.Stat(namedPath); err == nil {
			return namedPath
		}
	}
	return ""
}

// StartSession creates or attaches to a tmux session with the given name and path.
// If no local .sessionizer.yml exists, it resolves a layout from the direct
// layoutPath, or from a named layout file at configDir/layouts/<layout>.yml.
func StartSession(name string, path string, layout string, layoutPath string, configDir string) error {
	server := new(tmux.Server)

	var session tmux.Session
	var freshlyCreated bool

	sessionPtr, err := server.SessionByName(name)
	if err != nil {
		return err
	}

	if sessionPtr == nil {
		session, err = server.AddSession(name, path)
		if err != nil {
			return err
		}
		freshlyCreated = true
	} else {
		session = *sessionPtr
		freshlyCreated = false
	}

	if freshlyCreated {
		if resolvedPath := resolveLayoutPath(path, layout, layoutPath, configDir); resolvedPath != "" {
			l, err := tmuxp.ReadLayoutFromFile(resolvedPath)
			if err != nil {
				return err
			}
			err = ApplyLayout(server, session, *l)
			if err != nil {
				return err
			}
		}
	}

	err = server.AttachSession(session)
	if err != nil {
		return err
	}

	return nil
}
