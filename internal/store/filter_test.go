package store

import (
	"testing"

	"github.com/darkfactoryhq/tkr/internal/ticket"
)

func TestParseFilterStatusEq(t *testing.T) {
	f, err := ParseFilter([]string{"status:todo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Clauses) != 1 {
		t.Fatalf("expected 1 clause, got %d", len(f.Clauses))
	}
	c := f.Clauses[0]
	if c.Field != "status" {
		t.Errorf("Field = %q, want %q", c.Field, "status")
	}
	if c.Operator != OpEq {
		t.Errorf("Operator = %d, want OpEq (%d)", c.Operator, OpEq)
	}
	if len(c.Values) != 1 || c.Values[0] != "todo" {
		t.Errorf("Values = %v, want [todo]", c.Values)
	}
}

func TestParseFilterStatusIn(t *testing.T) {
	f, err := ParseFilter([]string{"status:todo,in-progress"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := f.Clauses[0]
	if c.Operator != OpIn {
		t.Errorf("Operator = %d, want OpIn (%d)", c.Operator, OpIn)
	}
	if len(c.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(c.Values))
	}
	if c.Values[0] != "todo" || c.Values[1] != "in-progress" {
		t.Errorf("Values = %v, want [todo in-progress]", c.Values)
	}
}

func TestParseFilterLabelContains(t *testing.T) {
	f, err := ParseFilter([]string{"+auth"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := f.Clauses[0]
	if c.Field != "labels" {
		t.Errorf("Field = %q, want %q", c.Field, "labels")
	}
	if c.Operator != OpContains {
		t.Errorf("Operator = %d, want OpContains (%d)", c.Operator, OpContains)
	}
	if len(c.Values) != 1 || c.Values[0] != "auth" {
		t.Errorf("Values = %v, want [auth]", c.Values)
	}
}

func TestParseFilterNegateStatus(t *testing.T) {
	f, err := ParseFilter([]string{"-blocked"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := f.Clauses[0]
	if c.Field != "status" {
		t.Errorf("Field = %q, want %q", c.Field, "status")
	}
	if !c.Negate {
		t.Error("expected Negate = true")
	}
	if c.Operator != OpEq {
		t.Errorf("Operator = %d, want OpEq (%d)", c.Operator, OpEq)
	}
	if len(c.Values) != 1 || c.Values[0] != "blocked" {
		t.Errorf("Values = %v, want [blocked]", c.Values)
	}
}

func TestParseFilterComplexityGt(t *testing.T) {
	f, err := ParseFilter([]string{"complexity:>5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := f.Clauses[0]
	if c.Field != "complexity" {
		t.Errorf("Field = %q, want %q", c.Field, "complexity")
	}
	if c.Operator != OpGt {
		t.Errorf("Operator = %d, want OpGt (%d)", c.Operator, OpGt)
	}
	if len(c.Values) != 1 || c.Values[0] != "5" {
		t.Errorf("Values = %v, want [5]", c.Values)
	}
}

func TestParseFilterComplexityRange(t *testing.T) {
	f, err := ParseFilter([]string{"complexity:1..5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := f.Clauses[0]
	if c.Field != "complexity" {
		t.Errorf("Field = %q, want %q", c.Field, "complexity")
	}
	if c.Operator != OpRange {
		t.Errorf("Operator = %d, want OpRange (%d)", c.Operator, OpRange)
	}
	if len(c.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(c.Values))
	}
	if c.Values[0] != "1" || c.Values[1] != "5" {
		t.Errorf("Values = %v, want [1 5]", c.Values)
	}
}

func TestMatchStatusMatch(t *testing.T) {
	tk := ticket.Ticket{
		ID:       "TKR-1",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
	}

	f, _ := ParseFilter([]string{"status:todo"})
	if !f.Match(tk) {
		t.Error("expected ticket with status:todo to match filter status:todo")
	}
}

func TestMatchStatusNoMatch(t *testing.T) {
	tk := ticket.Ticket{
		ID:       "TKR-1",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
	}

	f, _ := ParseFilter([]string{"status:done"})
	if f.Match(tk) {
		t.Error("expected ticket with status:todo to NOT match filter status:done")
	}
}

func TestMatchMultipleClausesAnd(t *testing.T) {
	tk := ticket.Ticket{
		ID:       "TKR-1",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityHigh,
	}

	f, _ := ParseFilter([]string{"status:todo", "priority:high"})
	if !f.Match(tk) {
		t.Error("expected ticket to match both clauses")
	}

	f2, _ := ParseFilter([]string{"status:todo", "priority:low"})
	if f2.Match(tk) {
		t.Error("expected ticket to NOT match when one clause fails")
	}
}

func TestMatchLabelContains(t *testing.T) {
	tk := ticket.Ticket{
		ID:       "TKR-1",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
		Labels:   []string{"auth", "backend"},
	}

	f, _ := ParseFilter([]string{"+auth"})
	if !f.Match(tk) {
		t.Error("expected ticket with label auth to match +auth")
	}

	f2, _ := ParseFilter([]string{"+frontend"})
	if f2.Match(tk) {
		t.Error("expected ticket without label frontend to NOT match +frontend")
	}
}

func TestMatchNegation(t *testing.T) {
	tk := ticket.Ticket{
		ID:       "TKR-1",
		Status:   ticket.StatusBlocked,
		Priority: ticket.PriorityMedium,
	}

	f, _ := ParseFilter([]string{"-blocked"})
	if f.Match(tk) {
		t.Error("expected blocked ticket to NOT match -blocked (negation)")
	}

	tk2 := ticket.Ticket{
		ID:       "TKR-2",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
	}
	if !f.Match(tk2) {
		t.Error("expected todo ticket to match -blocked (negation)")
	}
}

func TestMatchComplexityGt(t *testing.T) {
	tk := ticket.Ticket{
		ID:         "TKR-1",
		Status:     ticket.StatusTodo,
		Priority:   ticket.PriorityMedium,
		Complexity: 7,
	}

	f, _ := ParseFilter([]string{"complexity:>5"})
	if !f.Match(tk) {
		t.Error("expected complexity 7 > 5 to match")
	}

	tk2 := tk
	tk2.Complexity = 3
	if f.Match(tk2) {
		t.Error("expected complexity 3 > 5 to NOT match")
	}
}

func TestMatchComplexityRange(t *testing.T) {
	f, _ := ParseFilter([]string{"complexity:1..5"})

	tests := []struct {
		complexity int
		want       bool
	}{
		{0, false},
		{1, true},
		{3, true},
		{5, true},
		{6, false},
	}

	for _, tt := range tests {
		tk := ticket.Ticket{
			ID:         "TKR-1",
			Status:     ticket.StatusTodo,
			Priority:   ticket.PriorityMedium,
			Complexity: tt.complexity,
		}
		if got := f.Match(tk); got != tt.want {
			t.Errorf("complexity %d: Match = %v, want %v", tt.complexity, got, tt.want)
		}
	}
}

func TestMatchStatusIn(t *testing.T) {
	f, _ := ParseFilter([]string{"status:todo,in-progress"})

	tests := []struct {
		status ticket.Status
		want   bool
	}{
		{ticket.StatusTodo, true},
		{ticket.StatusInProgress, true},
		{ticket.StatusDone, false},
	}

	for _, tt := range tests {
		tk := ticket.Ticket{
			ID:       "TKR-1",
			Status:   tt.status,
			Priority: ticket.PriorityMedium,
		}
		if got := f.Match(tk); got != tt.want {
			t.Errorf("status %q: Match = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestParseFilterBareValue(t *testing.T) {
	// A bare value without field prefix defaults to status:eq.
	f, err := ParseFilter([]string{"todo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := f.Clauses[0]
	if c.Field != "status" {
		t.Errorf("Field = %q, want %q", c.Field, "status")
	}
	if c.Operator != OpEq {
		t.Errorf("Operator = %d, want OpEq (%d)", c.Operator, OpEq)
	}
}

func TestParseFilterLt(t *testing.T) {
	f, err := ParseFilter([]string{"complexity:<3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := f.Clauses[0]
	if c.Operator != OpLt {
		t.Errorf("Operator = %d, want OpLt (%d)", c.Operator, OpLt)
	}
}

func TestMatchActorFilter(t *testing.T) {
	tk := ticket.Ticket{
		ID:       "TKR-1",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
		Actor:    ticket.ActorAgent,
	}
	f, _ := ParseFilter([]string{"actor:agent"})
	if !f.Match(tk) {
		t.Error("expected actor:agent to match")
	}
	f2, _ := ParseFilter([]string{"actor:human"})
	if f2.Match(tk) {
		t.Error("expected actor:human to NOT match")
	}
}

func TestMatchTypeFilter(t *testing.T) {
	tk := ticket.Ticket{
		ID:       "TKR-1",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
		Type:     ticket.TypeBug,
	}
	f, _ := ParseFilter([]string{"type:bug"})
	if !f.Match(tk) {
		t.Error("expected type:bug to match")
	}
	f2, _ := ParseFilter([]string{"type:feature"})
	if f2.Match(tk) {
		t.Error("expected type:feature to NOT match")
	}
}

func TestEmptyFilter(t *testing.T) {
	f, err := ParseFilter(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty filter should match everything.
	tk := ticket.Ticket{
		ID:       "TKR-1",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
	}
	if !f.Match(tk) {
		t.Error("empty filter should match any ticket")
	}
}
