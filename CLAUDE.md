# gh-tuissue

`gh-tuissue` is a GitHub CLI extension that provides a kanban-style TUI for browsing and updating repository issues. It is written in Go and built on Bubble Tea / Lip Gloss.

## Project layout

- `main.go` — entry point for the `gh` extension
- `internal/agent` — launching background Claude Code sessions for an issue
- `internal/cli` — CLI flag parsing and entry wiring
- `internal/config` — local configuration (project binding, etc.)
- `internal/domain` — core types (issue, column, filter, …)
- `internal/github` — `gh` CLI / GitHub API access
- `internal/repo` — repository resolution from `origin`
- `internal/service` — application services that glue domain + github together
- `internal/ui` — Bubble Tea models, views, and keybindings

User-facing docs live in `README.md`; contributor docs in `CONTRIBUTING.md`.

## Development

- Go `1.25.0` or newer, and an authenticated `gh` CLI
- Build: `go build ./...`
- Test: `go test ./...`
- Run from source: `go run . --repo owner/name --project 1`

## Coding guidelines

- Follow existing Go conventions; keep changes small and focused.
- Prefer clear names and explicit error handling over clever shortcuts.
- Reuse existing packages and helpers before adding new abstractions.
- Keep TUI changes consistent with existing Bubble Tea / Lip Gloss patterns.
- Update `README.md` when user-visible behavior (flags, keybindings, views) changes.

## Language

- Think in English. Respond to the user in Japanese.
- Code, identifiers, commit messages, and documentation committed to the repository are written in English.
