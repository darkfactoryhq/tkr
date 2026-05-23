package bootstrap

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templates embed.FS

type InitOptions struct {
	Prefix  string
	NoSkill bool
}

type templateData struct {
	Prefix string
}

func Init(root string, opts InitOptions) error {
	if opts.Prefix == "" {
		opts.Prefix = "TKR"
	}

	tkrDir := filepath.Join(root, ".tkr")
	if _, err := os.Stat(tkrDir); err == nil {
		return fmt.Errorf(".tkr/ already exists in %s", root)
	}

	if err := os.MkdirAll(filepath.Join(tkrDir, "tickets"), 0755); err != nil {
		return fmt.Errorf("create .tkr/tickets: %w", err)
	}

	if err := os.WriteFile(filepath.Join(tkrDir, ".seq"), []byte("0\n"), 0644); err != nil {
		return fmt.Errorf("create .seq: %w", err)
	}

	configContent := fmt.Sprintf(`prefix: %s
default_priority: medium
default_status: todo
default_branch: main
auto_branch: true
branch_pattern: "feat/{id}-{slug}"
# default_definition_of_done:
#   - Code reviewed
#   - Tests written
#   - Documentation updated
`, opts.Prefix)
	if err := os.WriteFile(filepath.Join(tkrDir, "config.yml"), []byte(configContent), 0644); err != nil {
		return fmt.Errorf("create config.yml: %w", err)
	}

	data := templateData{Prefix: opts.Prefix}

	if err := os.WriteFile(filepath.Join(tkrDir, "tickets", ".gitkeep"), []byte(""), 0644); err != nil {
		return fmt.Errorf("create .gitkeep: %w", err)
	}

	if !opts.NoSkill {
		skillDir := filepath.Join(root, ".claude", "skills", "tkr")
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("create skill directory: %w", err)
		}
		if err := writeTemplate(filepath.Join(skillDir, "SKILL.md"), "templates/skill.md.tmpl", data); err != nil {
			return fmt.Errorf("create SKILL.md: %w", err)
		}
	}

	fmt.Println("Initialized tkr in .tkr/")
	fmt.Printf("  Ticket prefix: %s\n", opts.Prefix)
	fmt.Printf("  Tickets dir:   .tkr/tickets/\n")
	fmt.Printf("  Config:        .tkr/config.yml\n")
	if !opts.NoSkill {
		fmt.Printf("  Claude Skill:  .claude/skills/tkr/SKILL.md\n")
	}

	fmt.Println()
	fmt.Println("Add this to your AGENTS.md:")
	fmt.Println("---")
	if err := renderTemplate("templates/agents_snippet.md.tmpl", data); err != nil {
		return err
	}
	fmt.Println("---")

	fmt.Println()
	fmt.Println("Save this as .cursor/rules/tkr.mdc:")
	fmt.Println("---")
	if err := renderTemplate("templates/cursor_rule.mdc.tmpl", data); err != nil {
		return err
	}
	fmt.Println("---")

	return nil
}

func writeTemplate(dest, tmplPath string, data templateData) error {
	content, err := templates.ReadFile(tmplPath)
	if err != nil {
		return err
	}
	t, err := template.New(filepath.Base(tmplPath)).Parse(string(content))
	if err != nil {
		return err
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, data)
}

func renderTemplate(tmplPath string, data templateData) error {
	content, err := templates.ReadFile(tmplPath)
	if err != nil {
		return err
	}
	t, err := template.New(filepath.Base(tmplPath)).Parse(string(content))
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, data)
}
