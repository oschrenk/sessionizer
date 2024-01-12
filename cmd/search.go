package cmd

import (
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/oschrenk/sessionizer/tmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

type SearchEntry struct {
	Label string
	Path  string
}

func entries(defaultName string, defaultPath string, baseDir string, ignore []string) ([]SearchEntry, error) {
	projects := []SearchEntry{}
	// TODO this should not allow a session name with `.` or `:`
	projects = append(projects, SearchEntry{defaultName, defaultPath})

	filepath.WalkDir(baseDir, func(path string, file fs.DirEntry, err error) error {
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
				label := strings.ReplaceAll(path, baseDir+"/", "")
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

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search sessions",
	Run: func(cmd *cobra.Command, args []string) {
		defaultName := viper.GetString("default.name")
		defaultPath := os.ExpandEnv(viper.GetString("default.path"))
		baseDir := os.ExpandEnv(viper.GetString("projects.base_dir"))
		ignore := viper.GetStringSlice("base.ignore")

		// build entries
		projects, err := entries(defaultName, defaultPath, baseDir, ignore)
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
