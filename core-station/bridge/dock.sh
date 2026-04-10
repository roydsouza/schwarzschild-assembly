#!/usr/bin/env bash
# core-station/bridge/dock.sh
# Initializes a new docking bay for an application assembly line.

set -euo pipefail
STATION_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$STATION_ROOT"

APP_NAME=${1:-}

if [ -z "$APP_NAME" ]; then
    echo "Usage: ./core-station/bridge/dock.sh <app-name>"
    exit 1
fi

BASH_PATH="docking-bays/$APP_NAME"

if [ -d "$BASH_PATH" ]; then
    echo "Error: Docking Bay $APP_NAME is already occupied."
    exit 1
fi

echo "Initializing Docking Bay: $APP_NAME..."
mkdir -p "$BASH_PATH"

# 1. Link Core Station Protoplasm (Intelligence)
ln -s "$STATION_ROOT/core-station/protoplasm" "$BASH_PATH/protoplasm"

# 2. Setup Bay-Specific Machinery
mkdir -p "$BASH_PATH/machinery"
for engine in "$STATION_ROOT/core-station/machinery"/*; do
    engine_name=$(basename "$engine")
    ln -s "$engine" "$BASH_PATH/machinery/$engine_name"
done

# 3. Initialize Flight recorder (Briefings)
mkdir -p "$BASH_PATH/briefings"

# 4. Create App Blueprint Skeleton
mkdir -p "$BASH_PATH/blueprint"
cat <<EOM > "$BASH_PATH/blueprint/README.md"
# Spacecraft: $APP_NAME
Status: Docked (Assembly in Progress)
Bay: $APP_NAME
EOM

echo "Docking Bay $APP_NAME is now online."
echo "Location: $BASH_PATH"
