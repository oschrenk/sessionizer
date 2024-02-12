package cmd

import (
	"fmt"
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

func entriesFromDir(dir string, ignore []string) ([]SearchEntry, error) {
	projects := []SearchEntry{}

	filepath.WalkDir(dir, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if file.IsDir() {
			if _, err := os.Stat(path + "/.git"); os.IsNotExist(err) {
				// ignore directories
				if slices.Contains(ignore, file.Name()) {
					return filepath.SkipDir
				}
				// no .git, continue search
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
	if _, err := os.Stat(dir + "/.git"); os.IsNotExist(err) {
		return nil, fmt.Errorf("build search entries. entry is not a git project %s", dir)
	}
	entry := &SearchEntry{filepath.Base(dir), dir}

	return entry, nil
}

func entries(defaultName string, defaultPath string, searchDirs []string, entryPaths []string, ignore []string) ([]SearchEntry, error) {
	allProjects := []SearchEntry{}
	// TODO this should not allow a session name with `.` or `:`
	allProjects = append(allProjects, SearchEntry{defaultName, defaultPath})

	// search through directories
	for _, searchDir := range searchDirs {
		dirProjects, err := entriesFromDir(searchDir, ignore)
		if err != nil {
			return nil, err
		}
		allProjects = append(allProjects, dirProjects...)
	}

	// add specific entries
	for _, entryPath := range entryPaths {
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
		defaultName := viper.GetString("default.name")
		defaultPath := os.ExpandEnv(viper.GetString("default.path"))
		searchDirs := mapF(viper.GetStringSlice("search.directories"), os.ExpandEnv)
		searchEntries := mapF(viper.GetStringSlice("search.entries"), os.ExpandEnv)
		ignore := viper.GetStringSlice("base.ignore")

		// build entries
		projects, err := entries(defaultName, defaultPath, searchDirs, searchEntries, ignore)
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
