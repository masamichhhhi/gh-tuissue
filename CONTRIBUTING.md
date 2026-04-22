# Contributing to gh-tuissue

Thanks for your interest in improving `gh-tuissue`.

## Contribution flow

1. Open an issue first for bugs, enhancements, or larger changes.
2. Fork the repository and create a focused branch from `main`.
3. Make your changes with tests or validation where applicable.
4. Open a pull request that clearly explains the problem, approach, and verification steps.

## Development setup

### Prerequisites

- Go `1.25.0` or newer
- GitHub CLI (`gh`) authenticated against GitHub

### Getting started

```bash
git clone https://github.com/<your-account>/gh-tuissue.git
cd gh-tuissue
go test ./...
go build ./...
```

Run the extension from source during development:

```bash
go run . --repo owner/name --project 1
```

Build a local binary:

```bash
go build -o gh-tuissue .
```

## Coding guidelines

- Follow existing Go conventions and keep changes small and focused.
- Prefer clear names and explicit error handling over clever shortcuts.
- Reuse existing packages and helpers before adding new abstractions.
- Keep terminal UX changes consistent with the current Bubble Tea / Lip Gloss patterns.

## Commit and pull request guidelines

- Use descriptive commit messages that explain the intent of the change.
- Keep unrelated changes out of the same pull request.
- Include reproduction steps for bug fixes and usage notes for behavior changes.
- Update documentation when user-visible behavior changes.

## Review process

- Pull requests should pass CI before review is requested.
- Review feedback should be addressed with follow-up commits.
- Maintainers may ask for scope reduction if a change becomes too broad.
