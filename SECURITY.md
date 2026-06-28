# Security Policy

## Supported versions

tkr is pre-1.0 and ships from `main`. Security fixes are applied to the latest
release and the `main` branch. Older tagged releases are not patched —
please upgrade to the latest version.

| Version | Supported |
|---------|-----------|
| latest release / `main` | ✅ |
| older releases | ❌ |

## Reporting a vulnerability

**Please do not report security vulnerabilities through public GitHub issues,
discussions, or pull requests.**

Instead, use GitHub's private vulnerability reporting:

1. Go to the repository's **Security** tab.
2. Click **Report a vulnerability**.
3. Provide a description, reproduction steps, affected version (`tkr version`),
   and any relevant impact assessment.

If you cannot use private reporting, email the maintainers at
**security@darkfactoryhq.dev** <!-- TODO: set a real, monitored address -->.

## What to expect

- **Acknowledgement** within 3 business days.
- An initial assessment and severity triage within 7 days.
- Coordinated disclosure: we'll agree on a timeline and credit you (if you
  wish) in the release notes once a fix is available.

## Scope

tkr is a local CLI that reads and writes files in your repository and installs
git hooks. When reporting, the most relevant areas are:

- Path handling / writes outside the intended `.tkr/` directory.
- Git hook installation and commit-message parsing.
- Importers that parse untrusted Markdown/JSON (`tkr import`).
- Any code path that executes shell commands.

Thanks for helping keep tkr and its users safe.
