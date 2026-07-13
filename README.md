<div align="center">

# ✨ sparks

**Catch the thought before it disappears.**

A tiny, fast, local-first CLI for ideas, tasks, and nested thoughts.

[![CI](https://github.com/JuanCarlosAcostaPeraba/sparks-cli/actions/workflows/test.yml/badge.svg)](https://github.com/JuanCarlosAcostaPeraba/sparks-cli/actions/workflows/test.yml)
[![Latest release](https://img.shields.io/github/v/release/JuanCarlosAcostaPeraba/sparks-cli?display_name=tag&sort=semver)](https://github.com/JuanCarlosAcostaPeraba/sparks-cli/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/JuanCarlosAcostaPeraba/sparks-cli)](https://go.dev/)
[![License](https://img.shields.io/github/license/JuanCarlosAcostaPeraba/sparks-cli)](LICENSE)

[Website](https://sparks-web-nu.vercel.app/) · [Install](#install) · [Quick start](#quick-start) · [Command reference](docs/commands.md) · [Roadmap](docs/roadmap.md)

</div>

`sparks` is built for the moment when opening a full task manager would take longer than capturing the thought. It starts instantly, works without an account, and keeps everything in a local SQLite database.

```text
$ sparks + "Ship the launch page"
Added spark 1
$ sparks + --parent 1 "Polish the copy"
Added spark 2
$ sparks important 1
Marked spark 1 as important
$ sparks tree
└─ [!] 1) Ship the launch page
   └─ [ ] 1.1) Polish the copy
```

## Why sparks?

| | |
|---|---|
| ⚡ **Instant capture** | Add a thought from any terminal in a single command. |
| 🧭 **Keyboard-first TUI** | Browse, search, edit, prioritize, complete, and delete without leaving the interactive view. |
| 🌳 **Nested thoughts** | Turn an idea into a lightweight tree of tasks and sub-ideas. |
| 🔒 **Local by default** | Your data stays in a SQLite file on your machine. No account or cloud service required. |
| 🎨 **Useful visual feedback** | Consistent states and colors make active, important, completed, and selected sparks easy to spot. |
| 🤖 **Script friendly** | JSON output, redirected input, stable status markers, and shell completions fit naturally into terminal workflows. |
| 📦 **Native and cross-platform** | One small Go binary for macOS, Linux, and Windows, with verified self-updates. |

## Install

### macOS and Linux

```bash
curl -fsSL https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.sh | sh
```

The installer selects the correct release for your OS and architecture, verifies its SHA-256 checksum, and installs `sparks` in `~/.local/bin` by default.

### Windows

Run in PowerShell:

```powershell
irm https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.ps1 | iex
```

The installer downloads and verifies `sparks.exe`, installs it under `%LOCALAPPDATA%\Programs\sparks`, and adds that directory to your user `PATH`.

Need a specific version or install directory? See the [installation guide](docs/installation.md). Homebrew, Scoop, and winget packages are planned but are not available yet.

## Quick start

```bash
# Capture a thought
sparks add "Prepare release notes"

# The + alias is even faster
sparks + "Publish the release"

# Create a child spark
sparks + --parent 1 "Add installation notes"

# Prioritize and complete sparks
sparks important 1
sparks done 2

# See the hierarchy
sparks tree
```

Run `sparks` with no arguments to open the full-screen interactive experience.

## The interactive TUI

The TUI keeps the whole workflow under your fingertips. It starts with active sparks only, and lets you search, include completed items, or clear completed items without returning to the shell.

```text
  ███████╗██████╗  █████╗ ██████╗ ██╗  ██╗███████╗
  ██╔════╝██╔══██╗██╔══██╗██╔══██╗██║ ██╔╝██╔════╝
  ███████╗██████╔╝███████║██████╔╝█████╔╝ ███████╗
  ╚════██║██╔═══╝ ██╔══██║██╔══██╗██╔═██╗ ╚════██║
  ███████║██║     ██║  ██║██║  ██║██║  ██╗███████║
  ╚══════╝╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝

  View: active only

  SEL  ID     STATE          TITLE
   >   #1     important      Ship the launch page
       #2     active         Polish the copy

  a add · e edit · i important · c child · d done · x remove
  s search · v active/all · C clear completed · ? help · q quit
```

| Key | Action |
|---|---|
| `↑` / `↓` or `j` / `k` | Move through sparks |
| `a` | Add a spark |
| `s` or `/` | Search by title |
| `v` | Switch between active sparks and the complete list |
| `e` | Edit the selected spark |
| `i` | Toggle important status |
| `c` | Add a child to the selected spark |
| `d` | Mark the selected spark as completed |
| `x` | Remove the selected spark |
| `C` | Clear completed sparks after confirmation |
| `?` | Show keyboard help |
| `q` | Quit |

Status is always represented consistently: `[ ]` active, `[!]` important, and `[x]` completed. Colors add feedback in a terminal but are automatically omitted from JSON and redirected output. Set [`NO_COLOR`](https://no-color.org/) to disable them everywhere.

## Commands at a glance

| Command | What it does | Shortcut |
|---|---|---|
| `sparks add <text>` | Capture a new spark | `sparks +` |
| `sparks list` | List active sparks (`-a` includes completed) | `sparks ls` |
| `sparks edit <id> <text>` | Change a title | `sparks e` |
| `sparks done <id>` | Mark a spark as completed | `sparks ok` |
| `sparks important <id>` | Toggle important status | `sparks !` |
| `sparks remove <id>` | Remove a spark | `sparks rm`, `sparks -` |
| `sparks clear` | Delete completed sparks | — |
| `sparks tree` | Display nested sparks | — |
| `sparks search <query>` | Search titles | — |
| `sparks update` | Install the latest release | — |
| `sparks version` | Print the installed version | — |

Use `sparks <command> --help` for examples and every available flag, or read the full [command reference](docs/commands.md).

## Automation and piping

Use `--json` when another tool needs structured data:

```bash
sparks list --all --json
sparks search "release" --json
```

When input or output is redirected, `sparks` uses its line-based interactive fallback, so scripts can send regular commands followed by `exit` or `quit`.

## Your data stays yours

Sparks are stored in the operating system's application data directory, never beside the executable:

| Platform | Database location |
|---|---|
| macOS | `~/Library/Application Support/sparks/sparks.db` |
| Linux | `$XDG_DATA_HOME/sparks/sparks.db` or `~/.local/share/sparks/sparks.db` |
| Windows | `%APPDATA%\sparks\sparks.db` |

For an isolated workspace or test run, pass `--db /path/to/sparks.db`.

## Stay current

```bash
sparks update
```

The updater shows what it is doing, detects `bash`, `zsh`, `fish`, or another POSIX-compatible shell on macOS and Linux, and uses the official installer. On Windows it safely replaces the executable through PowerShell after the running process exits. Downloads are verified against the checksums published with each GitHub release.

Set `SPARKS_SHELL` if you need to override shell detection on Unix-like systems.

## Build and contribute

You need a current Go toolchain. Then:

```bash
git clone https://github.com/JuanCarlosAcostaPeraba/sparks-cli.git
cd sparks-cli
go test ./...
go vet ./...
go run . --db ./sparks-dev.db
```

Bug reports and focused pull requests are welcome. Browse the [open issues](https://github.com/JuanCarlosAcostaPeraba/sparks-cli/issues) or read the [roadmap](docs/roadmap.md) to see what is next.

Releases are built by GoReleaser from tag-triggered GitHub Actions workflows. Release archives and checksums are available on the [releases page](https://github.com/JuanCarlosAcostaPeraba/sparks-cli/releases).

## License

`sparks` is available under the [MIT License](LICENSE).
