# CLAUDE.md Amendment Proposal
**Date:** 2026-04-09 06:00:00 UTC
**Proposal ID:** amendment-2026-04-09-phase7-assembly-line-manager
**Proposed by:** Roy Peter D'Souza (Human) + Analyst Droid (design)
**Status:** APPROVED — explicit operator directive

## Summary

Add Phase 7 — Assembly Line Manager. Provides a cookie-cutter lifecycle for spinning up
new software services through the dark factory.

## Decisions locked

- Self-modification: Option A (skill versioning via UpdateSkill RPC) + Option B
  (code proposals through existing SubmitProposal pipeline). No live self-patching.
- Cross-service sharing / Service Registry: shelved, not in scope.
- Roy has final say on all specs. Advisor challenges are recorded but never block.
- INTAKE produces a draft spec. Roy explicitly approves it. Only after Roy's approval
  does the spec enter DESIGN review by Analyst Droid. Analyst Droid reviews Roy's
  approved requirements — never its own draft. Clean separation.
- Dialog medium: Claude Code via MCP tools. Control Panel UI deferred.
- Deployment target (LOCAL / CONTAINER / AWS / GCP) declared in spec at intake.
- Each assembly line is fully self-contained. No cross-assembly-line sharing for now.

## Changes to CLAUDE.md

- §2: Add assembly-lines/ to project structure; add factories/scaffold-engine/
- §3: Add Phase 7 spec
- §3 Phase 3 proto: Add CreateAssemblyLine, GetAssemblyLineStatus, AdvanceLifecycle,
  UpdateSkill RPCs to Orchestrator service
