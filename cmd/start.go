package cmd

import (
	"fmt"
	"os"

	"github.com/oschrenk/sessionizer/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tmux with default session",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		defaultName := viper.GetString("default.name")
		defaultPath := os.ExpandEnv(viper.GetString("default.path"))

		err := core.StartSession(defaultName, defaultPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error switching to session: %s", defaultName)
			os.Exit(1)
		}
	},
}

func init() {
	// add flags and such here
}
