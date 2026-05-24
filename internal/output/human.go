package output

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/tkr-cli/tkr/internal/ticket"
)

type humanFormatter struct{}

var (
	cyan    = color.New(color.FgCyan, color.Bold).SprintFunc()
	bold    = color.New(color.Bold).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	blue    = color.New(color.FgBlue).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	dim     = color.New(color.Faint).SprintFunc()
	white   = color.New(color.FgWhite).SprintFunc()
	redBold = color.New(color.FgRed, color.Bold).SprintFunc()
)

func statusColor(s ticket.Status) string {
	switch s {
	case ticket.StatusDone:
		return green(string(s))
	case ticket.StatusInProgress:
		return yellow(string(s))
	case ticket.StatusTodo:
		return blue(string(s))
	case ticket.StatusBlocked:
		return red(string(s))
	case ticket.StatusInReview:
		return magenta(string(s))
	default:
		return string(s)
	}
}

func priorityColor(p ticket.Priority) string {
	switch p {
	case ticket.PriorityCritical:
		return red(string(p))
	case ticket.PriorityHigh:
		return yellow(string(p))
	case ticket.PriorityMedium:
		return white(string(p))
	case ticket.PriorityLow:
		return dim(string(p))
	default:
		return string(p)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func (f *humanFormatter) FormatTicket(t ticket.Ticket) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s  %s\n", cyan(t.ID), bold(t.Title))
	fmt.Fprintf(&b, "%-14s%-15s  Priority: %s\n", "Status:", statusColor(t.Status), priorityColor(t.Priority))

	if t.Type != "" {
		fmt.Fprintf(&b, "%-14s%s\n", "Type:", string(t.Type))
	}
	if t.Actor != "" {
		fmt.Fprintf(&b, "%-14s%s\n", "Actor:", string(t.Actor))
	}
	if t.Assignee != "" {
		fmt.Fprintf(&b, "%-14s%s\n", "Assignee:", t.Assignee)
	}
	if len(t.Dependencies) > 0 {
		fmt.Fprintf(&b, "%-14s%s\n", "Dependencies:", strings.Join(t.Dependencies, ", "))
	}
	if t.ParentID != "" {
		fmt.Fprintf(&b, "%-14s%s\n", "Parent:", t.ParentID)
	}
	if len(t.Blocks) > 0 {
		fmt.Fprintf(&b, "%-14s%s\n", "Blocks:", strings.Join(t.Blocks, ", "))
	}
	if len(t.RelatedTo) > 0 {
		fmt.Fprintf(&b, "%-14s%s\n", "Related To:", strings.Join(t.RelatedTo, ", "))
	}
	if len(t.Duplicates) > 0 {
		fmt.Fprintf(&b, "%-14s%s\n", "Duplicates:", strings.Join(t.Duplicates, ", "))
	}
	if len(t.Labels) > 0 {
		fmt.Fprintf(&b, "%-14s%s\n", "Labels:", strings.Join(t.Labels, ", "))
	}
	if t.Estimate != "" {
		fmt.Fprintf(&b, "%-14s%s\n", "Estimate:", t.Estimate)
	}
	if t.Complexity > 0 {
		fmt.Fprintf(&b, "%-14s%d\n", "Complexity:", t.Complexity)
	}
	if t.Branch != "" {
		fmt.Fprintf(&b, "%-14s%s\n", "Branch:", t.Branch)
	}
	if t.PR != "" {
		fmt.Fprintf(&b, "%-14s%s\n", "PR:", t.PR)
	}

	created := t.Created.Format("2006-01-02")
	updated := t.Updated.Format("2006-01-02")
	fmt.Fprintf(&b, "%-14s%-17sUpdated: %s\n", "Created:", created, updated)

	if len(t.AcceptanceCriteria) > 0 {
		b.WriteString("\nAcceptance Criteria:\n")
		for _, ac := range t.AcceptanceCriteria {
			fmt.Fprintf(&b, "  %s %s\n", green("✓"), ac)
		}
	}

	if t.Body != "" {
		fmt.Fprintf(&b, "\n%s\n", t.Body)
	}

	return b.String()
}

func (f *humanFormatter) FormatTicketList(tickets []ticket.Ticket) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
		bold("ID"), bold("Title"), bold("Status"), bold("Priority"), bold("Type"), bold("Assignee"))
	_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
		"------", "----------------------------------------", "-----------", "--------", "-------", "--------")

	for _, t := range tickets {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			cyan(t.ID),
			truncate(t.Title, 40),
			statusColor(t.Status),
			priorityColor(t.Priority),
			string(t.Type),
			t.Assignee,
		)
	}

	_ = w.Flush()
	return b.String()
}

func (f *humanFormatter) FormatNext(t *ticket.Ticket) string {
	if t == nil {
		return dim("No tickets available") + "\n"
	}
	return bold("Next ticket:") + "\n\n" + f.FormatTicket(*t)
}

func (f *humanFormatter) FormatError(err error) string {
	return redBold("error: ") + err.Error() + "\n"
}
