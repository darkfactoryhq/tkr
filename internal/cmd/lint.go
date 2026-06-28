package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/darkfactoryhq/tkr/internal/dag"
	"github.com/darkfactoryhq/tkr/internal/ticket"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Validate all tickets and check for dependency cycles",
	RunE:  runLint,
}

func runLint(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	tickets, err := s.Load()
	if err != nil {
		return err
	}

	hasErrors := false

	for _, t := range tickets {
		errs := ticket.Validate(&t)
		for _, e := range errs {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s: %s: %s\n", t.ID, e.Field, e.Message)
			hasErrors = true
		}
	}

	idSet := make(map[string]bool)
	for _, t := range tickets {
		idSet[t.ID] = true
	}

	g := dag.BuildFromTickets(tickets)
	cr := dag.DetectCycles(g)
	if cr.HasCycle {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "dependency cycle detected involving: %v\n", cr.CycleNodes)
		hasErrors = true
	}

	for _, t := range tickets {
		for _, dep := range t.Dependencies {
			if !idSet[dep] {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: dependency %s does not exist\n", t.ID, dep)
				hasErrors = true
			}
		}

		if t.ParentID != "" && !idSet[t.ParentID] {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s: parent %s does not exist\n", t.ID, t.ParentID)
			hasErrors = true
		}

		for _, b := range t.Blocks {
			if !idSet[b] {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: blocks target %s does not exist\n", t.ID, b)
				hasErrors = true
			}
		}
		for _, r := range t.RelatedTo {
			if !idSet[r] {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: related target %s does not exist\n", t.ID, r)
				hasErrors = true
			}
		}
		for _, d := range t.Duplicates {
			if !idSet[d] {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: duplicate target %s does not exist\n", t.ID, d)
				hasErrors = true
			}
		}
	}

	parentCycles := dag.DetectParentCycles(g)
	if len(parentCycles) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "parent hierarchy cycle detected involving: %v\n", parentCycles)
		hasErrors = true
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("All %d tickets valid. No dependency cycles. No parent hierarchy cycles.\n", len(tickets))
	return nil
}
