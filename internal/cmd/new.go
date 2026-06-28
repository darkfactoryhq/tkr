package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/darkfactoryhq/tkr/internal/git"
	"github.com/darkfactoryhq/tkr/internal/store"
	"github.com/darkfactoryhq/tkr/internal/ticket"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new ticket",
	RunE:  runNew,
}

func init() {
	newCmd.Flags().StringP("title", "t", "", "Ticket title (required)")
	newCmd.Flags().StringP("priority", "p", "", "Priority: critical|high|medium|low")
	newCmd.Flags().String("actor", "", "Actor: agent or human")
	newCmd.Flags().String("type", "", "Type: feature, bug, spike, chore, docs")
	newCmd.Flags().StringSlice("dep", nil, "Dependencies (ticket IDs)")
	newCmd.Flags().StringSlice("label", nil, "Labels")
	newCmd.Flags().StringSlice("ac", nil, "Acceptance criteria")
	newCmd.Flags().Int("complexity", 0, "Complexity score (1-10)")
	newCmd.Flags().Bool("force", false, "Allow creation on non-default branch")
	newCmd.Flags().String("parent", "", "Parent ticket ID")
	newCmd.Flags().String("estimate", "", "Time estimate (e.g., 30m, 2h, 1d, 1w)")
	newCmd.Flags().String("assignee", "", "Human assignee")
	newCmd.Flags().StringSlice("blocks", nil, "Tickets this blocks")
	newCmd.Flags().StringSlice("related", nil, "Related tickets")
	newCmd.Flags().StringSlice("duplicates", nil, "Duplicate tickets")
}

func runNew(cmd *cobra.Command, args []string) error {
	title, _ := cmd.Flags().GetString("title")
	if title == "" {
		return fmt.Errorf("--title is required")
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force {
		isDefault, err := git.IsDefaultBranch()
		if err == nil && !isDefault {
			return fmt.Errorf("new tickets can only be created on the default branch (use --force to override)")
		}
	}

	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	id, err := s.NextID()
	if err != nil {
		return fmt.Errorf("generate ID: %w", err)
	}

	priority, _ := cmd.Flags().GetString("priority")
	if priority == "" {
		priority = string(s.Config().DefaultPriority)
	}
	if priority == "" {
		priority = "medium"
	}

	actorFlag, _ := cmd.Flags().GetString("actor")
	typeFlag, _ := cmd.Flags().GetString("type")
	deps, _ := cmd.Flags().GetStringSlice("dep")
	labels, _ := cmd.Flags().GetStringSlice("label")
	ac, _ := cmd.Flags().GetStringSlice("ac")
	complexity, _ := cmd.Flags().GetInt("complexity")
	parentID, _ := cmd.Flags().GetString("parent")
	estimate, _ := cmd.Flags().GetString("estimate")
	assignee, _ := cmd.Flags().GetString("assignee")
	blocks, _ := cmd.Flags().GetStringSlice("blocks")
	related, _ := cmd.Flags().GetStringSlice("related")
	duplicates, _ := cmd.Flags().GetStringSlice("duplicates")

	now := time.Now().UTC()
	t := &ticket.Ticket{
		ID:                 id,
		Title:              title,
		Status:             ticket.StatusTodo,
		Priority:           ticket.Priority(priority),
		Actor:              ticket.Actor(actorFlag),
		Type:               ticket.TicketType(typeFlag),
		Assignee:           assignee,
		Dependencies:       deps,
		ParentID:           parentID,
		Blocks:             blocks,
		RelatedTo:          related,
		Duplicates:         duplicates,
		Labels:             labels,
		Created:            now,
		Updated:            now,
		AcceptanceCriteria: ac,
		Complexity:         complexity,
		Estimate:           estimate,
	}

	if len(ac) == 0 && len(s.Config().DefaultDefinitionOfDone) > 0 {
		t.AcceptanceCriteria = make([]string, len(s.Config().DefaultDefinitionOfDone))
		copy(t.AcceptanceCriteria, s.Config().DefaultDefinitionOfDone)
	}

	errs := ticket.Validate(t)
	if len(errs) > 0 {
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
		}
		return fmt.Errorf("validation failed: %s", strings.Join(msgs, "; "))
	}

	if err := s.Save(t); err != nil {
		return fmt.Errorf("save ticket: %w", err)
	}

	f := formatter()
	fmt.Print(f.FormatTicket(*t))
	return nil
}

func storeFromCwd() (*store.Store, error) {
	root, err := store.FindRoot()
	if err != nil {
		return nil, fmt.Errorf("tkr not initialized (run 'tkr init')")
	}
	return store.New(root)
}
