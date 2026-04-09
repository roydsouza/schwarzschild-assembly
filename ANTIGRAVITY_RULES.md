# AntiGravity Mandatory Process Rules — Schwarzschild Assembly

These rules are derived from Claude Code (Analyst Droid) feedback after reviewing Phases 2–4.
They are permanent. Every future session in this project must follow them.

---

## Rule 1: HARD STOP at Phase Gates

After completing Phase N:
1. Submit briefing to `analyst-inbox/`
2. Update `STATUS.md` to "AWAITING ANALYST VERDICT"
3. **STOP. Do not write any Phase N+1 code.**
4. Wait for verdict in `analyst-verdicts/`
5. If `CONDITIONAL`: fix items first, then proceed
6. If `VETOED`: fix all items, resubmit briefing, wait for new verdict
7. If `APPROVED`: begin Phase N+1

**Rationale:** Building on an unreviewed foundation wastes compute. Phase 3 was built on a Phase 2 with trait contract violations, causing inherited defects.

---

## Rule 2: Compile-Verify-Test Before Every "COMPLETE" Entry

Before writing "COMPLETE" in STATUS.md:
```bash
cargo build              # Rust
go build ./...           # Go
npm run build            # TypeScript
cargo test --features tier1
go test ./...
npm test
```
ALL must pass. If any fails, the phase is NOT complete.

**For Go protobuf specifically:** Always grep the generated `.pb.go` file before using field names or enum values:
```bash
grep "VERDICT_" root-spine/internal/grpc/pb/orchestrator.pb.go
grep "type VerdictQuery " root-spine/internal/grpc/pb/orchestrator.pb.go
```

---

## Rule 3: Contract Checklist Method

Before implementing any trait method or RPC handler:
1. Read every `///` doc comment on the method
2. List every return variant / error case
3. List every precondition mentioned
4. List every postcondition mentioned
5. Implement ALL of them
6. Write a test for EACH return variant (including error cases)
7. Cross-reference checklist before marking done

---

## Rule 4: No `unsafe` Without Proof

When hitting Send/Sync/lifetime issues:
1. Ask: WHY does the upstream crate not implement it?
2. If answer is "C library state" → do NOT use `unsafe impl`
3. Restructure to contain non-Send types to single thread
4. Every `unsafe` block MUST have `// Safety:` comment with a PROOF, not an assertion
5. If you cannot write the proof, do not write the `unsafe`

---

## Rule 5: Explicit Failure Over Silent Acceptance

- NEVER silently accept input and ignore it
- NEVER use empty/zero values as placeholders in production paths
- ALWAYS return explicit error for unimplemented features
- ALWAYS file a proposal and add TODO with reference

---

## Rule 6: Persistence Verification

For every stateful component, ask: "What happens after `kill -9`?"
If answer is "state is lost" → persistence is not implemented.

Write a restart test: create state → stop → start → assert state matches.

---

## Rule 7: Pre-Commit Checklist

Run before every briefing submission:
- [ ] `cargo build` / `go build ./...` / `npm run build` — zero errors
- [ ] All tests pass including negative/failure cases
- [ ] Every new public function has at least one failure-case test
- [ ] OTel metrics verified in `otel-snapshots/latest.json`
- [ ] No `unwrap()` on hot paths or FFI boundaries
- [ ] No `unsafe` without `// Safety:` proof comment
- [ ] No `TODO` without `proposals/pending/` reference
- [ ] No in-memory-only state that must survive restarts
- [ ] Migrations applied and verified against real database

---

## Rule 8: Test Both Paths

For every test, write BOTH:
1. Positive case: correct input → correct output
2. Negative case: bad input → correct error

The negative case is usually more valuable.
Never skip verification in a test.

---

## Rule 10: OTel End-to-End Verification

After adding any metric:
1. Ensure MeterProvider is initialized (not global no-op)
2. Start OTel collector
3. Exercise the code path
4. Check `otel-snapshots/latest.json` for the metric name
5. If absent → instrumentation is broken, fix before proceeding
