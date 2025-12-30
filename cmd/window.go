package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/oschrenk/sessionizer/internal/tmux"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(windowCmd)
}

var windowCmd = &cobra.Command{
	Use:   "window",
	Short: "Print current window",
	Run: func(cmd *cobra.Command, args []string) {
		AsJson, _ := cmd.Flags().GetBool("json")

		server := new(tmux.Server)

		window, err := server.CurrentWindow()
		if err != nil {
			return
		}
		if AsJson {
			json, _ := json.MarshalIndent(window, "", "  ")
			fmt.Println(string(json))
		} else {
			fmt.Println(window.Name)
		}
	},
}

func init() {
	windowCmd.Flags().BoolP("json", "", false, "Print json")
}
