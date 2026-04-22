# Aethereum-Spine: Master Orchestration Layer

The Aethereum-Spine (formerly root-spine) is the station's central nervous system. It manages gRPC routing, proposal ingestion, and the Merkle Audit Log.

## Sections
- [cmd/](./cmd/CONTENTS.md) — Main entry points for the orchestration server and testing tools.
- [proto/](./proto/CONTENTS.md) — The gRPC service definitions that govern station communication.
- [internal/](./internal/CONTENTS.md) — Core logic implementation (gRPC server, Merkle engine, MCP routing).
- [analyst-inbox/](./analyst-inbox/CONTENTS.md) — Local replica/staging for mission briefings.

## Key Files
- [go.mod](./go.mod) — Go module definition.
- [go.sum](./go.sum) — Go dependency checksums.
- [README.md](./README.md) — Technical manual for the Spine infrastructure.
- [STATUS.md](./STATUS.md) — Operational status and version tracking for the Spine.
- [main](./main) — The compiled Aethereum-Spine binary.
- [sati-central](./sati-central) — Legacy binary symlink.
