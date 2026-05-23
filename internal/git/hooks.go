package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const hookMarker = "Installed by tkr"

const postCommitScript = `#!/bin/sh
# Installed by tkr -- do not edit manually
tkr _hook-post-commit
# Run previous hook if it existed
if [ -f "$(dirname "$0")/post-commit.pre-tkr" ]; then
    "$(dirname "$0")/post-commit.pre-tkr" "$@"
fi
`

func RepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func InstallHooks(repoRoot string) error {
	hooksDir := filepath.Join(repoRoot, ".git", "hooks")
	hookPath := filepath.Join(hooksDir, "post-commit")
	backupPath := filepath.Join(hooksDir, "post-commit.pre-tkr")

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("create hooks dir: %w", err)
	}

	if info, err := os.Stat(hookPath); err == nil && info.Size() > 0 {
		existing, err := os.ReadFile(hookPath)
		if err != nil {
			return fmt.Errorf("read existing hook: %w", err)
		}
		if !strings.Contains(string(existing), hookMarker) {
			if err := os.Rename(hookPath, backupPath); err != nil {
				return fmt.Errorf("backup existing hook: %w", err)
			}
		}
	}

	if err := os.WriteFile(hookPath, []byte(postCommitScript), 0755); err != nil {
		return fmt.Errorf("write post-commit hook: %w", err)
	}
	return nil
}

func UninstallHooks(repoRoot string) error {
	hooksDir := filepath.Join(repoRoot, ".git", "hooks")
	hookPath := filepath.Join(hooksDir, "post-commit")
	backupPath := filepath.Join(hooksDir, "post-commit.pre-tkr")

	if data, err := os.ReadFile(hookPath); err == nil {
		if strings.Contains(string(data), hookMarker) {
			if err := os.Remove(hookPath); err != nil {
				return fmt.Errorf("remove hook: %w", err)
			}
		}
	}

	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Rename(backupPath, hookPath); err != nil {
			return fmt.Errorf("restore backup hook: %w", err)
		}
	}

	return nil
}
