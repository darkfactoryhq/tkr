package output

import (
	"encoding/json"
	"time"

	"github.com/darkfactoryhq/tkr/internal/ticket"
)

type jsonFormatter struct{}

type jsonTicket struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Status             string   `json:"status"`
	Priority           string   `json:"priority"`
	Type               string   `json:"type,omitempty"`
	Actor              string   `json:"actor,omitempty"`
	Assignee           string   `json:"assignee,omitempty"`
	Dependencies       []string `json:"dependencies,omitempty"`
	ParentID           string   `json:"parent_id,omitempty"`
	Blocks             []string `json:"blocks,omitempty"`
	RelatedTo          []string `json:"related_to,omitempty"`
	Duplicates         []string `json:"duplicates,omitempty"`
	Labels             []string `json:"labels,omitempty"`
	Created            string   `json:"created"`
	Updated            string   `json:"updated"`
	AcceptanceCriteria []string `json:"acceptance_criteria,omitempty"`
	Complexity         int      `json:"complexity,omitempty"`
	Estimate           string   `json:"estimate,omitempty"`
	Branch             string   `json:"branch,omitempty"`
	PR                 string   `json:"pr,omitempty"`
	Body               string   `json:"body,omitempty"`
}

func toJSON(t ticket.Ticket) jsonTicket {
	return jsonTicket{
		ID:                 t.ID,
		Title:              t.Title,
		Status:             string(t.Status),
		Priority:           string(t.Priority),
		Type:               string(t.Type),
		Actor:              string(t.Actor),
		Assignee:           t.Assignee,
		Dependencies:       t.Dependencies,
		ParentID:           t.ParentID,
		Blocks:             t.Blocks,
		RelatedTo:          t.RelatedTo,
		Duplicates:         t.Duplicates,
		Labels:             t.Labels,
		Created:            t.Created.Format(time.RFC3339),
		Updated:            t.Updated.Format(time.RFC3339),
		AcceptanceCriteria: t.AcceptanceCriteria,
		Complexity:         t.Complexity,
		Estimate:           t.Estimate,
		Branch:             t.Branch,
		PR:                 t.PR,
		Body:               t.Body,
	}
}

func marshalIndent(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return `{"error": "failed to marshal JSON"}`
	}
	return string(data) + "\n"
}

func (f *jsonFormatter) FormatTicket(t ticket.Ticket) string {
	return marshalIndent(toJSON(t))
}

func (f *jsonFormatter) FormatTicketList(tickets []ticket.Ticket) string {
	items := make([]jsonTicket, len(tickets))
	for i, t := range tickets {
		items[i] = toJSON(t)
	}
	return marshalIndent(items)
}

func (f *jsonFormatter) FormatNext(t *ticket.Ticket) string {
	if t == nil {
		return marshalIndent(map[string]any{"result": nil})
	}
	return marshalIndent(toJSON(*t))
}

func (f *jsonFormatter) FormatError(err error) string {
	return marshalIndent(map[string]string{"error": err.Error()})
}
