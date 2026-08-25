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
	rootCmd.PersistentFlags().String("tmux-path", "", "path to the tmux executable")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		// non-fatal config read so base.* applies to every command, including
		// read-only ones that don't otherwise require a config file.
		configureViper()
		_ = viper.ReadInConfig()

		tmux.SetSocket(resolve(cmd, "socket-name", "SESSIONIZER_SOCKET_NAME", "base.socket_name"))
		tmux.SetBinary(resolve(cmd, "tmux-path", "SESSIONIZER_TMUX_PATH", "base.tmux_path"))
	}
}

// resolve picks a setting from, in order: the named flag, the env var, the
// config key. It expects the config to have been read already.
func resolve(cmd *cobra.Command, flag string, env string, key string) string {
	if s, _ := cmd.Flags().GetString(flag); s != "" {
		return s
	}
	if s := os.Getenv(env); s != "" {
		return s
	}
	return viper.GetString(key)
}

// configureViper sets the config file name, type, search paths and defaults.
// Shared by initConfig (fatal read) and the root PersistentPreRun (non-fatal read).
func configureViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$XDG_CONFIG_HOME/sessionizer")
	viper.AddConfigPath("$HOME/.config/sessionizer")

	viper.SetDefault("base.ignore", "")
	viper.SetDefault("base.tmux_path", "tmux")
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
