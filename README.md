# Schwarzschild Assembly: Automated Production for High-Integrity Systems

A centralized environment for the automated assembly and formal verification of high-integrity software systems.

## Project Role & Relationships
- **Purpose**: Implements an automated operational model for code production, ensuring all state changes meet formal safety constraints before deployment.
- **Substrate**: Utilizes the homoiconic logic substrate (STASIS) provided by the **[shapeshifter](../shapeshifter/)** project for reflecting on and refining assembly logic.
- **Governance**: Employs the standardized Forge/Crucible harnesses and protocols defined in the **[agents](../agents/)** library to govern agentic interactions.

## Technical Objectives

The Schwarzschild Assembly provides a rigorous production environment designed to eliminate non-deterministic or structurally fragile software through:
- **Formal Verification**: Integrating SMT solvers (e.g., Z3) into the production pipeline.
- **Adversarial Auditing**: Utilizing the Forge/Crucible protocol to identify and remediate corner cases during assembly.
- **Tiered Logic Constraints**: Restricting software functionality to predefined safety bounds (The Safety Horizon).

### Why "Schwarzschild-Radius"?
In physics, the **Schwarzschild Radius** defines the event horizon of a black hole: the boundary where escape is impossible. 

In this station, we apply this principle to code integrity. The Assembly creates a **Safety Horizon**. Once a spacecraft (application) enters a docking bay for refit:
1. It is compressed by **Formal Verification** (Z3 SMT solvers).
2. It is audited by an **Adversarial Protocol** (Forge/Crucible).
3. It is bound by a tiered logic substrate (**STASIS**).

Any code that fails to meet the station's physical laws (formal constraints) cannot cross the event horizon. It is effectively trapped within the gravity of the assembly until it is hardened, or it is purged.

---

## 🏗️ Station Architecture

The station is organized into a persistent **Core Station** and ephemeral **Docking Bays**.

### The Shell/Mind Dichotomy
A Schwarzschild Spacecraft is composed of two distinct layers:
- **Static Shell (Hardware/Native):** Hardened Go/Rust binaries compiled for Apple Silicon (M5) with native acceleration for safety evaluations.
- **Protoplasm Mind ([STASIS](./STASIS-LANGUAGE.md)):** A malleable, tiered logic substrate that provides the "Life Support" (intelligence) to the ship.

### Adversarial Auditing: Forge & Crucible
All operations follow a strict two-agent protocol to eliminate model-specific blind spots:
- **Forge (The Builder):** High-velocity implementation agent. Its only output is a structured briefing of proposed changes.
- **Crucible (The Auditor):** An adversarial auditor that assumes the Forge has cut corners. It independently re-runs all verification tools and SMT proofs from disk.

---

## 🚀 Operations & Lifecycle

Every spacecraft follows a rigorous lifecycle from initial docking to final launch:

1. **Docking (Provisioning):** `dock.sh` creates an ephemeral workspace (`docking-bays/`) and symlinks the shared Protoplasm.
2. **Refit (Assembly):** The **HOTL** (Human On The Loop) process governs the iterative build-audit-fix cycle.
3. **Launch (Deployment):** Once a ship reaches **0-FAIL status** and satisfies Z3 policy compliance, it is compressed into an archive and deployed.

**Deep Dives:**
- **[Operations Manual](./OPERATIONS.md)** — Industrial manual for build-release-repair cycles.
- **[Process Guide](./PROCESS.md)** — The Three-Agent protocol and routing rules.
- **[Spacecraft Architecture](./SPACECRAFTS.md)** — Technical spec for Shell/Mind anatomy.

---

## 🛠️ Flight Ops (Operational Commands)

### Docking a Ship
Spawn a new assembly line for a spacecraft:
```bash
./core-station/bridge/dock.sh <ship-name>
```

### Launching a Ship
Finalize the mission, archive the state, and deploy:
```bash
./core-station/bridge/launch.sh <ship-name>
```

### Station Health Check
Verify the core station integrity and Merkle-log consistency:
```bash
./core-station/bridge/pre-submit.sh
```

---
*Schwarzschild Assembly: Engineering beyond the Event Horizon.*

