package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tkr-cli/tkr/internal/ticket"
)

var txtCmd = &cobra.Command{
	Use:   "txt",
	Short: "Output all tickets in a single-file, grep-friendly todo.txt format",
	RunE:  runTxt,
}

func init() {
	txtCmd.Flags().Bool("write", false, "Write output to .tkr/tickets.txt instead of stdout")
}

func runTxt(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	tickets, err := s.Load()
	if err != nil {
		return err
	}

	sortTicketsForTxt(tickets)

	var lines []string
	for i := range tickets {
		lines = append(lines, formatTxtLine(&tickets[i]))
	}

	output := strings.Join(lines, "\n")
	if len(lines) > 0 {
		output += "\n"
	}

	writeFlag, _ := cmd.Flags().GetBool("write")
	if writeFlag {
		outPath := filepath.Join(s.Root(), ".tkr", "tickets.txt")
		if err := os.WriteFile(outPath, []byte(output), 0o644); err != nil {
			return fmt.Errorf("writing tickets.txt: %w", err)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "Wrote %s\n", outPath)
		return nil
	}

	fmt.Print(output)
	return nil
}

func sortTicketsForTxt(tickets []ticket.Ticket) {
	sort.SliceStable(tickets, func(i, j int) bool {
		di := tickets[i].Status == ticket.StatusDone
		dj := tickets[j].Status == ticket.StatusDone
		if di != dj {
			return !di // done tickets last
		}
		ri := tickets[i].Priority.Rank()
		rj := tickets[j].Priority.Rank()
		if ri != rj {
			return ri < rj // lower rank = higher priority
		}
		return tickets[i].ID < tickets[j].ID
	})
}

func priorityLetter(p ticket.Priority) string {
	switch p {
	case ticket.PriorityCritical:
		return "A"
	case ticket.PriorityHigh:
		return "B"
	case ticket.PriorityMedium:
		return "C"
	case ticket.PriorityLow:
		return "D"
	default:
		return "C"
	}
}

func formatTxtLine(t *ticket.Ticket) string {
	var parts []string

	// Prefix: done or priority
	if t.Status == ticket.StatusDone {
		parts = append(parts, "x")
	} else {
		parts = append(parts, fmt.Sprintf("(%s)", priorityLetter(t.Priority)))
	}

	// Creation date
	parts = append(parts, t.Created.Format("2006-01-02"))

	// Ticket ID
	parts = append(parts, t.ID)

	// Title
	parts = append(parts, t.Title)

	// Metadata tags
	for _, l := range t.Labels {
		parts = append(parts, "+"+l)
	}
	if t.Actor != "" {
		parts = append(parts, "actor:"+string(t.Actor))
	}
	if t.Type != "" {
		parts = append(parts, "type:"+string(t.Type))
	}
	if t.Assignee != "" {
		parts = append(parts, "assignee:"+t.Assignee)
	}
	if t.Estimate != "" {
		parts = append(parts, "est:"+t.Estimate)
	}
	if t.Complexity > 0 {
		parts = append(parts, fmt.Sprintf("cplx:%d", t.Complexity))
	}
	for _, d := range t.Dependencies {
		parts = append(parts, "dep:"+d)
	}
	for _, b := range t.Blocks {
		parts = append(parts, "blocks:"+b)
	}
	if t.ParentID != "" {
		parts = append(parts, "parent:"+t.ParentID)
	}
	// Status tag for non-todo/non-done statuses
	if t.Status != ticket.StatusTodo && t.Status != ticket.StatusDone {
		parts = append(parts, "status:"+string(t.Status))
	}

	return strings.Join(parts, " ")
}
