# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> Tagged releases and their notes are produced automatically by
> [GoReleaser](https://goreleaser.com/) from Conventional Commit history.
> This file tracks human-curated highlights.

## [Unreleased]

## [0.1.1] - 2026-06-29

### Fixed
- `tkr version` now reports the correct version for binaries built with
  `go install <module>@<version>`, by falling back to `runtime/debug` build
  metadata when GoReleaser's ldflags are absent.

## [0.1.0] - 2026-06-28

### Added
- Initial public release.
- Community health files: `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`,
  `SECURITY.md`, issue/PR templates, Dependabot config, and `CODEOWNERS`.

[Unreleased]: https://github.com/darkfactoryhq/tkr/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/darkfactoryhq/tkr/releases/tag/v0.1.1
[0.1.0]: https://github.com/darkfactoryhq/tkr/releases/tag/v0.1.0
