package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/oschrenk/sessionizer/tmux"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(windowsCmd)
}

var windowsCmd = &cobra.Command{
	Use:   "windows",
	Short: "Print windows",
	Run: func(cmd *cobra.Command, args []string) {
		AsJson, _ := cmd.Flags().GetBool("json")

		server := new(tmux.Server)
		windows, err := server.ListWindows()
		if err != nil {
			return
		}
		if AsJson {
			json, _ := json.MarshalIndent(windows, "", "  ")
			fmt.Println(string(json))
		} else {
			for _, s := range windows {
				fmt.Println(s.Name)
			}
		}
	},
}

func init() {
	windowsCmd.Flags().BoolP("json", "", false, "Print json")
}
