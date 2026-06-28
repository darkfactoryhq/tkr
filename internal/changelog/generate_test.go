package changelog

import (
	"strings"
	"testing"
	"time"

	"github.com/darkfactoryhq/tkr/internal/ticket"
)

func mkTicket(id string, status ticket.Status, ticketType ticket.TicketType) ticket.Ticket {
	return ticket.Ticket{
		ID:       id,
		Title:    "Ticket " + id,
		Status:   status,
		Priority: ticket.PriorityMedium,
		Type:     ticketType,
		Created:  time.Now(),
	}
}

func TestFeatureTypeGoesToAdded(t *testing.T) {
	tickets := []ticket.Ticket{
		mkTicket("TKR-1", ticket.StatusDone, ticket.TypeFeature),
	}

	date := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	section := Generate(tickets, "1.0.0", date)

	entries, ok := section.Categories["Added"]
	if !ok {
		t.Fatal("expected 'Added' category to exist")
	}
	if len(entries) != 1 || entries[0].ID != "TKR-1" {
		t.Errorf("expected TKR-1 in Added, got %v", entries)
	}
}

func TestBugTypeGoesToFixed(t *testing.T) {
	tickets := []ticket.Ticket{
		mkTicket("TKR-1", ticket.StatusDone, ticket.TypeBug),
	}

	section := Generate(tickets, "1.0.0", time.Now())

	entries, ok := section.Categories["Fixed"]
	if !ok {
		t.Fatal("expected 'Fixed' category to exist")
	}
	if len(entries) != 1 || entries[0].ID != "TKR-1" {
		t.Errorf("expected TKR-1 in Fixed, got %v", entries)
	}
}

func TestNoTypeGoesToOther(t *testing.T) {
	tickets := []ticket.Ticket{
		mkTicket("TKR-1", ticket.StatusDone, ""),
	}

	section := Generate(tickets, "1.0.0", time.Now())

	entries, ok := section.Categories["Other"]
	if !ok {
		t.Fatal("expected 'Other' category to exist")
	}
	if len(entries) != 1 || entries[0].ID != "TKR-1" {
		t.Errorf("expected TKR-1 in Other, got %v", entries)
	}
}

func TestOnlyDoneTicketsIncluded(t *testing.T) {
	tickets := []ticket.Ticket{
		mkTicket("TKR-1", ticket.StatusDone, ticket.TypeFeature),
		mkTicket("TKR-2", ticket.StatusTodo, ticket.TypeFeature),
		mkTicket("TKR-3", ticket.StatusInProgress, ticket.TypeBug),
		mkTicket("TKR-4", ticket.StatusBlocked, ticket.TypeBug),
	}

	section := Generate(tickets, "1.0.0", time.Now())

	totalEntries := 0
	for _, entries := range section.Categories {
		totalEntries += len(entries)
	}
	if totalEntries != 1 {
		t.Errorf("expected 1 total entry (only done), got %d", totalEntries)
	}
}

func TestStringOutputFormat(t *testing.T) {
	tickets := []ticket.Ticket{
		mkTicket("TKR-1", ticket.StatusDone, ticket.TypeFeature),
		mkTicket("TKR-2", ticket.StatusDone, ticket.TypeBug),
		mkTicket("TKR-3", ticket.StatusDone, ""),
	}

	date := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	section := Generate(tickets, "1.0.0", date)
	output := section.String()

	// Check version header.
	if !strings.Contains(output, "## [1.0.0] - 2025-06-01") {
		t.Errorf("expected version header, got:\n%s", output)
	}

	// Check category headers.
	if !strings.Contains(output, "### Added") {
		t.Errorf("expected '### Added' header, got:\n%s", output)
	}
	if !strings.Contains(output, "### Fixed") {
		t.Errorf("expected '### Fixed' header, got:\n%s", output)
	}
	if !strings.Contains(output, "### Other") {
		t.Errorf("expected '### Other' header, got:\n%s", output)
	}

	// Check entry format.
	if !strings.Contains(output, "- Ticket TKR-1 (TKR-1)") {
		t.Errorf("expected entry for TKR-1, got:\n%s", output)
	}
	if !strings.Contains(output, "- Ticket TKR-2 (TKR-2)") {
		t.Errorf("expected entry for TKR-2, got:\n%s", output)
	}
}

func TestStringCategoryOrdering(t *testing.T) {
	tickets := []ticket.Ticket{
		mkTicket("TKR-1", ticket.StatusDone, ticket.TypeBug),
		mkTicket("TKR-2", ticket.StatusDone, ticket.TypeFeature),
		mkTicket("TKR-3", ticket.StatusDone, ""),
	}

	section := Generate(tickets, "1.0.0", time.Now())
	output := section.String()

	addedIdx := strings.Index(output, "### Added")
	fixedIdx := strings.Index(output, "### Fixed")
	otherIdx := strings.Index(output, "### Other")

	if addedIdx == -1 || fixedIdx == -1 || otherIdx == -1 {
		t.Fatalf("missing expected category headers:\n%s", output)
	}

	// Added should appear before Fixed, and Fixed before Other.
	if addedIdx >= fixedIdx {
		t.Errorf("expected Added before Fixed")
	}
	if fixedIdx >= otherIdx {
		t.Errorf("expected Fixed before Other")
	}
}

func TestGenerateVersionAndDate(t *testing.T) {
	date := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	section := Generate(nil, "2.0.0-beta", date)

	if section.Version != "2.0.0-beta" {
		t.Errorf("Version = %q, want %q", section.Version, "2.0.0-beta")
	}
	if section.Date != "2025-03-15" {
		t.Errorf("Date = %q, want %q", section.Date, "2025-03-15")
	}
}

func TestGenerateEntriesSortedByID(t *testing.T) {
	tickets := []ticket.Ticket{
		mkTicket("TKR-3", ticket.StatusDone, ticket.TypeFeature),
		mkTicket("TKR-1", ticket.StatusDone, ticket.TypeFeature),
		mkTicket("TKR-2", ticket.StatusDone, ticket.TypeFeature),
	}

	section := Generate(tickets, "1.0.0", time.Now())

	entries := section.Categories["Added"]
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].ID != "TKR-1" || entries[1].ID != "TKR-2" || entries[2].ID != "TKR-3" {
		t.Errorf("expected sorted order [TKR-1, TKR-2, TKR-3], got [%s, %s, %s]",
			entries[0].ID, entries[1].ID, entries[2].ID)
	}
}

func TestSpikeTypeGoesToChanged(t *testing.T) {
	tickets := []ticket.Ticket{
		mkTicket("TKR-1", ticket.StatusDone, ticket.TypeSpike),
	}

	section := Generate(tickets, "1.0.0", time.Now())

	if _, ok := section.Categories["Changed"]; !ok {
		t.Fatal("expected 'Changed' category for 'spike' type")
	}
}

func TestChoreTypeGoesToChanged(t *testing.T) {
	tickets := []ticket.Ticket{
		mkTicket("TKR-1", ticket.StatusDone, ticket.TypeChore),
	}

	section := Generate(tickets, "1.0.0", time.Now())

	if _, ok := section.Categories["Changed"]; !ok {
		t.Fatal("expected 'Changed' category for 'chore' type")
	}
}

func TestDocsTypeGoesToOther(t *testing.T) {
	tickets := []ticket.Ticket{
		mkTicket("TKR-1", ticket.StatusDone, ticket.TypeDocs),
	}

	section := Generate(tickets, "1.0.0", time.Now())

	if _, ok := section.Categories["Other"]; !ok {
		t.Fatal("expected 'Other' category for 'docs' type")
	}
}
