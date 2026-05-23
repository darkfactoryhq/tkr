package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func CurrentBranch() (string, error) {
	out, err := exec.Command("git", "symbolic-ref", "--short", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git symbolic-ref: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func DefaultBranch() (string, error) {
	out, err := exec.Command("git", "config", "init.defaultBranch").Output()
	if err == nil {
		if branch := strings.TrimSpace(string(out)); branch != "" {
			return branch, nil
		}
	}

	// Check if "master" branch exists; if so, use it.
	if err := exec.Command("git", "rev-parse", "--verify", "refs/heads/master").Run(); err == nil {
		return "master", nil
	}

	return "main", nil
}

func IsDefaultBranch() (bool, error) {
	current, err := CurrentBranch()
	if err != nil {
		return false, err
	}
	def, err := DefaultBranch()
	if err != nil {
		return false, err
	}
	return current == def, nil
}

func IsWorktree() (bool, error) {
	root, err := RepoRoot()
	if err != nil {
		return false, err
	}
	info, err := os.Lstat(root + "/.git")
	if err != nil {
		return false, fmt.Errorf("stat .git: %w", err)
	}
	return !info.IsDir(), nil
}
