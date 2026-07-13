#!/usr/bin/env sh
set -eu

repo="JuanCarlosAcostaPeraba/sparks-cli"
version="${SPARKS_VERSION:-latest}"
install_dir="${SPARKS_INSTALL_DIR:-"$HOME/.local/bin"}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "sparks installer: missing required command: $1" >&2
    exit 1
  fi
}

need curl
need tar
need awk

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$os" in
  darwin|linux) ;;
  *)
    echo "sparks installer: unsupported OS: $os" >&2
    exit 1
    ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "sparks installer: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

if [ "$version" = "latest" ]; then
  version="$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" |
    sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' |
    head -n 1)"
fi

if [ -z "$version" ]; then
  echo "sparks installer: could not resolve the latest release" >&2
  exit 1
fi

tag="$version"
case "$tag" in
  v*) ;;
  *) tag="v$tag" ;;
esac

release_version="${tag#v}"
asset="sparks_${release_version}_${os}_${arch}.tar.gz"
url="https://github.com/$repo/releases/download/$tag/$asset"
checksums_url="https://github.com/$repo/releases/download/$tag/checksums.txt"
tmp_dir="$(mktemp -d)"

cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

echo "Installing sparks $tag for $os/$arch..."
curl -fL --retry 3 "$url" -o "$tmp_dir/$asset"
curl -fsSL --retry 3 "$checksums_url" -o "$tmp_dir/checksums.txt"

expected="$(awk -v asset="$asset" '$2 == asset || $2 == "*" asset { print $1; exit }' "$tmp_dir/checksums.txt")"
if [ -z "$expected" ]; then
  echo "sparks installer: checksum for $asset was not published" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp_dir/$asset" | awk '{ print $1 }')"
elif command -v shasum >/dev/null 2>&1; then
  actual="$(shasum -a 256 "$tmp_dir/$asset" | awk '{ print $1 }')"
else
  echo "sparks installer: missing SHA-256 tool (sha256sum or shasum)" >&2
  exit 1
fi

if [ "$actual" != "$expected" ]; then
  echo "sparks installer: checksum mismatch for $asset" >&2
  exit 1
fi

echo "Checksum verified."
tar -xzf "$tmp_dir/$asset" -C "$tmp_dir"

mkdir -p "$install_dir"
install "$tmp_dir/sparks" "$install_dir/sparks"

echo "sparks installed to $install_dir/sparks"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "Add $install_dir to your PATH to run sparks from any terminal." ;;
esac
