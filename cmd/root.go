package cmd

import (
	"fmt"
	"os"

	"github.com/oschrenk/sessionizer/internal/tmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:     "sessionizer",
	Short:   "Fuzzy finder for tmux sessions",
	Version: Version,
	Run:     func(cmd *cobra.Command, args []string) {},
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("socket-name", "s", "", "tmux socket name (tmux -L)")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		tmux.SetSocket(resolveSocketName(cmd))
	}
}

// resolveSocketName picks the tmux socket name: the --socket-name flag, else the
// SESSIONIZER_SOCKET_NAME env var. Empty means the default socket / ambient $TMUX.
// (base.socket_name config fallback is added in a follow-up commit.)
func resolveSocketName(cmd *cobra.Command) string {
	if s, _ := cmd.Flags().GetString("socket-name"); s != "" {
		return s
	}
	return os.Getenv("SESSIONIZER_SOCKET_NAME")
}

func initConfig() {
	// read config file here
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$XDG_CONFIG_HOME/sessionizer")
	viper.AddConfigPath("$HOME/.config/sessionizer")

	viper.SetDefault("base.ignore", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintln(os.Stderr, "Config file not found")
		} else {
			fmt.Fprintln(os.Stderr, "Other error")
		}
		os.Exit(1)
	}
}
