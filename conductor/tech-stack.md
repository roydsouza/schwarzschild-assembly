# Technology Stack: Schwarzschild Assembly

## 1. Primary Languages
- **Go (1.26):** High-concurrency "Root Spine" and gRPC control plane.
- **Rust (1.75):** Formal mathematical verification (Z3) and sandboxed execution (Wasmtime).
- **TypeScript (5.x):** Frontend "Control Panel" and dashboard interface.

## 2. Core Frameworks & Libraries
- **Orchestration:** gRPC with Protobuf for inter-service communication.
- **Verification:** Z3 SMT solver for Tier 1 safety proofs.
- **Sandboxing:** Wasmtime with WASI for unverified agent code.
- **Frontend:** Next.js (16.2.x) with React (19.2.x).
- **Communication:** Socket.io for real-time status updates from the orchestrator.

## 3. Data & Persistence
- **Audit Substrate:** Merkle Log compliant with RFC-6962 for tamper-proof accountability.
- **Primary Database:** PostgreSQL (via `pgx/v5` in Go) for persistent non-critical state.

## 4. Hardware Optimization
- **M5 Native Substrate:** Architecture optimized for Apple Silicon (M5 Series) hardware:
  - Unified Memory utilization (300+ GB/s bandwidth).
  - Neural Interconnect for low-latency safety inference acceleration.
  - Native NEON, AMX, and SVE instructions.

## 5. Observability
- **OpenTelemetry (OTLP):** Global fitness monitoring (Safety, Ethics, Integrity, Performance, Cost).