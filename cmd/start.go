package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oschrenk/sessionizer/core"
	"github.com/oschrenk/sessionizer/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tmux with the startup session",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		// -n overrides startup.name from config; one of them is required.
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			name = viper.GetString("startup.name")
		}
		if strings.TrimSpace(name) == "" {
			fmt.Fprintln(os.Stderr, "No session name: pass -n <name> or set startup.name in config")
			os.Exit(1)
		}

		// built through StartupEntry so `start` inherits default.layout on the
		// same terms as the finder does
		entry := core.StartupEntry(model.Config{
			StartupName:       name,
			StartupPath:       os.ExpandEnv(viper.GetString("startup.path")),
			StartupLayout:     viper.GetString("startup.layout"),
			StartupLayoutPath: viper.GetString("startup.layout_path"),
			DefaultLayout:     viper.GetString("default.layout"),
		})

		configDir := filepath.Dir(viper.ConfigFileUsed())
		err := core.StartSession(entry.Label, entry.Path, entry.Layout, entry.LayoutPath, configDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error switching to session: %s", name)
			os.Exit(1)
		}
	},
}

func init() {
	startCmd.Flags().StringP("name", "n", "", "Session name to start (overrides startup.name)")
}
