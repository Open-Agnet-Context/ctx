package cmd

import (
	"fmt"
	"os"

	"ctx/internal/agent"
	"github.com/spf13/cobra"
)

func init() {
	evidenceCmd.AddCommand(evidenceAddCmd)
	rootCmd.AddCommand(evidenceCmd)
}

var evidenceCmd = &cobra.Command{
	Use:   "evidence",
	Short: "Manage evidence for the active work item",
}

var evidenceAddCmd = &cobra.Command{
	Use:   "add <file>",
	Short: "Copy evidence into .agent/evidence/ and link it to the active work item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := agent.EnsureAgentExists(); err != nil {
			return err
		}
		state, err := agent.LoadState()
		if err != nil {
			return err
		}
		if state.ActiveWorkItem == "" {
			return fmt.Errorf("no active work item; start one with ctx work start <WI-XXX>")
		}

		src := args[0]
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("source evidence %q not found: %w", src, err)
		}

		destRel, err := agent.CopyEvidence(src)
		if err != nil {
			return err
		}

		wi, err := agent.LoadWorkItem(state.ActiveWorkItem)
		if err != nil {
			return err
		}
		wi.Meta.Evidence = append(wi.Meta.Evidence, destRel)
		if err := agent.SaveWorkItem(wi); err != nil {
			return err
		}

		fmt.Printf("Added evidence %s to %s.\n", destRel, wi.Meta.ID)
		return nil
	},
}
