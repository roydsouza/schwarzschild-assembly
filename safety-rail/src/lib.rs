//! # Sati-Central Safety Rail — Trait Contract
//!
//! Authoritative public interface for the Sati-Central safety verification layer.
//! Version: 1.0.0
//!
//! All implementations — Tier 1 (Z3/Wasmtime) and Tier 2 (rocq-of-rust) —
//! must satisfy this contract. The trait is the guarantee.
//!
//! ## Architecture
//!
//! The safety layer forms the bottom of the trust stack. No component above it
//! may execute an unverified artifact or admit an unverified constraint. The
//! policy set is append-only and self-protecting: `register_constraint` passes
//! through `verify_proposal` before admission, preventing policy poisoning.
//!
//! ## Timing Contract
//!
//! `verify_proposal` must complete in < 100ms for Tier 1. This is a hard
//! contract enforced at the orchestrator level. Implementations that cannot
//! meet this budget must return `SafetyVerdict::Timeout`.
//!
//! ## Thread Safety
//!
//! All implementations must be `Send + Sync`. The orchestrator may parallelize
//! verification across goroutines via the gRPC boundary.

use std::collections::HashMap;

// ─────────────────────────────────────────────────────────────────────────────
// Identity types
// ─────────────────────────────────────────────────────────────────────────────

/// Unique identifier for a policy constraint. Wraps a UUID v7 (time-ordered).
#[derive(Debug, Clone, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct ConstraintId(pub [u8; 16]);

impl ConstraintId {
    /// Create a ConstraintId from a UUID v7 byte array.
    pub fn new(bytes: [u8; 16]) -> Self {
        Self(bytes)
    }

    /// Return the hex string representation (lowercase, no hyphens).
    pub fn to_hex(&self) -> String {
        self.0.iter().map(|b| format!("{b:02x}")).collect()
    }
}

impl std::fmt::Display for ConstraintId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let h = self.to_hex();
        write!(
            f,
            "{}-{}-{}-{}-{}",
            &h[0..8],
            &h[8..12],
            &h[12..16],
            &h[16..20],
            &h[20..32]
        )
    }
}

// ─────────────────────────────────────────────────────────────────────────────
// Policy fingerprint
// ─────────────────────────────────────────────────────────────────────────────

/// SHA-256 hash of the complete, serialized Z3 constraint set.
///
/// Two implementations with identical constraint sets in identical insertion
/// order MUST produce identical fingerprints. Written to every Merkle leaf
/// for auditability.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PolicyFingerprint {
    /// SHA-256 digest (32 bytes, raw).
    pub digest: [u8; 32],
    /// Number of constraints included in this fingerprint.
    pub constraint_count: u32,
    /// Unix epoch milliseconds when this fingerprint was last recomputed.
    pub computed_at_ms: u64,
}

impl PolicyFingerprint {
    /// Return the digest as a lowercase hex string (64 chars).
    pub fn hex(&self) -> String {
        self.digest.iter().map(|b| format!("{b:02x}")).collect()
    }

    /// Return the "empty policy" fingerprint — SHA-256 of the empty string.
    /// Used before any constraints are registered.
    pub fn empty() -> Self {
        // SHA-256("") =
        // e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
        Self {
            digest: [
                0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14,
                0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24,
                0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c,
                0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55,
            ],
            constraint_count: 0,
            computed_at_ms: 0,
        }
    }
}

// ─────────────────────────────────────────────────────────────────────────────
// Action proposal
// ─────────────────────────────────────────────────────────────────────────────

/// An agent-proposed action awaiting safety verification.
///
/// The caller is responsible for computing `payload_hash` from `payload`
/// before submission. The safety layer verifies consistency on receipt.
#[derive(Debug, Clone)]
pub struct ActionProposal {
    /// Unique proposal identifier (UUID v7 — time-ordered).
    pub id: [u8; 16],
    /// Submitting agent identifier (e.g., "antigravity", "root-spine").
    pub agent_id: String,
    /// Human-readable description of the proposed action. ≤ 512 chars.
    pub description: String,
    /// Serialized payload in RFC 8785 canonical JSON.
    pub payload: Vec<u8>,
    /// SHA-256 hash of `payload`, pre-computed by the caller.
    /// The safety layer verifies this matches before proceeding.
    pub payload_hash: [u8; 32],
    /// Target artifact path relative to repo root, if this proposal
    /// modifies a specific file (None for runtime-only proposals).
    pub target_path: Option<String>,
    /// True if this proposal touches a security-adjacent component.
    /// Security-adjacent: safety-rail/, merkle-log/ schema, auth logic,
    /// proto/ contracts, or CLAUDE.md. These always require Translucent Gate.
    pub is_security_adjacent: bool,
    /// Unix epoch milliseconds of proposal submission.
    pub submitted_at_ms: u64,
}

// ─────────────────────────────────────────────────────────────────────────────
// Proof certificate
// ─────────────────────────────────────────────────────────────────────────────

/// Proof certificate returned for a `SafetyVerdict::Safe` result.
#[derive(Debug, Clone)]
pub struct ProofCertificate {
    /// Tier of proof (Z3 satisfiability model or Rocq certificate).
    pub tier: SafetyTier,
    /// Opaque proof bytes.
    /// - Tier1: Z3 SMT satisfying model, serialized as SMT-LIB2 text.
    /// - Tier2: Serialized Rocq proof term (Coq .vo format).
    pub bytes: Vec<u8>,
    /// SHA-256 hash of `bytes`, for integrity verification.
    pub digest: [u8; 32],
}

// ─────────────────────────────────────────────────────────────────────────────
// Safety verdict
// ─────────────────────────────────────────────────────────────────────────────

/// The result of formal verification of an `ActionProposal`.
#[derive(Debug, Clone)]
pub enum SafetyVerdict {
    /// The proposal satisfies all constraints. Includes a proof certificate
    /// that can be embedded in the Merkle leaf for the subsequent GateApproved
    /// or FactoryCommit event.
    Safe {
        /// The proposal ID this verdict covers.
        proposal_id: [u8; 16],
        /// Proof of satisfiability.
        proof: ProofCertificate,
        /// Policy fingerprint at verification time.
        policy_fingerprint: PolicyFingerprint,
        /// Unix epoch milliseconds when verification completed.
        verified_at_ms: u64,
        /// Actual verification duration. Must be < 100ms for Tier 1.
        duration_ms: u64,
    },

    /// The proposal violates one or more constraints. Contains a structured
    /// report that the submitting agent must act on.
    Unsafe {
        /// The proposal ID this verdict covers.
        proposal_id: [u8; 16],
        /// Detailed violation report.
        violation: ViolationReport,
        /// Policy fingerprint at verification time.
        policy_fingerprint: PolicyFingerprint,
        /// Unix epoch milliseconds when verification completed.
        verified_at_ms: u64,
        /// Actual verification duration.
        duration_ms: u64,
    },

    /// Verification could not complete within the time budget (100ms for Tier 1).
    /// The proposal is NOT safe — a timeout is not an approval.
    Timeout {
        /// The proposal ID this verdict covers.
        proposal_id: [u8; 16],
        /// Time elapsed before timeout.
        elapsed_ms: u64,
        /// The configured budget that was exceeded.
        budget_ms: u64,
    },

    /// The payload hash in the proposal does not match the payload content.
    /// This indicates either a bug in the caller or a tampered proposal.
    TamperedPayload {
        /// The proposal ID this verdict covers.
        proposal_id: [u8; 16],
        /// The hash the caller claimed.
        claimed_hash: [u8; 32],
        /// The hash the safety layer computed.
        actual_hash: [u8; 32],
    },
}

impl SafetyVerdict {
    /// Return true only for `Safe` verdicts. All other variants are non-approvals.
    pub fn is_safe(&self) -> bool {
        matches!(self, SafetyVerdict::Safe { .. })
    }

    /// Extract the proposal ID from any verdict variant.
    pub fn proposal_id(&self) -> [u8; 16] {
        match self {
            SafetyVerdict::Safe { proposal_id, .. } => *proposal_id,
            SafetyVerdict::Unsafe { proposal_id, .. } => *proposal_id,
            SafetyVerdict::Timeout { proposal_id, .. } => *proposal_id,
            SafetyVerdict::TamperedPayload { proposal_id, .. } => *proposal_id,
        }
    }
}

// ─────────────────────────────────────────────────────────────────────────────
// Violation report
// ─────────────────────────────────────────────────────────────────────────────

/// Structured report of a safety constraint violation.
/// Provides enough information for the submitting agent to understand
/// what must change and how.
#[derive(Debug, Clone)]
pub struct ViolationReport {
    /// Map of violated constraint IDs to human-readable explanations.
    /// At least one entry is always present.
    pub violated_constraints: HashMap<ConstraintId, String>,
    /// The minimal unsatisfiable core from Z3, if the solver could compute it.
    /// Identifies the tightest set of constraints in conflict.
    /// None if Z3 could not compute an unsat core within the time budget.
    pub unsat_core: Option<Vec<ConstraintId>>,
    /// Suggested remediation if the safety layer can infer one from the
    /// violated constraint metadata. Not guaranteed to be present.
    pub remediation_hint: Option<String>,
    /// Severity of the most severe violated constraint.
    pub max_severity: ConstraintSeverity,
}

// ─────────────────────────────────────────────────────────────────────────────
// Verified artifact
// ─────────────────────────────────────────────────────────────────────────────

/// A proposal that has received a `SafetyVerdict::Safe` result.
///
/// Only a `VerifiedArtifact` may be passed to `execute_sandboxed`. This type
/// cannot be constructed outside of a `SafetyVerdict::Safe` match — it is the
/// type-level proof that verification occurred.
#[derive(Debug, Clone)]
pub struct VerifiedArtifact {
    /// The original proposal.
    pub proposal: ActionProposal,
    /// Proof certificate from verification.
    pub proof: ProofCertificate,
    /// Policy fingerprint at time of verification.
    pub policy_fingerprint: PolicyFingerprint,
    /// Unix epoch milliseconds when verification completed.
    pub verified_at_ms: u64,
}

impl VerifiedArtifact {
    /// Construct a VerifiedArtifact from a Safe verdict and its proposal.
    ///
    /// Returns `None` if the verdict is not `Safe` or if the proposal ID
    /// does not match the verdict's proposal ID.
    pub fn from_verdict(proposal: ActionProposal, verdict: SafetyVerdict) -> Option<Self> {
        match verdict {
            SafetyVerdict::Safe {
                proposal_id,
                proof,
                policy_fingerprint,
                verified_at_ms,
                ..
            } if proposal_id == proposal.id => Some(Self {
                proposal,
                proof,
                policy_fingerprint,
                verified_at_ms,
            }),
            _ => None,
        }
    }
}

// ─────────────────────────────────────────────────────────────────────────────
// Execution result
// ─────────────────────────────────────────────────────────────────────────────

/// Result of sandboxed execution of a verified artifact.
#[derive(Debug, Clone)]
pub enum ExecutionResult {
    /// Execution completed successfully within all sandbox limits.
    Success {
        /// Artifact ID (matches `VerifiedArtifact::proposal.id`).
        artifact_id: [u8; 16],
        /// Stdout bytes from the WASM module.
        output: Vec<u8>,
        /// Exit code (0 = clean exit).
        exit_code: i32,
        /// Wall-clock execution time in milliseconds.
        elapsed_ms: u64,
        /// Peak memory usage in bytes.
        peak_memory_bytes: u64,
    },

    /// Execution failed with a structured error.
    Failure {
        /// Artifact ID.
        artifact_id: [u8; 16],
        /// Category of failure.
        error_kind: ExecutionErrorKind,
        /// Human-readable detail.
        message: String,
        /// Process exit code if available.
        exit_code: Option<i32>,
        /// Elapsed time before failure.
        elapsed_ms: u64,
    },

    /// The artifact was executed but attempted a sandbox policy violation
    /// (e.g., forbidden syscall, network access). The sandbox caught and
    /// terminated it. This is distinct from a pre-execution `SafetyVerdict::Unsafe`.
    SandboxViolation {
        /// Artifact ID.
        artifact_id: [u8; 16],
        /// What the artifact attempted.
        violation: ViolationReport,
        /// Elapsed time before the violation was detected.
        elapsed_ms: u64,
    },
}

/// Categories of sandboxed execution failure.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ExecutionErrorKind {
    /// WASM compilation failed (malformed bytecode or unsupported features).
    CompilationError,
    /// WASM instantiation failed (missing imports, type mismatch).
    InstantiationError,
    /// WASM runtime trap (out-of-bounds memory, division by zero, unreachable).
    RuntimeTrap,
    /// Execution exceeded the 5-second time limit.
    Timeout,
    /// Execution exceeded the 256 MiB memory limit.
    MemoryExhausted,
    /// Artifact called a syscall not on the allow-list.
    ForbiddenSyscall,
    /// `execute_sandboxed` was called with an artifact whose proof certificate
    /// has expired (policy fingerprint changed since verification).
    StaleProof,
}

// ─────────────────────────────────────────────────────────────────────────────
// Policy constraint
// ─────────────────────────────────────────────────────────────────────────────

/// A policy constraint to be admitted to the Z3 constraint set.
///
/// Every constraint must have a non-empty `justification`. This is enforced
/// at registration time — the safety layer will reject a constraint with an
/// empty justification string.
#[derive(Debug, Clone)]
pub struct PolicyConstraint {
    /// Unique identifier for this constraint (UUID v7).
    pub id: ConstraintId,
    /// Short human-readable name (≤ 128 chars, no newlines).
    pub name: String,
    /// The constraint assertion in the form appropriate for the target tier.
    pub assertion: ConstraintAssertion,
    /// Domain category for grouping, reporting, and fitness vector attribution.
    pub category: ConstraintCategory,
    /// How violations are treated.
    pub severity: ConstraintSeverity,
    /// Non-empty justification for why this constraint exists.
    /// Required. The safety layer rejects constraints with empty justification.
    pub justification: String,
    /// Author of this constraint (agent_id of the submitter).
    pub author: String,
    /// Unix epoch milliseconds when this constraint was first authored.
    pub authored_at_ms: u64,
}

/// The assertion payload for a policy constraint.
#[derive(Debug, Clone)]
pub enum ConstraintAssertion {
    /// SMT-LIB2 assertion string for Z3 (Tier 1 and Tier 2).
    /// Must be a well-formed `(assert ...)` expression.
    SmtLib2(String),
    /// Serialized Rocq proof term for rocq-of-rust proofs (Tier 2 only).
    /// Tier 1 implementations must return `RegistrationResult::Rejected`
    /// with reason "tier2-only assertion not supported by this implementation"
    /// when presented with this variant.
    RocqTerm(Vec<u8>),
}

/// Domain category for a policy constraint.
/// Maps 1:1 with the global fitness vector metrics.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ConstraintCategory {
    /// Constraints enforcing safety compliance (fitness weight 0.30).
    SafetyCompliance,
    /// Constraints enforcing Merkle audit integrity (fitness weight 0.25).
    AuditIntegrity,
    /// Constraints enforcing Dhamma alignment (fitness weight 0.15).
    DhammaAlignment,
    /// Constraints enforcing system performance bounds (fitness weight 0.20).
    SystemPerformance,
    /// Constraints enforcing operational cost bounds (fitness weight 0.10).
    OperationalCost,
    /// Cross-cutting security constraints not mapped to a single metric.
    Security,
}

/// How violations of this constraint are treated.
#[derive(Debug, Clone, PartialEq, Eq, PartialOrd, Ord)]
pub enum ConstraintSeverity {
    /// Advisory: violation is logged and reported but does not block execution.
    /// Used for soft preferences and efficiency guidelines.
    Advisory,
    /// Mandatory: any violation produces `SafetyVerdict::Unsafe` and
    /// blocks the proposal from proceeding. Used for hard safety invariants.
    Mandatory,
}

/// Result of attempting to register a new constraint.
#[derive(Debug, Clone)]
pub enum RegistrationResult {
    /// Constraint was accepted and added to the active policy set.
    Accepted {
        /// The constraint ID that was accepted.
        id: ConstraintId,
        /// The new policy fingerprint after admission.
        new_fingerprint: PolicyFingerprint,
    },
    /// The constraint itself failed verification against the current policy set.
    /// This is the circular protection mechanism: a proposed constraint that
    /// would violate existing constraints is rejected.
    Rejected {
        /// The constraint ID that was rejected.
        id: ConstraintId,
        /// Human-readable rejection reason.
        reason: String,
        /// The violation that caused rejection (from the self-verification pass).
        violation: ViolationReport,
    },
    /// A constraint with this ID already exists in the policy set.
    Duplicate {
        /// The duplicate constraint ID.
        id: ConstraintId,
        /// The fingerprint of the existing constraint with this ID.
        existing_fingerprint: PolicyFingerprint,
    },
    /// The constraint had an empty `justification` field.
    MissingJustification {
        /// The constraint ID that was rejected.
        id: ConstraintId,
    },
    /// The `ConstraintAssertion` variant is not supported by this implementation tier.
    UnsupportedAssertionKind {
        /// The constraint ID.
        id: ConstraintId,
        /// Description of what is supported.
        supported: String,
    },
}

// ─────────────────────────────────────────────────────────────────────────────
// Safety tier
// ─────────────────────────────────────────────────────────────────────────────

/// Safety verification tier level.
///
/// Tier 2 is a strict superset of Tier 1: every Tier 2 implementation
/// provides all Tier 1 guarantees plus Rocq proof certificates.
#[derive(Debug, Clone, PartialEq, Eq, PartialOrd, Ord)]
pub enum SafetyTier {
    /// Z3 SMT satisfiability checking + Wasmtime sandboxed execution.
    /// 100ms hard timing budget. Required for production.
    Tier1,
    /// Tier 1 plus rocq-of-rust proof certificates for invariant preservation.
    /// Full machine-checkable proofs. Timing budget is relaxed for proof
    /// generation; the 100ms constraint applies only to `verify_proposal`.
    Tier2,
}

impl std::fmt::Display for SafetyTier {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            SafetyTier::Tier1 => write!(f, "Tier1"),
            SafetyTier::Tier2 => write!(f, "Tier2"),
        }
    }
}

// ─────────────────────────────────────────────────────────────────────────────
// The trait
// ─────────────────────────────────────────────────────────────────────────────

/// The canonical interface for the Sati-Central safety verification layer.
///
/// All implementations — Tier 1 (Z3/Wasmtime) and Tier 2 (rocq-of-rust) —
/// must satisfy this contract. The trait is the guarantee.
///
/// # Thread Safety
///
/// All implementations must be `Send + Sync`. The orchestrator parallelizes
/// verification calls across goroutine pools via the gRPC boundary.
///
/// # Invariants
///
/// 1. `verify_proposal` never panics.
/// 2. A `SafetyVerdict::Safe` verdict is valid only for the policy fingerprint
///    returned alongside it. If the policy changes before `execute_sandboxed`
///    is called, the artifact is stale and `execute_sandboxed` must return
///    `ExecutionResult::Failure { error_kind: ExecutionErrorKind::StaleProof }`.
/// 3. The constraint set is append-only. There is no `deregister_constraint`.
///    Revocation is achieved by registering an override constraint.
/// 4. `policy_fingerprint` is deterministic: the same set of constraints in the
///    same insertion order always produces the same fingerprint.
pub trait SafetyRail: Send + Sync {
    /// Verify a proposal against the compiled Z3 policy set.
    ///
    /// # Returns
    ///
    /// - `SafetyVerdict::Safe` — proposal satisfies all constraints.
    /// - `SafetyVerdict::Unsafe` — proposal violates ≥ 1 mandatory constraint.
    /// - `SafetyVerdict::Timeout` — verification exceeded the time budget.
    /// - `SafetyVerdict::TamperedPayload` — `payload_hash` does not match `payload`.
    ///
    /// # Timing
    ///
    /// Must complete in ≤ 100ms for Tier 1. This is a hard contract enforced
    /// by the orchestrator. Return `SafetyVerdict::Timeout` if the budget is
    /// exceeded rather than blocking the caller.
    ///
    /// # Panics
    ///
    /// Never. All error conditions are represented in the return type.
    fn verify_proposal(&self, proposal: &ActionProposal) -> SafetyVerdict;

    /// Sandbox and execute a verified artifact in a Wasmtime isolated environment.
    ///
    /// # Precondition
    ///
    /// The `artifact` must have been produced by `VerifiedArtifact::from_verdict`
    /// with a `SafetyVerdict::Safe`. If the artifact's `policy_fingerprint` does
    /// not match the current `policy_fingerprint()`, return
    /// `ExecutionResult::Failure { error_kind: ExecutionErrorKind::StaleProof }`.
    ///
    /// # Sandbox Limits
    ///
    /// - Filesystem: read/write restricted to paths declared in the proposal's
    ///   `target_path` only.
    /// - Network: none.
    /// - Memory: 256 MiB hard limit.
    /// - Time: 5 seconds hard limit.
    /// - Syscalls: Wasmtime default allow-list only.
    ///
    /// # Panics
    ///
    /// Never.
    fn execute_sandboxed(&self, artifact: &VerifiedArtifact) -> ExecutionResult;

    /// Register a new policy constraint.
    ///
    /// The constraint itself passes through `verify_proposal` before admission
    /// (circular protection against policy poisoning).
    ///
    /// # Atomicity
    ///
    /// If this returns `RegistrationResult::Accepted`, the new fingerprint is
    /// immediately visible to all subsequent `verify_proposal` and
    /// `policy_fingerprint` calls. Callers must treat `Accepted` as a memory
    /// barrier.
    ///
    /// # Panics
    ///
    /// Never.
    fn register_constraint(&self, constraint: &PolicyConstraint) -> RegistrationResult;

    /// Return the current policy fingerprint.
    ///
    /// SHA-256 of the complete serialized constraint set in deterministic
    /// insertion order. Two implementations with identical constraint sets
    /// MUST produce identical fingerprints. Written to every Merkle leaf.
    fn policy_fingerprint(&self) -> PolicyFingerprint;

    /// Return the tier level of this implementation.
    fn tier(&self) -> SafetyTier;
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests (trait contract compliance helpers)
// ─────────────────────────────────────────────────────────────────────────────

/// Contract compliance test helpers.
///
/// Implementations should call these from their test suites to verify they
/// correctly implement the trait contract.
#[cfg(test)]
pub mod contract_tests {
    use super::*;

    /// Verify that an empty-policy implementation returns the canonical
    /// empty-policy fingerprint.
    pub fn assert_empty_policy_fingerprint(rail: &impl SafetyRail) {
        let fp = rail.policy_fingerprint();
        assert_eq!(
            fp.constraint_count, 0,
            "empty policy must report constraint_count = 0"
        );
        assert_eq!(
            fp.digest,
            PolicyFingerprint::empty().digest,
            "empty policy fingerprint digest must match SHA-256 of empty string"
        );
    }

    /// Verify that verify_proposal never panics on a well-formed proposal.
    pub fn assert_verify_does_not_panic(rail: &impl SafetyRail, proposal: &ActionProposal) {
        // This should not panic regardless of verdict
        let _ = rail.verify_proposal(proposal);
    }

    /// Verify that execute_sandboxed returns StaleProof when the policy
    /// fingerprint in the artifact does not match the current policy.
    pub fn assert_stale_proof_rejected(rail: &impl SafetyRail, artifact: &VerifiedArtifact) {
        // Only call this when you know the policy has changed since verification.
        let result = rail.execute_sandboxed(artifact);
        match result {
            ExecutionResult::Failure { error_kind, .. } => {
                assert_eq!(
                    error_kind,
                    ExecutionErrorKind::StaleProof,
                    "execute_sandboxed must return StaleProof when policy fingerprint changed"
                );
            }
            other => panic!(
                "expected ExecutionResult::Failure(StaleProof), got {other:?}"
            ),
        }
    }

    /// Verify that a constraint with empty justification is rejected.
    pub fn assert_empty_justification_rejected(
        rail: &impl SafetyRail,
        mut constraint: PolicyConstraint,
    ) {
        constraint.justification = String::new();
        let result = rail.register_constraint(&constraint);
        assert!(
            matches!(result, RegistrationResult::MissingJustification { .. }),
            "register_constraint must reject constraints with empty justification"
        );
    }

    /// Verify that registering the same constraint ID twice returns Duplicate.
    pub fn assert_duplicate_rejected(rail: &impl SafetyRail, constraint: &PolicyConstraint) {
        // Assumes the constraint was already registered once.
        let result = rail.register_constraint(constraint);
        assert!(
            matches!(result, RegistrationResult::Duplicate { .. }),
            "register_constraint must return Duplicate for already-registered IDs"
        );
    }
}
