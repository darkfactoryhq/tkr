package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/darkfactoryhq/tkr/internal/git"
	"github.com/darkfactoryhq/tkr/internal/ticket"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Install git hooks for automatic status updates",
	RunE:  runWatch,
}

func init() {
	watchCmd.Flags().Bool("uninstall", false, "Remove tkr git hooks")
}

func runWatch(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	uninstall, _ := cmd.Flags().GetBool("uninstall")
	if uninstall {
		if err := git.UninstallHooks(repoRoot); err != nil {
			return err
		}
		fmt.Println("Removed tkr git hooks")
		return nil
	}

	if err := git.InstallHooks(repoRoot); err != nil {
		return err
	}
	fmt.Println("Installed git hooks. Commit messages with 'Closes TKR-N' will auto-update ticket status.")
	return nil
}

var hookPostCommitCmd = &cobra.Command{
	Use:    "_hook-post-commit",
	Hidden: true,
	RunE:   runHookPostCommit,
}

func runHookPostCommit(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return nil
	}

	msg, err := git.LatestCommitMessage()
	if err != nil {
		return nil
	}

	refs := git.ParseCommitMessage(msg)
	if len(refs) == 0 {
		return nil
	}

	for _, ref := range refs {
		t, err := s.Get(ref.TicketID)
		if err != nil {
			continue
		}

		switch ref.Action {
		case git.RefClose:
			if t.Status != ticket.StatusDone {
				t.Status = ticket.StatusDone
				_ = s.Save(t)
			}
		case git.RefWorkOn:
			if t.Status == ticket.StatusTodo {
				t.Status = ticket.StatusInProgress
				_ = s.Save(t)
			}
		}
	}

	return nil
}
