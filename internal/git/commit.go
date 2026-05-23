package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type RefAction int

const (
	RefClose   RefAction = iota
	RefWorkOn
	RefMention
)

type CommitRef struct {
	TicketID string
	Action   RefAction
}

var (
	ticketPattern    = regexp.MustCompile(`[A-Z]+-\d+`)
	closePattern     = regexp.MustCompile(`(?i)\b(?:close|closes|fix|fixes|resolve|resolves)\s+([A-Z]+-\d+)`)
	conventionalPattern = regexp.MustCompile(`(?i)\b(?:feat|fix|chore|refactor|test|docs)\(([A-Z]+-\d+)\):`)
	wipPattern       = regexp.MustCompile(`(?i)\bWIP\s+([A-Z]+-\d+)`)
)

func ParseCommitMessage(msg string) []CommitRef {
	actions := make(map[string]RefAction)

	for _, m := range closePattern.FindAllStringSubmatch(msg, -1) {
		id := strings.ToUpper(m[1])
		actions[id] = RefClose
	}

	for _, m := range conventionalPattern.FindAllStringSubmatch(msg, -1) {
		id := strings.ToUpper(m[1])
		if prev, ok := actions[id]; !ok || prev > RefWorkOn {
			actions[id] = RefWorkOn
		}
	}

	for _, m := range wipPattern.FindAllStringSubmatch(msg, -1) {
		id := strings.ToUpper(m[1])
		if prev, ok := actions[id]; !ok || prev > RefWorkOn {
			actions[id] = RefWorkOn
		}
	}

	for _, m := range ticketPattern.FindAllString(msg, -1) {
		id := strings.ToUpper(m)
		if _, ok := actions[id]; !ok {
			actions[id] = RefMention
		}
	}

	refs := make([]CommitRef, 0, len(actions))
	for id, action := range actions {
		refs = append(refs, CommitRef{TicketID: id, Action: action})
	}
	return refs
}

func LatestCommitMessage() (string, error) {
	out, err := exec.Command("git", "log", "-1", "--format=%B").Output()
	if err != nil {
		return "", fmt.Errorf("git log: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
