# Command Bridge: Strategic Operations

The Bridge contains the primary shell scripts used to command the station's factories and manage spacecraft lifecycles.

## Tactical Tools
- [dock.sh](./dock.sh) — Provision a new Docking Bay and mission workspace.
- [launch.sh](./launch.sh) — Perform final verification and deploy a completed spacecraft.
- [pre-submit.sh](./pre-submit.sh) — The mandatory 0-FAIL validation script required for filing briefings.
- [checkpoint.sh](./checkpoint.sh) — Synchronize station state to the central umbra.
- [bootstrap.sh](./bootstrap.sh) — Install station dependencies and toolchains (Go, Rust, SWI-Prolog).
- [revert.sh](./revert.sh) — Roll back a mission to a previous Merkle leaf.
- [mcp-client.sh](./mcp-client.sh) — Interactive client for the Model Context Protocol grid.

## Specialized Tools
- [validate-stasis-tier1.pl](./validate-stasis-tier1.pl) — Perl-based verification for Tier 1 STASIS axioms.
