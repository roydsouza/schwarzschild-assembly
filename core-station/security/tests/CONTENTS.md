# Safety Rail: Formal Verification Tests

The Tests sector contains the suite of formal contract compliance and sandbox verification tests that ensure the integrity of the Safety Rail.

## Sub-Sectors
- [integration_tests/](./integration_tests/CONTENTS.md) — End-to-end verification of mission safety bounds.
- [unit_tests/](./unit_tests/CONTENTS.md) — Isolated tests for Safety Rail components.

## Key Files
- [contract_compliance_tests.rs](./contract_compliance_tests.rs) — Tests for Z3 policy adherence.
- [sandbox_tests.rs](./sandbox_tests.rs) — Tests for resource isolation and execution boundaries.
