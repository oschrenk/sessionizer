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
// SESSIONIZER_SOCKET_NAME env var, else base.socket_name from config. Empty means
// the default socket / ambient $TMUX.
func resolveSocketName(cmd *cobra.Command) string {
	if s, _ := cmd.Flags().GetString("socket-name"); s != "" {
		return s
	}
	if s := os.Getenv("SESSIONIZER_SOCKET_NAME"); s != "" {
		return s
	}
	// non-fatal config read so base.socket_name applies to every command,
	// including read-only ones that don't otherwise require a config file.
	configureViper()
	_ = viper.ReadInConfig()
	return viper.GetString("base.socket_name")
}

// configureViper sets the config file name, type, search paths and defaults.
// Shared by initConfig (fatal read) and resolveSocketName (non-fatal read).
func configureViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$XDG_CONFIG_HOME/sessionizer")
	viper.AddConfigPath("$HOME/.config/sessionizer")

	viper.SetDefault("base.ignore", "")
}

func initConfig() {
	configureViper()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintln(os.Stderr, "Config file not found")
		} else {
			fmt.Fprintln(os.Stderr, "Other error")
		}
		os.Exit(1)
	}
}
