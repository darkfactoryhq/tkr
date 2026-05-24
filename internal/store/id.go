package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (s *Store) NextID() (string, error) {
	seqPath := filepath.Join(s.root, ".tkr", ".seq")

	f, err := os.OpenFile(seqPath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return "", fmt.Errorf("opening seq file: %w", err)
	}
	defer f.Close()

	if err := lockFile(f); err != nil {
		return "", fmt.Errorf("acquiring lock: %w", err)
	}
	defer unlockFile(f)

	data, err := os.ReadFile(seqPath)
	if err != nil {
		return "", fmt.Errorf("reading seq file: %w", err)
	}

	current := 0
	if s := strings.TrimSpace(string(data)); s != "" {
		current, err = strconv.Atoi(s)
		if err != nil {
			return "", fmt.Errorf("parsing seq value %q: %w", s, err)
		}
	}

	next := current + 1

	if err := f.Truncate(0); err != nil {
		return "", fmt.Errorf("truncating seq file: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return "", fmt.Errorf("seeking seq file: %w", err)
	}
	if _, err := fmt.Fprintf(f, "%d\n", next); err != nil {
		return "", fmt.Errorf("writing seq file: %w", err)
	}

	return fmt.Sprintf("%s-%d", s.config.Prefix, next), nil
}
