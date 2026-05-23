package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// taskmasterFile is the top-level JSON structure for claude-task-master.
type taskmasterFile struct {
	Tasks []taskmasterTask `json:"tasks"`
}

type taskmasterTask struct {
	ID           int                `json:"id"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	Details      string             `json:"details"`
	TestStrategy string             `json:"testStrategy"`
	Status       string             `json:"status"`
	Dependencies []int              `json:"dependencies"`
	Priority     string             `json:"priority"`
	Tags         []string           `json:"tags"`
	Complexity   int                `json:"complexity"`
	Subtasks     []taskmasterSubtask `json:"subtasks"`
}

type taskmasterSubtask struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// ParseTaskmasterFile parses a claude-task-master tasks.json file.
func ParseTaskmasterFile(path string) ([]ImportedTask, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading taskmaster file: %w", err)
	}

	var tf taskmasterFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("parsing taskmaster JSON: %w", err)
	}

	var tasks []ImportedTask
	for _, t := range tf.Tasks {
		body := buildTaskmasterBody(t.Description, t.Details, t.TestStrategy)

		var deps []string
		for _, d := range t.Dependencies {
			deps = append(deps, fmt.Sprintf("%d", d))
		}

		tasks = append(tasks, ImportedTask{
			OriginalID:   fmt.Sprintf("%d", t.ID),
			Title:        t.Title,
			Status:       mapTaskmasterStatus(t.Status),
			Priority:     mapTaskmasterPriority(t.Priority),
			Labels:       t.Tags,
			Dependencies: deps,
			Complexity:   t.Complexity,
			Body:         body,
		})

		// Subtasks become separate tasks with ParentID set.
		for _, sub := range t.Subtasks {
			subBody := sub.Description
			tasks = append(tasks, ImportedTask{
				OriginalID: fmt.Sprintf("%d.%d", t.ID, sub.ID),
				Title:      sub.Title,
				Status:     mapTaskmasterStatus(sub.Status),
				Priority:   mapTaskmasterPriority(t.Priority),
				ParentID:   fmt.Sprintf("%d", t.ID),
				Body:       subBody,
			})
		}
	}

	return tasks, nil
}

func buildTaskmasterBody(description, details, testStrategy string) string {
	var parts []string

	if description != "" {
		parts = append(parts, "## Description\n\n"+description)
	}
	if details != "" {
		parts = append(parts, "## Details\n\n"+details)
	}
	if testStrategy != "" {
		parts = append(parts, "## Test Strategy\n\n"+testStrategy)
	}

	return strings.Join(parts, "\n\n")
}

func mapTaskmasterStatus(s string) string {
	switch strings.ToLower(s) {
	case "pending", "not-started":
		return "todo"
	case "in-progress":
		return "in-progress"
	case "done", "completed":
		return "done"
	case "deferred":
		return "blocked"
	case "review":
		return "in-review"
	default:
		return "todo"
	}
}

func mapTaskmasterPriority(p string) string {
	switch strings.ToLower(p) {
	case "critical", "urgent":
		return "critical"
	case "high":
		return "high"
	case "medium", "normal":
		return "medium"
	case "low":
		return "low"
	default:
		return "medium"
	}
}
