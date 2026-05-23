package ticket

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusInReview   Status = "in-review"
	StatusDone       Status = "done"
	StatusBlocked    Status = "blocked"
)

func ValidStatuses() []Status {
	return []Status{StatusTodo, StatusInProgress, StatusInReview, StatusDone, StatusBlocked}
}

func (s Status) IsValid() bool {
	for _, v := range ValidStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

func ValidPriorities() []Priority {
	return []Priority{PriorityCritical, PriorityHigh, PriorityMedium, PriorityLow}
}

func (p Priority) IsValid() bool {
	for _, v := range ValidPriorities() {
		if p == v {
			return true
		}
	}
	return false
}

func (p Priority) Rank() int {
	switch p {
	case PriorityCritical:
		return 0
	case PriorityHigh:
		return 1
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 3
	default:
		return -1
	}
}

type Actor string

const (
	ActorAgent Actor = "agent"
	ActorHuman Actor = "human"
)

func ValidActors() []Actor {
	return []Actor{ActorAgent, ActorHuman}
}

func (a Actor) IsValid() bool {
	if a == "" {
		return true // optional field
	}
	for _, v := range ValidActors() {
		if a == v {
			return true
		}
	}
	return false
}

type TicketType string

const (
	TypeFeature TicketType = "feature"
	TypeBug     TicketType = "bug"
	TypeSpike   TicketType = "spike"
	TypeChore   TicketType = "chore"
	TypeDocs    TicketType = "docs"
)

func ValidTypes() []TicketType {
	return []TicketType{TypeFeature, TypeBug, TypeSpike, TypeChore, TypeDocs}
}

func (tt TicketType) IsValid() bool {
	if tt == "" {
		return true // optional field
	}
	for _, v := range ValidTypes() {
		if tt == v {
			return true
		}
	}
	return false
}

type Ticket struct {
	ID                 string     `yaml:"id"`
	Title              string     `yaml:"title"`
	Status             Status     `yaml:"status"`
	Priority           Priority   `yaml:"priority"`
	Type               TicketType `yaml:"type,omitempty"`
	Actor              Actor      `yaml:"actor,omitempty"`
	Assignee           string     `yaml:"assignee,omitempty"`
	Dependencies       []string  `yaml:"dependencies,omitempty"`
	ParentID           string    `yaml:"parent_id,omitempty"`
	Blocks             []string  `yaml:"blocks,omitempty"`
	RelatedTo          []string  `yaml:"related_to,omitempty"`
	Duplicates         []string  `yaml:"duplicates,omitempty"`
	Labels             []string  `yaml:"labels,omitempty"`
	Created            time.Time `yaml:"created"`
	Updated            time.Time `yaml:"updated"`
	AcceptanceCriteria []string  `yaml:"acceptance_criteria,omitempty"`
	Complexity         int       `yaml:"complexity,omitempty"`
	Estimate           string    `yaml:"estimate,omitempty"`
	Branch             string    `yaml:"branch,omitempty"`
	PR                 string    `yaml:"pr,omitempty"`
	Body               string    `yaml:"-"`
	FilePath           string    `yaml:"-"`
}

func ParseEstimate(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	suffix := s[len(s)-1]
	numStr := s[:len(s)-1]

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid estimate %q: %w", s, err)
	}

	switch suffix {
	case 'm':
		return time.Duration(num * float64(time.Minute)), nil
	case 'h':
		return time.Duration(num * float64(time.Hour)), nil
	case 'd':
		return time.Duration(num * 8 * float64(time.Hour)), nil
	case 'w':
		return time.Duration(num * 40 * float64(time.Hour)), nil
	default:
		return 0, fmt.Errorf("invalid estimate %q: unknown suffix %q (use m, h, d, w)", s, string(suffix))
	}
}

func (t *Ticket) EstimateDuration() time.Duration {
	d, _ := ParseEstimate(t.Estimate)
	return d
}
