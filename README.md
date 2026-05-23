# tkr

Markdown-native, git-resident project management CLI for AI coding agents.

Each ticket is a Markdown file with YAML frontmatter, stored in `.tkr/tickets/` and version-controlled alongside your code. No server, no database, no accounts.

## Install

```
go install github.com/tkr-cli/tkr/cmd/tkr@latest
```

Or build from source:

```
git clone https://github.com/tkr-cli/tkr.git
cd tkr
make build
```

## Quick start

```bash
tkr init                                    # scaffold .tkr/ in the repo
tkr new -t "Add OAuth2 login" -p high       # create a ticket
tkr next                                    # get highest-priority unblocked ticket
tkr claim TKR-1                             # mark in-progress, create feature branch
tkr done TKR-1                              # mark done
tkr release v1.0.0                          # generate CHANGELOG from done tickets
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize tkr in the current repository |
| `new` | Create a new ticket |
| `next` | Get the highest-priority unblocked ticket |
| `claim` | Mark a ticket as in-progress (auto-creates feature branch) |
| `done` | Mark a ticket as done |
| `delete` | Permanently remove a ticket |
| `list` | List tickets with optional filtering |
| `show` | Show ticket details and children |
| `lint` | Validate all tickets and check for dependency cycles |
| `graph` | Output dependency graph as Mermaid diagram |
| `release` | Generate CHANGELOG section from completed tickets |
| `watch` | Install git hooks for automatic status updates |
| `sync` | Sync ticket status from GitHub PR state |
| `import` | Import from Backlog.md or Taskmaster |
| `txt` | Export as grep-friendly todo.txt |

## Ticket format

```markdown
---
id: TKR-1
title: Add OAuth2 login
status: todo
priority: high
type: feature
actor: agent
labels:
  - auth
  - mvp
dependencies:
  - TKR-2
acceptance_criteria:
  - Tests pass
  - Docs updated
complexity: 5
estimate: 4h
---

## Implementation notes

OAuth2 PKCE flow with Google and GitHub providers.
```

## Filtering

```bash
tkr list status:todo                        # by status
tkr list priority:high,critical             # multiple values
tkr list status:todo priority:>2            # comparisons
tkr list +backend                           # shorthand for labels:contains:backend
tkr list status:todo,in-progress --limit 5  # combine with limit
```

## Dependencies and DAG

Tickets can declare dependencies, blocks, and parent-child relationships. The `next` command uses topological ordering to surface only unblocked tickets, sorted by priority, complexity, and estimate.

```bash
tkr new -t "Setup DB" -p high
tkr new -t "Add API" -p high --dep TKR-1
tkr next                                    # returns TKR-1 (TKR-2 is blocked)
tkr lint                                    # detects cycles
tkr graph                                   # Mermaid diagram
```

## Git integration

```bash
tkr watch install                           # install post-commit hook
```

The hook parses commit messages and updates ticket status automatically:

- `Closes TKR-1` or `Fixes TKR-1` → marks done
- `feat(TKR-1): ...` (conventional commits) → marks in-progress
- `WIP TKR-1` → marks in-progress

## Branch auto-creation

When `auto_branch` is enabled (default), `tkr claim` creates a feature branch:

```bash
tkr claim TKR-1                             # creates feat/tkr-1-add-oauth2-login
tkr claim TKR-1 --no-branch                 # skip branch creation
```

The branch pattern is configurable in `.tkr/config.yml`:

```yaml
auto_branch: true
branch_pattern: "feat/{id}-{slug}"
```

## Agent integration

tkr ships templates for AI coding agents:

- **Claude Code**: `tkr init` generates a Claude Skill in `.claude/skills/tkr/`
- **Codex/AGENTS.md**: prints an AGENTS.md snippet at init
- **Cursor**: prints a `.cursor/rules/` rule at init

The `--plain` flag outputs LLM-friendly text (auto-detected when piped). The `--json` flag outputs structured JSON.

```bash
tkr next --plain                            # deterministic output for agents
tkr list status:todo --json                 # structured JSON
```

## Output formats

All commands support three output modes:

- **Human** (default in TTY): colored, tabulated
- **Plain** (`--plain` or piped): key-value pairs, no ANSI codes
- **JSON** (`--json`): structured, machine-readable

## Configuration

`.tkr/config.yml`:

```yaml
prefix: TKR
default_priority: medium
default_status: todo
default_branch: main
auto_branch: true
branch_pattern: "feat/{id}-{slug}"
# default_definition_of_done:
#   - Code reviewed
#   - Tests written
```

## License

MIT
