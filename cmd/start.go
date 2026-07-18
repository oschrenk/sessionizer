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
		// -n overrides default.name from config; one of them is required.
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			name = viper.GetString("default.name")
		}
		if strings.TrimSpace(name) == "" {
			fmt.Fprintln(os.Stderr, "No session name: pass -n <name> or set default.name in config")
			os.Exit(1)
		}

		defaultPath := os.ExpandEnv(viper.GetString("default.path"))
		defaultLayoutPath := viper.GetString("default.layout_path")

		configDir := filepath.Dir(viper.ConfigFileUsed())
		err := core.StartSession(name, defaultPath, "", defaultLayoutPath, configDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error switching to session: %s", name)
			os.Exit(1)
		}
	},
}

func init() {
	startCmd.Flags().StringP("name", "n", "", "Session name to start (overrides default.name)")
}
