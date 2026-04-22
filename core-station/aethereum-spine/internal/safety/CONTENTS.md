# Aethereum-Spine: Safety Rail Interface

This sector implements the Go-side bridge to the Rust-based Tier 1 Safety Rail. It manages the CGO calls and policy enforcement during mission orchestration.

## Key Files
- [bridge.go](./bridge.go) — The CGO interface for the Safety Rail.
- [enforcer.go](./enforcer.go) — High-level logic for applying safety policies to gRPC streams.
- [CONTENTS.md](./CONTENTS.md) — This atlas file.
