# Commands

## Interactive mode

Run `sparks` without a subcommand to open the full-screen interactive table.
Use the arrow keys or `j`/`k` to select a spark. The available actions are:

- `a`: add a root spark
- `e`: edit the selected spark
- `i`: toggle the selected spark's importance
- `c`: add a child under the selected spark
- `d`: complete the selected spark
- `x`: remove the selected spark
- `r`: refresh, `?`: help, `q`: quit

Redirected input or output uses the compatible line prompt, which accepts any
regular command without the leading `sparks` and exits with `exit` or `quit`.

## List

```bash
sparks
sparks list
sparks list -a
sparks list -j
```

## Add

```bash
sparks add Create GoReleaser config
sparks + Create Homebrew tap
sparks add -p 1 "Create child spark"
```

Quotes remain supported but are optional; every positional word after `add` is
joined into the spark title.

## Edit

```bash
sparks edit 3 "Ship v0.2.0"
sparks e 3 "Ship v0.2.0"
```

## Complete

```bash
sparks done 3
sparks ok 3
```

## Important

```bash
sparks important 3
sparks ! 3
```

## Remove

```bash
sparks remove 3
sparks rm 3
sparks - 3
```

## Clear

```bash
sparks clear
sparks clear -a -y
```

## Tree

```bash
sparks tree
sparks tree -j
```

## Search

```bash
sparks search "codex"
sparks search "codex" -j
```

The long forms remain available. Short aliases are `-a` for `--all`, `-j` for
`--json`, `-p` for `--parent`, `-y` for `--yes`, and `-d` for `--db`.

## Version

```bash
sparks version
```

## Update

```bash
sparks update
```

The command downloads the latest platform archive from GitHub Releases,
verifies it against `checksums.txt`, and replaces the current executable.
