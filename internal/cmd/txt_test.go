package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/darkfactoryhq/tkr/internal/ticket"
)

func TestFormatTxtLine_Done(t *testing.T) {
	tk := &ticket.Ticket{
		ID:       "TKR-3",
		Title:    "Setup CI",
		Status:   ticket.StatusDone,
		Priority: ticket.PriorityMedium,
		Labels:   []string{"devops"},
		Created:  time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
	}

	line := formatTxtLine(tk)
	if !strings.HasPrefix(line, "x ") {
		t.Errorf("done ticket should start with 'x', got %q", line)
	}
	if strings.Contains(line, "(") {
		t.Errorf("done ticket should not have priority letter, got %q", line)
	}
	if !strings.Contains(line, "+devops") {
		t.Errorf("expected +devops label, got %q", line)
	}
	expected := "x 2026-05-24 TKR-3 Setup CI +devops"
	if line != expected {
		t.Errorf("got %q, want %q", line, expected)
	}
}

func TestFormatTxtLine_Critical(t *testing.T) {
	tk := &ticket.Ticket{
		ID:       "TKR-1",
		Title:    "Build auth module",
		Status:   ticket.StatusInProgress,
		Priority: ticket.PriorityCritical,
		Labels:   []string{"auth"},
		Actor:    ticket.ActorAgent,
		Assignee: "bob",
		Estimate: "2h",
		Complexity: 5,
		Dependencies: []string{"TKR-3"},
		Created:  time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
	}

	line := formatTxtLine(tk)
	if !strings.HasPrefix(line, "(A) ") {
		t.Errorf("critical ticket should start with '(A)', got %q", line)
	}
	expected := "(A) 2026-05-24 TKR-1 Build auth module +auth actor:agent assignee:bob est:2h cplx:5 dep:TKR-3 status:in-progress"
	if line != expected {
		t.Errorf("got %q, want %q", line, expected)
	}
}

func TestFormatTxtLine_HighPriority(t *testing.T) {
	tk := &ticket.Ticket{
		ID:       "TKR-2",
		Title:    "Fix login bug",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityHigh,
		Labels:   []string{"security"},
		Dependencies: []string{"TKR-1"},
		Created:  time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
	}

	line := formatTxtLine(tk)
	if !strings.HasPrefix(line, "(B) ") {
		t.Errorf("high priority ticket should start with '(B)', got %q", line)
	}
	expected := "(B) 2026-05-24 TKR-2 Fix login bug +security dep:TKR-1"
	if line != expected {
		t.Errorf("got %q, want %q", line, expected)
	}
}

func TestFormatTxtLine_MediumPriority(t *testing.T) {
	tk := &ticket.Ticket{
		ID:       "TKR-4",
		Title:    "Add docs",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
		Created:  time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
	}

	line := formatTxtLine(tk)
	if !strings.HasPrefix(line, "(C) ") {
		t.Errorf("medium priority ticket should start with '(C)', got %q", line)
	}
}

func TestFormatTxtLine_LowPriority(t *testing.T) {
	tk := &ticket.Ticket{
		ID:       "TKR-5",
		Title:    "Cleanup",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityLow,
		Created:  time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
	}

	line := formatTxtLine(tk)
	if !strings.HasPrefix(line, "(D) ") {
		t.Errorf("low priority ticket should start with '(D)', got %q", line)
	}
}

func TestFormatTxtLine_AllMetadata(t *testing.T) {
	tk := &ticket.Ticket{
		ID:           "TKR-10",
		Title:        "Full ticket",
		Status:       ticket.StatusBlocked,
		Priority:     ticket.PriorityCritical,
		Labels:       []string{"backend", "api"},
		Actor:        ticket.ActorAgent,
		Type:         ticket.TypeBug,
		Assignee:     "alice",
		Estimate:     "4h",
		Complexity:   8,
		Dependencies: []string{"TKR-5", "TKR-6"},
		Blocks:       []string{"TKR-11"},
		ParentID:     "TKR-1",
		Created:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	line := formatTxtLine(tk)

	checks := []string{
		"(A)",
		"2026-01-15",
		"TKR-10",
		"Full ticket",
		"+backend",
		"+api",
		"actor:agent",
		"type:bug",
		"assignee:alice",
		"est:4h",
		"cplx:8",
		"dep:TKR-5",
		"dep:TKR-6",
		"blocks:TKR-11",
		"parent:TKR-1",
		"status:blocked",
	}
	for _, check := range checks {
		if !strings.Contains(line, check) {
			t.Errorf("expected %q in line, got %q", check, line)
		}
	}
}

func TestFormatTxtLine_NoStatusTagForTodo(t *testing.T) {
	tk := &ticket.Ticket{
		ID:       "TKR-1",
		Title:    "Test",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
		Created:  time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
	}

	line := formatTxtLine(tk)
	if strings.Contains(line, "status:") {
		t.Errorf("todo tickets should not have status: tag, got %q", line)
	}
}

func TestSortTicketsForTxt(t *testing.T) {
	now := time.Now()
	tickets := []ticket.Ticket{
		{ID: "TKR-3", Status: ticket.StatusDone, Priority: ticket.PriorityCritical, Created: now},
		{ID: "TKR-1", Status: ticket.StatusTodo, Priority: ticket.PriorityLow, Created: now},
		{ID: "TKR-2", Status: ticket.StatusTodo, Priority: ticket.PriorityCritical, Created: now},
		{ID: "TKR-4", Status: ticket.StatusTodo, Priority: ticket.PriorityCritical, Created: now},
	}

	sortTicketsForTxt(tickets)

	// Non-done first, then by priority (critical before low), then by ID
	expected := []string{"TKR-2", "TKR-4", "TKR-1", "TKR-3"}
	for i, id := range expected {
		if tickets[i].ID != id {
			t.Errorf("position %d: got %s, want %s", i, tickets[i].ID, id)
		}
	}
}
