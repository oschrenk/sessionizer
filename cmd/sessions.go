package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/oschrenk/sessionizer/tmux"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sessionsCmd)
}

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Print sessions",
	Run: func(cmd *cobra.Command, args []string) {
		detachedOnly, _ := cmd.Flags().GetBool("detached-only")
		AsJson, _ := cmd.Flags().GetBool("json")

		server := new(tmux.Server)
		sessions, err := server.ListSessions(detachedOnly)
		if err != nil {
			// empty array
			fmt.Println("[]")
			return
		}
		if AsJson {
			json, _ := json.MarshalIndent(sessions, "", "  ")
			fmt.Println(string(json))
		} else {
			for _, s := range sessions {
				fmt.Println(s.Name)
			}
		}
	},
}

func init() {
	sessionsCmd.Flags().BoolP("detached-only", "d", false, "Show detached sessions only")
	sessionsCmd.Flags().BoolP("json", "", false, "Print json")
}
