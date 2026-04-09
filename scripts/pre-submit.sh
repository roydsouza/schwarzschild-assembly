#!/usr/bin/env bash
# scripts/pre-submit.sh
# Mandatory pre-submission verification for AntiGravity briefings.
# Run from the schwarzschild-assembly project root.
# Copy the COMPLETE output verbatim into the briefing's ## Verification Output section.
# If this script exits non-zero, the briefing cannot be filed.

set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
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
echo "  Sati-Central Pre-Submission Verification"
echo "  $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "============================================================"

# ── 1. BUILD ──────────────────────────────────────────────────────
echo ""
echo "── BUILD ──"

if [ -f root-spine/go.mod ]; then
  if (cd root-spine && go build ./... 2>&1); then
    ok "root-spine: go build ./..."
  else
    fail "root-spine: go build ./... FAILED"
  fi
fi

if [ -f safety-rail/Cargo.toml ]; then
  if (cd safety-rail && cargo build --features tier1 2>&1); then
    ok "safety-rail: cargo build --features tier1"
  else
    fail "safety-rail: cargo build FAILED"
  fi
fi

if [ -f control-panel/package.json ]; then
  if (cd control-panel && npx tsc --noEmit 2>&1); then
    ok "control-panel: tsc --noEmit"
  else
    fail "control-panel: tsc --noEmit FAILED"
  fi
fi

for factory_mod in factories/*/go.mod; do
  factory_dir=$(dirname "$factory_mod")
  if (cd "$factory_dir" && go build ./... 2>&1); then
    ok "$factory_dir: go build ./..."
  else
    fail "$factory_dir: go build FAILED"
  fi
done

# ── 2. TESTS (cumulative — all phases) ───────────────────────────
echo ""
echo "── TESTS (cumulative) ──"

if [ -f root-spine/go.mod ]; then
  if (cd root-spine && go test ./... 2>&1); then
    ok "root-spine: go test ./..."
  else
    fail "root-spine: go test ./... FAILED"
  fi
fi

if [ -f safety-rail/Cargo.toml ]; then
  if (cd safety-rail && cargo test --features tier1 2>&1); then
    ok "safety-rail: cargo test --features tier1"
  else
    fail "safety-rail: cargo test FAILED"
  fi
fi

if [ -f control-panel/package.json ]; then
  if (cd control-panel && npx vitest run 2>&1); then
    ok "control-panel: vitest run"
  else
    fail "control-panel: vitest run FAILED"
  fi
fi

if [ -d agents/prolog-substrate/tests ]; then
  for test_file in agents/prolog-substrate/tests/test_*.pl; do
    if swipl -g "consult('$test_file'), run_tests, halt." 2>&1; then
      ok "prolog: $test_file"
    else
      fail "prolog: $test_file FAILED"
    fi
  done
fi

# ── 3. INTERFACE CONSISTENCY ──────────────────────────────────────
echo ""
echo "── INTERFACE CONSISTENCY ──"

# Metric IDs: every ID declared in domain-fitness/ must appear verbatim in mcp-server/
for metrics_file in factories/*/domain-fitness/metrics.go; do
  factory_dir=$(dirname "$(dirname "$metrics_file")")
  mcp_server="$factory_dir/mcp-server/server.go"
  if [ ! -f "$mcp_server" ]; then continue; fi
  while IFS= read -r metric_id; do
    metric_id="${metric_id//\"/}"
    if grep -q "\"$metric_id\"" "$mcp_server"; then
      ok "metric ID '$metric_id' consistent across $(basename $factory_dir)"
    else
      fail "metric ID '$metric_id' defined in $metrics_file but NOT found in $mcp_server"
    fi
  done < <(perl -ne 'print "$1\n" while /MetricId:\s*"\K([^"]+)/g' "$metrics_file")
done

# Proto types used in TypeScript must exist in generated bindings
if [ -f control-panel/src/types/orchestrator_pb.d.ts ]; then
  while IFS= read -r msg_type; do
    if grep -q "class $msg_type " control-panel/src/types/orchestrator_pb.d.ts; then
      ok "TypeScript proto type '$msg_type' exists in orchestrator_pb.d.ts"
    else
      fail "TypeScript proto type '$msg_type' used but NOT found in orchestrator_pb.d.ts"
    fi
  done < <(grep -rh 'new [A-Z][A-Za-z]+()' control-panel/src/ 2>/dev/null | perl -ne 'print "$1\n" while /new ([A-Z][A-Za-z]+)/g' | sort -u)
fi

# ── 4. ANATOMY CHECK ──────────────────────────────────────────────
echo ""
echo "── ANATOMY CHECK ──"

for factory_dir in factories/*/; do
  factory_name=$(basename "$factory_dir")
  for required in worker domain-fitness mcp-server analyst-briefing README.md; do
    if [ -e "$factory_dir$required" ]; then
      ok "$factory_name: $required exists"
    else
      fail "$factory_name: $required MISSING"
    fi
  done
done

# ── 5. PROLOG SAFETY (if prolog-substrate exists) ─────────────────
if [ -d agents/prolog-substrate ]; then
  echo ""
  echo "── PROLOG SAFETY ──"

  # No bare assertz/retract in production code (only safe_assert/safe_retract allowed).
  # Exception: core/safety_bridge.pl is the ONE authorized mutation point — it calls
  # assertz/retract directly by design (it IS the safe_assert implementation).
  violations=$(set +e; grep -rn 'assertz(\|retract(' agents/prolog-substrate/ \
    --include='*.pl' \
    | grep -v 'safe_assert\|safe_retract\|%\|test_\|_test\.pl\|core/safety_bridge\.pl' \
    | wc -l | tr -d ' '; set -e)
  if [ "$violations" -eq 0 ]; then
    ok "No bare assertz/retract in production Prolog code"
  else
    fail "Found $violations bare assertz/retract calls — use safe_assert/safe_retract"
  fi
fi

# ── 6. HYGIENE ────────────────────────────────────────────────────
echo ""
echo "── HYGIENE ──"

# go mod tidy must produce no diff
for go_mod in root-spine/go.mod factories/*/go.mod; do
  mod_dir=$(dirname "$go_mod")
  if (cd "$mod_dir" && go mod tidy && git diff --exit-code go.mod go.sum 2>&1); then
    ok "$mod_dir: go mod tidy produces no diff"
  else
    fail "$mod_dir: go mod tidy produced a diff — run go mod tidy and commit"
  fi
done

# No committed binaries (files with no extension or known binary extensions)
untracked_binaries=$(git ls-files --others --exclude-standard | \
  grep -vE '\.(go|rs|pl|py|ts|tsx|js|json|sql|yaml|yml|toml|md|sh|proto|css|lock|sum|mod|txt|html|d\.ts|chr)$' | \
  grep -v '\.git' || true)
if [ -z "$untracked_binaries" ]; then
  ok "No untracked binaries"
else
  fail "Untracked files that may be binaries: $untracked_binaries"
fi

# ── 7. BRIEFING PATH REMINDER ─────────────────────────────────────
echo ""
echo "── BRIEFING PATH ──"
echo "  File your briefing to:"
echo "  analyst-inbox/$(date -u '+%Y-%m-%d-%H%M%S')-<topic>.md"
echo "  NOT to any subdirectory."

# ── SUMMARY ───────────────────────────────────────────────────────
echo ""
echo "============================================================"
echo "  PASS: $PASS   FAIL: $FAIL"
echo "============================================================"

if [ "$FAIL" -gt 0 ]; then
  echo "  PRE-SUBMIT FAILED — fix failures before filing briefing"
  exit 1
else
  echo "  PRE-SUBMIT PASSED — copy this output into ## Verification Output"
  exit 0
fi
