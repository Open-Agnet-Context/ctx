package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ctx",
	Short: "Repository-local context and intent manager for coding agents",
	Long:  "ctx manages project context, state, work items, evidence, and prompt assembly offline inside the repo.",
}

// Execute runs the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
