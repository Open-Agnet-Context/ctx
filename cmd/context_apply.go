package cmd

import (
	"fmt"

	"ctx/internal/agent"
	"github.com/spf13/cobra"
)

func init() {
	contextCmd.AddCommand(contextApplyCmd)
}

var contextApplyCmd = &cobra.Command{
	Use:   "apply <template>",
	Short: "Apply a template to .agent/context.yaml",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]
		if err := agent.EnsureAgentExists(); err != nil {
			return err
		}
		resolved, err := agent.ApplyTemplateToContext(templateName)
		if err != nil {
			return err
		}
		fmt.Printf("Applied template %q to .agent/context.yaml\n", resolved)
		return nil
	},
}
