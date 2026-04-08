#!/usr/bin/env bash
# revert.sh — Revert an approved change by proposal ID.
#
# Every approved change writes a RevertArtifact to proposals/approved/.
# This script takes a proposal ID, locates its RevertArtifact, restores
# the prior state, and appends a Reversion leaf to the Merkle log.
#
# Usage: ./scripts/revert.sh <proposal-id>
#
# Example: ./scripts/revert.sh 2026-04-08-143022-antigravity-tier1-z3-policy
#
# Implementation: stubbed. The root-spine internal/merkle package will
# implement the Merkle append. The revert logic itself is proposal-specific
# and is defined in each RevertArtifact.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if [[ $# -lt 1 ]]; then
  echo "Usage: ./scripts/revert.sh <proposal-id>"
  echo "Example: ./scripts/revert.sh 2026-04-08-143022-antigravity-tier1-z3-policy"
  exit 1
fi

PROPOSAL_ID="$1"
REVERT_ARTIFACT="${PROJECT_ROOT}/proposals/approved/${PROPOSAL_ID}-revert.md"

echo "revert.sh: reverting proposal ${PROPOSAL_ID}"
echo "Project root: ${PROJECT_ROOT}"
echo ""

if [[ ! -f "$REVERT_ARTIFACT" ]]; then
  echo "ERROR: RevertArtifact not found at ${REVERT_ARTIFACT}"
  echo "Only approved proposals can be reverted."
  exit 1
fi

echo "Found RevertArtifact: ${REVERT_ARTIFACT}"
echo "Status: STUB — Phase 3 implementation required"
echo ""
echo "When implemented, this script will:"
echo "  1. Parse the RevertArtifact at proposals/approved/${PROPOSAL_ID}-revert.md"
echo "  2. Execute the revert procedure defined in the artifact"
echo "  3. Submit a Reversion proposal via the Orchestrator gRPC endpoint"
echo "  4. Append a Reversion leaf to the Merkle log"
echo "  5. Update STATUS.md with the reversion record"
echo ""
echo "For now, perform the revert manually and document it in STATUS.md."
