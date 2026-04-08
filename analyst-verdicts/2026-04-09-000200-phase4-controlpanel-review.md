# Analyst Verdict
**Date:** 2026-04-09 00:02:00 UTC
**Artifact:** control-panel/ — Phase 4 Translucent Gate UI
**Verdict:** CONDITIONAL

---

## Verdict Rationale

Phase 4 was executed without an Analyst Droid verdict on Phase 3 — a phase gate
protocol violation. Despite this, the UI work itself has merit and is being evaluated
rather than discarded. The component architecture, visual design, and information
layout are sound. The `TranslucentGate` component correctly gates the Approve button
behind an explicit signature checkbox. The dark utilitarian aesthetic is appropriate.

The UI is currently entirely mocked and disconnected from the Root Spine. This is
acceptable AS LONG AS Phase 3 is completed correctly first — the UI integration is
the final step. The required changes below are integration work, not redesign.

---

## What Was Done Well

- `TranslucentGate`: explicit signature gate before approval is correct behavior
- Component decomposition (Gate, FitnessVector, MerkleInspector, AnalystVerdict) matches
  the component contracts from CLAUDE.md Section 3 Phase 4 exactly
- `DhammaReflection` section renders citations and moral root classification, not decorative prose
- TypeScript types in `src/types/index.ts` match the proto message shapes reasonably well
- Dark high-density layout matches the "operational instrument, not consumer product" requirement

---

## Required Changes

### [BLOCKER-1] Phase gate protocol violated

Phase 4 was done while Phase 3 is VETOED and Phase 2 is CONDITIONAL. The UI cannot
be accepted until Phase 3 is approved. No further UI work should begin until:
1. Phase 2 CONDITIONAL items resolved + Analyst APPROVED verdict issued
2. Phase 3 VETOED items resolved + Analyst APPROVED verdict issued

The existing UI code may remain as-is (do not delete it) — it will be integrated
once the backend is ready.

---

### [REQUIRED-2] `approveProposal` must call `ApproveAction` RPC, not `console.log`

**File:** `control-panel/src/hooks/useOrchestrator.ts` lines 79–85

```typescript
const approveProposal = useCallback(async (id: string, signature: string) => {
    console.log(`[ORCHESTRATOR] Approving proposal ${id}...`);  // ← not acceptable
    setProposals(prev => ...);  // local state only
}, []);
```

Once Phase 3 implements `ApproveAction`, this must call the gRPC-web or REST
endpoint with the `ApprovalRequest` proto message. The signature string must be
passed as `approval_signature`. A console.log is not a production approval path.

---

### [REQUIRED-3] WebSocket endpoint is wrong

**File:** `control-panel/src/hooks/useOrchestrator.ts` line 67

```typescript
const socket = new WebSocket('ws://localhost:50051/ws');
```

Port 50051 is the gRPC port — it does not serve WebSocket connections. Once Phase 3
implements the WebSocket signaling plane in `root-spine/internal/websocket/`, update
this to the correct port (suggest 50052 or 8080 — document in root-spine README and
here).

---

### [REQUIRED-4] `useMock` must be environment-driven, not hardcoded

**File:** `control-panel/src/app/page.tsx` line ~10

```typescript
const { proposals, ... } = useOrchestrator();  // defaults useMock=true
```

Mock mode must be controlled by an environment variable so that production deployments
connect to the real backend without code changes:

```typescript
const useMock = process.env.NEXT_PUBLIC_USE_MOCK === 'true';
const { proposals, ... } = useOrchestrator(useMock);
```

Document `NEXT_PUBLIC_USE_MOCK=true` in a `.env.local.example` file (not committed).

---

### [REQUIRED-5] `DhammaReflection` component missing

**File:** `control-panel/src/components/DhammaReflection/` — not created

CLAUDE.md Section 3 Phase 4 requires a `DhammaReflection` component that is
standalone (not just a section inside `TranslucentGate`). The Dhamma section in
`TranslucentGate.tsx` is inline JSX, not a separate component. Extract it:

```
control-panel/src/components/DhammaReflection/
├── DhammaReflection.tsx
└── DhammaReflection.module.css
```

The component receives `MoralWeighting` (matching the proto message shape) and renders
`score`, `root` (kusala/akusala/neutral), `citations` as clickable Bilara segment IDs,
and `reasoning`. Citations must link to `https://suttacentral.net/<segment_id>` — they
are evidence references, not decorative text.

---

### [SIGNIFICANT-6] No tests

**File:** `control-panel/` — zero test files

Anti-Slop Rule 3 applies to UI component contracts too. Minimum required:

- Unit test for `TranslucentGate`: verify that the Approve button is disabled when
  `signatureChecked = false`, disabled when `analystVerdict.status = 'VETOED'`, and
  enabled only when both conditions are satisfied.
- Unit test for `FitnessVector`: verify that metrics above threshold render `red` status.
- Unit test for `useOrchestrator`: verify `approveProposal` transitions proposal to
  `COMMITTED` state.

Use `@testing-library/react` + `vitest` or `jest`.

---

### [SIGNIFICANT-7] `AnalystVerdict` component does not render `CONDITIONAL` state

**File:** `control-panel/src/components/AnalystVerdict/AnalystVerdict.tsx`

The component handles `APPROVED` and `VETOED` but has no case for `CONDITIONAL`
(which is a valid `VerdictDecision` in the proto). Add a distinct visual treatment for
`CONDITIONAL` (amber/warning color, showing required changes) so the human operator
can see that an artifact was approved with conditions.

---

## Integration Sequence (after Phase 3 approval)

1. Point `useOrchestrator` WebSocket URL at the Phase 3 WebSocket port
2. Replace `MOCK_PROPOSALS` with real proposals from the WebSocket stream
3. Wire `approveProposal` to `ApproveAction` RPC
4. Wire `denyProposal` to `VetoAction` RPC
5. Wire `FitnessVector` to `GetFitnessSnapshot` RPC (poll every 5s initially)
6. Wire `MerkleInspector` to Merkle inclusion proof endpoint
7. Set `NEXT_PUBLIC_USE_MOCK=false` in CI

---

## Fitness Vector Impact Assessment

| Metric | Impact | Notes |
|--------|--------|-------|
| Safety compliance | Neutral (now) | Approve button gated behind signature checkbox — correct. Will be positive once connected to real backend. |
| Audit integrity | Neutral | No Merkle interaction yet. |
| Dhamma alignment | Neutral | Dhamma display present but read-only mock data. |
| System performance | Not measurable | No real data flow. |
| Operational cost | Neutral | Static UI, no LLM calls. |

---

## Conditions for Final Acceptance

- [ ] BLOCKER-1: Phase 2 APPROVED + Phase 3 APPROVED first
- [ ] REQUIRED-2: `approveProposal` calls real `ApproveAction` RPC
- [ ] REQUIRED-3: WebSocket URL updated to correct port
- [ ] REQUIRED-4: `useMock` driven by environment variable
- [ ] REQUIRED-5: `DhammaReflection` extracted as standalone component with Bilara links
- [ ] SIGNIFICANT-6: Tests for TranslucentGate, FitnessVector, useOrchestrator
- [ ] SIGNIFICANT-7: CONDITIONAL verdict state rendered

---

## Merkle Log Entry

```json
{
  "event_type": "VetoIssued",
  "agent_id": "claude-code",
  "payload_hash": "<SHA-256 of this verdict file>",
  "safety_cert": "<PolicyFingerprint.empty()>",
  "dhamma_ref": null,
  "fitness_delta": null,
  "model_version": "claude-sonnet-4-6"
}
```
