package cmd

import (
	"fmt"
	"os"

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
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// read config file here
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$XDG_CONFIG_HOME/sessionizer")
	viper.AddConfigPath("$HOME/.config/sessionizer")

	viper.SetDefault("base.ignore", "")
	viper.SetDefault("default.name", "default")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintln(os.Stderr, "Config file not found")
		} else {
			fmt.Fprintln(os.Stderr, "Other error")
		}
		os.Exit(1)
	}
}
