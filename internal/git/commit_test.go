package git

import (
	"testing"
)

func TestParseCommitMessageCloses(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{"Closes", "Closes TKR-42", "TKR-42"},
		{"Fixes", "Fixes TKR-42", "TKR-42"},
		{"Resolves", "Resolves TKR-42", "TKR-42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refs := ParseCommitMessage(tt.msg)
			found := findRef(refs, tt.want, RefClose)
			if !found {
				t.Errorf("ParseCommitMessage(%q) did not produce RefClose for %s, got %v", tt.msg, tt.want, refs)
			}
		})
	}
}

func TestParseCommitMessageCaseInsensitive(t *testing.T) {
	refs := ParseCommitMessage("closes tkr-42")
	found := findRef(refs, "TKR-42", RefClose)
	if !found {
		t.Errorf("expected case-insensitive match to produce RefClose for TKR-42, got %v", refs)
	}
}

func TestParseCommitMessageConventional(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{"feat", "feat(TKR-42): add auth", "TKR-42"},
		{"fix", "fix(TKR-42): bug", "TKR-42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refs := ParseCommitMessage(tt.msg)
			found := findRef(refs, tt.want, RefWorkOn)
			if !found {
				t.Errorf("ParseCommitMessage(%q) did not produce RefWorkOn for %s, got %v", tt.msg, tt.want, refs)
			}
		})
	}
}

func TestParseCommitMessageWIP(t *testing.T) {
	refs := ParseCommitMessage("WIP TKR-42")
	found := findRef(refs, "TKR-42", RefWorkOn)
	if !found {
		t.Errorf("expected WIP to produce RefWorkOn for TKR-42, got %v", refs)
	}
}

func TestParseCommitMessageMention(t *testing.T) {
	refs := ParseCommitMessage("TKR-42 mentioned")
	found := findRef(refs, "TKR-42", RefMention)
	if !found {
		t.Errorf("expected plain mention to produce RefMention for TKR-42, got %v", refs)
	}
}

func TestParseCommitMessageMultipleRefs(t *testing.T) {
	refs := ParseCommitMessage("Closes TKR-42, Closes TKR-43")
	if len(refs) < 2 {
		t.Fatalf("expected at least 2 refs, got %d: %v", len(refs), refs)
	}
	found42 := findRef(refs, "TKR-42", RefClose)
	found43 := findRef(refs, "TKR-43", RefClose)
	if !found42 {
		t.Errorf("expected RefClose for TKR-42, got %v", refs)
	}
	if !found43 {
		t.Errorf("expected RefClose for TKR-43, got %v", refs)
	}
}

func TestParseCommitMessageNoRefs(t *testing.T) {
	refs := ParseCommitMessage("random commit no refs")
	if len(refs) != 0 {
		t.Errorf("expected 0 refs, got %d: %v", len(refs), refs)
	}
}

func TestParseCommitMessageNonTKRPrefix(t *testing.T) {
	refs := ParseCommitMessage("Closes PROJ-1")
	found := findRef(refs, "PROJ-1", RefClose)
	if !found {
		t.Errorf("expected RefClose for PROJ-1, got %v", refs)
	}
}

func TestParseCommitMessageDeduplication(t *testing.T) {
	// "Closes TKR-42 TKR-42" - the close pattern matches once, the ticket pattern
	// matches twice, but deduplication should yield a single entry with the highest
	// priority action (RefClose < RefMention, so RefClose wins).
	refs := ParseCommitMessage("Closes TKR-42 TKR-42")

	count := 0
	for _, r := range refs {
		if r.TicketID == "TKR-42" {
			count++
			if r.Action != RefClose {
				t.Errorf("expected RefClose (highest priority) for TKR-42, got %d", r.Action)
			}
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 entry for TKR-42, got %d: %v", count, refs)
	}
}

func TestParseCommitMessageCloseOverridesMention(t *testing.T) {
	// When a ticket is both closed and mentioned, Close (higher priority) wins.
	refs := ParseCommitMessage("Closes TKR-42, also see TKR-42 in docs")
	found := findRef(refs, "TKR-42", RefClose)
	if !found {
		t.Errorf("expected RefClose to override RefMention for TKR-42, got %v", refs)
	}
}

func findRef(refs []CommitRef, ticketID string, action RefAction) bool {
	for _, r := range refs {
		if r.TicketID == ticketID && r.Action == action {
			return true
		}
	}
	return false
}
