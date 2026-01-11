package cmd

import (
	"fmt"

	"ctx/internal/agent"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init <template>",
	Short: "Initialize .agent/ with starter files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]
		if err := agent.EnsureAgentLayout(templateName); err != nil {
			return err
		}
		fmt.Println("Initialized .agent/ with starter context.")
		return nil
	},
}
