package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Permanently remove a ticket",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().Bool("force", false, "Delete without confirmation")
}

func runDelete(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	t, err := s.Get(args[0])
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force {
		fmt.Fprintf(os.Stderr, "Delete %s (%s)? [y/N] ", t.ID, t.Title)
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			return fmt.Errorf("aborted")
		}
	}

	if err := os.Remove(t.FilePath); err != nil {
		return fmt.Errorf("removing ticket file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Deleted %s\n", t.ID)
	return nil
}
