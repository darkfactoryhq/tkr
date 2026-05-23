package changelog

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tkr-cli/tkr/internal/ticket"
)

type Section struct {
	Version    string
	Date       string
	Categories map[string][]Entry
}

type Entry struct {
	ID    string
	Title string
}

var typeToCategory = map[string]string{
	"feature": "Added",
	"bug":     "Fixed",
	"spike":   "Changed",
	"chore":   "Changed",
	"docs":    "Other",
}

var categoryOrder = []string{
	"Added",
	"Changed",
	"Deprecated",
	"Removed",
	"Fixed",
	"Security",
	"Other",
}

func Generate(tickets []ticket.Ticket, version string, date time.Time) Section {
	s := Section{
		Version:    version,
		Date:       date.Format("2006-01-02"),
		Categories: make(map[string][]Entry),
	}

	for _, t := range tickets {
		if t.Status != ticket.StatusDone {
			continue
		}

		cat := categorize(string(t.Type))
		s.Categories[cat] = append(s.Categories[cat], Entry{
			ID:    t.ID,
			Title: t.Title,
		})
	}

	for cat := range s.Categories {
		sort.Slice(s.Categories[cat], func(i, j int) bool {
			return s.Categories[cat][i].ID < s.Categories[cat][j].ID
		})
	}

	return s
}

func categorize(ticketType string) string {
	if cat, ok := typeToCategory[ticketType]; ok {
		return cat
	}
	return "Other"
}

func (s Section) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "## [%s] - %s\n", s.Version, s.Date)

	for _, cat := range categoryOrder {
		entries, ok := s.Categories[cat]
		if !ok {
			continue
		}

		fmt.Fprintf(&b, "\n### %s\n", cat)
		for _, e := range entries {
			fmt.Fprintf(&b, "- %s (%s)\n", e.Title, e.ID)
		}
	}

	return b.String()
}
