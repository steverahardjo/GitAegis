#!/usr/bin/env bash
set -e

CLI_NAME="gitaegis"
BIN_DIR="/usr/local/bin"
URL="https://github.com/steverahardjo/GitAegis/releases/download/v2.0.0/gitaegis-linux-amd64"

echo "Installing ${CLI_NAME}..."

# Download binary
curl -L "$URL" -o "$BIN_DIR/$CLI_NAME"

# Make it executable
chmod +x "$BIN_DIR/$CLI_NAME"

echo "Running initial setup..."
"$BIN_DIR/$CLI_NAME" init --bash

echo "Installation complete! You can now run: $CLI_NAME"