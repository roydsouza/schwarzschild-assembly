# Scaffold Engine Factory

The Scaffold Engine is a core component of the Phase 7 "Assembly Line Manager" architecture. It is responsible for autonomously initializing software service repositories from validated specifications.

## Overview

In the Schwarzschild Assembly lifecycle, the Scaffold Engine becomes active during the **SCAFFOLD** phase. After the **Requirements Advisor** (Root Spine MCP) has finalized a `SpecDocument` and the **Analyst Droid** has issued an `APPROVED` verdict, the Scaffold Engine translates those requirements into a tangible codebase.

## Key Capabilities

- **Boilerplate Generation**: Initializes Go/Rust/Python services with standardized directory structures.
- **Safety Injection**: Automatically configures the `safety-rail/` local hooks and `CLAUDE.md` prime directives for new services.
- **Interface Consistency**: Ensures generated proto bindings and API contracts align with the Root Spine.

## Metrics

The Scaffold Engine reports the following domain fitness metrics:
- `scaffold_success_rate`: % of created specs that reach DELIVERED. Escalation: < 60%.
- `scaffold_latency`: p99 ms from `finalize_spec` to scaffold artifacts ready. Escalation: > 30,000ms.
- `acceptance_criterion_coverage`: % of spec acceptance criteria with non-stub test implementations at VERIFY. Escalation: < 100%.
- `rework_rate`: % of assembly lines requiring > 2 CONDITIONAL verdicts before BUILD. Escalation: > 30%.

## Anatomy

- `worker/`: Main execution loop and gRPC client.
- `domain-fitness/`: Metric declaration and reporting logic.
- `mcp-server/`: Tool definitions for autonomous scaffolding actions.
- `analyst-briefing/`: Storage for phase-completion reports ready for Analyst Droid review.
