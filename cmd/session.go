package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/oschrenk/sessionizer/internal/tmux"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sessionCmd)
}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Print current session",
	Run: func(cmd *cobra.Command, args []string) {
		AsJson, _ := cmd.Flags().GetBool("json")

		server := new(tmux.Server)
		session, err := server.CurrentSession()
		if err != nil {
			// empty session
			fmt.Println("{}")
			return
		}
		if AsJson {
			json, _ := json.MarshalIndent(session, "", "  ")
			fmt.Println(string(json))
		} else {
			fmt.Println(session.Name)
		}
	},
}

func init() {
	sessionCmd.Flags().BoolP("json", "", false, "Print json")
}
