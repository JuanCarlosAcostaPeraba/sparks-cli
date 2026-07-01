# sparks-cli

A tiny, fast CLI to capture ideas, tasks and nested thoughts without leaving your terminal.

`sparks` is a native Go command-line tool for quickly recording ideas and tasks. It stores data in a local SQLite database under the operating system's application data directory, not beside the executable.

## Installation

### Homebrew

```bash
brew tap JuanCarlosAcostaPeraba/tap
brew install sparks
```

Or:

```bash
brew install JuanCarlosAcostaPeraba/tap/sparks
```

### Manual binaries

Download the archive for your platform from the GitHub Releases page, extract it, and place the `sparks` binary on your `PATH`.

### Windows

Windows binaries are built by GoReleaser. Scoop and winget packaging are planned after the MVP.

## Usage

```bash
sparks
sparks list
sparks add "Prepare Codex prompt"
sparks + "Create Homebrew tap"
sparks done 3
sparks ok 3
sparks important 3
sparks ! 3
sparks remove 3
sparks rm 3
sparks - 3
sparks clear
sparks clear --all --yes
sparks tree
sparks search "codex"
sparks version
```

The default command lists active sparks.

```txt
□ 1) Prepare Codex prompt
❗ 2) Publish Homebrew tap
☑ 3) Initial README
```

Most list-style commands support `--json`.

## Data Location

- macOS: `~/Library/Application Support/sparks/sparks.db`
- Linux: `$XDG_DATA_HOME/sparks/sparks.db`, or `~/.local/share/sparks/sparks.db`
- Windows: `%APPDATA%\sparks\sparks.db`

For tests or isolated runs, pass `--db /path/to/sparks.db`.

## Development

```bash
go mod tidy
go test ./...
go vet ./...
go run . add "Try sparks"
go run . list
```

## Releases

Releases are handled by GoReleaser. Tag-based GitHub Actions builds publish archives, checksums and Homebrew tap updates.

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Roadmap

### v0.1.0

- Core CLI
- Add/list/done/important/remove/search/tree
- SQLite storage
- JSON output
- Tests
- GitHub Actions
- GoReleaser config
- Homebrew Tap config

### v0.2.0

- Edit command
- Tags
- Export/import
- Shell completions
- Better themes

### v0.3.0

- TUI mode
- Optional encrypted sync
- Raycast/Alfred integration ideas
- More package managers: Scoop, winget, AUR, Nix

## License

MIT
