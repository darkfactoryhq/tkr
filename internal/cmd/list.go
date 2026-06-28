package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/darkfactoryhq/tkr/internal/store"
	"github.com/darkfactoryhq/tkr/internal/ticket"
)

var listCmd = &cobra.Command{
	Use:   "list [filters...]",
	Short: "List tickets with optional filtering",
	RunE:  runList,
}

func init() {
	listCmd.Flags().String("sort", "", "Sort field")
	listCmd.Flags().Int("limit", 0, "Limit number of results")
}

func runList(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	tickets, err := s.Load()
	if err != nil {
		return err
	}

	if len(args) > 0 {
		filter, err := store.ParseFilter(args)
		if err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}
		var filtered []ticket.Ticket
		for _, t := range tickets {
			if filter.Match(t) {
				filtered = append(filtered, t)
			}
		}
		tickets = filtered
	}

	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 && len(tickets) > limit {
		tickets = tickets[:limit]
	}

	f := formatter()
	fmt.Print(f.FormatTicketList(tickets))
	return nil
}
