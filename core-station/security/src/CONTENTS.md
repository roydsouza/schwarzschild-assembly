# Safety Rail: Source Implementation

This sector contains the Rust source code for the station's Safety Rail. It implements the formal verification logic and hardware-native bindings for Apple Silicon.

## Sub-Sectors
- [tier1/](./tier1/CONTENTS.md) — Fundamental safety axioms and Z3-based contract enforcement.
- [tier2/](./tier2/CONTENTS.md) — Extended stateful verification and complex policy checking.

## Key Files
- [lib.rs](./lib.rs) — Primary entry point for the Safety Rail library.
- [main.rs](./main.rs) — Entry point for the standalone security auditor.
- [mod.rs](./mod.rs) — Module definitions for security sub-systems.
