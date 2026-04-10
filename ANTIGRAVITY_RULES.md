# Schwarzschild Assembly: Operator Rules

## 0. Prime Directives

Every significant state change or session close must follow the centralized synchronization protocol:

1. **Local SYNC_LOG.md**: Always update `{PROJECT_ROOT}/SYNC_LOG.md` with a summary of the current session's accomplishments and the immediate next steps.
2. **Central Persistence**: Peer directories like `umbra/`, `penumbra/`, or `darkmatter/` are NO LONGER relevant for this project. The station boundary is absolute.
3. **Checkpoint Protocol**: When the user says "checkpoint", perform the following:
   - Update the local `./SYNC_LOG.md`.
   - Run `git add . && git commit -m "checkpoint: <summary>"` in the current project root.
   - The `umbra` persistence layer is decommissioned. All synchronization is localized.

## 1. Safety Rail Enforcements

- No raw `assert/retract` in STASIS code. Use `safe_assert/1`.
- All CGO logic must include Z3 policy verification.
- OTel metrics are mandatory for all production execution paths.

---

