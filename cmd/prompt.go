package cmd

import (
	"fmt"

	"ctx/internal/agent"
	"github.com/spf13/cobra"
)

func init() {
	promptCmd.Flags().StringP("profile", "p", "cheap", "Prompt profile to use (cheap|standard|deep)")
	rootCmd.AddCommand(promptCmd)
}

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Generate the agent prompt for the active work item",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := agent.EnsureAgentExists(); err != nil {
			return err
		}
		profile, _ := cmd.Flags().GetString("profile")
		dest, err := agent.BuildPrompt(profile)
		if err != nil {
			return err
		}
		fmt.Printf("Prompt written to %s\n", dest)
		return nil
	},
}
