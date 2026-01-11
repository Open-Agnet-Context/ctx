package cmd

import (
	"fmt"
	"strings"

	"ctx/internal/agent"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(issueCmd)
}

var issueCmd = &cobra.Command{
	Use:   "issue <text>",
	Short: "Create a new work item from natural language",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := agent.EnsureAgentExists(); err != nil {
			return err
		}
		title := strings.TrimSpace(strings.Join(args, " "))
		if title == "" {
			return fmt.Errorf("work item text cannot be empty")
		}
		id, err := agent.NextWorkItemID()
		if err != nil {
			return err
		}
		intents := agent.ClassifyIntent(title)
		wi := agent.NewWorkItemFile(id, title, intents)
		if err := agent.SaveWorkItem(wi); err != nil {
			return err
		}

		state, err := agent.LoadState()
		if err != nil {
			return err
		}
		state.ActiveWorkItem = id
		state.BranchSuggestion = ""
		state.LastSummary = ""
		if err := agent.SaveState(state); err != nil {
			return err
		}

		fmt.Printf("Created %s (%s) and set as active.\n", id, title)
		return nil
	},
}
