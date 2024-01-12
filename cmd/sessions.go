package cmd

import (
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

		server := new(tmux.Server)
		sessions, err := server.ListSessions(detachedOnly)
		if err != nil {
			return
		}
		for _, s := range sessions {
			fmt.Println(s.Name)
		}
	},
}

func init() {
	sessionsCmd.Flags().BoolP("detached-only", "d", false, "Show detached sessions only")
}
