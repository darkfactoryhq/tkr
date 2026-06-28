package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/darkfactoryhq/tkr/internal/output"
)

var (
	flagPlain   bool
	flagJSON    bool
	flagNoColor bool

	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func SetVersion(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
}

func formatter() output.Formatter {
	return output.New(output.DetectMode(flagPlain, flagJSON))
}

var rootCmd = &cobra.Command{
	Use:   "tkr",
	Short: "Markdown-native, git-resident project management for AI coding agents",
	Long:  "tkr stores tickets as Markdown files with YAML frontmatter in .tkr/tickets/.\nDesigned for AI coding agents: deterministic CLI, dependency-aware task selection, agent routing.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagPlain, "plain", false, "Output deterministic plain text for LLM consumption")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output structured JSON")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable ANSI color codes")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(nextCmd)
	rootCmd.AddCommand(claimCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(releaseCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(hookPostCommitCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(txtCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
