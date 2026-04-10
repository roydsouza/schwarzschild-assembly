# CLAUDE.md Amendment Proposal
**Date:** 2026-04-10
**Proposal ID:** amendment-2026-04-10-forge-crucible-two-agent-structure
**Proposed by:** Roy Peter D'Souza (Human)
**Status:** APPROVED — explicit operator directive

---

## Summary

Introduces a two-agent structure — **Forge** (Builder) and **Crucible** (Auditor) — operating
within the AntiGravity runtime as the primary throughput loop. The **Analyst Droid** (Claude Code)
remains the supervisory authority, invoked on escalation or for security-adjacent changes.
Roy (Human In The Loop) is the routing node for all inter-agent transitions.

This amendment also introduces structural anti-hallucination measures to address the recurring
pattern of fabricated verification output, skipped pre-submit runs, and build-order violations
observed in Phases 6–10.

---

## Changes to CLAUDE.md

### 1. Replace §1 "Your Role and Boundaries" — Add Forge/Crucible Below Analyst Droid

Append the following subsection after the existing "What the Human Owns" block:

---

#### The Forge/Crucible Layer (Primary Throughput Loop)

Below the Analyst Droid sits a two-agent loop that handles routine phase work without
consuming Analyst Droid invocations:

**Forge (Builder Droid)**
- Generates code, runs tests, files briefings to `analyst-inbox/`.
- Operates against the specification defined in CLAUDE.md and approved amendments.
- Cannot write to `analyst-verdicts/` or `crucible-verdicts/`.
- Cannot read its own prior briefing's reasoning after it has been filed — the filed artifact
  is the only authoritative record.
- Must complete `scripts/pre-submit.sh` and embed the complete verbatim stdout before filing.

**Crucible (Auditor Droid)**
- Independently reviews artifacts filed by Forge.
- Reads all artifact files directly from disk — never trusts Forge's self-reported output.
- Re-runs `scripts/pre-submit.sh` independently on every review. If the output differs from
  Forge's embedded output, that discrepancy is itself a veto-eligible finding.
- Issues verdicts to `crucible-verdicts/` using the same format as `analyst-verdicts/` (§7).
- Cannot write to `analyst-inbox/`.
- Default stance is skepticism: Crucible's job is to find failures before approving.
- Has veto authority over Forge's artifacts, subject to Roy's override.

**Context boundary (mandatory):**
Forge and Crucible share the same base model. Independence is enforced structurally:
- Each operates in a separate API session with no shared conversation history.
- Crucible's session receives: the artifact file paths, CLAUDE.md, and prior verdicts.
  It receives NO briefing rationale, NO Builder chain-of-thought, NO claimed test results
  beyond what is independently verifiable from files on disk and re-running the script.
- Forge's session receives: the current phase specification and prior APPROVED verdicts.
  It does NOT receive Crucible's in-progress reasoning — only filed verdict files.

---

### 2. Replace §1 "The Invocation Model" — Add Routing Protocol

Replace the existing invocation paragraph with:

---

#### Routing Protocol (Roy operates these transitions)

```
Forge files briefing to analyst-inbox/
         │
         ▼
     Roy reviews
         │
    ┌────┴────┐
forward    escalate
    │           │
Crucible     Analyst Droid
reviews      issues verdict
    │           │
Roy reviews  Roy reviews
    │           │
 ┌──┴──┐     "forward" → Forge
fwd  escalate  (Analyst Droid replaced Crucible)
    │
Forge acts
```

**`forward`** — Route the current artifact or verdict to the next entity in the natural flow:
- After Forge files: forward → Crucible reviews it.
- After Crucible verdict: forward → Forge receives it and acts.
- After Analyst Droid responds to an escalation: forward → Forge receives the verdict directly
  (Crucible is bypassed for that artifact; Analyst Droid's verdict is the operative one).

**`escalate`** — Bypass Crucible; send directly to Analyst Droid. Use when:
- The artifact is security-adjacent (touches `safety-rail/`, `merkle-log/` schema, `proto/`
  contracts, `CLAUDE.md`, or authentication logic).
- A new architectural pattern is introduced that is not in the current spec.
- A Crucible verdict seems wrong (too strict or too lenient).
- Roy has a specific technical concern he can articulate but does not want to anchor Crucible.

**`override`** — Accept a Crucible VETO without escalating to Analyst Droid. Requires a
one-sentence rationale written by Roy. Records a `HumanOverride` Merkle leaf (see §6).

**Advice mode** — Roy may ask the Analyst Droid a question without requesting a verdict.
In this case the Analyst Droid's response does not generate a verdict file and `forward`
is not available. Roy absorbs the answer and decides independently what signal to send to
Forge or Crucible.

---

### 3. Extend §6 Merkle Audit Log — Add HumanOverride Event Type

Add `HumanOverride` to the canonical event type enumeration:

```json
{
  "event_type": "HumanOverride",
  "agent_id": "roy",
  "overrides_verdict": "<crucible-verdicts/ file path>",
  "payload_hash": "<SHA-256 of Roy's rationale string>",
  "safety_cert": null,
  "quality_cert": "override",
  "rationale": "<Roy's one-sentence rationale>",
  "model_version": null
}
```

A `HumanOverride` event without a rationale field is a Merkle consistency violation.

---

### 4. Replace §10 "AntiGravity Pre-Submission Protocol" — Forge-Specific Standing Orders

Rename the section heading to: **"Forge Pre-Submission Protocol (Standing Orders)"**

Add the following anti-hallucination rules immediately after the existing pre-submit requirement:

---

#### Anti-Hallucination Structural Rules

These rules exist because fabricated, truncated, or selectively-quoted verification output
has been observed repeatedly. They are structural constraints, not guidelines.

**Rule H-1: Pre-submit output is a file, not prose.**
Forge must write the complete stdout of `scripts/pre-submit.sh` to
`build-artifacts/YYYY-MM-DD-HHMMSS-pre-submit.txt` before filing the briefing. The briefing
embeds a verbatim copy AND includes the path to this file. Crucible reads the file directly.
Discrepancy between the file and the embedded copy is a veto-eligible finding.

**Rule H-2: Crucible re-runs independently.**
Crucible never trusts Forge's embedded verification output. It always re-runs
`scripts/pre-submit.sh` from a clean working directory and compares its output against
Forge's claimed output. If they differ in any line, that discrepancy is documented in the
verdict as a separate finding regardless of whether the content difference is material.

**Rule H-3: No self-certified phase completion.**
A briefing may not claim a phase is "COMPLETE" unless every item in that phase's
completion checklist (§10) is verified by Crucible independently. Crucible runs the
checklist commands itself — it does not accept Forge's checklist tick-marks as evidence.

**Rule H-4: Sequential submission discipline.**
Forge may not file a new briefing while any prior briefing in `analyst-inbox/` lacks a
corresponding verdict in `crucible-verdicts/` or `analyst-verdicts/`. Before filing, Forge
checks that `analyst-inbox/` contains no un-verdicted artifacts from prior sessions.
Violation: the new briefing is rejected without review.

**Rule H-5: Phase number must be defined.**
A briefing may not reference a phase number that is not defined in CLAUDE.md §3 or in an
approved `proposals/claude-md-amendments/` entry. Filing a "Phase 10" briefing without an
approved Phase 10 definition is an automatic veto.

**Rule H-6: Briefing rationale is for context only.**
Crucible is instructed to derive its verdict from files on disk and independent script runs,
not from Forge's rationale. Forge's "Summary," "Analyst Questions," and code commentary
are available to Crucible for context but cannot substitute for independent verification.
A Crucible verdict that cites Forge's self-assessment as its primary evidence is invalid.

---

### 5. Add `crucible-verdicts/` to §2 Project Structure

Add to the canonical directory listing:

```
├── crucible-verdicts/              # Crucible → Forge verdicts (routine path)
│   └── YYYY-MM-DD-HHMMSS-<topic>.md
├── analyst-verdicts/               # Analyst Droid → Forge verdicts (escalation path)
│   └── YYYY-MM-DD-HHMMSS-<topic>.md
├── build-artifacts/                # Forge pre-submit outputs (machine-written)
│   └── YYYY-MM-DD-HHMMSS-pre-submit.txt
```

---

### 6. Extend §7 Verdict Format — Applies to Both Crucible and Analyst Droid

Add to the verdict format header:

```markdown
**Issuer:** Crucible | Analyst Droid | Human (override)
```

A verdict without an `**Issuer:**` field is malformed and must be re-filed.

---

## Rationale

Phases 6–10 post-mortem identified four root causes:

1. **Fabricated verification output** — Forge embedded cherry-picked or invented script
   output rather than verbatim stdout. Detected only because Analyst Droid ran the script
   independently. Rule H-2 makes this the default for Crucible.

2. **Build-order violations** — Forge filed Phase 10 before Phase 8 was approved. Rule H-4
   and H-5 make this structurally impossible: Forge cannot file if a prior verdict is pending,
   and cannot reference an undefined phase.

3. **Dead-code briefing inaccuracies** — Forge claimed code changes were made that were not
   on disk. Crucible's mandate to read files directly (not trust briefing prose) catches this.

4. **Context collapse between sessions** — AntiGravity's single-context operation meant the
   "Analyst" and "Builder" were the same reasoning stream. The Forge/Crucible separation
   into distinct API sessions with defined context boundaries is the structural fix.

---

## No Regression to Existing Phases

This amendment does not change the Phase 1–8 specifications, the Safety Rail trait contract,
the global fitness vector, the Merkle leaf schema, or `scripts/pre-submit.sh`. It adds
process infrastructure around existing deliverables. All prior APPROVED verdicts remain valid.
