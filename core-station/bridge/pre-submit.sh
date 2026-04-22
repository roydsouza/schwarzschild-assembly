#!/usr/bin/env bash
# core-station/bridge/pre-submit.sh
# Mandatory pre-submission verification for AntiGravity briefings.
# Run from the schwarzschild-assembly project root.

set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

# Ensure all required tools are on PATH for this machine
export PATH="/opt/homebrew/bin:/usr/local/go/bin:$HOME/go/bin:$HOME/.cargo/bin:/usr/local/bin:$PATH"
export CPATH="${CPATH:-}:/opt/homebrew/include:/usr/local/include"
export LIBRARY_PATH="${LIBRARY_PATH:-}:/opt/homebrew/lib:/usr/local/lib"

PASS=0
FAIL=0
report() { local status=$1; shift; echo "[$status] $*"; }
ok()   { report "PASS" "$@"; ((PASS++)) || true; }
fail() { report "FAIL" "$@"; ((FAIL++)) || true; }

echo "============================================================"
echo "  Schwarzschild Assembly: Pre-Submission Verification"
echo "  $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "============================================================"

# ── 1. BUILD ──────────────────────────────────────────────────────
echo ""
echo "── BUILD ──"

if [ -f core-station/aethereum-spine/go.mod ]; then
  if (cd core-station/aethereum-spine && go build ./... 2>&1); then
    ok "governance: go build ./..."
  else
    fail "governance: go build FAILED"
  fi
fi

if [ -f core-station/security/Cargo.toml ]; then
  if (cd core-station/security && cargo build --features tier1 2>&1); then
    ok "security: cargo build --features tier1"
  else
    fail "security: cargo build FAILED"
  fi
fi

for factory_dir in core-station/machinery/*; do
  if [ -f "$factory_dir/go.mod" ]; then
    if (cd "$factory_dir" && go build ./... 2>&1); then
      ok "$(basename "$factory_dir"): go build ./..."
    else
      fail "$(basename "$factory_dir"): go build FAILED"
    fi
  fi
done

# ── 2. STASIS VALIDATION ──────────────────────────────────────────
echo ""
echo "── STASIS VALIDATION ──"

if [ -f core-station/bridge/validate-stasis-tier1.pl ]; then
  if swipl -l core-station/bridge/validate-stasis-tier1.pl -g "main(['core-station/protoplasm/policies/invariants.pl']), halt." 2>&1; then
    ok "stasis: Tier 1 syntactic validation passed"
  else
    fail "stasis: Tier 1 syntactic validation FAILED"
  fi
fi

# ── 3. TESTS ─────────────────────────────────────────────────────
echo ""
echo "── TESTS ──"

if [ -d core-station/protoplasm/tests ]; then
  for test_file in core-station/protoplasm/tests/test_*.pl; do
    if swipl -g "consult('$test_file'), run_tests, halt." 2>&1; then
      ok "prolog: $(basename "$test_file")"
    else
      fail "prolog: $(basename "$test_file") FAILED"
    fi
  done
fi

if [ -f core-station/security/Cargo.toml ]; then
  if (cd core-station/security && cargo test --features tier1 2>&1); then
    ok "security: cargo test"
  else
    fail "security: cargo test FAILED"
  fi
fi

# ── 4. ANATOMY CHECK ──────────────────────────────────────────────
echo ""
echo "── ANATOMY CHECK ──"

for factory_dir in core-station/machinery/*/; do
  if [ ! -f "${factory_dir}README.md" ]; then continue; fi
  factory_name=$(basename "$factory_dir")
  for required in worker domain-fitness mcp-server analyst-briefing README.md; do
    if [ -e "$factory_dir$required" ]; then
      ok "$factory_name: $required exists"
    else
      fail "$factory_name: $required MISSING"
    fi
  done
done

# ── 5. HYGIENE ────────────────────────────────────────────────────
echo ""
echo "── HYGIENE ──"

for go_mod in core-station/aethereum-spine/go.mod core-station/machinery/*/go.mod; do
  mod_dir=$(dirname "$go_mod")
  if (cd "$mod_dir" && go mod tidy && git diff --exit-code go.mod go.sum 2>&1); then
    ok "$mod_dir: hygiene passed"
  else
    fail "$mod_dir: hygiene FAILED (run go mod tidy)"
  fi
done

# ── SUMMARY ───────────────────────────────────────────────────────
echo ""
echo "============================================================"
echo "  PASS: $PASS   FAIL: $FAIL"
echo "============================================================"

if [ "$FAIL" -gt 0 ]; then exit 1; else exit 0; fi
