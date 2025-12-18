package cmd

import (
	"fmt"
	"os"

	"github.com/oschrenk/sessionizer/core"
	"github.com/oschrenk/sessionizer/internal/tmux"
	"github.com/oschrenk/sessionizer/internal/tmuxp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(layoutCmd)
	layoutCmd.Flags().BoolP("apply", "a", false, "Apply the layout")
}

var layoutCmd = &cobra.Command{
	Use:    "layout <file>",
	Short:  "Reads layout from a yaml file (beta)",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		apply, _ := cmd.Flags().GetBool("apply")

		layout, err := tmuxp.ReadLayoutFromFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading layout file: %v\n", err)
			os.Exit(1)
		}

		if apply {
			server := &tmux.Server{}
			err := core.ApplyLayout(server, layout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error applying layout: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Applied")
			return
		}

		yamlData, err := yaml.Marshal(layout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting to YAML: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(yamlData))
	},
}
