# The Schwarzschild Assembly

The Schwarzschild Assembly is a **Meta-Factory** (Space Station) designed to build, repair, and retrofit autonomous software applications (**Spacecraft**) using self-evolving assembly lines gated by rigorous safety protocols. This is a Dark Factory for building applications (spacecraft) without the "AI Slop" of Vibe Coding.

## 🛰️ Architecture: The Space Station Model

A Schwarzschild Spacecraft is composed of a hardened **Static Shell** (Go/Rust) and a malleable **Protoplasm Mind** (Prolog). This "Shell/Mind" dichotomy ensures that applications are both physically secure and logically self-evolving.

**Deep Dive:** For the technical specification of spacecraft anatomy, see **[Spacecraft Architecture](./SPACECRAFTS.md)**.

The station is organized into a persistent **Core Station** and ephemeral **Docking Bays**.

### 🏗️ Station Naming Convention
- **Space Station (Meta-Factory):** The main repository (`schwarzschild-assembly`). It provides the shared intelligence, machinery, and governance.
- **Spacecraft (Application):** A modular software product (e.g., a CLI tool, a Web App) built within the station.
- **Docking Bay (Assembly Line):** A temporary workspace where a specific spacecraft is actively modularized or repaired.
- **Protoplasm (Intelligence):** The shared Prolog reasoning substrate that provides the "Life Support" to all spacecraft.
- **Machinery (Engines):** The reusable Go/Rust factories that perform the actual code manufacturing.
- **Dry Dock (Archives):** Storage for completed mission snapshots and flight data.

## 🚀 Operations & Lifecycle

Every spacecraft follows a rigorous lifecycle from initial docking to final launch and subsequent retrofitting. 

**Deep Dive:** For a comprehensive technical manual on build-release-repair cycles, see the **[Operations Manual](./OPERATIONS.md)**.

### 🛡️ The HOTL Process (High-Order Test Loop)
All assembly and refit work is governed by the **HOTL Process**. This Two-Agent protocol ensures that no code is launched without an independent audit and safety-rail verification.
- **Forge (The Builder):** Proposes and implements modifications.
- **Crucible (The Auditor):** Verifies modifications against the station's safety policies and gives final launch approval.

**Protocol Details:** See **[PROCESS.md](./PROCESS.md)** for governed execution rules.

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
Verify the core station integrity:
```bash
./core-station/bridge/pre-submit.sh
```

---
*Schwarzschild Assembly: Engineering the Event Horizon.*
