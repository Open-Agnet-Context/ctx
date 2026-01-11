package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"ctx/internal/agent"
	"github.com/spf13/cobra"
)

func init() {
	workCmd.AddCommand(workStartCmd)
	workCmd.AddCommand(workStopCmd)
	rootCmd.AddCommand(workCmd)
}

var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Manage active work items",
}

var workStartCmd = &cobra.Command{
	Use:   "start <WI-XXX>",
	Short: "Mark a work item as active and suggest a branch name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := agent.EnsureAgentExists(); err != nil {
			return err
		}
		id := args[0]
		wi, err := agent.LoadWorkItem(id)
		if err != nil {
			return fmt.Errorf("could not load %s: %w", id, err)
		}

		state, err := agent.LoadState()
		if err != nil {
			return err
		}
		state.ActiveWorkItem = id
		state.BranchSuggestion = agent.SuggestBranchName(wi.Meta)
		if err := agent.SaveState(state); err != nil {
			return err
		}

		wi.Meta.Status = "active"
		wi.Meta.BranchSuggestion = state.BranchSuggestion
		if err := agent.SaveWorkItem(wi); err != nil {
			return err
		}

		fmt.Printf("Set %s as active. Suggested branch: %s\n", id, state.BranchSuggestion)
		return nil
	},
}

var workStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop active work and capture a one-line handoff summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := agent.EnsureAgentExists(); err != nil {
			return err
		}
		state, err := agent.LoadState()
		if err != nil {
			return err
		}
		if state.ActiveWorkItem == "" {
			return fmt.Errorf("no active work item to stop")
		}

		fmt.Print("One-line summary: ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		summary := strings.TrimSpace(line)

		wi, err := agent.LoadWorkItem(state.ActiveWorkItem)
		if err != nil {
			return err
		}
		wi.Meta.LastSummary = summary
		wi.Meta.Status = "paused"
		if err := agent.SaveWorkItem(wi); err != nil {
			return err
		}

		state.LastSummary = summary
		state.ActiveWorkItem = ""
		state.BranchSuggestion = ""
		if err := agent.SaveState(state); err != nil {
			return err
		}

		fmt.Println("Work stopped and summary captured.")
		return nil
	},
}
