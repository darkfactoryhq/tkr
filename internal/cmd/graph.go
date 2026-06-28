package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/darkfactoryhq/tkr/internal/dag"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Output dependency graph as Mermaid diagram",
	RunE:  runGraph,
}

func runGraph(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	tickets, err := s.Load()
	if err != nil {
		return err
	}

	g := dag.BuildFromTickets(tickets)
	fmt.Println(dag.ToMermaid(tickets, g))
	return nil
}
