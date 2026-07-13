# Roadmap

## v1.2.0 — shell-aware updates

- Shell-aware delegation to the official installers
- Active executable targeting on every supported platform
- Deferred PowerShell replacement on Windows
- SHA-256 verification in both installation scripts

## v1.1.1 — updater reliability

- Visible update progress and actionable errors
- Resilient Windows release downloads with retries and a `curl.exe` fallback

## v1.1.0 — TUI workflow

- Consistent fixed-width `[ ]`, `[!]`, and `[x]` status indicators
- Search inside the TUI with `s` or `/`
- Active-only and complete listing views toggled with `v`
- Completed-spark cleanup with `C` and confirmation

## v1.0.0 — stable

- Native Go CLI with local SQLite storage
- Complete command workflow for capturing and organizing sparks
- Nested parent-child thoughts and hierarchical tree output
- Navigable full-screen TUI with keyboard actions
- Color and visual feedback with accessible plain-output fallbacks
- JSON output, redirected command mode, and self-update support
- Tested GoReleaser artifacts for Windows, macOS, and Linux

## Distribution

- Homebrew
- Scoop, winget, AUR, and Nix

## Future

- Export/import and optional encrypted sync
- Raycast/Alfred integration ideas
- Rust implementation on `main` while retaining the Go implementation on a dedicated branch
