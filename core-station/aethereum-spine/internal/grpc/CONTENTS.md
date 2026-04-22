# Aethereum-Spine: gRPC Infrastructure

This sector implements the station's central gRPC framework, handling inter-droid communication and factory control.

## Sections
- [pb/](./pb/CONTENTS.md) — Protocol Buffer generated code and service interfaces.

## Key Files
- [server.go](./server.go) — The master gRPC server implementation.
- [interceptors.go](./interceptors.go) — Logging, tracing, and security middleware.
- [CONTENTS.md](./CONTENTS.md) — This atlas file.
