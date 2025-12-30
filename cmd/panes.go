package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/oschrenk/sessionizer/internal/tmux"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(panesCmd)
}

var panesCmd = &cobra.Command{
	Use:   "panes",
	Short: "Print panes",
	Run: func(cmd *cobra.Command, args []string) {
		AsJson, _ := cmd.Flags().GetBool("json")

		server := new(tmux.Server)
		currentWindow, err := server.CurrentWindow()
		if err != nil {
			return
		}

		panes, err := server.ListPanes(currentWindow.Id)
		if err != nil {
			return
		}
		if AsJson {
			json, _ := json.MarshalIndent(panes, "", "  ")
			fmt.Println(string(json))
		} else {
			for _, p := range panes {
				fmt.Println(p.Id)
			}
		}
	},
}

func init() {
	panesCmd.Flags().BoolP("json", "", false, "Print json")
}
