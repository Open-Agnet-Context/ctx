package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(contextCmd)
}

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage project context",
}
