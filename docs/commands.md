# Commands

## List

```bash
sparks
sparks list
sparks list -a
sparks list -j
```

## Add

```bash
sparks add "Create GoReleaser config"
sparks + "Create Homebrew tap"
sparks add -p 1 "Create child spark"
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
