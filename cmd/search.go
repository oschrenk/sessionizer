package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/oschrenk/sessionizer/core"
	"github.com/oschrenk/sessionizer/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

func search(projects []model.Entry) (model.Entry, error) {
	idx, err := fuzzyfinder.Find(
		projects,
		func(i int) string {
			return projects[i].Label
		},
	)
	if err != nil {
		return model.Entry{}, err
	}
	return projects[idx], nil
}

func startSession(project model.Entry) {
	configDir := filepath.Dir(viper.ConfigFileUsed())
	err := core.StartSession(project.Label, project.Path, project.Layout, configDir)
	if err != nil {
		panic(err)
	}
}

func mapF[T, V interface{}](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

// parseSearchEntries parses search.entries which can be a mix of strings and objects.
// Strings are treated as paths (name auto-derived). Objects must have a "path" key
// and an optional "name" key.
func parseSearchEntries(raw interface{}) ([]model.SearchEntry, error) {
	if raw == nil {
		return nil, nil
	}
	slice, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("search.entries: expected array, got %T", raw)
	}
	entries := make([]model.SearchEntry, 0, len(slice))
	for i, item := range slice {
		switch v := item.(type) {
		case string:
			entries = append(entries, model.SearchEntry{Path: os.ExpandEnv(v)})
		case map[string]interface{}:
			path, ok := v["path"].(string)
			if !ok {
				return nil, fmt.Errorf("search.entries[%d]: missing or invalid 'path'", i)
			}
			entry := model.SearchEntry{Path: os.ExpandEnv(path)}
			if name, ok := v["name"].(string); ok {
				entry.Name = name
			}
			if layout, ok := v["layout"].(string); ok {
				entry.Layout = layout
			}
			entries = append(entries, entry)
		default:
			return nil, fmt.Errorf("search.entries[%d]: expected string or object, got %T", i, item)
		}
	}
	return entries, nil
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search sessions",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		searchEntries, err := parseSearchEntries(viper.Get("search.entries"))
		if err != nil {
			log.Fatal(err)
		}

		config := model.Config{
			DefaultName:    viper.GetString("default.name"),
			DefaultPath:    os.ExpandEnv(viper.GetString("default.path")),
			SearchDirs:     mapF(viper.GetStringSlice("search.directories"), os.ExpandEnv),
			SearchEntries:  searchEntries,
			Ignore:         viper.GetStringSlice("base.ignore"),
			RooterPatterns: viper.GetStringSlice("base.rooter_patterns"),
		}

		// build entries
		projects, err := core.BuildEntries(config)
		if err != nil {
			log.Fatal(err)
		}

		// search, select entry
		project, err := search(projects)
		if err != nil {
			printPath, _ := cmd.Flags().GetBool("print-path")
			if printPath {
				return
			}
			log.Fatal(err)
		}

		printPath, _ := cmd.Flags().GetBool("print-path")
		if printPath {
			fmt.Println(project.Path)
			return
		}
		startSession(project)
	},
}

func init() {
	searchCmd.Flags().Bool("print-path", false, "Print selected path to stdout instead of starting a tmux session")
}
