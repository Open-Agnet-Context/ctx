package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(templateCmd)
}

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage ctx templates",
}
