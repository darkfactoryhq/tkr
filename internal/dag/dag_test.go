package dag

import (
	"strings"
	"testing"
	"time"

	"github.com/tkr-cli/tkr/internal/ticket"
)

func newTicket(id string, status ticket.Status, priority ticket.Priority, complexity int, deps []string, labels []string) ticket.Ticket {
	return ticket.Ticket{
		ID:           id,
		Title:        "Ticket " + id,
		Status:       status,
		Priority:     priority,
		Complexity:   complexity,
		Dependencies: deps,
		Labels:       labels,
		Created:      time.Now(),
	}
}

func TestNextTicketEmptyGraph(t *testing.T) {
	g := New()
	result, err := NextTicket(nil, g, NextOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty graph, got %v", result)
	}
}

func TestNextTicketSingleNoDeps(t *testing.T) {
	tk := newTicket("TKR-1", ticket.StatusTodo, ticket.PriorityMedium, 3, nil, nil)
	tickets := []ticket.Ticket{tk}
	g := BuildFromTickets(tickets)

	result, err := NextTicket(tickets, g, NextOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected a ticket, got nil")
	}
	if result.ID != "TKR-1" {
		t.Errorf("expected TKR-1, got %s", result.ID)
	}
}

func TestNextTicketLinearChain(t *testing.T) {
	// A depends on nothing, B depends on A, C depends on B.
	a := newTicket("TKR-1", ticket.StatusTodo, ticket.PriorityMedium, 3, nil, nil)
	b := newTicket("TKR-2", ticket.StatusTodo, ticket.PriorityMedium, 3, []string{"TKR-1"}, nil)
	c := newTicket("TKR-3", ticket.StatusTodo, ticket.PriorityMedium, 3, []string{"TKR-2"}, nil)

	tickets := []ticket.Ticket{a, b, c}
	g := BuildFromTickets(tickets)

	result, err := NextTicket(tickets, g, NextOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "TKR-1" {
		t.Errorf("expected TKR-1 (no deps), got %v", result)
	}
}

func TestNextTicketLinearChainAfterDone(t *testing.T) {
	// A is done, so next should be B.
	a := newTicket("TKR-1", ticket.StatusDone, ticket.PriorityMedium, 3, nil, nil)
	b := newTicket("TKR-2", ticket.StatusTodo, ticket.PriorityMedium, 3, []string{"TKR-1"}, nil)
	c := newTicket("TKR-3", ticket.StatusTodo, ticket.PriorityMedium, 3, []string{"TKR-2"}, nil)

	tickets := []ticket.Ticket{a, b, c}
	g := BuildFromTickets(tickets)

	result, err := NextTicket(tickets, g, NextOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "TKR-2" {
		t.Errorf("expected TKR-2 (A done), got %v", result)
	}
}

func TestNextTicketDiamond(t *testing.T) {
	// A depends on B and C. B and C depend on D.
	d := newTicket("TKR-4", ticket.StatusTodo, ticket.PriorityMedium, 3, nil, nil)
	b := newTicket("TKR-2", ticket.StatusTodo, ticket.PriorityMedium, 3, []string{"TKR-4"}, nil)
	c := newTicket("TKR-3", ticket.StatusTodo, ticket.PriorityMedium, 3, []string{"TKR-4"}, nil)
	a := newTicket("TKR-1", ticket.StatusTodo, ticket.PriorityMedium, 3, []string{"TKR-2", "TKR-3"}, nil)

	tickets := []ticket.Ticket{a, b, c, d}
	g := BuildFromTickets(tickets)

	result, err := NextTicket(tickets, g, NextOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "TKR-4" {
		t.Errorf("expected TKR-4 (leaf), got %v", result)
	}
}

func TestDetectCyclesSelfLoop(t *testing.T) {
	g := New()
	g.AddEdge("TKR-1", "TKR-1")

	result := DetectCycles(g)
	if !result.HasCycle {
		t.Error("expected cycle for self-loop")
	}
	if len(result.CycleNodes) == 0 {
		t.Error("expected cycle nodes to be non-empty")
	}
}

func TestDetectCyclesTwoNodeCycle(t *testing.T) {
	g := New()
	g.AddEdge("TKR-1", "TKR-2")
	g.AddEdge("TKR-2", "TKR-1")

	result := DetectCycles(g)
	if !result.HasCycle {
		t.Error("expected cycle for two-node cycle")
	}
	found := make(map[string]bool)
	for _, n := range result.CycleNodes {
		found[n] = true
	}
	if !found["TKR-1"] || !found["TKR-2"] {
		t.Errorf("expected both TKR-1 and TKR-2 in cycle nodes, got %v", result.CycleNodes)
	}
}

func TestDetectCyclesNoCycle(t *testing.T) {
	g := New()
	g.AddEdge("TKR-1", "TKR-2")
	g.AddNode("TKR-3")

	result := DetectCycles(g)
	if result.HasCycle {
		t.Error("expected no cycle")
	}
	if len(result.TopoOrder) != 3 {
		t.Errorf("expected 3 nodes in topo order, got %d", len(result.TopoOrder))
	}
}

func TestNextTicketPriorityOrdering(t *testing.T) {
	low := newTicket("TKR-1", ticket.StatusTodo, ticket.PriorityLow, 3, nil, nil)
	critical := newTicket("TKR-2", ticket.StatusTodo, ticket.PriorityCritical, 3, nil, nil)
	high := newTicket("TKR-3", ticket.StatusTodo, ticket.PriorityHigh, 3, nil, nil)

	tickets := []ticket.Ticket{low, critical, high}
	g := BuildFromTickets(tickets)

	result, err := NextTicket(tickets, g, NextOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "TKR-2" {
		t.Errorf("expected TKR-2 (critical priority), got %v", result)
	}
}

func TestNextTicketComplexityTiebreaker(t *testing.T) {
	// Same priority, different complexity. Lower complexity wins.
	big := newTicket("TKR-1", ticket.StatusTodo, ticket.PriorityMedium, 8, nil, nil)
	small := newTicket("TKR-2", ticket.StatusTodo, ticket.PriorityMedium, 2, nil, nil)

	tickets := []ticket.Ticket{big, small}
	g := BuildFromTickets(tickets)

	result, err := NextTicket(tickets, g, NextOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "TKR-2" {
		t.Errorf("expected TKR-2 (lower complexity), got %v", result)
	}
}

func TestNextTicketIDNaturalSort(t *testing.T) {
	// Same priority and complexity. TKR-2 < TKR-10 in natural order.
	t10 := newTicket("TKR-10", ticket.StatusTodo, ticket.PriorityMedium, 3, nil, nil)
	t2 := newTicket("TKR-2", ticket.StatusTodo, ticket.PriorityMedium, 3, nil, nil)

	tickets := []ticket.Ticket{t10, t2}
	g := BuildFromTickets(tickets)

	result, err := NextTicket(tickets, g, NextOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "TKR-2" {
		t.Errorf("expected TKR-2 (natural sort), got %v", result)
	}
}

func TestNextTicketActorFilter(t *testing.T) {
	tkAgent := ticket.Ticket{
		ID: "TKR-1", Title: "Ticket TKR-1", Status: ticket.StatusTodo,
		Priority: ticket.PriorityCritical, Complexity: 1,
		Actor: ticket.ActorAgent, Created: time.Now(),
	}
	tkHuman := ticket.Ticket{
		ID: "TKR-2", Title: "Ticket TKR-2", Status: ticket.StatusTodo,
		Priority: ticket.PriorityMedium, Complexity: 3,
		Actor: ticket.ActorHuman, Created: time.Now(),
	}
	tickets := []ticket.Ticket{tkAgent, tkHuman}
	g := BuildFromTickets(tickets)

	result, err := NextTicket(tickets, g, NextOptions{Actor: "human"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "TKR-2" {
		t.Errorf("expected TKR-2 (human actor), got %v", result)
	}
}

func TestNextTicketLabelFilter(t *testing.T) {
	// Only tickets with the required label pass.
	noLabel := newTicket("TKR-1", ticket.StatusTodo, ticket.PriorityCritical, 1, nil, nil)
	hasLabel := newTicket("TKR-2", ticket.StatusTodo, ticket.PriorityMedium, 3, nil, []string{"backend"})

	tickets := []ticket.Ticket{noLabel, hasLabel}
	g := BuildFromTickets(tickets)

	result, err := NextTicket(tickets, g, NextOptions{Labels: []string{"backend"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "TKR-2" {
		t.Errorf("expected TKR-2 (has label), got %v", result)
	}
}

func TestToMermaid(t *testing.T) {
	a := newTicket("TKR-1", ticket.StatusTodo, ticket.PriorityMedium, 3, []string{"TKR-2"}, nil)
	b := newTicket("TKR-2", ticket.StatusDone, ticket.PriorityMedium, 3, nil, nil)

	tickets := []ticket.Ticket{a, b}
	g := BuildFromTickets(tickets)

	output := ToMermaid(tickets, g)

	if !strings.HasPrefix(output, "graph TD\n") {
		t.Errorf("expected output to start with 'graph TD\\n', got prefix: %q", output[:20])
	}

	// Check that nodes appear.
	if !strings.Contains(output, `TKR-1["TKR-1: Ticket TKR-1"]:::todo`) {
		t.Errorf("missing TKR-1 node in output:\n%s", output)
	}
	if !strings.Contains(output, `TKR-2["TKR-2: Ticket TKR-2"]:::done`) {
		t.Errorf("missing TKR-2 node in output:\n%s", output)
	}

	// Check that edge appears.
	if !strings.Contains(output, "TKR-1 --> TKR-2") {
		t.Errorf("missing edge TKR-1 --> TKR-2 in output:\n%s", output)
	}

	// Check classDef lines.
	if !strings.Contains(output, "classDef done") {
		t.Errorf("missing classDef done in output:\n%s", output)
	}
}

func TestBuildFromTickets(t *testing.T) {
	tickets := []ticket.Ticket{
		newTicket("TKR-1", ticket.StatusTodo, ticket.PriorityMedium, 3, []string{"TKR-2"}, nil),
		newTicket("TKR-2", ticket.StatusTodo, ticket.PriorityMedium, 3, nil, nil),
	}

	g := BuildFromTickets(tickets)

	nodes := g.Nodes()
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}

	deps := g.Dependencies("TKR-1")
	if len(deps) != 1 || deps[0] != "TKR-2" {
		t.Errorf("expected TKR-1 to depend on TKR-2, got %v", deps)
	}

	dependents := g.Dependents("TKR-2")
	if len(dependents) != 1 || dependents[0] != "TKR-1" {
		t.Errorf("expected TKR-2 to have dependent TKR-1, got %v", dependents)
	}
}
