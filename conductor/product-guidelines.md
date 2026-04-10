# Product Guidelines: Schwarzschild Assembly

## 1. Document Style
- **Technical Precision:** Use explicit, unambiguous language. Avoid vague terms like "fast" or "secure"; use "low-latency" and "formally verified" where appropriate.
- **Structured Hierarchy:** Use standard markdown headings (H1 for titles, H2 for sections, H3 for sub-sections) and consistent list styles.
- **Reference Integrity:** All internal links between documentation files must use relative paths and be verified.

## 2. Technical Standards
- **Mathematical Determinism:** Every architectural change must consider its impact on the safety substrate. Documentation should highlight potential Z3 constraint implications.
- **Verification First:** Prefer design patterns that simplify formal verification. Avoid complex, non-deterministic state machines where a simpler, verifiable alternative exists.
- **Hardware Awareness:** Design and document components with Apple Silicon (M5) optimizations in mind, specifically unified memory bandwidth and neural accelerator utilization.

## 3. Communication Guidelines
- **Audit-Ready Logs:** All system events and error messages must be structured for Merkle-log inclusion. Log messages should be concise and actionable.
- **Security-First Reporting:** Security assessments and bug reports must be prioritized. Any deviation from safety constraints must be documented as a critical event.
- **Review Protocol:** Code and documentation changes follow a strictly iterative review cycle between the AntiGravity substrate and independent AI analysts.