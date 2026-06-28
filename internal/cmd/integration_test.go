package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/darkfactoryhq/tkr/internal/bootstrap"
	"github.com/darkfactoryhq/tkr/internal/store"
	"github.com/darkfactoryhq/tkr/internal/ticket"
)

func setupTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := bootstrap.Init(dir, bootstrap.InitOptions{Prefix: "TEST", NoSkill: true}); err != nil {
		t.Fatalf("init: %v", err)
	}
	return dir
}

func createTicket(t *testing.T, dir, title string) *ticket.Ticket {
	t.Helper()
	s, err := store.New(dir)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	id, err := s.NextID()
	if err != nil {
		t.Fatalf("next id: %v", err)
	}
	tk := &ticket.Ticket{
		ID:       id,
		Title:    title,
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
	}
	if err := s.Save(tk); err != nil {
		t.Fatalf("save: %v", err)
	}
	return tk
}

func TestIntegration_NewAndShow(t *testing.T) {
	dir := setupTestProject(t)

	s, err := store.New(dir)
	if err != nil {
		t.Fatalf("store: %v", err)
	}

	id, err := s.NextID()
	if err != nil {
		t.Fatalf("next id: %v", err)
	}

	if id != "TEST-1" {
		t.Errorf("first ID = %q, want TEST-1", id)
	}

	tk := &ticket.Ticket{
		ID:       id,
		Title:    "Implement login",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityHigh,
		Type:     ticket.TypeFeature,
		Labels:   []string{"auth", "mvp"},
	}
	if err := s.Save(tk); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := s.Get("TEST-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Implement login" {
		t.Errorf("title = %q, want %q", got.Title, "Implement login")
	}
	if got.Priority != ticket.PriorityHigh {
		t.Errorf("priority = %q, want high", got.Priority)
	}
	if got.Type != ticket.TypeFeature {
		t.Errorf("type = %q, want feature", got.Type)
	}
}

func TestIntegration_ClaimAndDone(t *testing.T) {
	dir := setupTestProject(t)
	tk := createTicket(t, dir, "Fix bug")

	s, err := store.New(dir)
	if err != nil {
		t.Fatalf("store: %v", err)
	}

	tk.Status = ticket.StatusInProgress
	tk.Actor = ticket.ActorAgent
	if err := s.Save(tk); err != nil {
		t.Fatalf("save in-progress: %v", err)
	}

	got, _ := s.Get(tk.ID)
	if got.Status != ticket.StatusInProgress {
		t.Errorf("status = %q, want in-progress", got.Status)
	}
	if got.Actor != ticket.ActorAgent {
		t.Errorf("actor = %q, want agent", got.Actor)
	}

	got.Status = ticket.StatusDone
	if err := s.Save(got); err != nil {
		t.Fatalf("save done: %v", err)
	}

	final, _ := s.Get(tk.ID)
	if final.Status != ticket.StatusDone {
		t.Errorf("status = %q, want done", final.Status)
	}
}

func TestIntegration_ListAndFilter(t *testing.T) {
	dir := setupTestProject(t)
	createTicket(t, dir, "Task A")
	createTicket(t, dir, "Task B")

	s, _ := store.New(dir)

	tk, _ := s.Get("TEST-2")
	tk.Status = ticket.StatusInProgress
	if err := s.Save(tk); err != nil {
		t.Fatalf("save: %v", err)
	}

	tickets, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(tickets) != 2 {
		t.Fatalf("got %d tickets, want 2", len(tickets))
	}

	filter, _ := store.ParseFilter([]string{"status:todo"})
	var filtered []ticket.Ticket
	for _, tk := range tickets {
		if filter.Match(tk) {
			filtered = append(filtered, tk)
		}
	}
	if len(filtered) != 1 {
		t.Errorf("got %d todo tickets, want 1", len(filtered))
	}
}

func TestIntegration_Dependencies(t *testing.T) {
	dir := setupTestProject(t)

	s, _ := store.New(dir)
	id1, _ := s.NextID()
	id2, _ := s.NextID()

	t1 := &ticket.Ticket{
		ID: id1, Title: "Setup DB", Status: ticket.StatusTodo, Priority: ticket.PriorityHigh,
	}
	t2 := &ticket.Ticket{
		ID: id2, Title: "Add API", Status: ticket.StatusTodo, Priority: ticket.PriorityHigh,
		Dependencies: []string{id1},
	}
	if err := s.Save(t1); err != nil {
		t.Fatalf("save t1: %v", err)
	}
	if err := s.Save(t2); err != nil {
		t.Fatalf("save t2: %v", err)
	}

	got, _ := s.Get(id2)
	if len(got.Dependencies) != 1 || got.Dependencies[0] != id1 {
		t.Errorf("deps = %v, want [%s]", got.Dependencies, id1)
	}
}

func TestIntegration_ParentChild(t *testing.T) {
	dir := setupTestProject(t)
	s, _ := store.New(dir)

	parentID, _ := s.NextID()
	childID, _ := s.NextID()

	parent := &ticket.Ticket{
		ID: parentID, Title: "Epic", Status: ticket.StatusTodo, Priority: ticket.PriorityHigh,
	}
	child := &ticket.Ticket{
		ID: childID, Title: "Subtask", Status: ticket.StatusTodo, Priority: ticket.PriorityMedium,
		ParentID: parentID,
	}
	if err := s.Save(parent); err != nil {
		t.Fatalf("save parent: %v", err)
	}
	if err := s.Save(child); err != nil {
		t.Fatalf("save child: %v", err)
	}

	tickets, _ := s.Load()
	var children []ticket.Ticket
	for _, tk := range tickets {
		if tk.ParentID == parentID {
			children = append(children, tk)
		}
	}
	if len(children) != 1 {
		t.Errorf("got %d children, want 1", len(children))
	}
	if children[0].ID != childID {
		t.Errorf("child ID = %q, want %q", children[0].ID, childID)
	}
}

func TestIntegration_Delete(t *testing.T) {
	dir := setupTestProject(t)
	tk := createTicket(t, dir, "Temp ticket")

	path := filepath.Join(dir, ".tkr", "tickets", tk.ID+".md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("ticket file should exist: %v", err)
	}

	if err := os.Remove(path); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("ticket file should not exist after delete")
	}

	s, _ := store.New(dir)
	tickets, _ := s.Load()
	if len(tickets) != 0 {
		t.Errorf("got %d tickets, want 0 after delete", len(tickets))
	}
}

func TestIntegration_SequentialIDs(t *testing.T) {
	dir := setupTestProject(t)
	s, _ := store.New(dir)

	var ids []string
	for i := 0; i < 5; i++ {
		id, err := s.NextID()
		if err != nil {
			t.Fatalf("next id %d: %v", i, err)
		}
		ids = append(ids, id)
	}

	expected := []string{"TEST-1", "TEST-2", "TEST-3", "TEST-4", "TEST-5"}
	for i, want := range expected {
		if ids[i] != want {
			t.Errorf("id[%d] = %q, want %q", i, ids[i], want)
		}
	}
}

func TestIntegration_Validation(t *testing.T) {
	tk := &ticket.Ticket{
		ID:       "TEST-1",
		Title:    "",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
	}
	errs := ticket.Validate(tk)
	if len(errs) == 0 {
		t.Error("expected validation error for empty title")
	}

	found := false
	for _, e := range errs {
		if e.Field == "title" {
			found = true
		}
	}
	if !found {
		t.Error("expected title validation error")
	}
}

func TestIntegration_EstimateParsing(t *testing.T) {
	tests := []struct {
		input string
		hours float64
	}{
		{"30m", 0.5},
		{"2h", 2},
		{"1d", 8},
		{"1w", 40},
	}
	for _, tt := range tests {
		d, err := ticket.ParseEstimate(tt.input)
		if err != nil {
			t.Errorf("ParseEstimate(%q): %v", tt.input, err)
			continue
		}
		got := d.Hours()
		if got != tt.hours {
			t.Errorf("ParseEstimate(%q) = %.1f hours, want %.1f", tt.input, got, tt.hours)
		}
	}
}

func TestIntegration_BranchPattern(t *testing.T) {
	tk := &ticket.Ticket{
		ID:    "TKR-42",
		Title: "Add user authentication flow",
	}
	branch := expandBranchPattern("feat/{id}-{slug}", tk)
	if branch != "feat/tkr-42-add-user-authentication-flow" {
		t.Errorf("branch = %q", branch)
	}
}

func TestIntegration_BranchPatternTruncation(t *testing.T) {
	tk := &ticket.Ticket{
		ID:    "TKR-1",
		Title: "This is a really long title that should be truncated to avoid overly long branch names in git",
	}
	branch := expandBranchPattern("feat/{id}-{slug}", tk)
	slug := strings.TrimPrefix(branch, "feat/tkr-1-")
	if len(slug) > 50 {
		t.Errorf("slug length = %d, want <= 50", len(slug))
	}
}

func TestIntegration_BranchPatternSpecialChars(t *testing.T) {
	tk := &ticket.Ticket{
		ID:    "TKR-5",
		Title: "Fix: login (OAuth2) & signup!",
	}
	branch := expandBranchPattern("feat/{id}-{slug}", tk)
	if strings.ContainsAny(branch, ":()&! ") {
		t.Errorf("branch contains special chars: %q", branch)
	}
}

func TestIntegration_TicketRoundTrip(t *testing.T) {
	dir := setupTestProject(t)
	s, _ := store.New(dir)
	id, _ := s.NextID()

	original := &ticket.Ticket{
		ID:                 id,
		Title:              "Full roundtrip",
		Status:             ticket.StatusTodo,
		Priority:           ticket.PriorityCritical,
		Type:               ticket.TypeBug,
		Actor:              ticket.ActorAgent,
		Assignee:           "alice",
		Labels:             []string{"urgent", "backend"},
		Complexity:         7,
		Estimate:           "4h",
		AcceptanceCriteria: []string{"Tests pass", "No regressions"},
		Body:               "## Details\n\nThis needs careful handling.",
	}
	if err := s.Save(original); err != nil {
		t.Fatalf("save original: %v", err)
	}

	loaded, err := s.Get(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if loaded.Title != original.Title {
		t.Errorf("title mismatch")
	}
	if loaded.Priority != original.Priority {
		t.Errorf("priority = %q, want %q", loaded.Priority, original.Priority)
	}
	if loaded.Type != original.Type {
		t.Errorf("type = %q, want %q", loaded.Type, original.Type)
	}
	if loaded.Actor != original.Actor {
		t.Errorf("actor = %q, want %q", loaded.Actor, original.Actor)
	}
	if loaded.Assignee != original.Assignee {
		t.Errorf("assignee = %q, want %q", loaded.Assignee, original.Assignee)
	}
	if loaded.Complexity != original.Complexity {
		t.Errorf("complexity = %d, want %d", loaded.Complexity, original.Complexity)
	}
	if loaded.Estimate != original.Estimate {
		t.Errorf("estimate = %q, want %q", loaded.Estimate, original.Estimate)
	}
	if len(loaded.Labels) != 2 {
		t.Errorf("labels count = %d, want 2", len(loaded.Labels))
	}
	if len(loaded.AcceptanceCriteria) != 2 {
		t.Errorf("AC count = %d, want 2", len(loaded.AcceptanceCriteria))
	}
	if !strings.Contains(loaded.Body, "## Details") {
		t.Errorf("body missing markdown content")
	}
}

func TestIntegration_OutputFormats(t *testing.T) {
	tk := ticket.Ticket{
		ID:       "TEST-1",
		Title:    "Test ticket",
		Status:   ticket.StatusTodo,
		Priority: ticket.PriorityMedium,
	}

	// Plain format should not contain ANSI codes
	flagPlain = true
	flagJSON = false
	f := formatter()
	plain := f.FormatTicket(tk)
	if strings.Contains(plain, "\033[") {
		t.Error("plain output contains ANSI codes")
	}

	// JSON format should be valid
	flagPlain = false
	flagJSON = true
	f = formatter()
	jsonOut := f.FormatTicket(tk)
	if !strings.Contains(jsonOut, `"id"`) {
		t.Error("JSON output missing id field")
	}
	if !strings.Contains(jsonOut, `"TEST-1"`) {
		t.Error("JSON output missing ticket ID value")
	}

	// Reset
	flagPlain = false
	flagJSON = false
}

func TestIntegration_EmptyProject(t *testing.T) {
	dir := setupTestProject(t)
	s, _ := store.New(dir)

	tickets, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(tickets) != 0 {
		t.Errorf("got %d tickets, want 0", len(tickets))
	}
}

func TestIntegration_ConfigDefaults(t *testing.T) {
	dir := setupTestProject(t)
	s, _ := store.New(dir)
	cfg := s.Config()

	if cfg.Prefix != "TEST" {
		t.Errorf("prefix = %q, want TEST", cfg.Prefix)
	}
	if cfg.DefaultPriority != "medium" {
		t.Errorf("default_priority = %q, want medium", cfg.DefaultPriority)
	}
	if cfg.DefaultBranch != "main" {
		t.Errorf("default_branch = %q, want main", cfg.DefaultBranch)
	}
	if !cfg.AutoBranch {
		t.Error("auto_branch should be true by default")
	}
}

// Suppress bootstrap stdout in tests
func init() {
	var buf bytes.Buffer
	_ = buf
}
