> **DEPRECATED (2026-04-30):** This assembly has been superseded by `dark-factory/`.
> The factory core is now at `dark-factory/factory/`. The process law is at `dark-factory/CONTEXT.md`.
> Do not do new work in schwarzschild-assembly — redirect to `dark-factory/assembly-lines/`.

# Schwarzschild Assembly: Claude System Context

## Scope Boundary

**When operating inside `schwarzschild-assembly/`, your context is this directory and
nothing outside it.** Do not read, reference, or act on sibling projects unless Roy
explicitly asks you to.

---

## Process

The authoritative Forge/Crucible process is at `~/antigravity/agents/PROCESS.md`.
Read it on every session open if you haven't already.

**Assembly-specific escalation triggers (always escalate to Claude Code):**
- Any artifact modifying `core-station/security/` (Safety Rail, Z3 policy)
- Any artifact modifying `core-station/protoplasm/policies/` (STASIS Tier 1/2)
- Any artifact modifying `core-station/protoplasm/core/merkle_bridge.pl`
- Any amendment to phase definitions in this file
- Any change to `dock.sh`, `launch.sh`, or `pre-submit.sh`

**Routine work (forward to Crucible):**
- New spacecraft docking bay work
- Go/Rust build fixes
- Prolog test additions and fixes
- Documentation updates
- Anatomy additions to existing docking bays

---

## Build and Verification

Run all from the `schwarzschild-assembly/` root:

| Command | What it checks |
|:--------|:---------------|
| `./core-station/bridge/pre-submit.sh` | Full: Go build, Rust build, STASIS validation, Prolog tests, Cargo tests, anatomy |
| `./core-station/bridge/dock.sh <name>` | New spacecraft docking |
| `./core-station/bridge/launch.sh <name>` | Spacecraft launch (final Z3 check + archive) |
| `swipl -s core-station/protoplasm/tests/test_safe_assert.pl -g "run_tests, halt."` | STASIS safe_assert unit tests |
| `cd core-station/security && LIBRARY_PATH=/opt/homebrew/lib CPATH=/opt/homebrew/include cargo test --features tier1` | Z3 policy tests |
| `cd core-station/aethereum-spine && go build ./...` | Spine build only |

**The pre-submit script must exit 0 (FAIL: 0) before any briefing is filed.**

---

## Directory Hierarchy

| Path | Role |
|:-----|:-----|
| `core-station/` | Permanent high-integrity infrastructure — do not move or delete |
| `core-station/security/` | Rust Safety Rail (Z3 + Wasmtime) — always escalate changes |
| `core-station/protoplasm/` | STASIS Prolog substrate — policies/ and core/ always escalate |
| `core-station/aethereum-spine/` | Go orchestration spine |
| `core-station/machinery/` | Per-factory toolchain modules |
| `core-station/bridge/` | Lifecycle scripts (dock, launch, pre-submit) — always escalate |
| `docking-bays/` | Ephemeral spacecraft assembly lines (git-ignored) |
| `dry-dock-archives/` | Compressed completed spacecraft data (git-ignored) |
| `analyst-inbox/` | Briefings escalated to Claude Code (Roy routes here) |
| `analyst-verdicts/` | Claude Code audit verdicts |
| `proposals/` | Phase amendment proposals and pending changes |
| `merkle-log/` | Append-only audit log (never delete entries) |

---

## Phase Completion Checklist

Before calling any phase COMPLETE in a briefing:

- [ ] `pre-submit.sh` exits 0 with FAIL: 0
- [ ] All new Prolog tests pass: `swipl -g "run_tests, halt." <test_file>`
- [ ] All Rust tests pass: `cargo test --features tier1`
- [ ] Go builds clean: `go build ./...`
- [ ] STASIS Tier 1 validation passes: `swipl -l core-station/bridge/validate-stasis-tier1.pl ...`
- [ ] Anatomy check passes for all affected factory modules
- [ ] Briefing contains full verbatim `pre-submit.sh` output in `## Verification Output`
- [ ] Briefing contains GATE-PASS block from `python3 forge/gate.py pre-submit`

---

## Current State (as of 2026-04-22)

The factory is operational. Foundation cleanup (root `TASKS.md` fnd-001 through fnd-008)
is complete. No spacecraft is currently docked.

**What's ready:**
- STASIS three-tier architecture (Tier 1 linter, Tier 2 CHR, Tier 3 meta)
- Z3 Safety Rail (5 constraints including `stasis_tier2_calls_tier1_only`)
- Forge/Crucible harness (imports from `~/antigravity/agents/`)
- `dock.sh` auto-copies gate templates and creates standard inbox dirs
- `pre-submit.sh` covers Go, Rust, Prolog, anatomy, hygiene

**What's not yet ready:**
- No spacecraft has been docked and launched end-to-end (docking time unverified)
- `stasis-language/` extraction deferred (trigger: Shapeshifter Phase 4)
- STASIS generative mutation (`propose_improvement/2` is a stub — dsl-005)

---

## Synchronization

Update `./SYNC_LOG.md` at every context switch or session end.
Use `git` in this repository root for versioned state capture.
