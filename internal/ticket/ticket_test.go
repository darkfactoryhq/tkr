package ticket

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStatusIsValid(t *testing.T) {
	tests := []struct {
		status Status
		want   bool
	}{
		{StatusTodo, true},
		{StatusInProgress, true},
		{StatusInReview, true},
		{StatusDone, true},
		{StatusBlocked, true},
		{Status("invalid"), false},
		{Status(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("Status(%q).IsValid() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestPriorityIsValid(t *testing.T) {
	tests := []struct {
		priority Priority
		want     bool
	}{
		{PriorityCritical, true},
		{PriorityHigh, true},
		{PriorityMedium, true},
		{PriorityLow, true},
		{Priority("invalid"), false},
		{Priority(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			if got := tt.priority.IsValid(); got != tt.want {
				t.Errorf("Priority(%q).IsValid() = %v, want %v", tt.priority, got, tt.want)
			}
		})
	}
}

func TestPriorityRank(t *testing.T) {
	tests := []struct {
		priority Priority
		want     int
	}{
		{PriorityCritical, 0},
		{PriorityHigh, 1},
		{PriorityMedium, 2},
		{PriorityLow, 3},
		{Priority("invalid"), -1},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			if got := tt.priority.Rank(); got != tt.want {
				t.Errorf("Priority(%q).Rank() = %d, want %d", tt.priority, got, tt.want)
			}
		})
	}
}

func TestParseFileMarshalRoundtrip(t *testing.T) {
	dir := t.TempDir()

	created := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2025, 1, 16, 12, 0, 0, 0, time.UTC)

	original := &Ticket{
		ID:           "TKR-42",
		Title:        "Implement auth",
		Status:       StatusTodo,
		Priority:     PriorityHigh,
		Actor:        ActorAgent,
		Type:         TypeFeature,
		Dependencies: []string{"TKR-40", "TKR-41"},
		Labels:       []string{"auth", "security"},
		Created:      created,
		Updated:      updated,
		AcceptanceCriteria: []string{
			"Users can log in",
			"Sessions are managed",
		},
		Complexity: 5,
		Branch:     "feat/TKR-42-auth",
		PR:         "https://github.com/example/pr/42",
		Body:       "This ticket covers the auth implementation.\n",
	}

	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	path := filepath.Join(dir, "TKR-42.md")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	parsed, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if parsed.ID != original.ID {
		t.Errorf("ID = %q, want %q", parsed.ID, original.ID)
	}
	if parsed.Title != original.Title {
		t.Errorf("Title = %q, want %q", parsed.Title, original.Title)
	}
	if parsed.Status != original.Status {
		t.Errorf("Status = %q, want %q", parsed.Status, original.Status)
	}
	if parsed.Priority != original.Priority {
		t.Errorf("Priority = %q, want %q", parsed.Priority, original.Priority)
	}
	if parsed.Actor != original.Actor {
		t.Errorf("Actor = %q, want %q", parsed.Actor, original.Actor)
	}
	if parsed.Type != original.Type {
		t.Errorf("Type = %q, want %q", parsed.Type, original.Type)
	}
	if len(parsed.Dependencies) != len(original.Dependencies) {
		t.Errorf("Dependencies len = %d, want %d", len(parsed.Dependencies), len(original.Dependencies))
	}
	for i, dep := range parsed.Dependencies {
		if dep != original.Dependencies[i] {
			t.Errorf("Dependencies[%d] = %q, want %q", i, dep, original.Dependencies[i])
		}
	}
	if len(parsed.Labels) != len(original.Labels) {
		t.Errorf("Labels len = %d, want %d", len(parsed.Labels), len(original.Labels))
	}
	for i, l := range parsed.Labels {
		if l != original.Labels[i] {
			t.Errorf("Labels[%d] = %q, want %q", i, l, original.Labels[i])
		}
	}
	if !parsed.Created.Equal(original.Created) {
		t.Errorf("Created = %v, want %v", parsed.Created, original.Created)
	}
	if parsed.Complexity != original.Complexity {
		t.Errorf("Complexity = %d, want %d", parsed.Complexity, original.Complexity)
	}
	if parsed.Branch != original.Branch {
		t.Errorf("Branch = %q, want %q", parsed.Branch, original.Branch)
	}
	if parsed.PR != original.PR {
		t.Errorf("PR = %q, want %q", parsed.PR, original.PR)
	}
	if parsed.Body != original.Body {
		t.Errorf("Body = %q, want %q", parsed.Body, original.Body)
	}
	if parsed.FilePath != path {
		t.Errorf("FilePath = %q, want %q", parsed.FilePath, path)
	}

	// Marshal again and verify output is stable.
	data2, err := Marshal(parsed)
	if err != nil {
		t.Fatalf("second Marshal() error = %v", err)
	}
	if string(data) != string(data2) {
		t.Errorf("second marshal differs:\n--- first ---\n%s\n--- second ---\n%s", string(data), string(data2))
	}
}

func TestParseFileMissingFrontmatterDelimiters(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing opening delimiter", func(t *testing.T) {
		path := filepath.Join(dir, "no-open.md")
		_ = os.WriteFile(path, []byte("no frontmatter here"), 0o644)
		_, err := ParseFile(path)
		if err == nil {
			t.Fatal("expected error for missing opening delimiter")
		}
	})

	t.Run("missing closing delimiter", func(t *testing.T) {
		path := filepath.Join(dir, "no-close.md")
		_ = os.WriteFile(path, []byte("---\nid: TKR-1\ntitle: test\n"), 0o644)
		_, err := ParseFile(path)
		if err == nil {
			t.Fatal("expected error for missing closing delimiter")
		}
	})
}

func TestValidate(t *testing.T) {
	now := time.Now()

	validTicket := Ticket{
		ID:       "TKR-1",
		Title:    "Valid ticket",
		Status:   StatusTodo,
		Priority: PriorityMedium,
		Created:  now,
	}

	t.Run("valid ticket", func(t *testing.T) {
		errs := Validate(&validTicket)
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
	})

	t.Run("missing ID", func(t *testing.T) {
		tk := validTicket
		tk.ID = ""
		errs := Validate(&tk)
		if !hasValidationField(errs, "id") {
			t.Errorf("expected validation error for 'id', got %v", errs)
		}
	})

	t.Run("missing Title", func(t *testing.T) {
		tk := validTicket
		tk.Title = ""
		errs := Validate(&tk)
		if !hasValidationField(errs, "title") {
			t.Errorf("expected validation error for 'title', got %v", errs)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		tk := validTicket
		tk.Status = Status("nope")
		errs := Validate(&tk)
		if !hasValidationField(errs, "status") {
			t.Errorf("expected validation error for 'status', got %v", errs)
		}
	})

	t.Run("invalid priority", func(t *testing.T) {
		tk := validTicket
		tk.Priority = Priority("nope")
		errs := Validate(&tk)
		if !hasValidationField(errs, "priority") {
			t.Errorf("expected validation error for 'priority', got %v", errs)
		}
	})

	t.Run("complexity over 10", func(t *testing.T) {
		tk := validTicket
		tk.Complexity = 11
		errs := Validate(&tk)
		if !hasValidationField(errs, "complexity") {
			t.Errorf("expected validation error for 'complexity', got %v", errs)
		}
	})

	t.Run("zero created time", func(t *testing.T) {
		tk := validTicket
		tk.Created = time.Time{}
		errs := Validate(&tk)
		if !hasValidationField(errs, "created") {
			t.Errorf("expected validation error for 'created', got %v", errs)
		}
	})

	t.Run("invalid actor", func(t *testing.T) {
		tk := validTicket
		tk.Actor = Actor("robot")
		errs := Validate(&tk)
		if !hasValidationField(errs, "actor") {
			t.Errorf("expected validation error for 'actor', got %v", errs)
		}
	})

	t.Run("valid actor agent", func(t *testing.T) {
		tk := validTicket
		tk.Actor = ActorAgent
		errs := Validate(&tk)
		if hasValidationField(errs, "actor") {
			t.Errorf("unexpected validation error for 'actor', got %v", errs)
		}
	})

	t.Run("valid actor human", func(t *testing.T) {
		tk := validTicket
		tk.Actor = ActorHuman
		errs := Validate(&tk)
		if hasValidationField(errs, "actor") {
			t.Errorf("unexpected validation error for 'actor', got %v", errs)
		}
	})

	t.Run("empty actor valid", func(t *testing.T) {
		tk := validTicket
		tk.Actor = ""
		errs := Validate(&tk)
		if hasValidationField(errs, "actor") {
			t.Errorf("unexpected validation error for 'actor', got %v", errs)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		tk := validTicket
		tk.Type = TicketType("epic")
		errs := Validate(&tk)
		if !hasValidationField(errs, "type") {
			t.Errorf("expected validation error for 'type', got %v", errs)
		}
	})

	t.Run("valid type feature", func(t *testing.T) {
		tk := validTicket
		tk.Type = TypeFeature
		errs := Validate(&tk)
		if hasValidationField(errs, "type") {
			t.Errorf("unexpected validation error for 'type', got %v", errs)
		}
	})

	t.Run("empty type valid", func(t *testing.T) {
		tk := validTicket
		tk.Type = ""
		errs := Validate(&tk)
		if hasValidationField(errs, "type") {
			t.Errorf("unexpected validation error for 'type', got %v", errs)
		}
	})

	t.Run("validation error string", func(t *testing.T) {
		ve := ValidationError{Field: "id", Message: "must not be empty"}
		want := "id: must not be empty"
		if ve.Error() != want {
			t.Errorf("Error() = %q, want %q", ve.Error(), want)
		}
	})
}

func TestMarshalEmptyBody(t *testing.T) {
	tk := &Ticket{
		ID:       "TKR-1",
		Title:    "No body",
		Status:   StatusTodo,
		Priority: PriorityMedium,
		Created:  time.Now(),
	}

	data, err := Marshal(tk)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	content := string(data)
	// Should have opening and closing delimiters, but no extra blank body section.
	if !strings.HasPrefix(content, "---\n") {
		t.Error("expected opening ---")
	}
	// The closing delimiter should be the last --- line.
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if lines[len(lines)-1] != "---" {
		t.Errorf("expected closing --- as last line, got %q", lines[len(lines)-1])
	}
}

func hasValidationField(errs []ValidationError, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}
