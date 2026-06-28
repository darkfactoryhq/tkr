package dag

import (
	"fmt"
	"strings"

	"github.com/darkfactoryhq/tkr/internal/ticket"
)

func ToMermaid(tickets []ticket.Ticket, g *Graph) string {
	var b strings.Builder

	statusByID := make(map[string]ticket.Status)
	titleByID := make(map[string]string)
	for _, t := range tickets {
		statusByID[t.ID] = t.Status
		titleByID[t.ID] = t.Title
	}

	b.WriteString("graph TD\n")

	// Determine which nodes are children (rendered inside a subgraph).
	inSubgraph := make(map[string]bool)
	for _, id := range g.Nodes() {
		for _, child := range g.Children(id) {
			inSubgraph[child] = true
		}
	}

	// Render parent nodes as subgraphs containing their children.
	rendered := make(map[string]bool)
	for _, id := range g.Nodes() {
		kids := g.Children(id)
		if len(kids) == 0 {
			continue
		}
		title := titleByID[id]
		status := statusClass(statusByID[id])
		fmt.Fprintf(&b, "    subgraph %s [\"%s: %s\"]\n", id, id, title)
		// Declare the parent node inside its own subgraph for styling.
		fmt.Fprintf(&b, "        %s_node[\"%s\"]:::%s\n", id, id, status)
		rendered[id] = true
		for _, child := range kids {
			ct := titleByID[child]
			cs := statusClass(statusByID[child])
			fmt.Fprintf(&b, "        %s[\"%s: %s\"]:::%s\n", child, child, ct, cs)
			rendered[child] = true
		}
		b.WriteString("    end\n")
	}

	// Render remaining nodes not already in a subgraph.
	for _, id := range g.Nodes() {
		if rendered[id] {
			continue
		}
		title := titleByID[id]
		status := statusClass(statusByID[id])
		fmt.Fprintf(&b, "    %s[\"%s: %s\"]:::%s\n", id, id, title, status)
	}

	// Render dependency/blocks edges (real DAG edges).
	for _, id := range g.Nodes() {
		for _, dep := range g.Dependencies(id) {
			et := g.GetEdgeType(id, dep)
			switch et {
			case EdgeBlocks:
				fmt.Fprintf(&b, "    %s ==>|blocks| %s\n", id, dep)
			default:
				fmt.Fprintf(&b, "    %s --> %s\n", id, dep)
			}
		}
	}

	// Render informational edges (related, duplicates).
	for from, targets := range g.edgeTypes {
		for to, et := range targets {
			switch et {
			case EdgeRelated:
				fmt.Fprintf(&b, "    %s -.->|related| %s\n", from, to)
			case EdgeDuplicate:
				fmt.Fprintf(&b, "    %s -.->|duplicates| %s\n", from, to)
			}
		}
	}

	// Render parent-child edges.
	for _, id := range g.Nodes() {
		parent := g.Parent(id)
		if parent != "" {
			fmt.Fprintf(&b, "    %s ---|child of| %s\n", id, parent)
		}
	}

	b.WriteString("    classDef done fill:#2d6,stroke:#333\n")
	b.WriteString("    classDef todo fill:#69f,stroke:#333\n")
	b.WriteString("    classDef in-progress fill:#f92,stroke:#333\n")
	b.WriteString("    classDef blocked fill:#f33,stroke:#333\n")

	return b.String()
}

func statusClass(s ticket.Status) string {
	switch s {
	case ticket.StatusDone:
		return "done"
	case ticket.StatusTodo:
		return "todo"
	case ticket.StatusInProgress:
		return "in-progress"
	case ticket.StatusBlocked:
		return "blocked"
	default:
		return "todo"
	}
}
