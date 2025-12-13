package cmd

import (
	"log"
	"os"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/oschrenk/sessionizer/core"
	"github.com/oschrenk/sessionizer/internal/tmux"
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
		log.Fatal(err)
	}
	return projects[idx], nil
}

func startSession(project model.Entry) {
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
		config := model.Config{
			DefaultName:    viper.GetString("default.name"),
			DefaultPath:    os.ExpandEnv(viper.GetString("default.path")),
			SearchDirs:     mapF(viper.GetStringSlice("search.directories"), os.ExpandEnv),
			SearchEntries:  mapF(viper.GetStringSlice("search.entries"), os.ExpandEnv),
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
			log.Fatal(err)
		}

		startSession(project)
	},
}

func init() {
	// add flags and such here
}
