#!/bin/sh
# mdoc installer — downloads the latest (or a pinned) release binary from
# GitHub and installs it into a bin directory on your PATH.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/hinkolas/mdoc/main/install.sh | sh
#
# Environment overrides:
#   MDOC_VERSION   Tag to install (e.g. v0.1.0). Default: latest release.
#   MDOC_BIN_DIR   Install directory. Default: $HOME/.local/bin.

set -eu

REPO="hinkolas/mdoc"
BIN_NAME="mdoc"
VERSION="${MDOC_VERSION:-latest}"
BIN_DIR="${MDOC_BIN_DIR:-$HOME/.local/bin}"

info() { printf '  %s\n' "$1"; }
err() { printf 'error: %s\n' "$1" >&2; exit 1; }

need() { command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"; }
need uname
need tar
need mktemp

# Pick a downloader.
if command -v curl >/dev/null 2>&1; then
  download() { curl -fsSL "$1" -o "$2"; }
  fetch() { curl -fsSL "$1"; }
elif command -v wget >/dev/null 2>&1; then
  download() { wget -qO "$2" "$1"; }
  fetch() { wget -qO- "$1"; }
else
  err "need curl or wget"
fi

# Detect OS.
os="$(uname -s)"
case "$os" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *) err "unsupported OS: $os (mdoc ships linux and darwin builds)" ;;
esac

# Detect arch (must match goarch values used in .goreleaser.yaml).
arch="$(uname -m)"
case "$arch" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) err "unsupported architecture: $arch" ;;
esac

# Resolve the version tag.
if [ "$VERSION" = "latest" ]; then
  info "resolving latest release..."
  VERSION="$(fetch "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | head -n1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  [ -n "$VERSION" ] || err "could not determine the latest release tag"
fi

# Archive name must match .goreleaser.yaml's name_template:
#   {{ .ProjectName }}_{{ .Os }}_{{ .Arch }}
asset="${BIN_NAME}_${os}_${arch}.tar.gz"
base="https://github.com/${REPO}/releases/download/${VERSION}"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

info "downloading ${asset} (${VERSION})..."
download "${base}/${asset}" "${tmp}/${asset}" \
  || err "download failed: ${base}/${asset}"

# Verify checksum when available (non-fatal if the file is absent).
if download "${base}/checksums.txt" "${tmp}/checksums.txt" 2>/dev/null; then
  info "verifying checksum..."
  if command -v sha256sum >/dev/null 2>&1; then
    (cd "$tmp" && grep " ${asset}\$" checksums.txt | sha256sum -c -) >/dev/null \
      || err "checksum verification failed"
  elif command -v shasum >/dev/null 2>&1; then
    (cd "$tmp" && grep " ${asset}\$" checksums.txt | shasum -a 256 -c -) >/dev/null \
      || err "checksum verification failed"
  else
    info "no sha256 tool found; skipping verification"
  fi
fi

info "extracting..."
tar -xzf "${tmp}/${asset}" -C "$tmp"
[ -f "${tmp}/${BIN_NAME}" ] || err "binary ${BIN_NAME} not found in archive"

mkdir -p "$BIN_DIR"
install -m 0755 "${tmp}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}" 2>/dev/null \
  || { cp "${tmp}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}" && chmod 0755 "${BIN_DIR}/${BIN_NAME}"; }

info "installed ${BIN_NAME} ${VERSION} -> ${BIN_DIR}/${BIN_NAME}"

# PATH hint.
case ":${PATH}:" in
  *":${BIN_DIR}:"*) ;;
  *) info "note: ${BIN_DIR} is not on your PATH — add it to your shell profile" ;;
esac

if [ -t 1 ] && { [ -t 0 ] || [ -r /dev/tty ]; }; then
  info "starting interactive setup..."
  if [ -t 0 ]; then
    "${BIN_DIR}/${BIN_NAME}" install
  else
    "${BIN_DIR}/${BIN_NAME}" install </dev/tty
  fi
else
  info "run 'mdoc install' once to set up Chromium and optional agent skills."
fi
