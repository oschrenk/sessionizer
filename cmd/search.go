package cmd

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/oschrenk/sessionizer/tmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

type SearchEntry struct {
	Label string
	Path  string
}

func entriesFromDir(dir string, ignore []string, rooterPatterns []string) ([]SearchEntry, error) {
	projects := []SearchEntry{}

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
				projects = append(projects, SearchEntry{label, path})

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

func entryFromPath(dir string) (*SearchEntry, error) {
	entry := &SearchEntry{filepath.Base(dir), dir}
	return entry, nil
}

func entries(config Config) ([]SearchEntry, error) {
	allProjects := []SearchEntry{}
	// TODO this should not allow a session name with `.` or `:`
	allProjects = append(allProjects, SearchEntry{config.DefaultName, config.DefaultPath})

	// search through directories
	for _, searchDir := range config.SearchDirs {
		dirProjects, err := entriesFromDir(searchDir, config.Ignore, config.RooterPatterns)
		if err != nil {
			return nil, err
		}
		allProjects = append(allProjects, dirProjects...)
	}

	// add specific entries
	for _, entryPath := range config.SearchEntries {
		searchEntry, err := entryFromPath(entryPath)
		if err != nil {
			return nil, err
		}
		allProjects = append(allProjects, *searchEntry)
	}

	return allProjects, nil
}

func search(projects []SearchEntry) (SearchEntry, error) {
	idx, err := fuzzyfinder.Find(
		projects,
		func(i int) string {
			return projects[i].Label
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	return projects[idx], nil
}

func startSession(project SearchEntry) {
	server := new(tmux.Server)
	server.CreateOrAttachSession(project.Label, project.Path)
}

func mapF[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search sessions",
	Run: func(cmd *cobra.Command, args []string) {
		config := Config{
			DefaultName:    viper.GetString("default.name"),
			DefaultPath:    os.ExpandEnv(viper.GetString("default.path")),
			SearchDirs:     mapF(viper.GetStringSlice("search.directories"), os.ExpandEnv),
			SearchEntries:  mapF(viper.GetStringSlice("search.entries"), os.ExpandEnv),
			Ignore:         viper.GetStringSlice("base.ignore"),
			RooterPatterns: viper.GetStringSlice("base.rooter_patterns"),
		}

		// build entries
		projects, err := entries(config)
		if err != nil {
			log.Fatal(err)
		}

		// search, select entry
		project, err := search(projects)
		if err != nil {
			log.Fatal(err)
		}

		startSession(project)
	},
}

func init() {
	// add flags and such here
}
