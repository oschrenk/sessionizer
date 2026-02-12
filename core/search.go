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

// BuildEntries creates a list of all searchable entries based on configuration
func BuildEntries(config model.Config) ([]model.Entry, error) {
	allProjects := []model.Entry{}
	// TODO this should not allow a session name with `.` or `:`
	allProjects = append(allProjects, model.Entry{Label: config.DefaultName, Path: config.DefaultPath})

	// search through directories
	for _, searchDir := range config.SearchDirs {
		dirProjects, err := EntriesFromDir(searchDir, config.Ignore, config.RooterPatterns)
		if err != nil {
			return nil, err
		}
		allProjects = append(allProjects, dirProjects...)
	}

	// add specific entries
	for _, entryPath := range config.SearchEntries {
		allProjects = append(allProjects, EntryFromSearchEntry(entryPath))
	}

	return allProjects, nil
}

// resolveLayoutPath returns the path to the layout file to apply, or "" if none.
// Precedence: local .sessionizer.yml > named layout from configDir/layouts/ > none.
func resolveLayoutPath(sessionPath string, layout string, configDir string) string {
	localPath := filepath.Join(sessionPath, layoutFileName)
	if _, err := os.Stat(localPath); err == nil {
		return localPath
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
// If layout is non-empty and no local .sessionizer.yml exists, it resolves a named
// layout file from configDir/layouts/<layout>.yml.
func StartSession(name string, path string, layout string, configDir string) error {
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
		if layoutPath := resolveLayoutPath(path, layout, configDir); layoutPath != "" {
			l, err := tmuxp.ReadLayoutFromFile(layoutPath)
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
