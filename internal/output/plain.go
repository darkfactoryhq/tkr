package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/darkfactoryhq/tkr/internal/ticket"
)

type plainFormatter struct{}

func formatPlainTicket(t ticket.Ticket) string {
	var b strings.Builder

	fmt.Fprintf(&b, "id: %s\n", t.ID)
	fmt.Fprintf(&b, "title: %s\n", t.Title)
	fmt.Fprintf(&b, "status: %s\n", t.Status)
	fmt.Fprintf(&b, "priority: %s\n", t.Priority)

	if t.Type != "" {
		fmt.Fprintf(&b, "type: %s\n", string(t.Type))
	}
	if t.Actor != "" {
		fmt.Fprintf(&b, "actor: %s\n", string(t.Actor))
	}
	if t.Assignee != "" {
		fmt.Fprintf(&b, "assignee: %s\n", t.Assignee)
	}
	if len(t.Dependencies) > 0 {
		fmt.Fprintf(&b, "dependencies: %s\n", strings.Join(t.Dependencies, ", "))
	}
	if t.ParentID != "" {
		fmt.Fprintf(&b, "parent_id: %s\n", t.ParentID)
	}
	if len(t.Blocks) > 0 {
		fmt.Fprintf(&b, "blocks: %s\n", strings.Join(t.Blocks, ", "))
	}
	if len(t.RelatedTo) > 0 {
		fmt.Fprintf(&b, "related_to: %s\n", strings.Join(t.RelatedTo, ", "))
	}
	if len(t.Duplicates) > 0 {
		fmt.Fprintf(&b, "duplicates: %s\n", strings.Join(t.Duplicates, ", "))
	}
	if len(t.Labels) > 0 {
		fmt.Fprintf(&b, "labels: %s\n", strings.Join(t.Labels, ", "))
	}
	if t.Complexity > 0 {
		fmt.Fprintf(&b, "complexity: %d\n", t.Complexity)
	}
	if t.Estimate != "" {
		fmt.Fprintf(&b, "estimate: %s\n", t.Estimate)
	}
	if t.Branch != "" {
		fmt.Fprintf(&b, "branch: %s\n", t.Branch)
	}
	if t.PR != "" {
		fmt.Fprintf(&b, "pr: %s\n", t.PR)
	}

	fmt.Fprintf(&b, "created: %s\n", t.Created.Format(time.RFC3339))
	fmt.Fprintf(&b, "updated: %s\n", t.Updated.Format(time.RFC3339))

	if len(t.AcceptanceCriteria) > 0 {
		b.WriteString("acceptance_criteria:\n")
		for _, ac := range t.AcceptanceCriteria {
			fmt.Fprintf(&b, "  - %s\n", ac)
		}
	}

	if t.Body != "" {
		fmt.Fprintf(&b, "\n---\n\n%s\n", t.Body)
	}

	return b.String()
}

func (f *plainFormatter) FormatTicket(t ticket.Ticket) string {
	return formatPlainTicket(t)
}

func (f *plainFormatter) FormatTicketList(tickets []ticket.Ticket) string {
	var parts []string
	for _, t := range tickets {
		parts = append(parts, formatPlainTicket(t))
	}
	return strings.Join(parts, "\n")
}

func (f *plainFormatter) FormatNext(t *ticket.Ticket) string {
	if t == nil {
		return "no tickets available\n"
	}
	return formatPlainTicket(*t)
}

func (f *plainFormatter) FormatError(err error) string {
	return "error: " + err.Error() + "\n"
}
