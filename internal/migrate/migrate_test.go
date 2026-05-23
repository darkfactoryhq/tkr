package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseBacklogDir(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "task-1.md"), `---
id: task-1
title: First task
status: To Do
assignee: "@claude"
labels: [frontend, auth]
priority: high
dependencies: [task-2]
acceptance_criteria:
  - Some criterion
parent_task_id: task-3
created_at: 2026-01-15T10:00:00Z
---
## Description
Some body text
`)

	writeFile(t, filepath.Join(dir, "task-2.md"), `---
id: task-2
title: Second task
status: Done
priority: low
---
Another body
`)

	tasks, err := ParseBacklogDir(dir)
	if err != nil {
		t.Fatalf("ParseBacklogDir() error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	// Find task-1 and task-2 (order may vary).
	var t1, t2 *ImportedTask
	for i := range tasks {
		switch tasks[i].OriginalID {
		case "task-1":
			t1 = &tasks[i]
		case "task-2":
			t2 = &tasks[i]
		}
	}

	if t1 == nil || t2 == nil {
		t.Fatal("expected to find task-1 and task-2")
	}

	// Verify task-1 fields.
	if t1.Title != "First task" {
		t.Errorf("task-1 title = %q, want %q", t1.Title, "First task")
	}
	if t1.Status != "todo" {
		t.Errorf("task-1 status = %q, want %q", t1.Status, "todo")
	}
	if t1.Priority != "high" {
		t.Errorf("task-1 priority = %q, want %q", t1.Priority, "high")
	}
	if t1.Assignee != "claude" {
		t.Errorf("task-1 assignee = %q, want %q", t1.Assignee, "claude")
	}
	if len(t1.Labels) != 2 || t1.Labels[0] != "frontend" || t1.Labels[1] != "auth" {
		t.Errorf("task-1 labels = %v, want [frontend auth]", t1.Labels)
	}
	if len(t1.Dependencies) != 1 || t1.Dependencies[0] != "task-2" {
		t.Errorf("task-1 dependencies = %v, want [task-2]", t1.Dependencies)
	}
	if len(t1.AcceptanceCriteria) != 1 || t1.AcceptanceCriteria[0] != "Some criterion" {
		t.Errorf("task-1 acceptance_criteria = %v, want [Some criterion]", t1.AcceptanceCriteria)
	}
	if t1.ParentID != "task-3" {
		t.Errorf("task-1 parent_id = %q, want %q", t1.ParentID, "task-3")
	}
	if t1.Body != "Some body text" {
		t.Errorf("task-1 body = %q, want %q", t1.Body, "Some body text")
	}
	if t1.Created.IsZero() {
		t.Error("task-1 created should not be zero")
	}

	// Verify task-2 fields.
	if t2.Title != "Second task" {
		t.Errorf("task-2 title = %q, want %q", t2.Title, "Second task")
	}
	if t2.Status != "done" {
		t.Errorf("task-2 status = %q, want %q", t2.Status, "done")
	}
	if t2.Priority != "low" {
		t.Errorf("task-2 priority = %q, want %q", t2.Priority, "low")
	}
}

func TestParseBacklogStatusMapping(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"To Do", "todo"},
		{"In Progress", "in-progress"},
		{"Blocked", "blocked"},
		{"Done", "done"},
		{"In Review", "in-review"},
		{"unknown", "todo"},
	}

	for _, tt := range tests {
		got := mapBacklogStatus(tt.input)
		if got != tt.want {
			t.Errorf("mapBacklogStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseBacklogPriorityMapping(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"critical", "critical"},
		{"urgent", "critical"},
		{"high", "high"},
		{"medium", "medium"},
		{"normal", "medium"},
		{"low", "low"},
		{"unknown", "medium"},
	}

	for _, tt := range tests {
		got := mapBacklogPriority(tt.input)
		if got != tt.want {
			t.Errorf("mapBacklogPriority(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseTaskmasterFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	data := taskmasterFile{
		Tasks: []taskmasterTask{
			{
				ID:           1,
				Title:        "Setup project",
				Description:  "Description text",
				Details:      "Implementation details",
				TestStrategy: "Test approach",
				Status:       "pending",
				Dependencies: []int{2},
				Priority:     "high",
				Tags:         []string{"frontend"},
				Complexity:   5,
				Subtasks: []taskmasterSubtask{
					{ID: 1, Title: "Sub item", Status: "pending", Description: "Sub desc"},
				},
			},
			{
				ID:          2,
				Title:       "Second task",
				Description: "Another description",
				Status:      "done",
				Priority:    "low",
			},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	writeFile(t, path, string(raw))

	tasks, err := ParseTaskmasterFile(path)
	if err != nil {
		t.Fatalf("ParseTaskmasterFile() error: %v", err)
	}

	// 2 top-level tasks + 1 subtask = 3 total.
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	// Task 1.
	if tasks[0].OriginalID != "1" {
		t.Errorf("task 1 original_id = %q, want %q", tasks[0].OriginalID, "1")
	}
	if tasks[0].Title != "Setup project" {
		t.Errorf("task 1 title = %q, want %q", tasks[0].Title, "Setup project")
	}
	if tasks[0].Status != "todo" {
		t.Errorf("task 1 status = %q, want %q", tasks[0].Status, "todo")
	}
	if tasks[0].Priority != "high" {
		t.Errorf("task 1 priority = %q, want %q", tasks[0].Priority, "high")
	}
	if len(tasks[0].Dependencies) != 1 || tasks[0].Dependencies[0] != "2" {
		t.Errorf("task 1 dependencies = %v, want [2]", tasks[0].Dependencies)
	}
	if tasks[0].Complexity != 5 {
		t.Errorf("task 1 complexity = %d, want 5", tasks[0].Complexity)
	}
	if len(tasks[0].Labels) != 1 || tasks[0].Labels[0] != "frontend" {
		t.Errorf("task 1 labels = %v, want [frontend]", tasks[0].Labels)
	}

	// Verify body contains all three sections.
	if !contains(tasks[0].Body, "## Description") {
		t.Error("task 1 body missing Description heading")
	}
	if !contains(tasks[0].Body, "## Details") {
		t.Error("task 1 body missing Details heading")
	}
	if !contains(tasks[0].Body, "## Test Strategy") {
		t.Error("task 1 body missing Test Strategy heading")
	}

	// Task 2 (second top-level).
	if tasks[2].OriginalID != "2" {
		t.Errorf("task 2 original_id = %q, want %q", tasks[2].OriginalID, "2")
	}
	if tasks[2].Status != "done" {
		t.Errorf("task 2 status = %q, want %q", tasks[2].Status, "done")
	}
}

func TestParseTaskmasterStatusMapping(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"pending", "todo"},
		{"not-started", "todo"},
		{"in-progress", "in-progress"},
		{"done", "done"},
		{"completed", "done"},
		{"deferred", "blocked"},
		{"review", "in-review"},
		{"unknown", "todo"},
	}

	for _, tt := range tests {
		got := mapTaskmasterStatus(tt.input)
		if got != tt.want {
			t.Errorf("mapTaskmasterStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseTaskmasterSubtasks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	data := taskmasterFile{
		Tasks: []taskmasterTask{
			{
				ID:       5,
				Title:    "Parent task",
				Status:   "in-progress",
				Priority: "high",
				Subtasks: []taskmasterSubtask{
					{ID: 1, Title: "First sub", Status: "done", Description: "Sub 1 desc"},
					{ID: 2, Title: "Second sub", Status: "pending", Description: "Sub 2 desc"},
				},
			},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	writeFile(t, path, string(raw))

	tasks, err := ParseTaskmasterFile(path)
	if err != nil {
		t.Fatalf("ParseTaskmasterFile() error: %v", err)
	}

	// 1 parent + 2 subtasks = 3.
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	// Subtask 1.
	sub1 := tasks[1]
	if sub1.OriginalID != "5.1" {
		t.Errorf("subtask 1 original_id = %q, want %q", sub1.OriginalID, "5.1")
	}
	if sub1.ParentID != "5" {
		t.Errorf("subtask 1 parent_id = %q, want %q", sub1.ParentID, "5")
	}
	if sub1.Title != "First sub" {
		t.Errorf("subtask 1 title = %q, want %q", sub1.Title, "First sub")
	}
	if sub1.Status != "done" {
		t.Errorf("subtask 1 status = %q, want %q", sub1.Status, "done")
	}

	// Subtask 2.
	sub2 := tasks[2]
	if sub2.OriginalID != "5.2" {
		t.Errorf("subtask 2 original_id = %q, want %q", sub2.OriginalID, "5.2")
	}
	if sub2.ParentID != "5" {
		t.Errorf("subtask 2 parent_id = %q, want %q", sub2.ParentID, "5")
	}
	if sub2.Status != "todo" {
		t.Errorf("subtask 2 status = %q, want %q", sub2.Status, "todo")
	}

	// Subtasks inherit parent priority.
	if sub1.Priority != "high" {
		t.Errorf("subtask 1 priority = %q, want %q (inherited from parent)", sub1.Priority, "high")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing test file %s: %v", path, err)
	}
}
