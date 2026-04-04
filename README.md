# gh-tuissue

`gh-tuissue` is a GitHub CLI extension for browsing and updating repository issues from the terminal. It provides a kanban-style view backed by GitHub Issues and, when configured, GitHub Projects columns.

## Features

- Browse issues in a terminal UI
- Switch columns in a kanban-style board
- Open issue details and update title, body, labels, assignees, milestone, and status
- Create new issues without leaving the terminal
- Optionally map board columns to a GitHub Project

## Preview

`gh-tuissue` is a terminal application. Once launched, it shows a multi-column issue board with detail, filter, and help views inside an alternate screen.

## Requirements

- `gh` CLI installed and authenticated
- Go `1.25.0` or newer for local development
- Access to the target GitHub repository and, if used, its GitHub Project

## Installation

Install as a GitHub CLI extension:

```bash
gh extension install masamichhhhi/gh-tuissue
```

Update later with:

```bash
gh extension upgrade gh-tuissue
```

## Usage

Inside a cloned Git repository with `origin` configured:

```bash
gh tuissue --project 1
```

Or specify a repository explicitly:

```bash
gh tuissue --repo owner/name --project 1
```

Show the current version:

```bash
gh tuissue --version
```

### Options

- `--repo`, `-R`: repository in `owner/name` format. If omitted, `origin` is used.
- `--project`, `-p`: GitHub Project number used to derive kanban columns.
- `--version`: print version and exit.

## Keyboard shortcuts

### Board view

- `h` / `l` / `←` / `→`: switch column
- `j` / `k` / `↑` / `↓`: move cursor
- `H` / `L`: move issue status left or right
- `Enter`: open issue detail
- `f`: open filter panel
- `r`: refresh issues
- `n`: create a new issue
- `?`: show help

### Detail view

- `j` / `k` / `↑` / `↓`: scroll
- `s`: toggle open or closed status
- `e`: edit body with `$EDITOR`
- `t`: edit title
- `l`: edit labels
- `a`: edit assignees
- `m`: edit milestone
- `c`: add comment with `$EDITOR`
- `Esc`: return to the board

### Filter view

- `j` / `k` / `↑` / `↓`: move cursor
- `Tab`: switch pane
- `Space`: toggle selection
- `Enter`: apply filters
- `Esc`: cancel

## Development

Clone the repository and run the standard Go checks:

```bash
go test ./...
go build ./...
```

Build a local binary:

```bash
go build -o gh-tuissue .
```

## Contributing

Contributions are welcome. Please read [`CONTRIBUTING.md`](./CONTRIBUTING.md) before opening a pull request.

## License

This project is released under the [`MIT`](./LICENSE) license.
