# Security: The Safety Rail Sector

The Security sector contains the station's formal verification machinery. It enforces the **Schwarzschild Radius**—the event horizon where code is compressed by the **Tier 1 Safety Rail** (Rust/Z3) to ensure mathematical invariants are never violated.

## Sections
- [src/](./src/CONTENTS.md) — The Rust implementation of the Safety Rail and Z3 integration.
- [include/](./include/CONTENTS.md) — C-API headers for the Go/Rust CGO bridge.
- [proofs/](./proofs/CONTENTS.md) — Serialization for cryptographic proof certificates.
- [tests/](./tests/CONTENTS.md) — Formal contract compliance and sandbox verification tests.

## Key Files
- [Cargo.toml](./Cargo.toml) — Rust package definition.
- [Cargo.lock](./Cargo.lock) — Rust dependency lockfile.
- [README.md](./README.md) — Technical manual for the Safety Rail architecture.
