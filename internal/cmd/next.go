package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tkr-cli/tkr/internal/dag"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Get the highest-priority unblocked ticket",
	RunE:  runNext,
}

func init() {
	nextCmd.Flags().String("actor", "", "Filter by actor (agent or human)")
	nextCmd.Flags().String("assignee", "", "Filter by human assignee")
	nextCmd.Flags().StringSlice("label", nil, "Filter by labels")
	nextCmd.Flags().String("type", "", "Filter by ticket type")
}

func runNext(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	tickets, err := s.Load()
	if err != nil {
		return err
	}

	g := dag.BuildFromTickets(tickets)

	opts := dag.NextOptions{}
	opts.Actor, _ = cmd.Flags().GetString("actor")
	opts.Assignee, _ = cmd.Flags().GetString("assignee")
	opts.Labels, _ = cmd.Flags().GetStringSlice("label")
	opts.Type, _ = cmd.Flags().GetString("type")

	result, err := dag.NextTicket(tickets, g, opts)
	if err != nil {
		return err
	}

	f := formatter()
	fmt.Print(f.FormatNext(result))

	if result == nil {
		os.Exit(10)
	}
	return nil
}
