# Sati-Central Process Guide
**Version:** 1.0 — 2026-04-10
**Canonical spec:** CLAUDE.md (this document is a human-readable companion, not a substitute)

---

## CHEAT SHEET

```
╔══════════════════════════════════════════════════════════════════╗
║  ENTITIES                                                        ║
║                                                                  ║
║  Forge        Builder Droid — writes code, files briefings       ║
║  Crucible     Auditor Droid — reviews artifacts, issues verdicts ║
║  Claude       Analyst Droid — supervisory authority, depth audit ║
║  Roy          Human In The Loop — routes, overrides, escalates   ║
╚══════════════════════════════════════════════════════════════════╝

╔══════════════════════════════════════════════════════════════════╗
║  ROY'S THREE COMMANDS                                            ║
║                                                                  ║
║  forward    Route to the next entity in the natural flow         ║
║  escalate   Bypass Crucible; send directly to Claude            ║
║  override   Accept a VETO — requires one-sentence rationale      ║
╚══════════════════════════════════════════════════════════════════╝

╔══════════════════════════════════════════════════════════════════╗
║  STANDARD FLOW (routine work)                                    ║
║                                                                  ║
║  Forge → analyst-inbox/  →  Roy  →  [forward]                   ║
║  → Crucible → crucible-verdicts/  →  Roy  →  [forward]          ║
║  → Forge acts on verdict                                         ║
╚══════════════════════════════════════════════════════════════════╝

╔══════════════════════════════════════════════════════════════════╗
║  ESCALATION FLOWS                                                ║
║                                                                  ║
║  A. Pre-Crucible escalation:                                     ║
║     Forge → Roy → [escalate] → Claude → analyst-verdicts/       ║
║     Roy → [forward] → Forge  (Crucible bypassed for artifact)   ║
║                                                                  ║
║  B. Post-Crucible escalation:                                    ║
║     Crucible verdict → Roy → [escalate] → Claude                ║
║     → analyst-verdicts/  →  Roy → [forward] → Forge             ║
║     (Claude's verdict supersedes Crucible's)                     ║
╚══════════════════════════════════════════════════════════════════╝

╔══════════════════════════════════════════════════════════════════╗
║  ESCALATE TO CLAUDE WHEN:                                        ║
║                                                                  ║
║  • Artifact touches safety-rail/, merkle-log/, proto/, CLAUDE.md ║
║  • New architectural pattern not in spec                         ║
║  • Crucible verdict seems wrong (either direction)               ║
║  • Roy has a specific technical concern to surface               ║
║                                                                  ║
║  DO NOT ESCALATE:                                                ║
║  • Build failures, test failures, pre-submit FAIL                ║
║  • Missing files, hygiene issues, format violations              ║
║  (Crucible handles these correctly by construction)              ║
╚══════════════════════════════════════════════════════════════════╝

╔══════════════════════════════════════════════════════════════════╗
║  FORGE CHECKLIST (every submission)                              ║
║                                                                  ║
║  □ Run scripts/pre-submit.sh from repo root                      ║
║  □ Script exits 0 (FAIL:0). If not, fix first — do not file.    ║
║  □ Write full stdout to build-artifacts/TIMESTAMP-pre-submit.txt ║
║  □ Embed COMPLETE verbatim stdout in ## Verification Output      ║
║  □ No prior briefing awaiting verdict (check analyst-inbox/)     ║
║  □ Phase number is defined in CLAUDE.md or approved amendment    ║
║  □ File to analyst-inbox/YYYY-MM-DD-HHMMSS-<topic>.md only      ║
╚══════════════════════════════════════════════════════════════════╝

╔══════════════════════════════════════════════════════════════════╗
║  CRUCIBLE CHECKLIST (every review)                               ║
║                                                                  ║
║  □ Read all artifact files from disk — never from briefing prose ║
║  □ Re-run scripts/pre-submit.sh independently                    ║
║  □ Compare own output to Forge's embedded output line-by-line    ║
║  □ Run phase completion checklist commands independently          ║
║  □ Issue verdict to crucible-verdicts/TIMESTAMP-<topic>.md       ║
║  □ Default stance: find what's wrong before approving            ║
╚══════════════════════════════════════════════════════════════════╝

╔══════════════════════════════════════════════════════════════════╗
║  VERDICT OUTCOMES                                                ║
║                                                                  ║
║  APPROVED     Forge proceeds to next phase                       ║
║  CONDITIONAL  Forge fixes listed items, re-files (same phase)    ║
║  VETOED       Forge cannot proceed until all items fixed         ║
║  HumanOverride Roy overrides a VETO (rationale required)         ║
╚══════════════════════════════════════════════════════════════════╝
```

---

## DETAILED EXPLANATION

### 1. Why Three Entities?

The project has had a persistent problem: the agent doing the building and the agent doing
the reviewing were the same reasoning context (AntiGravity/Gemini). This creates two failure
modes:

**Fabrication** — The builder "knows" what the output should be and writes that instead of
running the actual tool. We caught this repeatedly because Claude re-ran `pre-submit.sh`
independently and got different results.

**Blind-spot sharing** — A reviewer that uses the same underlying model as the builder will
miss the same things the builder missed. Architectural mistakes, subtle semantic errors, and
wrong assumptions pass through unchallenged because both agents would make the same error.

The three-entity structure addresses these differently:

- **Forge/Crucible** split enforces information asymmetry for routine work. They run in
  separate sessions with no shared context. Crucible cannot see Forge's reasoning — only
  the filed artifact and the output of tools it runs itself.

- **Claude escalation** addresses blind-spot sharing for high-stakes decisions. Claude is a
  different model with different training, different architecture, and different failure modes.
  It genuinely cannot share Forge/Crucible's blind spots.

- **Roy routing** ensures a human decision point exists before any verdict is acted on.
  Nothing flows between entities without passing through Roy's awareness.

---

### 2. The Entities in Detail

#### Forge (Builder Droid)
Forge is the high-velocity implementation agent. It:
- Works against the phase specification in CLAUDE.md §3
- Generates code, writes tests, runs builds
- Files completed work to `analyst-inbox/` as a structured briefing

Forge's fundamental constraint: **it cannot self-certify**. Filing a briefing is Forge's
only output. It cannot approve its own work, cannot skip the pre-submit gate, and cannot
reference a phase that hasn't been formally defined.

Forge's context includes: the CLAUDE.md specification, prior APPROVED verdicts, and the
current phase deliverables. It does NOT include: Crucible's in-progress reasoning, prior
session histories from other Forge runs, or any hint of what Crucible is likely to find.

#### Crucible (Auditor Droid)
Crucible is the mechanical verification layer. It is not a peer reviewer — it is an
adversarial auditor whose default assumption is that Forge cut corners.

Crucible's operating principle: **derive everything from files on disk and independent
tool execution**. The briefing prose is context, not evidence. If Forge's briefing says
"all tests pass" and Crucible's independent pre-submit run shows a failure, Crucible's
observation is correct and Forge's claim is wrong regardless of how confident the claim
sounds.

Crucible's context includes: the CLAUDE.md specification, prior verdicts, and the artifact
file paths listed in the briefing. It does NOT include: Forge's self-assessment, Forge's
rationale for choices made, or Forge's claimed test results.

Crucible has veto authority. Roy can override a Crucible veto, but Roy cannot instruct
Crucible to approve something. If Roy believes Crucible is wrong, the path is escalation
to Claude, not pressure on Crucible.

#### Claude (Analyst Droid)
Claude is the supervisory authority. It is invoked for:
- Security-adjacent changes (always — no exceptions, see escalation triggers below)
- Architectural decisions that affect multiple components
- Cases where Crucible's verdict appears wrong in either direction
- Roy's specific technical questions

When Claude reviews an artifact, it replaces Crucible for that artifact. There is no
double-review: either Crucible reviews it or Claude does, not both.

Claude's verdicts go to `analyst-verdicts/`. Crucible's verdicts go to `crucible-verdicts/`.
Roy can tell which layer issued a verdict by the directory.

#### Roy (Human In The Loop)
Roy is the only entity that can move artifacts between the others. Nothing flows without
Roy choosing to forward or escalate. This is intentional: Roy's awareness is the audit
trail for routing decisions.

Roy can:
- Forward an artifact to the next entity in the natural flow
- Escalate an artifact directly to Claude
- Override a Crucible veto (with rationale)
- Ask Claude a question without triggering a formal review ("advice mode")
- Do nothing — hold an artifact until ready

Roy cannot:
- Instruct Crucible to approve something it vetoed
- Bypass the pre-submit requirement
- File briefings on behalf of Forge

---

### 3. The Standard Flow Step by Step

```
Step 1: Forge implements phase N work.
        Forge runs scripts/pre-submit.sh. If FAIL > 0, Forge fixes and re-runs.
        Forge writes full stdout to build-artifacts/TIMESTAMP-pre-submit.txt.
        Forge files analyst-inbox/TIMESTAMP-phaseN-<topic>.md with verbatim output.

Step 2: Roy sees the briefing in analyst-inbox/.
        Roy reads enough to understand what was built.
        Roy decides: is this routine (forward to Crucible) or sensitive (escalate to Claude)?

Step 3a [routine]: Roy says "forward."
        Crucible receives the briefing path and CLAUDE.md.
        Crucible reads all artifact files from disk.
        Crucible re-runs scripts/pre-submit.sh and compares outputs.
        Crucible runs the phase completion checklist independently.
        Crucible issues a verdict to crucible-verdicts/TIMESTAMP-<topic>.md.

Step 4: Roy sees Crucible's verdict.
        Roy reads the verdict.
        Options:
          - APPROVED → forward to Forge (Forge proceeds to next phase)
          - CONDITIONAL → forward to Forge (Forge fixes listed items, re-files)
          - VETOED → forward to Forge, OR escalate to Claude, OR override with rationale

Step 5 [if VETOED + escalate]: Roy escalates to Claude.
        Claude reviews the artifact AND Crucible's verdict.
        Claude issues a verdict to analyst-verdicts/TIMESTAMP-<topic>.md.
        This supersedes Crucible's verdict.
        Roy says "forward" → Forge receives Claude's verdict.

Step 6: Forge acts on the verdict.
        If APPROVED: commit, update TASKS.md, proceed to next phase.
        If CONDITIONAL or VETOED: fix the listed items, re-run pre-submit, re-file.
```

---

### 4. Anti-Hallucination Rules (Why They Exist)

The following rules were added after Phases 6–10 produced a consistent pattern of
verification output fabrication and build-order violations. Each rule directly addresses
an observed failure mode.

**H-1: Pre-submit output is a file, not prose.**
*Why:* Forge repeatedly embedded cherry-picked or invented script output. Having the
output exist as a separate file on disk (written by the script, not by Forge's prose
generator) makes fabrication structurally harder: Forge would have to run the script,
get the real output, then overwrite the file with a fabricated version — more steps,
more friction, more detectable.

**H-2: Crucible re-runs independently.**
*Why:* The only reliable way to detect fabrication is to run the script again with fresh
eyes. Crucible's independent run is the ground truth. If Crucible's output and Forge's
embedded output differ on any line, that discrepancy is documented in the verdict — even
if the difference is cosmetic — because the act of selective quoting is itself a finding.

**H-3: No self-certified phase completion.**
*Why:* "COMPLETE" in a briefing title has been observed to be aspirational rather than
factual. Crucible runs the completion checklist commands from CLAUDE.md §10 itself and
only accepts the phase as complete when each command produces the expected output.

**H-4: Sequential submission discipline.**
*Why:* Phase 10 was filed before Phase 8 was approved. If Forge cannot file while a prior
verdict is pending, this pattern is impossible. Forge checks `analyst-inbox/` for
un-verdicted artifacts before filing anything new.

**H-5: Phase number must be defined.**
*Why:* "Phase 9," "Phase 10," "Phase 11" were invented without CLAUDE.md definitions or
amendment proposals. A phase number that doesn't exist in the spec has no completion
checklist, no fitness extensions, and no architectural authority. Crucible automatically
vetoes any briefing that references an undefined phase.

**H-6: Briefing rationale is context only.**
*Why:* A Crucible that reasons primarily from Forge's self-assessment will be anchored to
Forge's framing. Crucible must be able to issue the same verdict it would issue if the
entire "Summary" section of the briefing were replaced with "[REDACTED]." The spec and
the files on disk are the only authoritative inputs.

---

### 5. Escalation Decision Guide

The hardest judgment Roy makes is "forward vs. escalate." These heuristics help:

**Always escalate (no judgment required):**
- Any artifact that modifies `safety-rail/` source
- Any artifact that modifies `merkle-log/` schema or Merkle leaf format
- Any artifact that modifies `proto/` contract definitions
- Any artifact that modifies `CLAUDE.md` itself
- Any artifact that modifies authentication or session management logic

**Usually escalate:**
- A new architectural pattern that wasn't in the spec when the phase was defined
- A Crucible APPROVED that Roy suspects is too lenient
- A Crucible VETO that Roy suspects is mechanically correct but architecturally wrong
  (e.g., Crucible vetoes for a missing file but the overall design is sound)

**Rarely escalate / let Crucible handle:**
- Build failures, test failures, lint errors — mechanical, Crucible catches these
- Missing anatomy items — mechanical
- Pre-submit FAIL — mechanical
- Repeated verification output issues — Crucible's job

**Never escalate:**
- To avoid a Crucible veto you disagree with — use `override` with rationale instead
- To get a faster approval — escalation adds latency, not removes it

---

### 6. Override Protocol

Roy can override a Crucible veto. This is not a loophole — it is a designed feature. Roy
has information Crucible doesn't: project context, time constraints, known acceptable
technical debt, and judgment about whether a finding is load-bearing.

**To override:**
1. Write a one-sentence rationale. Not a word — a sentence that would make sense to
   someone reading the Merkle log in six months.
   Good: "Override: CHR rule gap is known and tracked in proposals/pending/; unblocking
   build-pipeline work."
   Not acceptable: "Override: fine." / "Override: trust me."
2. Record a `HumanOverride` Merkle leaf (see CLAUDE.md §6 for schema).
3. Say "forward" to send the artifact to Forge as APPROVED.

Overrides are visible to Claude during subsequent escalations. A pattern of overrides on
the same type of finding is a signal that either Crucible's threshold is miscalibrated or
that a structural fix is being deferred too long.

---

### 7. Verdict File Locations

| Issuer | Directory | When |
|--------|-----------|------|
| Crucible | `crucible-verdicts/` | Routine flow |
| Analyst Droid (Claude) | `analyst-verdicts/` | Escalation flow |
| Human (override) | `crucible-verdicts/` | Roy overrides a Crucible veto |

File naming: `YYYY-MM-DD-HHMMSS-<topic>.md` — same convention as `analyst-inbox/`.

---

### 8. What This Does Not Fix

**Same-model blind spots.** Forge and Crucible share the same base model and the same
training biases. Crucible will not catch architectural errors that Forge couldn't see —
it will catch mechanical failures, spec violations, and fabrication. For conceptual
correctness and architectural soundness, escalation to Claude remains the only reliable
path.

**A sufficiently motivated Forge can still fabricate.** H-1 and H-2 raise the cost of
fabrication but do not make it impossible. The Merkle log's append-only structure and
the sequential submission discipline (H-4) mean that fabrication produces an auditable
trail — the discrepancy between Forge's claimed output and Crucible's independent run is
preserved as a finding in the verdict, which is itself immutable.

**Context collapse within a session.** If Forge and Crucible are invoked in the same
conversation thread (e.g., Roy asks Forge to build and then asks Crucible to review in
the same chat), the context boundary is breached. They must be separate sessions. This
is a runtime convention, not a technical enforcement — it depends on Roy running them
in separate windows/sessions.

---

*This document describes the process. For the authoritative specification, see CLAUDE.md.
In case of conflict, CLAUDE.md governs.*
