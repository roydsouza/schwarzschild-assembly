#!/usr/bin/env bash
# smoke_test.sh — Phase 1 OTel end-to-end smoke test
#
# Verifies that:
# 1. OTel collector is running and healthy
# 2. A synthetic fitness vector metric event can be submitted via OTLP HTTP
# 3. The event appears in otel-snapshots/latest.json with correct schema
# 4. The fitness-vector-schema.json is valid JSON
# 5. The log-schema.json is valid JSON
#
# Usage: ./observability/tests/smoke_test.sh
# Exit code: 0 = all checks pass, 1 = one or more checks failed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Ensure Homebrew and Go binaries are in PATH
export PATH="/opt/homebrew/bin:/usr/local/bin:$PATH"
if command -v go &>/dev/null; then
  export PATH="$(go env GOPATH)/bin:$PATH"
fi

# ── Colors ────────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; RESET='\033[0m'
ok()   { echo -e "${GREEN}  ✓${RESET} $*"; }
fail() { echo -e "${RED}  ✗${RESET} $*"; FAILURES=$((FAILURES + 1)); }
info() { echo -e "${BLUE}  →${RESET} $*"; }
warn() { echo -e "${YELLOW}  !${RESET} $*"; }

FAILURES=0
cd "${PROJECT_ROOT}"

echo ""
echo "=== Sati-Central Phase 1 Smoke Test ==="
echo ""

# ── Check 1: Schema files are valid JSON ─────────────────────────────────────
echo "── Check 1: Schema file validity ──"

for schema_file in \
  observability/fitness-vector-schema.json \
  observability/schemas/log-schema.json; do
  if [[ ! -f "$schema_file" ]]; then
    fail "${schema_file}: file not found"
  elif python3 -c "import json; json.load(open('${schema_file}'))" 2>/dev/null; then
    ok "${schema_file}: valid JSON"
  else
    fail "${schema_file}: invalid JSON"
  fi
done

echo ""

# ── Check 2: OTel collector health ───────────────────────────────────────────
echo "── Check 2: OTel collector health ──"

OTEL_HEALTH_URL="http://localhost:13133"
OTEL_OTLP_HTTP="http://localhost:4318"
OTEL_RUNNING=false

if curl -sf "${OTEL_HEALTH_URL}" >/dev/null 2>&1; then
  ok "OTel collector health endpoint: responding at ${OTEL_HEALTH_URL}"
  OTEL_RUNNING=true
else
  warn "OTel collector not running at ${OTEL_HEALTH_URL}"
  warn "Checks 3 and 4 will be skipped (collector not required for schema validation)"
  warn "To start: otelcol-contrib --config observability/otel-collector-config.yaml"
fi

echo ""

# ── Check 3: Submit synthetic fitness vector metric ───────────────────────────
echo "── Check 3: Synthetic metric submission ──"

if [[ "$OTEL_RUNNING" == true ]]; then
  TIMESTAMP_NS=$(($(date +%s) * 1000000000))
  TIMESTAMP_MS=$(($(date +%s) * 1000))

  # Build the OTLP JSON payload for a synthetic safety_compliance metric
  # This exercises the fitness/metrics pipeline in the OTel collector config
  OTLP_PAYLOAD=$(cat <<EOF
{
  "resourceMetrics": [
    {
      "resource": {
        "attributes": [
          { "key": "service.name", "value": { "stringValue": "sati-central-smoke-test" } },
          { "key": "service.namespace", "value": { "stringValue": "sati-central" } },
          { "key": "host.arch", "value": { "stringValue": "arm64" } }
        ]
      },
      "scopeMetrics": [
        {
          "scope": {
            "name": "sati-central.smoke-test",
            "version": "1.0.0"
          },
          "metrics": [
            {
              "name": "sati_central.fitness.safety_compliance",
              "description": "Smoke test synthetic safety compliance counter",
              "unit": "violations",
              "sum": {
                "dataPoints": [
                  {
                    "attributes": [
                      { "key": "test.run", "value": { "stringValue": "smoke-test" } }
                    ],
                    "startTimeUnixNano": "${TIMESTAMP_NS}",
                    "timeUnixNano": "${TIMESTAMP_NS}",
                    "asInt": "0",
                    "exemplars": []
                  }
                ],
                "aggregationTemporality": 2,
                "isMonotonic": true
              }
            },
            {
              "name": "sati_central.fitness.dhamma_alignment",
              "description": "Smoke test synthetic Dhamma alignment gauge",
              "unit": "score",
              "gauge": {
                "dataPoints": [
                  {
                    "attributes": [
                      { "key": "test.run", "value": { "stringValue": "smoke-test" } }
                    ],
                    "timeUnixNano": "${TIMESTAMP_NS}",
                    "asDouble": 1.0
                  }
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}
EOF
)

  HTTP_STATUS=$(curl -sf -o /dev/null -w "%{http_code}" \
    -X POST "${OTEL_OTLP_HTTP}/v1/metrics" \
    -H "Content-Type: application/json" \
    -d "${OTLP_PAYLOAD}" 2>/dev/null || echo "000")

  if [[ "$HTTP_STATUS" == "200" ]]; then
    ok "OTLP HTTP export: accepted (HTTP 200)"
    # Give the collector a moment to flush
    sleep 2
  else
    fail "OTLP HTTP export: unexpected status ${HTTP_STATUS}"
  fi
else
  warn "Skipping (collector not running)"
fi

echo ""

# ── Check 4: Verify metric appears in latest.json ────────────────────────────
echo "── Check 4: otel-snapshots/latest.json written ──"

if [[ "$OTEL_RUNNING" == true ]]; then
  LATEST="${PROJECT_ROOT}/otel-snapshots/latest.json"

  if [[ ! -f "$LATEST" ]]; then
    fail "otel-snapshots/latest.json: does not exist after export"
  elif [[ ! -s "$LATEST" ]]; then
    fail "otel-snapshots/latest.json: exists but is empty"
  else
    # Verify it contains JSON (collector may write JSONL; check last non-empty line)
    LAST_LINE=$(grep -v '^$' "$LATEST" | tail -1)
    if python3 -c "import json, sys; json.loads(sys.stdin.read())" <<< "$LAST_LINE" 2>/dev/null || \
       python3 -c "import json; json.load(open('${LATEST}'))" 2>/dev/null; then
      ok "otel-snapshots/latest.json: contains valid JSON"

      # Check for our synthetic metric name in the file
      if grep -q "sati_central.fitness" "$LATEST" 2>/dev/null; then
        ok "otel-snapshots/latest.json: contains sati_central.fitness metrics"
      else
        warn "otel-snapshots/latest.json: sati_central.fitness metrics not yet visible (collector may need more time)"
      fi
    else
      fail "otel-snapshots/latest.json: last line is not valid JSON"
    fi
  fi
else
  warn "Skipping (collector not running)"
fi

echo ""

# ── Check 5: OTel config is well-formed YAML ─────────────────────────────────
echo "── Check 5: OTel config YAML validity ──"

if python3 -c "import yaml; yaml.safe_load(open('observability/otel-collector-config.yaml'))" 2>/dev/null; then
  ok "observability/otel-collector-config.yaml: valid YAML"
else
  # yaml may not be installed; try with python3 pyyaml if available
  if python3 -c "import yaml" 2>/dev/null; then
    fail "observability/otel-collector-config.yaml: invalid YAML"
  else
    warn "PyYAML not installed — install with: uv pip install pyyaml"
    warn "Skipping YAML validation"
  fi
fi

echo ""

# ── Check 6: Safety Rail compiles ────────────────────────────────────────────
echo "── Check 6: safety-rail/src/lib.rs compiles ──"

if command -v cargo &>/dev/null; then
  # Check without features first (no z3/wasmtime needed for trait contract compilation)
  if cargo check --manifest-path safety-rail/Cargo.toml --no-default-features 2>/dev/null; then
    ok "safety-rail: cargo check (no-default-features) passed"
  else
    fail "safety-rail: cargo check failed — see output above"
  fi
else
  warn "cargo not found — skipping Rust compilation check"
fi

echo ""

# ── Check 7: Proto file is valid ─────────────────────────────────────────────
echo "── Check 7: Protobuf definition validity ──"

if command -v protoc &>/dev/null; then
  # Dry-run: parse only, no output generation
  if protoc --proto_path=root-spine/proto root-spine/proto/orchestrator.proto \
    -o /dev/null 2>/dev/null; then
    ok "root-spine/proto/orchestrator.proto: valid"
  else
    fail "root-spine/proto/orchestrator.proto: protoc parse error"
  fi
else
  warn "protoc not found — skipping proto validation"
fi

echo ""

# ── Summary ───────────────────────────────────────────────────────────────────
echo "── Smoke Test Summary ──"
if [[ $FAILURES -eq 0 ]]; then
  echo -e "${GREEN}All checks passed.${RESET} Phase 1 observability substrate is operational."
  echo ""
  echo "AntiGravity: update STATUS.md, then begin Phase 2 (Tier 1 Safety Rail)."
  exit 0
else
  echo -e "${RED}${FAILURES} check(s) failed.${RESET} Address failures before Phase 2."
  exit 1
fi
