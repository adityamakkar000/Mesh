#!/usr/bin/env bash

set -euo pipefail

REPO="https://github.com/adityamakkar000/mesh"
BIN_NAME="mesh"
INSTALL_DIR="${HOME}/.local/bin"
GO_MIN_VERSION="1.21"  

echo "==> Installing $BIN_NAME CLI"

mkdir -p "$INSTALL_DIR"
echo "Created install directory: $INSTALL_DIR"

if ! command -v go >/dev/null 2>&1; then
  echo "Go not found — installing Go"
  if [[ "$OSTYPE" == "darwin"* ]]; then
    brew install go
  else
    echo "Unsupported OS for automatic Go install"
    exit 1
  fi
fi

echo "Go version: $(go version)"

WORKDIR="$(mktemp -d)"
echo "Cloning repo into $WORKDIR"
git clone "$REPO" "$WORKDIR/mesh"
cd "$WORKDIR/mesh"

echo "Building $BIN_NAME"
go build -trimpath -o "${BIN_NAME}" ./cmd/mesh/main.go

echo "Installing to $INSTALL_DIR"
mv "${BIN_NAME}" "$INSTALL_DIR/"

if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  SHELL_RC="$HOME/.zshrc"
  echo "Adding $INSTALL_DIR to PATH in $SHELL_RC"
  echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"
  echo "⇨ Restart your shell or run: source $SHELL_RC"
fi

echo "==> $BIN_NAME installation complete!"
echo "Run '$BIN_NAME --help' to get started"
