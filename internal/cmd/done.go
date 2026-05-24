package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tkr-cli/tkr/internal/ticket"
)

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a ticket as done",
	Args:  cobra.ExactArgs(1),
	RunE:  runDone,
}

func init() {
	doneCmd.Flags().Bool("force", false, "Skip acceptance criteria check")
}

func runDone(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	t, err := s.Get(args[0])
	if err != nil {
		return err
	}

	if t.Status == ticket.StatusDone {
		return fmt.Errorf("ticket %s is already done", t.ID)
	}

	force, _ := cmd.Flags().GetBool("force")

	if !force {
		allTickets, err := s.Load()
		if err == nil {
			var incompleteChildren []string
			for _, child := range allTickets {
				if child.ParentID == t.ID && child.Status != ticket.StatusDone {
					incompleteChildren = append(incompleteChildren, child.ID)
				}
			}
			if len(incompleteChildren) > 0 {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "WARNING: ticket %s has incomplete children: %s\n", t.ID, strings.Join(incompleteChildren, ", "))
			}
		}
	}

	t.Status = ticket.StatusDone

	if err := s.Save(t); err != nil {
		return err
	}

	f := formatter()
	fmt.Print(f.FormatTicket(*t))
	return nil
}
