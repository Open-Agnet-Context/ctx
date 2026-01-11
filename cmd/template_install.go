package cmd

import (
	"fmt"

	"ctx/internal/agent"
	"github.com/spf13/cobra"
)

var (
	forceTemplateInstall bool
)

func init() {
	templateInstallCmd.Flags().BoolVar(&forceTemplateInstall, "force", false, "Overwrite existing template file")
	templateCmd.AddCommand(templateInstallCmd)
}

var templateInstallCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a built-in template into .agent/templates",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		dest, err := agent.InstallTemplate(name, forceTemplateInstall)
		if err != nil {
			return err
		}
		fmt.Printf("Installed template %q to %s\n", name, dest)
		return nil
	},
}
