# 🔘 Schwarzschild Assembly

![Sati-Central Space Factory](assets/hero.png)

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
**[📍 Project Roadmap](CONTENTS.md)**

---

## 🌌 The Problem: AI Slop & The Verification Gap

Developing agentic systems through "vibe coding"—relying on probabilistic LLM outputs without a formal safety substrate—results in **AI Slop**. This is characterized by codebases that are fragmented, unverified, and prone to "ghost regressions" where one agent's fix accidentally compromises another's logic.

For technically sophisticated environments, the challenge is not just automation; it is **The Verification Gap**. There is a fundamental disconnect between the high-velocity generation of an agent and the deterministic requirements of a production runtime. Without a strictly governed event horizon, autonomous evolution inevitably drifts toward entropy.

---

## 🏗️ The Solution: Adversarial Self-Evolution

**Schwarzschild Assembly** implements a "Dark Factory" model—a lights-out production environment where agency is constrained by a mathematical proof-of-safety. 

The core integrity mechanism is the **Adversarial Loop** between two specialized supervisory agents:

### 🔨 The Forge (Worker Droid / AntiGravity)
The Forge is the primary constructive force. It manages the assembly lines, ingests requirements (Specs), and generates high-velocity implementations in Go, Rust, or Python. It is optimized for throughput and domain-specific accuracy.

### ⚖️ The Crucible (Analyst Droid / Claude Code)
The Crucible is the adversarial challenger. It holds architectural authority and maintains the **Safety Rail**. It reviews every artifact produced by The Forge, holding the power of a **unilateral veto**. The Crucible is not a generator; it is a validator that ensures no proposal crosses the event horizon without a deterministic safety certificate.

**Crucible keeps Forge honest:** By requiring every Forge commit to pass both formal Z3 verification and a nuanced architectural audit, the assembly ensures that self-evolution is always upwardly mobile and never corrosive.

---

## 🎮 The Operator and the HOTL Model

Currently, the Schwarzschild Assembly operates in a **HOTL (Human-On-The-Loop)** configuration.

- **The Operator (You):** Acts as the final arbiter. Proposals that pass both the Safety Rail and The Crucible's audit are staged at the **Translucent Gate**.
- **Push the Button:** Evolution is currently gated by your signature. You review the evidence—Merkle proofs, fitness deltas, and audit logs—and "push the button" to commit the change to the production substrate.
- **The North Star:** We are moving toward **Full Lights Out**. As our Tier 2 formal proofs (Rocq-of-Rust) and Crucible audit density increase, the need for human intervention will diminish, leaving the Operator to set high-level strategic "Standing Orders" while the factory executes autonomously.

---

## 🚀 Creating an Assembly Line

The Schwarzschild Assembly is designed to spin up new autonomous services via the **Scaffold Engine**.

### 1. The Intake Phase
An assembly line begins with an **INTAKE** conversation with The Crucible. You define the requirements, and a `SpecDocument` is generated. Once the spec is finalized and approved, the line advances.

### 2. Supported Substrates
The assembly currently supports a multi-language substrate optimized for the Apple M5 architecture:
- **Go**: High-concurrency API servers and data pipelines. (Default)
- **Rust**: Performance-critical modules, safety rails, and cryptography.
- **Python**: Agentic logic, RAG ingestion, and LLM orchestration.
- **TypeScript**: High-density Control Panel UIs.
- **Prolog (SWI)**: Self-modifying knowledge bases and homoiconic skill loops.
- **Haskell**: High-assurance formal logic and property-based verification.

### 3. The Lifecycle Gate
Every service moves through a non-linear but strictly gated lifecycle:
`INTAKE → DESIGN → SCAFFOLD → BUILD → VERIFY → DELIVERED`

New assembly lines can be initiated by defining a target service and its primary language. The Scaffold Engine will then generate a local factory anatomy for that service's autonomous maintenance.

---
*Part of the [AntiGravity](https://chromewebstore.google.com/detail/antigravity-browser-exten/eeijfnjmjelapieobjiielcpmhhchbkg) station infrastructure.*
