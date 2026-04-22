# Aethereum-Spine: Internal Logic

The Internal sector contains the core logic implementations for the master orchestration layer, including communication, security, and persistence.

## Modules
- [grpc/](./grpc/CONTENTS.md) — Implementation of the station's gRPC services.
- [merkle/](./merkle/CONTENTS.md) — Implementation of the cryptographically secured mission engine.
- [mcp/](./mcp/CONTENTS.md) — Implementation of the Model Context Protocol grid.
- [persistence/](./persistence/CONTENTS.md) — Database and state management logic.
- [safety/](./safety/CONTENTS.md) — Interface logic for the Rust-based Safety Rail.
- [orchestrator/](./orchestrator/CONTENTS.md) — Master mission coordination logic.
- [gate/](./gate/CONTENTS.md) — Authentication and authorization infrastructure.
- [websocket/](./websocket/CONTENTS.md) — Real-time telemetry and command streams.
