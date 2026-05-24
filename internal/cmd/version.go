package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print tkr version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tkr %s (commit: %s, built: %s)\n", buildVersion, buildCommit, buildDate)
	},
}
