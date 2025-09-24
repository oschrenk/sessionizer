package cmd

import (
	"fmt"
	"os"

	"github.com/oschrenk/sessionizer/tmuxp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(layoutCmd)
}

var layoutCmd = &cobra.Command{
	Use:    "layout <file>",
	Short:  "Reads layout from a yaml file (beta)",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		var layout tmuxp.Layout
		err = yaml.Unmarshal(data, &layout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing YAML: %v\n", err)
			os.Exit(1)
		}

		yamlData, err := yaml.Marshal(layout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting to YAML: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(yamlData))
	},
}
