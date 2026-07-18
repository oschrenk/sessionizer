package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		if strings.TrimSpace(defaultName) == "" {
			fmt.Fprintln(os.Stderr, "No session name: set default.name in config")
			os.Exit(1)
		}

		defaultPath := os.ExpandEnv(viper.GetString("default.path"))
		defaultLayoutPath := viper.GetString("default.layout_path")

		configDir := filepath.Dir(viper.ConfigFileUsed())
		err := core.StartSession(defaultName, defaultPath, "", defaultLayoutPath, configDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error switching to session: %s", defaultName)
			os.Exit(1)
		}
	},
}

func init() {
	// add flags and such here
}
