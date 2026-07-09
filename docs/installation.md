# Installation

## Homebrew

Homebrew distribution is planned. The correct path for an open-source Go CLI is
a formula submitted to Homebrew's `homebrew/core`, subject to Homebrew's
acceptance criteria.

## Manual

Download a release archive for your platform from the [GitHub Releases page](https://github.com/JuanCarlosAcostaPeraba/sparks-cli/releases), extract it, and place `sparks` on your `PATH`.

For Apple Silicon macOS:

```bash
curl -L -o sparks.tar.gz https://github.com/JuanCarlosAcostaPeraba/sparks-cli/releases/download/v0.1.0/sparks_0.1.0_darwin_arm64.tar.gz
tar -xzf sparks.tar.gz
mkdir -p ~/.local/bin
mv sparks ~/.local/bin/sparks
chmod +x ~/.local/bin/sparks
```

For Intel macOS:

```bash
curl -L -o sparks.tar.gz https://github.com/JuanCarlosAcostaPeraba/sparks-cli/releases/download/v0.1.0/sparks_0.1.0_darwin_amd64.tar.gz
tar -xzf sparks.tar.gz
mkdir -p ~/.local/bin
mv sparks ~/.local/bin/sparks
chmod +x ~/.local/bin/sparks
```

Make sure `~/.local/bin` is on your `PATH`.

## Windows

Download the Windows archive from GitHub Releases and add the extracted directory to your `PATH`.
