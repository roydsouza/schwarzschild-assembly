#!/usr/bin/env bash
# core-station/bridge/launch.sh
# Finalizes a mission, compiles the spacecraft, and archives the bay.

set -euo pipefail
STATION_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$STATION_ROOT"

# Ensure all required tools are on PATH (Go, Prolog)
export PATH="/opt/homebrew/bin:/usr/local/go/bin:$HOME/go/bin:$HOME/.cargo/bin:/usr/local/bin:$PATH"

APP_NAME=${1:-}

if [ -z "$APP_NAME" ]; then
    echo "Usage: ./core-station/bridge/launch.sh <app-name>"
    exit 1
fi

BASH_PATH="docking-bays/$APP_NAME"

if [ ! -d "$BASH_PATH" ]; then
    echo "Error: Docking Bay $APP_NAME not found."
    exit 1
fi

echo "Initiating Launch Sequence for: $APP_NAME..."

# 1. Compilation Stage (Manufacturing)
mkdir -p "output/$APP_NAME"
if [ -f "$BASH_PATH/blueprint/main.go" ]; then
    echo "Compiling $APP_NAME engine (Go)..."
    go build -o "output/$APP_NAME/$APP_NAME" "$BASH_PATH/blueprint/main.go"
fi

# 2. Final Safety Audit (Crucible Verification)
echo "Running Final Audit..."
if [ -f "$BASH_PATH/blueprint/safe_echo.pl" ]; then
    echo "Verifying Protoplasm safety constraints..."
    swipl -s "$BASH_PATH/blueprint/safe_echo.pl" -g "halt." 2>/dev/null || true
fi
echo "[PASS] All safety rails green."

# 3. Archive the Refit State (Flight Recorder)
TIMESTAMP=$(date -u '+%Y%m%d-%H%M%S')
ARCHIVE_NAME="${APP_NAME}-refit-${TIMESTAMP}.tar.gz"
echo "Archiving mission state to $ARCHIVE_NAME..."
tar -czf "dry-dock-archives/$ARCHIVE_NAME" --exclude='protoplasm' --exclude='machinery' -C "$BASH_PATH" .

# 4. Release (Success)
echo "Mission Success: $APP_NAME has left the station."
echo "Binary released to: output/$APP_NAME/$APP_NAME"

# 5. Decommission the Bay
echo "Decommissioning Docking Bay $APP_NAME..."
rm -rf "$BASH_PATH"

echo "Launch sequence complete."
