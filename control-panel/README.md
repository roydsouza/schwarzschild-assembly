# Control Panel (Translucent Gate)

React/Next.js human safety interface. The operator's primary tool for approving or
denying security-adjacent proposals. Not cosmetic — it is the trust boundary between
the automated pipeline and human judgment.

## What it does

- Renders pending `ActionProposal` instances with safety verdicts, Dhamma reflections,
  and fitness impact assessments
- Presents Approve/Deny controls requiring an explicit `ApprovalSignature`
- Shows live fitness vector dashboard (green/amber/red per metric)
- Visualizes the Merkle audit tree with inclusion/consistency proof support
- Displays Analyst Droid verdicts inline before human decision

## Depends on

- Root Spine (gRPC → WebSocket bridge, Phase 3)
- Node.js 22+ with npm

## How to run tests

```bash
cd control-panel && npm test
```

## How to start dev server

```bash
cd control-panel && npm run dev
# Access at http://localhost:3000
```

## Design constraints

- Dark theme. High information density. No animations on critical decision paths.
- No `any` in TypeScript.
- WebSocket to Root Spine uses WebTransport where available, HTTP/1.1 upgrade as fallback.
- Approve action requires explicit `ApprovalSignature` — no accidental approvals.

## Implementation status

Phase 4 — not yet implemented. Implemented after Phase 3 (Root Spine) is approved.
