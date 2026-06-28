# Contributing to tkr

Thanks for your interest in improving **tkr** — a Markdown-native, git-resident
project manager for AI coding agents. Contributions of all kinds are welcome:
bug reports, feature ideas, docs, and code.

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). By
participating, you agree to uphold it.

## Getting started

You'll need **Go 1.26+** (the exact version is pinned in [`go.mod`](go.mod)).

```bash
git clone https://github.com/darkfactoryhq/tkr.git
cd tkr
make build        # builds ./tkr
./tkr --help
```

Common tasks (see the [`Makefile`](Makefile)):

```bash
make build        # go build -o tkr ./cmd/tkr
make test         # go test -race ./...
make lint         # go vet ./...
make install      # go install ./cmd/tkr
make clean
```

CI additionally runs [`golangci-lint`](https://golangci-lint.run/) — match it locally with:

```bash
golangci-lint run
```

## Project layout

| Path | Responsibility |
|------|----------------|
| `cmd/tkr/` | Entry point |
| `internal/cmd/` | Cobra commands (one per file) |
| `internal/ticket/` | Core data model, frontmatter parse/serialize, validation |
| `internal/store/` | Filesystem I/O, ID generation, filter DSL |
| `internal/dag/` | Dependency graph, cycle detection, next-ticket selection |
| `internal/git/` | Hooks, commit-message parsing, worktree safety |
| `internal/changelog/` | Changelog generation from done tickets |
| `internal/output/` | Human / plain / JSON formatters |
| `internal/bootstrap/` | `tkr init` scaffolding and templates |
| `internal/migrate/` | Importers (Backlog.md, Taskmaster) |

## Making a change

1. **Open an issue first** for anything non-trivial, so we can agree on the
   approach before you invest time.
2. Fork and create a topic branch off `main` (e.g. `feat/agent-routing` or
   `fix/cycle-detection`).
3. Keep changes focused; one logical change per PR.
4. Add or update tests — every package with logic has a `_test.go`, and new
   behavior should ship with coverage.
5. Run `make test` and `golangci-lint run` locally before pushing.
6. Update docs (`README.md`, command help text) when behavior changes.

## Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org/). This is
also what `tkr`'s own git hook parses, so it keeps the project self-consistent:

```
feat(store): add labels:contains filter shorthand
fix(dag): detect single-node self-dependency cycles
docs: clarify --plain output contract
```

Type prefixes: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`.
The release changelog is generated from these, so write them for a human reader.

## Pull request checklist

- [ ] Tests pass: `make test`
- [ ] Lint is clean: `golangci-lint run` and `go vet ./...`
- [ ] New/changed behavior is covered by tests
- [ ] Docs updated where relevant
- [ ] Commits follow Conventional Commits

## Reporting bugs & requesting features

Use the [issue templates](https://github.com/darkfactoryhq/tkr/issues/new/choose).
For security issues, **do not** open a public issue — see [SECURITY.md](SECURITY.md).

## License

By contributing, you agree that your contributions will be licensed under the
[MIT License](LICENSE) that covers this project.
