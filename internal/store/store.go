package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/darkfactoryhq/tkr/internal/ticket"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Prefix                string   `yaml:"prefix"`
	DefaultPriority       string   `yaml:"default_priority"`
	DefaultStatus         string   `yaml:"default_status"`
	DefaultBranch         string   `yaml:"default_branch"`
	AutoBranch            bool     `yaml:"auto_branch"`
	BranchPattern         string   `yaml:"branch_pattern"`
	DefaultDefinitionOfDone []string `yaml:"default_definition_of_done"`
}

type Store struct {
	root   string
	config Config
}

func New(root string) (*Store, error) {
	tkrDir := filepath.Join(root, ".tkr")
	info, err := os.Stat(tkrDir)
	if err != nil {
		return nil, fmt.Errorf("tkr root not found: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", tkrDir)
	}

	s := &Store{root: root}
	if err := s.LoadConfig(); err != nil {
		return nil, err
	}
	return s, nil
}

func FindRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	for {
		tkrDir := filepath.Join(dir, ".tkr")
		if info, err := os.Stat(tkrDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a tkr project (no .tkr/ directory found)")
		}
		dir = parent
	}
}

func (s *Store) Root() string {
	return s.root
}

func (s *Store) Config() Config {
	return s.config
}

func (s *Store) LoadConfig() error {
	s.config = Config{
		Prefix:          "TKR",
		DefaultPriority: "medium",
		DefaultStatus:   "todo",
		DefaultBranch:   "main",
		AutoBranch:      true,
		BranchPattern:   "feat/{id}-{slug}",
	}

	data, err := os.ReadFile(filepath.Join(s.root, ".tkr", "config.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Prefix != "" {
		s.config.Prefix = cfg.Prefix
	}
	if cfg.DefaultPriority != "" {
		s.config.DefaultPriority = cfg.DefaultPriority
	}
	if cfg.DefaultStatus != "" {
		s.config.DefaultStatus = cfg.DefaultStatus
	}
	if cfg.DefaultBranch != "" {
		s.config.DefaultBranch = cfg.DefaultBranch
	}
	if cfg.BranchPattern != "" {
		s.config.BranchPattern = cfg.BranchPattern
	}
	// AutoBranch: yaml unmarshals false as zero value, so we only override
	// when the file explicitly sets it. We detect this by re-checking the raw YAML.
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err == nil {
		if v, ok := raw["auto_branch"]; ok {
			if b, ok := v.(bool); ok {
				s.config.AutoBranch = b
			}
		}
	}

	if cfg.DefaultDefinitionOfDone != nil {
		s.config.DefaultDefinitionOfDone = cfg.DefaultDefinitionOfDone
	}

	return nil
}

func (s *Store) Load() ([]ticket.Ticket, error) {
	ticketDir := filepath.Join(s.root, ".tkr", "tickets")
	entries, err := os.ReadDir(ticketDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading tickets directory: %w", err)
	}

	var tickets []ticket.Ticket
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || strings.HasPrefix(e.Name(), ".") {
			continue
		}

		t, err := ticket.ParseFile(filepath.Join(ticketDir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", e.Name(), err)
		}
		tickets = append(tickets, *t)
	}

	return tickets, nil
}

func (s *Store) Get(id string) (*ticket.Ticket, error) {
	filename := id + ".md"
	path := filepath.Join(s.root, ".tkr", "tickets", filename)

	t, err := ticket.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading ticket %s: %w", id, err)
	}
	return t, nil
}

func (s *Store) Save(t *ticket.Ticket) error {
	ticketDir := filepath.Join(s.root, ".tkr", "tickets")
	if err := os.MkdirAll(ticketDir, 0o755); err != nil {
		return fmt.Errorf("creating tickets directory: %w", err)
	}

	t.Updated = time.Now()

	data, err := ticket.Marshal(t)
	if err != nil {
		return fmt.Errorf("marshaling ticket: %w", err)
	}

	finalPath := filepath.Join(ticketDir, t.ID+".md")
	tmpPath := filepath.Join(ticketDir, "."+t.ID+".md.tmp")

	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	t.FilePath = finalPath
	return nil
}
