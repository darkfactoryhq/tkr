package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkr-cli/tkr/internal/migrate"
	"github.com/tkr-cli/tkr/internal/ticket"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import tickets from other tools",
}

var importBacklogCmd = &cobra.Command{
	Use:   "backlog [path]",
	Short: "Import tickets from Backlog.md format",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runImportBacklog,
}

var importTaskmasterCmd = &cobra.Command{
	Use:   "taskmaster [path]",
	Short: "Import tickets from claude-task-master format",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runImportTaskmaster,
}

func init() {
	importBacklogCmd.Flags().Bool("dry-run", false, "Preview import without creating tickets")
	importTaskmasterCmd.Flags().Bool("dry-run", false, "Preview import without creating tickets")

	importCmd.AddCommand(importBacklogCmd)
	importCmd.AddCommand(importTaskmasterCmd)
}

func runImportBacklog(cmd *cobra.Command, args []string) error {
	dir := "backlog/tasks/"
	if len(args) > 0 {
		dir = args[0]
	}

	tasks, err := migrate.ParseBacklogDir(dir)
	if err != nil {
		return fmt.Errorf("parsing backlog: %w", err)
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return importTasks(tasks, "backlog", dryRun)
}

func runImportTaskmaster(cmd *cobra.Command, args []string) error {
	path := ".taskmaster/tasks/tasks.json"
	if len(args) > 0 {
		path = args[0]
	}

	tasks, err := migrate.ParseTaskmasterFile(path)
	if err != nil {
		return fmt.Errorf("parsing taskmaster: %w", err)
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return importTasks(tasks, "taskmaster", dryRun)
}

func importTasks(tasks []migrate.ImportedTask, source string, dryRun bool) error {
	if len(tasks) == 0 {
		fmt.Println("No tickets found to import.")
		return nil
	}

	var s interface {
		NextID() (string, error)
		Save(t *ticket.Ticket) error
	}

	if !dryRun {
		var err error
		s, err = storeFromCwd()
		if err != nil {
			return err
		}
	}

	// Phase 1: assign new IDs and build the mapping.
	idMap := make(map[string]string)
	tickets := make([]*ticket.Ticket, 0, len(tasks))

	for _, it := range tasks {
		var newID string
		if dryRun {
			newID = fmt.Sprintf("TKR-?%s", it.OriginalID)
		} else {
			var err error
			newID, err = s.NextID()
			if err != nil {
				return fmt.Errorf("generating ID: %w", err)
			}
		}
		idMap[it.OriginalID] = newID

		now := time.Now().UTC()
		created := it.Created
		if created.IsZero() {
			created = now
		}

		t := &ticket.Ticket{
			ID:                 newID,
			Title:              it.Title,
			Status:             ticket.Status(it.Status),
			Priority:           ticket.Priority(it.Priority),
			Actor:              ticket.Actor(it.Actor),
			Type:               ticket.TicketType(it.Type),
			Assignee:           it.Assignee,
			Labels:             it.Labels,
			AcceptanceCriteria: it.AcceptanceCriteria,
			Complexity:         it.Complexity,
			Estimate:           it.Estimate,
			Body:               it.Body,
			Created:            created,
			Updated:            now,
		}
		tickets = append(tickets, t)
	}

	// Phase 2: remap references using idMap.
	for i, it := range tasks {
		t := tickets[i]

		// Dependencies
		t.Dependencies = nil
		for _, dep := range it.Dependencies {
			if mapped, ok := idMap[dep]; ok {
				t.Dependencies = append(t.Dependencies, mapped)
			}
		}

		// ParentID
		t.ParentID = ""
		if it.ParentID != "" {
			if mapped, ok := idMap[it.ParentID]; ok {
				t.ParentID = mapped
			}
		}

		// Blocks
		t.Blocks = nil
		for _, b := range it.Blocks {
			if mapped, ok := idMap[b]; ok {
				t.Blocks = append(t.Blocks, mapped)
			}
		}

		// RelatedTo
		t.RelatedTo = nil
		for _, r := range it.RelatedTo {
			if mapped, ok := idMap[r]; ok {
				t.RelatedTo = append(t.RelatedTo, mapped)
			}
		}

		// Duplicates
		t.Duplicates = nil
		for _, d := range it.Duplicates {
			if mapped, ok := idMap[d]; ok {
				t.Duplicates = append(t.Duplicates, mapped)
			}
		}
	}

	// Phase 3: save tickets.
	if !dryRun {
		for _, t := range tickets {
			if err := s.Save(t); err != nil {
				return fmt.Errorf("saving ticket %s: %w", t.ID, err)
			}
		}
	}

	// Phase 4: print summary.
	fmt.Printf("Imported %d tickets from %s.\n", len(tickets), source)
	for i, it := range tasks {
		fmt.Printf("  %s → %s  %s\n", it.OriginalID, tickets[i].ID, tickets[i].Title)
	}

	return nil
}
