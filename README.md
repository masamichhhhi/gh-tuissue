# gh-tuissue

![tuissue-demo](./demo.gif)


`gh-tuissue` is a GitHub CLI extension that provides a TUI (Text User Interface) for browsing and updating repository issues from the terminal. It offers a kanban-style view backed by GitHub Issues and, when configured, GitHub Projects columns.

## Installation

Install as a GitHub CLI extension:

```bash
gh extension install masamichhhhi/gh-tuissue
```

Update later with:

```bash
gh extension upgrade gh-tuissue
```

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
- Access to the target GitHub repository and, if used, its GitHub Project


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
- `--config`: re-select the GitHub Project binding.
- `--version`: print version and exit.

## Keyboard shortcuts

### Global

- `Ctrl+C`: quit
- `Esc`: back (on board view, quit)
- `?`: show help

### Board view

- `h` / `l` / `←` / `→`: switch column
- `j` / `k` / `↑` / `↓`: move cursor
- `H` / `L` / `Shift+←` / `Shift+→`: move issue status left or right
- `d`: hide current column
- `D`: show all hidden columns
- `Enter`: open issue detail
- `f`: open filter panel
- `r`: refresh issues
- `n`: create a new issue (assigned to the currently selected column's status)

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

## Contributing

Contributions are welcome. Please read [`CONTRIBUTING.md`](./CONTRIBUTING.md) before opening a pull request.

## License

This project is released under the [`MIT`](./LICENSE) license.
