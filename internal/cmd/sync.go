package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tkr-cli/tkr/internal/ticket"
)

var syncCmd = &cobra.Command{
	Use:   "sync [id]",
	Short: "Sync ticket status from GitHub PR state",
	Long:  "Check GitHub PR status for tickets with a pr: field and update ticket status accordingly.\nRequires the gh CLI (https://cli.github.com/).",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSync,
}

func init() {
	syncCmd.Flags().String("link", "", "Link a PR to a ticket before syncing (format: #123, 123, owner/repo#123, or full URL)")
}

func runSync(cmd *cobra.Command, args []string) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found. Install from https://cli.github.com/")
	}

	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	linkRef, _ := cmd.Flags().GetString("link")

	// --link mode: link a PR to a ticket, then sync it
	if linkRef != "" {
		if len(args) == 0 {
			return fmt.Errorf("--link requires a ticket ID argument")
		}
		id := args[0]
		t, err := s.Get(id)
		if err != nil {
			return err
		}
		t.PR = linkRef
		if err := s.Save(t); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s: linked PR %s\n", t.ID, linkRef)
		return syncTicket(cmd, s, t)
	}

	// Single ticket sync
	if len(args) == 1 {
		t, err := s.Get(args[0])
		if err != nil {
			return err
		}
		if t.PR == "" {
			return fmt.Errorf("ticket %s has no PR linked", t.ID)
		}
		return syncTicket(cmd, s, t)
	}

	// Sync all tickets with PR field
	tickets, err := s.Load()
	if err != nil {
		return err
	}

	synced := 0
	for i := range tickets {
		if tickets[i].PR == "" {
			continue
		}
		t, err := s.Get(tickets[i].ID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "WARNING: could not load %s: %v\n", tickets[i].ID, err)
			continue
		}
		if err := syncTicket(cmd, s, t); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "WARNING: %s: %v\n", t.ID, err)
			continue
		}
		synced++
	}

	if synced == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No tickets with PR links found")
	}
	return nil
}

func syncTicket(cmd *cobra.Command, s interface{ Save(*ticket.Ticket) error }, t *ticket.Ticket) error {
	ghArg, repoFlag := parsePRRef(t.PR)

	state, err := queryPRState(ghArg, repoFlag)
	if err != nil {
		return fmt.Errorf("querying PR %s: %w", t.PR, err)
	}

	oldStatus := t.Status
	var newStatus ticket.Status

	switch state {
	case "MERGED":
		newStatus = ticket.StatusDone
	case "OPEN":
		newStatus = ticket.StatusInReview
	case "CLOSED":
		fmt.Fprintf(cmd.ErrOrStderr(), "WARNING: %s: PR %s is closed (not merged), skipping\n", t.ID, t.PR)
		return nil
	default:
		return fmt.Errorf("unexpected PR state %q", state)
	}

	if oldStatus == newStatus {
		fmt.Fprintf(cmd.OutOrStdout(), "%s: already %s\n", t.ID, newStatus)
		return nil
	}

	t.Status = newStatus
	if err := s.Save(t); err != nil {
		return fmt.Errorf("saving %s: %w", t.ID, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s: %s → %s (PR %s %s)\n",
		t.ID, oldStatus, newStatus, t.PR, strings.ToLower(state))
	return nil
}

// parsePRRef parses a PR reference into arguments for `gh pr view`.
// Supported formats:
//   - Full URL: https://github.com/owner/repo/pull/123
//   - Repo-qualified: owner/repo#123
//   - Short number: #123 or 123
func parsePRRef(ref string) (ghArg string, repoFlag string) {
	ref = strings.TrimSpace(ref)

	// Full URL — pass directly to gh
	if strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "http://") {
		return ref, ""
	}

	// owner/repo#123
	if idx := strings.Index(ref, "#"); idx > 0 {
		repo := ref[:idx]
		num := ref[idx+1:]
		if strings.Contains(repo, "/") {
			return num, repo
		}
	}

	// #123 → strip the #
	if strings.HasPrefix(ref, "#") {
		return ref[1:], ""
	}

	// Plain number
	return ref, ""
}

// queryPRState runs `gh pr view` and returns the PR state.
func queryPRState(ghArg, repoFlag string) (string, error) {
	args := []string{"pr", "view", ghArg, "--json", "state"}
	if repoFlag != "" {
		args = append(args, "--repo", repoFlag)
	}

	out, err := exec.Command("gh", args...).Output()
	if err != nil {
		return "", fmt.Errorf("gh pr view failed: %w", err)
	}

	var result struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return "", fmt.Errorf("parsing gh output: %w", err)
	}

	return result.State, nil
}
