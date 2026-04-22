#!/usr/bin/env bash
# core-station/bridge/retrofit.sh
# Restores a spacecraft from an archive back to a docking bay.
# Phase 11 Distribution & Scaling.

set -euo pipefail
STATION_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$STATION_ROOT"

ARCHIVE=$1
NEW_NAME=${2:-}

if [ -z "$ARCHIVE" ]; then
    echo "Usage: ./core-station/bridge/retrofit.sh <archive-path> [<new-app-name>]"
    exit 1
fi

if [ ! -f "$ARCHIVE" ]; then
    echo "Error: Archive $ARCHIVE not found."
    exit 1
fi

# Extract app name from archive if not provided
if [ -z "$NEW_NAME" ]; then
    # Assumes filename format: app-name-timestamp.tar.gz
    NEW_NAME=$(basename "$ARCHIVE" | sed -E 's/-[0-9]{12}\.tar\.gz$//')
fi

BASH_PATH="docking-bays/$NEW_NAME"

if [ -d "$BASH_PATH" ]; then
    echo "Error: Docking Bay $NEW_NAME is already occupied."
    exit 1
fi

echo "Retrofitting Spacecraft to Bay: $NEW_NAME..."
# Do NOT mkdir -p "$BASH_PATH" here, otherwise mv will nest the directory.
tar -xzf "$ARCHIVE" -C "docking-bays/"

# Since the archive contains the original name, we might need to rename it
# Assumes the archive contains the directory as the first entry
ORIGINAL_NAME=$(tar -tzf "$ARCHIVE" | head -n 1 | cut -f1 -d"/")
if [ "$ORIGINAL_NAME" != "$NEW_NAME" ]; then
    if [ -d "$BASH_PATH" ]; then
        rm -rf "$BASH_PATH"
    fi
    mv "docking-bays/$ORIGINAL_NAME" "$BASH_PATH"
fi

echo "Re-docking completed. Refreshing symlinks..."
# Refresh symlinks to ensure they point to the current station root (in case of transfer)
rm -f "$BASH_PATH/protoplasm"
ln -s "$STATION_ROOT/core-station/protoplasm" "$BASH_PATH/protoplasm"

rm -rf "$BASH_PATH/machinery"
mkdir -p "$BASH_PATH/machinery"
if [ -d "$STATION_ROOT/core-station/machinery" ]; then
    for engine in "$STATION_ROOT/core-station/machinery"/*; do
        engine_name=$(basename "$engine")
        ln -s "$engine" "$BASH_PATH/machinery/$engine_name"
    done
fi

echo "Retrofit Successful. Bay $NEW_NAME is ready for refit."
