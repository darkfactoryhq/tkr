package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ImportedTask is the shared intermediate representation used by all importers.
type ImportedTask struct {
	OriginalID         string
	Title              string
	Status             string
	Priority           string
	Actor              string
	Type               string
	Assignee           string
	Labels             []string
	Dependencies       []string // original IDs
	ParentID           string   // original ID
	Blocks             []string
	RelatedTo          []string
	Duplicates         []string
	AcceptanceCriteria []string
	Complexity         int
	Estimate           string
	Body               string
	Created            time.Time
}

// backlogFrontmatter mirrors the YAML frontmatter in Backlog.md task files.
type backlogFrontmatter struct {
	ID                 string   `yaml:"id"`
	Title              string   `yaml:"title"`
	Status             string   `yaml:"status"`
	Assignee           string   `yaml:"assignee"`
	Labels             []string `yaml:"labels"`
	Priority           string   `yaml:"priority"`
	Dependencies       []string `yaml:"dependencies"`
	AcceptanceCriteria []string `yaml:"acceptance_criteria"`
	ParentTaskID       string   `yaml:"parent_task_id"`
	CreatedAt          string   `yaml:"created_at"`
}

// ParseBacklogDir reads all .md files from dir and returns ImportedTasks.
func ParseBacklogDir(dir string) ([]ImportedTask, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading backlog directory: %w", err)
	}

	var tasks []ImportedTask
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || strings.HasPrefix(e.Name(), ".") {
			continue
		}

		t, err := parseBacklogFile(filepath.Join(dir, e.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", e.Name(), err)
			continue
		}
		tasks = append(tasks, *t)
	}

	return tasks, nil
}

func parseBacklogFile(path string) (*ImportedTask, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("missing frontmatter delimiter")
	}

	rest := content[4:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return nil, fmt.Errorf("missing closing frontmatter delimiter")
	}

	fmRaw := rest[:end]
	body := strings.TrimSpace(rest[end+4:])

	var fm backlogFrontmatter
	if err := yaml.Unmarshal([]byte(fmRaw), &fm); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	// Strip "## Description" heading if the body starts with it.
	body = strings.TrimPrefix(body, "## Description\n")
	body = strings.TrimPrefix(body, "## Description\r\n")
	body = strings.TrimSpace(body)

	assignee := fm.Assignee
	assignee = strings.TrimPrefix(assignee, "@")

	var created time.Time
	if fm.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, fm.CreatedAt); err == nil {
			created = t
		}
	}

	return &ImportedTask{
		OriginalID:         fm.ID,
		Title:              fm.Title,
		Status:             mapBacklogStatus(fm.Status),
		Priority:           mapBacklogPriority(fm.Priority),
		Assignee:           assignee,
		Labels:             fm.Labels,
		Dependencies:       fm.Dependencies,
		ParentID:           fm.ParentTaskID,
		AcceptanceCriteria: fm.AcceptanceCriteria,
		Body:               body,
		Created:            created,
	}, nil
}

func mapBacklogStatus(s string) string {
	switch strings.ToLower(s) {
	case "to do":
		return "todo"
	case "in progress":
		return "in-progress"
	case "blocked":
		return "blocked"
	case "done":
		return "done"
	case "in review":
		return "in-review"
	default:
		return "todo"
	}
}

func mapBacklogPriority(p string) string {
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
