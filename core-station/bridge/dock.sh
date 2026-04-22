#!/usr/bin/env bash
# core-station/bridge/dock.sh
# Initializes a new docking bay for an application assembly line.

set -euo pipefail
STATION_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
AGENTS_ROOT="$(cd "$STATION_ROOT/../agents" && pwd)"
cd "$STATION_ROOT"

APP_NAME=${1:-}

if [ -z "$APP_NAME" ]; then
    echo "Usage: ./core-station/bridge/dock.sh <app-name>"
    exit 1
fi

BAY_PATH="docking-bays/$APP_NAME"

if [ -d "$BAY_PATH" ]; then
    echo "Error: Docking Bay $APP_NAME is already occupied."
    exit 1
fi

echo "Initializing Docking Bay: $APP_NAME..."
mkdir -p "$BAY_PATH"

# 1. Link Core Station Protoplasm (Intelligence)
ln -s "$STATION_ROOT/core-station/protoplasm" "$BAY_PATH/protoplasm"

# 2. Setup Bay-Specific Machinery
mkdir -p "$BAY_PATH/machinery"
if [ -d "$STATION_ROOT/core-station/machinery" ]; then
    for engine in "$STATION_ROOT/core-station/machinery"/*; do
        engine_name=$(basename "$engine")
        ln -s "$engine" "$BAY_PATH/machinery/$engine_name"
    done
fi

# 3. Copy Forge/Crucible gate harness from agents/
# Templates import harness_lib from ~/antigravity/agents — no local copy needed.
mkdir -p "$BAY_PATH/forge" "$BAY_PATH/crucible"
for f in gate.py charter.py protocol.py; do
    cp "$AGENTS_ROOT/forge/$f"    "$BAY_PATH/forge/$f"
    cp "$AGENTS_ROOT/crucible/$f" "$BAY_PATH/crucible/$f"
done
# Stamp PROJECT_NAME in both gate.py copies
sed -i '' "s/PROJECT_NAME = \"unknown-project\"/PROJECT_NAME = \"$APP_NAME\"/" "$BAY_PATH/forge/gate.py"
sed -i '' "s/PROJECT_NAME = \"unknown-project\"/PROJECT_NAME = \"$APP_NAME\"/" "$BAY_PATH/crucible/gate.py"

# 4. Create standard Forge/Crucible inbox and verdict directories
mkdir -p "$BAY_PATH/crucible-inbox"
mkdir -p "$BAY_PATH/crucible-verdicts"
mkdir -p "$BAY_PATH/analyst-inbox"
mkdir -p "$BAY_PATH/analyst-verdicts"
mkdir -p "$BAY_PATH/build-artifacts"

# 5. OTel service registration
export OTEL_SERVICE_NAME="spacecraft.$APP_NAME"
mkdir -p "$BAY_PATH/logs"
touch "$BAY_PATH/logs/observation.otel"

# 6. Create App Blueprint Skeleton
mkdir -p "$BAY_PATH/blueprint"
cat <<EOM > "$BAY_PATH/blueprint/README.md"
# Spacecraft: $APP_NAME
Status: Docked (Assembly in Progress)
Bay: $APP_NAME
EOM

# 7. Automatic Bootstrap (non-fatal — environment issues don't block docking)
if [ -f "core-station/bridge/bootstrap.sh" ]; then
    echo "Bootstrapping bay environment..."
    if ! ./core-station/bridge/bootstrap.sh "$APP_NAME"; then
        echo "Warning: bootstrap reported issues — review output above, but docking continues."
    fi
fi

echo ""
echo "Docking Bay $APP_NAME is now online."
echo "Location: $BAY_PATH"
echo ""
echo "Next steps:"
echo "  1. Define Phase 1 in $BAY_PATH/CLAUDE.md"
echo "  2. python3 $BAY_PATH/forge/gate.py session-start"
echo "  3. python3 $BAY_PATH/forge/gate.py lock phase-1"
echo "  4. Build, then: python3 $BAY_PATH/forge/gate.py pre-submit"
