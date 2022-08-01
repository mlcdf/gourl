#!/bin/bash
set -e

EXE=$(ls dist -I "*.gz")
DEST=~/.local/bin/gourl

echo "Installing..."
cp "dist/${EXE}" $DEST
chmod +x $DEST
echo "Installed at ${DEST}"