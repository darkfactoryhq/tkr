# tkr

Markdown-native, git-resident project management CLI for AI coding agents.

## Build

```
go build -o tkr ./cmd/tkr
```

## Test

```
go test -race ./...
```

## Architecture

- `cmd/tkr/main.go` — entry point
- `internal/cmd/` — Cobra commands (one per file)
- `internal/ticket/` — core data model, frontmatter parse/serialize, validation
- `internal/store/` — filesystem I/O, ID generation, filter DSL
- `internal/dag/` — dependency graph, cycle detection (Kahn's), next-ticket selection
- `internal/git/` — hooks, commit message parsing, worktree safety
- `internal/changelog/` — changelog generation from done tickets
- `internal/output/` — human/plain/JSON formatters
- `internal/bootstrap/` — `tkr init` scaffolding and templates

## Conventions

- Tickets stored as Markdown + YAML frontmatter in `.tkr/tickets/`
- ID prefix configurable (default TKR), monotonic counter in `.tkr/.seq`
- All output supports `--plain` (auto-detect non-TTY) and `--json`
- New tickets must be created on the default branch (enforced, `--force` to override)
