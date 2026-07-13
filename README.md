# sparks-cli

A tiny, fast CLI to capture ideas, tasks and nested thoughts without leaving your terminal.

`sparks` is a native Go command-line tool for quickly recording ideas and tasks. It stores data in a local SQLite database under the operating system's application data directory, not beside the executable.

## Installation

### macOS and Linux

Install the latest release with one command:

```bash
curl -fsSL https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.sh | sh
```

The installer downloads the matching release archive and places `sparks` in
`~/.local/bin`.

### Windows

Install the latest release from PowerShell with one command:

```powershell
irm https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.ps1 | iex
```

The installer downloads the matching release archive, installs `sparks.exe` under
`%LOCALAPPDATA%\Programs\sparks`, and adds that directory to your user `PATH`.

### Options

```bash
SPARKS_VERSION=0.1.0 curl -fsSL https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.sh | sh
SPARKS_INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.sh | sh
```

```powershell
$env:SPARKS_VERSION = "0.1.0"; irm https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.ps1 | iex
$env:SPARKS_INSTALL_DIR = "$HOME\bin"; irm https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.ps1 | iex
```

Homebrew, Scoop, and winget packages are planned after the MVP.

## Usage

```bash
sparks
sparks list
sparks list -a
sparks add "Prepare Codex prompt"
sparks + "Create Homebrew tap"
sparks add --parent 1 "Add release notes"
sparks done 3
sparks ok 3
sparks important 3
sparks ! 3
sparks remove 3
sparks rm 3
sparks - 3
sparks clear
sparks clear -a -y
sparks tree
sparks search "codex"
sparks version
```

The default command lists active sparks.

```txt
STATUS  ID  TITLE
□       1   Prepare Codex prompt
❗       2   Publish Homebrew tap
☑       3   Initial README
```

Flags have short aliases: `-a` for `--all`, `-j` for `--json`, `-p` for
`--parent`, `-y` for `--yes`, and `-d` for `--db`.

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

Releases are handled by GoReleaser. Tag-based GitHub Actions builds publish archives and checksums.

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
