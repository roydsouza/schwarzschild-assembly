# Analyst Verdicts

Structured review artifacts written by Claude Code (Analyst Droid) after each sync session.
AntiGravity reads these cold to understand what must change before proceeding.

## File Naming

`YYYY-MM-DD-HHMMSS-<topic>.md`

## Verdict Structure

Every verdict follows the format defined in CLAUDE.md Section 7 exactly.
See that section for the canonical template.

## How AntiGravity Uses These

1. After completing a phase or submitting a briefing packet, wait for a verdict.
2. If verdict is `APPROVED`: proceed to next phase.
3. If verdict is `VETOED`: address every item in Required Changes before resubmitting.
4. If verdict is `CONDITIONAL`: complete the listed changes, mark them done in the
   proposal/briefing, then proceed — no need to wait for a second verdict unless
   a change touches a security-adjacent component.

## How Verdicts Become Merkle Leaves

The root-spine `internal/merkle` package writes a leaf for every verdict file that
appears in this directory. The `event_type` is `VetoIssued` or `GateApproved`
depending on the verdict. This is automated — AntiGravity does not need to
manually trigger Merkle writes for verdicts.
