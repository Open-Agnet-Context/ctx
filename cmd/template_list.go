package cmd

import (
	"fmt"

	"ctx/internal/agent"
	"github.com/spf13/cobra"
)

func init() {
	templateCmd.AddCommand(templateListCmd)
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		builtIns := agent.BuiltInTemplateNames()
		repoTemplates, err := agent.ListRepoTemplates()
		if err != nil {
			return err
		}

		fmt.Println("Built-in templates:")
		if len(builtIns) == 0 {
			fmt.Println("- none")
		} else {
			for _, name := range builtIns {
				fmt.Printf("- %s\n", name)
			}
		}

		fmt.Println("Repo templates (.agent/templates):")
		if len(repoTemplates) == 0 {
			fmt.Println("- none")
		} else {
			for _, name := range repoTemplates {
				fmt.Printf("- %s\n", name)
			}
		}
		return nil
	},
}
