# Schwarzschild Station: Operational Flight Manual

This document defines the lifecycle of a **Spacecraft** (Application) within the **Schwarzschild Space Station** (Meta-Factory).

## 1. The Spacecraft Lifecycle

A Spacecraft is a modular assembly of intelligence and execution, bound by the station's safety protocols.

### Stage 1: Docking (Provisioning)
A mission begins by spawning a new **Docking Bay**. 
- **Tool:** `./core-station/bridge/dock.sh <ship-name>`
- **Result:** An ephemeral workspace in `docking-bays/` is created.
- **Intelligence:** The shared **Protoplasm** (Prolog Substrate) is symlinked, granting the new ship access to the station's collective reasoning.

### Stage 2: Refit (Modular Assembly)
The ship is built or repaired using the **HOTL** (High-Order Test Loop) process.
- **Forge (Builder):** Implements features and fixes.
- **Crucible (Auditor):** Verifies all code against the **Safety Rail** and **Formal constraints**.
- **Audit Logs:** Every refit decision is recorded in the bay's `briefings/` folder.

### Stage 3: Launch (Deployment)
When a ship reaches 0-FAIL status, it is ready to leave the station.
- **Tool:** `./core-station/bridge/launch.sh <ship-name>`
- **Verification:** Final Crucible checkout for Z3 policy compliance.
- **Archiving:** The bay's unique blueprints and briefings are compressed into `dry-dock-archives/`.
- **Release:** The spacecraft is deployed to its destination (Local, AWS, GCP).

### Stage 4: Retrofit (Repair & Enhancement)
If a ship requires maintenance, it returns to the station.
- **Re-Docking:** The archive is unzipped back into a `docking-bay/`.
- **Delta-Briefing:** Forge and Crucible identify the gap between the archived state and the new requirement.
- **Evolution:** The ship is enhanced and re-launched.

## 2. Shared Station Intelligence

### The Protoplasm (Prolog)
See **[Spacecraft Architecture](./SPACECRAFTS.md)** for a deep dive into the Shell/Mind dichotomy.
The "Life Support" of the station. It contains the core philosophies, safety heuristics, and architectural patterns shared by all ships. If a ship learns a new safety pattern, it is committed to the Protoplasm for the benefit of future missions.

### The Machinery (Go/Rust/Z3)
The heavy industrial equipment used to manufacture ships. This includes the gRPC orchestrators, Wasm sandboxes, and symbolic solvers that ensure code integrity.

## 3. The "Safe-Echo" Mission (Hello World)

The **Safe-Echo** is the standard diagnostic spacecraft for new station refits. It exercises the following systems:
- **Scaffold Machinery:** Generates a basic CLI tool.
- **Prolog Protoplasm:** Enforces a "Safe String" policy (e.g., banning specific keywords).
- **HOTL Process:** Demonstrates Forge building and Crucible auditing the string-filtering logic.
- **Launch:** Final deployment of the verified binary.

---
*Reference: [PROCESS.md](./PROCESS.md) for the HOTL Agent Protocol.*
