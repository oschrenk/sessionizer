package cmd

import (
	"github.com/oschrenk/sessionizer/tmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tmux with default session",
	Run: func(cmd *cobra.Command, args []string) {
		defaultName := viper.GetString("default.name")
		defaultPath := os.ExpandEnv(viper.GetString("default.path"))

		server := new(tmux.Server)
		server.CreateOrAttachSession(defaultName, defaultPath)
	},
}

func init() {
	// add flags and such here
}
