#!/usr/bin/env bash
# bootstrap.sh — Sati-Central environment setup
# Target: macOS Tahoe, Apple M5 Pro/Max
# Run from: project root (schwarzschild-assembly/)
#
# Usage: ./scripts/bootstrap.sh [--check-only]
#
# --check-only: verify environment without installing anything

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# ── Colors ────────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

ok()   { echo -e "${GREEN}  ✓${RESET} $*"; }
warn() { echo -e "${YELLOW}  !${RESET} $*"; }
fail() { echo -e "${RED}  ✗${RESET} $*"; }
info() { echo -e "${BLUE}  →${RESET} $*"; }

CHECK_ONLY=false
for arg in "$@"; do
  [[ "$arg" == "--check-only" ]] && CHECK_ONLY=true
done

echo ""
echo "=== Sati-Central Bootstrap ==="
echo "Project root: ${PROJECT_ROOT}"
echo "Mode: $([ "$CHECK_ONLY" = true ] && echo 'check-only' || echo 'install')"
echo ""

ERRORS=0

# ── Helper ────────────────────────────────────────────────────────────────────
require_version() {
  local name="$1"
  local cmd="$2"
  local required="$3"
  local actual
  actual="$($cmd 2>/dev/null || echo "NOT FOUND")"
  if [[ "$actual" == "NOT FOUND" ]]; then
    fail "${name}: not found (required: ${required})"
    ERRORS=$((ERRORS + 1))
    return 1
  else
    ok "${name}: ${actual}"
    return 0
  fi
}

brew_install() {
  local pkg="$1"
  local display="${2:-$1}"
  if [[ "$CHECK_ONLY" == true ]]; then
    fail "${display}: not installed (would install via: brew install ${pkg})"
    ERRORS=$((ERRORS + 1))
    return 1
  fi
  info "Installing ${display} via Homebrew..."
  brew install "$pkg"
  ok "${display}: installed"
}

# ── 1. Core toolchain ─────────────────────────────────────────────────────────
echo "── Xcode License ──"

# Xcode command line tools and license must be accepted for Rust/C compilation
if ! xcode-select -p &>/dev/null; then
  fail "Xcode CLI tools: not installed (run: xcode-select --install)"
  ERRORS=$((ERRORS + 1))
else
  ok "Xcode CLI tools: $(xcode-select -p)"
fi

# Check if Xcode license has been accepted (required for cc/linker)
if ! clang --version &>/dev/null 2>&1; then
  fail "Xcode license: not accepted (run: sudo xcodebuild -license)"
  warn "Until accepted, Rust compilation will fail with 'exit status: 69'"
  ERRORS=$((ERRORS + 1))
else
  ok "Xcode license: accepted (clang available)"
fi

echo ""
echo "── Core Toolchain ──"

require_version "Go" "go version" "1.22+" || true
require_version "Rust" "rustc --version" "1.75+" || true
require_version "Python" "python3 --version" "3.12+" || true
require_version "Node.js" "node --version" "20+" || true
require_version "protoc" "protoc --version" "any" || true
require_version "uv" "uv --version" "any" || true

echo ""

# ── 2. Safety Rail dependencies ───────────────────────────────────────────────
echo "── Safety Rail Dependencies ──"

# z3 SMT solver
if ! command -v z3 &>/dev/null; then
  warn "z3 not found"
  brew_install z3 "Z3 SMT solver" || true
else
  ok "z3: $(z3 --version 2>/dev/null | head -1)"
fi

# wasmtime WASM runtime
if ! command -v wasmtime &>/dev/null; then
  warn "wasmtime not found"
  if [[ "$CHECK_ONLY" == true ]]; then
    fail "wasmtime: not installed (would install via: brew install wasmtime)"
    ERRORS=$((ERRORS + 1))
  else
    info "Installing wasmtime..."
    brew install wasmtime
    ok "wasmtime: $(wasmtime --version 2>/dev/null)"
  fi
else
  ok "wasmtime: $(wasmtime --version 2>/dev/null)"
fi

echo ""

# ── 3. OpenTelemetry Collector ────────────────────────────────────────────────
echo "── OpenTelemetry Collector ──"

if ! command -v otelcol-contrib &>/dev/null; then
  warn "otelcol-contrib not found"
  if [[ "$CHECK_ONLY" == true ]]; then
    fail "otelcol-contrib: not installed (would install via: brew install opentelemetry-collector)"
    ERRORS=$((ERRORS + 1))
  else
    info "Installing OpenTelemetry Collector Contrib..."
    brew install opentelemetry-collector
    ok "otelcol-contrib: $(otelcol-contrib --version 2>/dev/null | head -1)"
  fi
else
  ok "otelcol-contrib: $(otelcol-contrib --version 2>/dev/null | head -1)"
fi

echo ""

# ── 4. PostgreSQL ─────────────────────────────────────────────────────────────
echo "── PostgreSQL ──"

if ! command -v psql &>/dev/null; then
  warn "psql not found"
  if [[ "$CHECK_ONLY" == true ]]; then
    fail "PostgreSQL: not installed (required: 16+)"
    ERRORS=$((ERRORS + 1))
  else
    info "Installing PostgreSQL 16 via Homebrew..."
    brew install postgresql@16
    brew services start postgresql@16
    ok "PostgreSQL 16: installed and started"
  fi
else
  PG_VERSION=$(psql --version | grep -o '[0-9]*\.[0-9]*' | head -1)
  PG_MAJOR=$(echo "$PG_VERSION" | cut -d. -f1)
  if [[ "$PG_MAJOR" -lt 16 ]]; then
    fail "PostgreSQL: found version ${PG_VERSION}, required 16+"
    ERRORS=$((ERRORS + 1))
  else
    ok "PostgreSQL: ${PG_VERSION}"
  fi
fi

# Initialize sati_central database
echo ""
info "Checking sati_central database..."
if psql -lqt 2>/dev/null | cut -d \| -f 1 | grep -qw sati_central; then
  ok "sati_central database: exists"
else
  if [[ "$CHECK_ONLY" == true ]]; then
    warn "sati_central database: does not exist (run without --check-only to create)"
  else
    info "Creating sati_central database..."
    createdb sati_central 2>/dev/null || \
      psql -U postgres -c "CREATE DATABASE sati_central;" 2>/dev/null || \
      warn "Could not create sati_central — ensure PostgreSQL is running and try manually: createdb sati_central"
    ok "sati_central database: created"
  fi
fi

echo ""

# ── 5. Rust toolchain setup ───────────────────────────────────────────────────
echo "── Rust Setup ──"

if command -v rustup &>/dev/null; then
  # Ensure we have the aarch64-apple-darwin target (we're already on it, but be explicit)
  if ! rustup target list --installed 2>/dev/null | grep -q "aarch64-apple-darwin"; then
    info "Adding aarch64-apple-darwin target..."
    rustup target add aarch64-apple-darwin
  fi
  ok "Rust target: aarch64-apple-darwin"

  # Clippy and rustfmt are required
  if ! rustup component list --installed 2>/dev/null | grep -q "clippy"; then
    info "Installing clippy..."
    rustup component add clippy
  fi
  ok "clippy: installed"

  if ! rustup component list --installed 2>/dev/null | grep -q "rustfmt"; then
    info "Installing rustfmt..."
    rustup component add rustfmt
  fi
  ok "rustfmt: installed"
else
  warn "rustup not found — install from https://rustup.rs"
  ERRORS=$((ERRORS + 1))
fi

echo ""

# ── 6. Go tools ───────────────────────────────────────────────────────────────
echo "── Go Tools ──"

# protoc-gen-go and protoc-gen-go-grpc for code generation
if ! command -v protoc-gen-go &>/dev/null; then
  if [[ "$CHECK_ONLY" == true ]]; then
    warn "protoc-gen-go: not installed (would install via: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest)"
  else
    info "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    ok "protoc-gen-go: installed"
  fi
else
  ok "protoc-gen-go: $(protoc-gen-go --version 2>/dev/null)"
fi

if ! command -v protoc-gen-go-grpc &>/dev/null; then
  if [[ "$CHECK_ONLY" == true ]]; then
    warn "protoc-gen-go-grpc: not installed"
  else
    info "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    ok "protoc-gen-go-grpc: installed"
  fi
else
  ok "protoc-gen-go-grpc: $(protoc-gen-go-grpc --version 2>/dev/null || echo 'found')"
fi

echo ""

# ── 7. Python environment ─────────────────────────────────────────────────────
echo "── Python Environment ──"

if command -v uv &>/dev/null; then
  ok "uv: $(uv --version)"
  # dhamma-adviser will create its own venv via uv when implemented
else
  warn "uv not found — install from https://docs.astral.sh/uv/"
  ERRORS=$((ERRORS + 1))
fi

echo ""

# ── 8. OTel collector startup ─────────────────────────────────────────────────
echo "── OTel Collector ──"

cd "${PROJECT_ROOT}"

if pgrep -f "otelcol-contrib" &>/dev/null; then
  ok "OTel collector: already running"
else
  if command -v otelcol-contrib &>/dev/null; then
    if [[ "$CHECK_ONLY" == true ]]; then
      warn "OTel collector: not running (start with: otelcol-contrib --config observability/otel-collector-config.yaml)"
    else
      info "Starting OTel collector in background..."
      mkdir -p otel-snapshots
      otelcol-contrib --config observability/otel-collector-config.yaml \
        > otel-snapshots/collector.log 2>&1 &
      OTEL_PID=$!
      sleep 2
      if kill -0 "$OTEL_PID" 2>/dev/null; then
        ok "OTel collector: started (PID ${OTEL_PID})"
        echo "$OTEL_PID" > otel-snapshots/collector.pid
      else
        fail "OTel collector: failed to start — check otel-snapshots/collector.log"
        ERRORS=$((ERRORS + 1))
      fi
    fi
  else
    warn "OTel collector not installed — skipping startup"
  fi
fi

echo ""

# ── 9. Git submodule registration (manual step) ───────────────────────────────
echo "── Git Registration ──"
warn "MANUAL STEP REQUIRED: Register this repo as a submodule in ~/antigravity"
warn "After creating a GitHub remote, run from ~/antigravity:"
warn "  git submodule add <remote-url> schwarzschild-assembly"
warn "  git submodule update --init --recursive"
warn "  git add .gitmodules schwarzschild-assembly"
warn "  git commit -m 'feat: add schwarzschild-assembly submodule'"
echo ""

# ── Summary ───────────────────────────────────────────────────────────────────
echo "── Summary ──"
if [[ $ERRORS -eq 0 ]]; then
  ok "Environment ready. Run smoke test: ./observability/tests/smoke_test.sh"
else
  fail "${ERRORS} issue(s) found. Address them before proceeding."
  exit 1
fi
