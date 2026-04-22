# Schwarzschild Assembly: Process

The authoritative Forge/Crucible/Claude Code process is defined in the station-wide
canonical document:

**`~/antigravity/agents/PROCESS.md`**

All agents working in this project must follow that document. It supersedes any prior
local version of PROCESS.md.

## Schwarzschild Assembly–specific additions

The assembly follows the standard process with these additions:

### Additional escalation triggers (always escalate)

- Any artifact that modifies `core-station/security/` (Safety Rail)
- Any artifact that modifies `core-station/protoplasm/policies/` (STASIS Tier 1/2)
- Any artifact that modifies the Merkle log schema in `core-station/protoplasm/core/merkle_bridge.pl`
- Any amendment to phase definitions in this project's `CLAUDE.md`

### Pre-submit script

Run `./core-station/bridge/pre-submit.sh` from the assembly root. This script checks:
- Go build (aethereum-spine)
- Rust build (safety-rail, tier1 feature)
- STASIS Tier 1 syntactic validation
- Prolog tests
- Cargo tests
- Anatomy checks

The GATE-PASS from `python3 forge/gate.py pre-submit` and the full output of
`./core-station/bridge/pre-submit.sh` must both be embedded in every briefing.
