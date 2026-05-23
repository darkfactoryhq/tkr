package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tkr-cli/tkr/internal/ticket"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show ticket details",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	t, err := s.Get(args[0])
	if err != nil {
		return err
	}

	f := formatter()
	fmt.Print(f.FormatTicket(*t))

	allTickets, err := s.Load()
	if err == nil {
		var children []ticket.Ticket
		for _, c := range allTickets {
			if c.ParentID == t.ID {
				children = append(children, c)
			}
		}
		if len(children) > 0 {
			fmt.Printf("\nChildren (%d):\n", len(children))
			for _, c := range children {
				fmt.Printf("  %s  %s  [%s]\n", c.ID, c.Title, c.Status)
			}
		}
	}

	return nil
}
