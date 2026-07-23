# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`audit` is a CLI tool that audits a project's Node packages (via Yarn) and/or Ruby gems (via bundler-audit) for known CVEs, optionally upgrades vulnerable (or all) packages, and commits/pushes the result to a dated audit branch (e.g. `audit-2026-07-23`).

## Commands

```bash
go build          # build the `audit` binary
go run .          # run without building
go run . -a       # run with --upgrade-all
go test ./...     # run all tests
go test -run TestGitCurrentBranchReturnsCurrentBranch ./...  # run a single test
```

## Architecture

Single `main` package, files split by responsibility:

- `main.go` — entry point. Parses flags, enforces the git preconditions (must be on `main`/`master`/the audit branch, warns about dirty working tree), checks Node/Ruby version minimums, creates/checks out the `audit-YYYY-MM-DD` branch, runs the Node and/or Ruby audit flows, prints a results table, and offers to commit + push lockfile changes.
- `audit.go` — orchestration layer (`auditNodePackages`, `auditRubyGems`): runs an audit, prompts the user (via `huh` multi-select) to choose which packages/gems to upgrade unless `--upgrade-all` is set, performs upgrades, re-audits, and returns `[]UpgradeResult` for display.
- `yarn.go` — shells out to `yarn npm audit --json` and `yarn up`/`yarn info`; parses yarn's line-delimited JSON output into `YarnIssue`/`YarnInfo`.
- `bundler.go` — shells out to `bundle exec bundle-audit check --format=json`; note it tolerates exit code 1 (means vulnerabilities found, not a tool failure) and strips any leading non-JSON text from the output before parsing.
- `node.go` / `ruby.go` — version checks against `MIN_NODE_VERSION`/`MIN_RUBY_VERSION` (declared in `main.go`) using `golang.org/x/mod/semver`.
- `git.go` — thin wrappers around `git` subcommands (status, branch, add, commit, push) used by `main.go`.
- `command.go` — shared helper that runs an external command while showing a `huh/spinner` progress indicator.
- `utils.go` — small helpers (`confirm` prompt, `fileExists`, `unique`).

Presence of `package.json` / `Gemfile` in the working directory determines whether the Node and/or Ruby audit flow runs — a repo can have either or both.

Interactive prompts (multi-select, confirm) are built with `charm.land/huh/v2`; the results table uses `charm.land/lipgloss/v2`'s `table` package.

## Notes

- External tool dependencies at runtime: `git`, `node`, `yarn` (for Node projects), `ruby`, `bundle`/`bundle-audit` (for Ruby projects). These aren't Go dependencies — they must be present on PATH when running `audit`.
- Go module path is `github.com/friendsoftheweb/audit`; installed via `go install github.com/friendsoftheweb/audit@latest`.
