package cmd

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tkr-cli/tkr/internal/ticket"
)

var claimCmd = &cobra.Command{
	Use:   "claim <id>",
	Short: "Mark a ticket as in-progress",
	Args:  cobra.ExactArgs(1),
	RunE:  runClaim,
}

func init() {
	claimCmd.Flags().String("actor", "", "Set the actor: agent or human")
	claimCmd.Flags().String("assignee", "", "Set the human assignee")
	claimCmd.Flags().Bool("no-branch", false, "Skip automatic branch creation")
}

func runClaim(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	t, err := s.Get(args[0])
	if err != nil {
		return err
	}

	if t.Status == ticket.StatusInProgress {
		return fmt.Errorf("ticket %s is already in-progress", t.ID)
	}
	if t.Status == ticket.StatusDone {
		return fmt.Errorf("ticket %s is already done", t.ID)
	}

	t.Status = ticket.StatusInProgress

	actorFlag, _ := cmd.Flags().GetString("actor")
	if actorFlag != "" {
		t.Actor = ticket.Actor(actorFlag)
	}

	assignee, _ := cmd.Flags().GetString("assignee")
	if assignee != "" {
		t.Assignee = assignee
	}

	noBranch, _ := cmd.Flags().GetBool("no-branch")
	cfg := s.Config()
	if cfg.AutoBranch && !noBranch {
		branch := expandBranchPattern(cfg.BranchPattern, t)
		if err := exec.Command("git", "checkout", "-b", branch).Run(); err == nil {
			t.Branch = branch
		}
	}

	if err := s.Save(t); err != nil {
		return err
	}

	f := formatter()
	fmt.Print(f.FormatTicket(*t))
	return nil
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func expandBranchPattern(pattern string, t *ticket.Ticket) string {
	slug := strings.ToLower(t.Title)
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 50 {
		slug = slug[:50]
		slug = strings.TrimRight(slug, "-")
	}

	branch := strings.ReplaceAll(pattern, "{id}", strings.ToLower(t.ID))
	branch = strings.ReplaceAll(branch, "{slug}", slug)
	return branch
}
