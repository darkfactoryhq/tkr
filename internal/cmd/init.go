package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tkr-cli/tkr/internal/bootstrap"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize tkr in the current repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix, _ := cmd.Flags().GetString("prefix")
		noSkill, _ := cmd.Flags().GetBool("no-skill")

		return bootstrap.Init(".", bootstrap.InitOptions{
			Prefix:  prefix,
			NoSkill: noSkill,
		})
	},
}

func init() {
	initCmd.Flags().String("prefix", "TKR", "Ticket ID prefix")
	initCmd.Flags().Bool("no-skill", false, "Skip generating Claude Code Skill")
}
