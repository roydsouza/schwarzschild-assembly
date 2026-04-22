# Schwarzschild Station: Operational Flight Manual

This document defines the lifecycle of a **Spacecraft** (Application) within the **Schwarzschild Space Station** (Meta-Factory).

## 1. The Spacecraft Lifecycle: From Vibe to Verification

A Spacecraft is a modular assembly of intelligence and execution. The lifecycle is designed to take a "vibe-coded" concept and compress it through formal verification until it stabilizes as a high-integrity asset within the station's safety protocols.

### Stage 1: Docking (Provisioning)
A mission begins by spawning a new **Docking Bay**. 
- **Tool:** \`./core-station/bridge/dock.sh <ship-name>\`
- **Result:** An ephemeral workspace in \`docking-bays/\` is created.
- **Intelligence:** The shared **Protoplasm** ([STASIS](./STASIS-LANGUAGE.md) substrate) is symlinked, granting the new ship access to the station's collective reasoning.

### Stage 2: Refit (Modular Assembly)
The ship is built or repaired using the **HOTL** (Human On The Loop) process.
- **Forge (Builder):** Implements features and fixes.
- **Crucible (Auditor):** Verifies all code against the **Safety Rail** and **Formal constraints**.
- **Audit Logs:** Every refit decision is recorded in the bay's \`briefings/\` folder.

### Stage 3: Launch (Deployment)
When a ship reaches 0-FAIL status, it is ready to leave the station.
- **Tool:** \`./core-station/bridge/launch.sh <ship-name>\`
- **Verification:** Final Crucible checkout for Z3 policy compliance.
- **Archiving:** The bay's unique blueprints and briefings are compressed into \`dry-dock-archives/\`.
- **Release:** The spacecraft is deployed to its destination (Local, AWS, GCP).

### Stage 4: Retrofit (Repair & Enhancement)
If a ship requires maintenance, it returns to the station.
- **Re-Docking:** The archive is unzipped back into a \`docking-bay/\`.
- **Delta-Briefing:** Forge and Crucible identify the gap between the archived state and the new requirement.
- **Evolution:** The ship is enhanced and re-launched.

## 2. Shared Station Intelligence

### The Protoplasm ([STASIS](./STASIS-LANGUAGE.md))
See **[Spacecraft Architecture](./SPACECRAFTS.md)** for a deep dive into the Shell/Mind dichotomy.
The "Life Support" of the station. It contains the core philosophies, safety heuristics, and architectural patterns shared by all ships. If a ship learns a new safety pattern, it is committed to the Protoplasm for the benefit of future missions.

### The Machinery (Go/Rust/Z3)
The heavy industrial equipment used to manufacture ships. This includes the gRPC orchestrators, Wasm sandboxes, and symbolic solvers that ensure code integrity.

## 3. The "Safe-Echo" Mission (Hello World)

The **Safe-Echo** is the standard diagnostic spacecraft for new station refits. It exercises the following systems:
- **Scaffold Machinery:** Generates a basic CLI tool.
- **STASIS Protoplasm:** Enforces a "Safe String" policy (e.g., banning specific keywords).
- **HOTL Process:** Demonstrates Forge building and Crucible auditing the string-filtering logic.
- **Launch:** Final deployment of the verified binary.

## 4. Operational Findings (2026-04-22)

### Docking Sequence End-to-End Test

Executed full `dock.sh <name>` sequence against `test-probe` on 2026-04-22.
Result: **PASS — 2.782 seconds** (target: under 5 minutes).

Bay contents verified correct:
- `forge/`, `crucible/` — gate harness files copied from `~/antigravity/agents/`
- `crucible-inbox/`, `crucible-verdicts/`, `analyst-inbox/`, `analyst-verdicts/`, `build-artifacts/` — inbox/verdict directories
- `protoplasm` symlink → `core-station/protoplasm`
- `machinery/` symlinks → `core-station/machinery/*`
- `blueprint/README.md` — project scaffold
- `logs/observation.otel` — OTel stub

### Gap Found and Fixed: bootstrap.sh Non-Fatal

**Issue:** `bootstrap.sh` exits non-zero when the OTel collector is not running (attempts to register a service and fails). With `set -euo pipefail` in `dock.sh`, this cascaded into a full docking failure.

**Fix (committed):** `bootstrap.sh` call in `dock.sh` is now wrapped in an `if/else` block:
```bash
if ! ./core-station/bridge/bootstrap.sh "$APP_NAME"; then
    echo "Warning: bootstrap reported issues — review output above, but docking continues."
fi
```
Docking completes and prints a warning. The bay is fully usable even when bootstrap environment setup is partial (e.g., OTel collector offline during development).

**Impact:** Zero — all critical bay setup (harness copy, symlinks, inbox dirs) runs before bootstrap. Bootstrap only performs runtime environment registration (OTel, service discovery) which can be retried independently.

### Remaining Gaps (not blocking)

| Gap | Severity | Tracker |
|:----|:---------|:--------|
| No end-to-end `launch.sh` test verified | Low | fnd-006 deferred (no spacecraft docked yet) |
| `propose_improvement/2` in STASIS Tier 3 is a stub | Low | dsl-005 |
| `stasis-language/` extraction not yet done | Low | Trigger: Shapeshifter Phase 4 |

---
*Reference: [PROCESS.md](./PROCESS.md) for the Human On The Loop Agent Protocol.*
