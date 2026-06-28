package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/darkfactoryhq/tkr/internal/changelog"
	"github.com/darkfactoryhq/tkr/internal/ticket"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Generate CHANGELOG section from completed tickets",
	RunE:  runRelease,
}

func init() {
	releaseCmd.Flags().String("tag", "", "Version tag (e.g., 1.2.0)")
	releaseCmd.Flags().Bool("write", false, "Prepend to CHANGELOG.md")
}

func runRelease(cmd *cobra.Command, args []string) error {
	s, err := storeFromCwd()
	if err != nil {
		return err
	}

	tickets, err := s.Load()
	if err != nil {
		return err
	}

	var done []ticket.Ticket
	for _, t := range tickets {
		if t.Status == ticket.StatusDone {
			done = append(done, t)
		}
	}

	if len(done) == 0 {
		return fmt.Errorf("no completed tickets to include in release")
	}

	tag, _ := cmd.Flags().GetString("tag")
	if tag == "" {
		tag = "Unreleased"
	}

	section := changelog.Generate(done, tag, time.Now())
	output := section.String()

	write, _ := cmd.Flags().GetBool("write")
	if write {
		existing, _ := os.ReadFile("CHANGELOG.md")
		content := output + "\n" + string(existing)
		if err := os.WriteFile("CHANGELOG.md", []byte(content), 0644); err != nil {
			return fmt.Errorf("write CHANGELOG.md: %w", err)
		}
		fmt.Println("Updated CHANGELOG.md")
	} else {
		fmt.Print(output)
	}

	return nil
}
