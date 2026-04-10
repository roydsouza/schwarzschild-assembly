#!/usr/bin/env bash
# checkpoint.sh — Write a signed checkpoint of the current Merkle tree state.
#
# Called automatically by the aethereum-spine on a scheduled interval, and manually
# before any significant operation. Writes a Signed Tree Head (STH) to
# merkle-log/sth/ and verifies the root integrity.
#
# Usage: ./scripts/checkpoint.sh [--verify-only]
#
# --verify-only: verify last STH without writing a new one
#
# Implementation: stubbed. The aethereum-spine internal/merkle package will
# implement this fully in Phase 3. This script is the shell entry point.
# See proposals/README.md for the proposal that will track Phase 3 work.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "checkpoint.sh: aethereum-spine merkle checkpoint"
echo "Project root: ${PROJECT_ROOT}"
echo ""

if [[ "${1:-}" == "--verify-only" ]]; then
  echo "Mode: verify-only"
  echo "Status: STUB — Phase 3 implementation required"
  echo "The aethereum-spine internal/merkle package will implement this."
  exit 0
fi

echo "Mode: write checkpoint"
echo "Status: STUB — Phase 3 implementation required"
echo "The aethereum-spine internal/merkle package will implement this."
echo ""
echo "When implemented, this script will:"
echo "  1. Call the Orchestrator.ApproveAction RPC with a checkpoint proposal"
echo "  2. Receive a MerkleProof response"
echo "  3. Verify the inclusion path"
echo "  4. Write the STH to merkle-log/sth/YYYY-MM-DD-HHMMSS.json"
echo "  5. Update otel-snapshots/latest.json with audit_integrity metric"
