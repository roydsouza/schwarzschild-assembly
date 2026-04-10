# Schwarzschild Assembly: Claude System Context

## Scope Boundary

**When operating inside `schwarzschild-assembly/`, ignore all other projects in
`~/antigravity/` entirely.** The global `CLAUDE.md` previously described umbra,
darkmatter, penumbra, tachyon_tongs, and other projects — none of those are relevant
here. Your context is this directory and nothing outside it. Do not reference, read,
or act on any sibling project unless Roy explicitly asks you to.

## Synchronization Protocol

Every session must maintain the audit trail locally:
- **Local State**: Update `./SYNC_LOG.md` before every context switch or session end.
- **Global Context**: The `umbra/` legacy persistence layer is decommissioned. All logs are localized to the station boundary.
- **Checkpoints**: Use `git` in this repository root for versioned state capture.

## Build and Test Tools

- **Core Station Health**: `./core-station/bridge/pre-submit.sh`
- **Mission Docking**: `./core-station/bridge/dock.sh <app-name>`
- **Mission Launch**: `./core-station/bridge/launch.sh <app-name>`
- **Safety Rail (Rust)**: `cd core-station/security && cargo build --features tier1`
- **Aethereum-Spine (Go)**: `cd core-station/aethereum-spine && go build ./...`
- **STASIS Substrate (Prolog)**: `swipl -s core-station/protoplasm/tests/test_safe_assert.pl -g "run_tests, halt."`

## Directory Hierarchy

- `core-station/`: Permanent high-integrity infrastructure.
- `docking-bays/`: Ephemeral spacecraft assembly lines (GIT IGNORED).
- `dry-dock-archives/`: Compressed flight data (GIT IGNORED).
- `output/`: Compiled spacecraft binaries (GIT IGNORED).

---

