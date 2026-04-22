In the Schwarzschild meta-factory, a **Spacecraft** (Application) is defined by a dual-layered existence: a hardened, static **Shell** and a malleable, evolving **Protoplasm**. This architecture is the physical embodiment of the **Schwarzschild Radius** policy—creating a structural "event horizon" that prevents non-deterministic **AI Slop** from corrupting system invariants.

## 1. The Execution Shell (The "Body")

The Shell is the spacecraft's interface with the physical and digital universe.

- **Substrate:** Compiled languages (Go, Rust, C++) or hardened runtimes (Node.js, Python).
- **Nature:** **Static and Deterministic.** The Shell is manufactured by the station's machinery and deployed as a fixed asset. It does not change its own source code.
- **Responsibilities:**
    - High-performance I/O and System APIs.
    - Network transport and protocol handling.
    - User Interface (UI/UX) rendering.
    - Enforcement of the **Safety Rail** boundaries.
- **Metaphor:** The armored hull, life-support hardware, and sensory arrays.

## 2. The Protoplasm (The "Mind")

The Protoplasm is the spacecraft's reasoning engine and the seat of its self-evolution.

- **Substrate:** [STASIS](./STASIS-LANGUAGE.md) (Tiered Symbolic Logic, executed by SWI-Prolog). See [STASIS-LANGUAGE.md](./STASIS-LANGUAGE.md).
- **Nature:** **Malleable and Self-Evolving.** Because [STASIS](./STASIS-LANGUAGE.md) code is represented as symbolic data (homoiconicity), the mind can reason about its own rules and modify them at runtime.
- **Responsibilities:**
    - Business logic and complex decision-making.
    - Safety constraint heuristics.
    - Performance self-analysis and optimization.
    - Pattern recognition and knowledge persistence.
- **Metaphor:** The AI autopilot, mission computer, and linguistic core.

## 3. The Interface: The Aethereum-Spine

The relationship between Shell and Mind is governed by the **Aethereum-Spine**.

### The Flow of Awareness
1. **Perception:** The Shell receives a signal (e.g., a user request or a sensor reading).
2. **Consultation:** The Shell queries the Protoplasm: *"Is this action safe and consistent with our mission?"*
3. **Reasoning:** The Protoplasm evaluates the query against its current knowledge base and safety constraints.
4. **Conclusion:** The Protoplasm returns a symbolic result (e.g., \`accept\`, \`reject\`, or \`propose_retry\`).
5. **Execution:** The Shell performs the physical action based on the Mind's decision.

### The Self-Evolution Loop (Phases 11+)
Under the supervision of the **Crucible**, the Protoplasm is permitted to "improve" its own reasoning rules. This recursive self-analysis is what allows a Schwarzschild Spacecraft to become smarter over time without requiring a manual code refactor of the Shell.

## 4. Why This Architecture?

| Feature | Static Shell | Malleable Mind |
| :--- | :--- | :--- |
| **Safety** | Prevents physical system corruption. | Enforces logical and ethical boundaries. |
| **Speed** | Handles high-throughput operations. | Handles high-complexity reasoning. |
| **Agility** | Updated via Station Refit (Launch). | Updated via Self-Evolution (Knowledge). |
| **Audit** | Fixed during manufacturing. | Every mental shift is logged to the Merkle-tree. |

---
*Reference: [OPERATIONS.md](./OPERATIONS.md) for the mission lifecycle.*
