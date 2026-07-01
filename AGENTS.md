# Agent Instructions

## Project Goal

Build `sparks`, a tiny, fast native CLI for capturing ideas, tasks and nested thoughts without leaving the terminal.

## Stack

- Go
- Cobra for CLI commands
- SQLite for local storage
- GoReleaser for releases
- GitHub Actions for CI

Do not add Node.js. Do not use shell as the main implementation.

## Architecture Rules

- Keep Cobra command files thin.
- Keep business logic in `internal/`.
- Keep persistent data in OS-appropriate application data directories.
- Do not store data beside the executable.
- Use parameterized SQL only.
- Prefer CGO-free dependencies where practical for cross-platform builds.

## Testing

Run:

```bash
go test ./...
go vet ./...
```

Tests must use temporary databases and must not touch the user's real local database.

## Releases

GoReleaser owns release builds and packaging metadata. Release publishing should only happen from tag-triggered workflows.
