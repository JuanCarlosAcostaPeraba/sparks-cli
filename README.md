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
SPARKS_VERSION=1.0.0 curl -fsSL https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.sh | sh
SPARKS_INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.sh | sh
```

```powershell
$env:SPARKS_VERSION = "1.0.0"; irm https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.ps1 | iex
$env:SPARKS_INSTALL_DIR = "$HOME\bin"; irm https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.ps1 | iex
```

Homebrew, Scoop, and winget packages are planned after the MVP.

## Usage

```bash
sparks
sparks list
sparks list -a
sparks add Prepare Codex prompt
sparks + Create Homebrew tap
sparks add --parent 1 "Add release notes"
sparks edit 1 "Prepare release notes"
sparks e 1 "Prepare release notes"
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
sparks update
sparks version
```

The default command opens a full-screen interactive table when the terminal is
interactive. Navigate with the arrow keys or `j`/`k`; use `a` to add, `e` to
edit, `i` to toggle importance, `c` to add a child to the selected spark, `d`
to complete, and `x` to remove. Use `s` or `/` to search, `v` to switch
between active-only and all sparks, and `C` to clear completed sparks after a
confirmation. Press `?` for help or `q` to quit.

```txt
  SEL  ID     STATE       TITLE
   >   #1     active      Prepare release notes
       #2     important   Publish Homebrew tap
```

When input or output is redirected, `sparks` keeps the line-based interactive
fallback so scripts can send regular commands followed by `exit` or `quit`.

```txt
STATUS  ID  TITLE
[ ]     1   Prepare Codex prompt
[!]     2   Publish Homebrew tap
[x]     3   Initial README
```

Flags have short aliases: `-a` for `--all`, `-j` for `--json`, `-p` for
`--parent`, `-y` for `--yes`, and `-d` for `--db`.

Interactive terminal output uses color to distinguish IDs, important and
completed sparks, selections, shortcuts, and action feedback. Color is omitted
from JSON and redirected output; set the standard `NO_COLOR` environment
variable to disable it in a terminal as well.

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

Installed release binaries can update themselves with `sparks update`. The
command verifies the GoReleaser SHA-256 checksum before replacing the current
executable.

Releases are handled by GoReleaser. Tag-based GitHub Actions builds publish archives and checksums.

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Roadmap

### v1.0.0 — stable

- Native Go CLI with local SQLite storage
- Add, list, edit, complete, prioritize, remove, search, clear, and tree commands
- Nested thoughts with parent-child relationships
- Full-screen navigable TUI with inline actions and keyboard help
- Color-coded IDs, states, selections, and action feedback with `NO_COLOR` support
- JSON and redirected output suitable for scripts
- Verified self-update command and cross-platform installers
- GoReleaser archives and checksums for Windows, macOS, and Linux

### Next

- Homebrew distribution
- Optional encrypted sync
- Raycast/Alfred integration ideas
- More package managers: Scoop, winget, AUR, and Nix
- Long-term Rust implementation while retaining the Go branch

## License

MIT
