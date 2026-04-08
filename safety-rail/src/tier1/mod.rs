use crate::{
    ActionProposal, ExecutionResult, PolicyConstraint, PolicyFingerprint, SafetyRail, 
    SafetyVerdict, VerifiedArtifact, ViolationReport, ConstraintSeverity
};
use crate::tier1::z3_policy::{Z3PolicyEngine, extract_facts};
use crate::tier1::sandbox::WasmSandbox;
use crate::tier1::fingerprint::compute_fingerprint;
use opentelemetry::{global, metrics::{Counter, Histogram}};
use std::time::Instant;

pub mod fingerprint;
pub mod sandbox;
pub mod z3_policy;

/// Tier 1 Safety Rail implementation using Z3 SMT and Wasmtime.
pub struct Tier1SafetyRail {
    z3_engine: Z3PolicyEngine,
    sandbox: WasmSandbox,
    // Metrics
    m_latency: Histogram<u64>,
    m_violations: Counter<u64>,
    m_sandbox_runs: Counter<u64>,
}

impl Tier1SafetyRail {
    pub fn new() -> Result<Self, String> {
        let mut z3_engine = Z3PolicyEngine::new();
        z3_engine.register_initial_constraints();
        
        let sandbox = WasmSandbox::new()?;
        
        let meter = global::meter("sati-central-safety-rail");
        let m_latency = meter
            .u64_histogram("sati_central.safety.verification_latency")
            .with_description("Latency of Z3 policy verification in microseconds")
            .init();
        let m_violations = meter
            .u64_counter("sati_central.safety.violations_total")
            .with_description("Total number of policy violations detected")
            .init();
        let m_sandbox_runs = meter
            .u64_counter("sati_central.safety.sandbox_executions_total")
            .with_description("Total number of sandboxed executions")
            .init();

        Ok(Self {
            z3_engine,
            sandbox,
            m_latency,
            m_violations,
            m_sandbox_runs,
        })
    }
}


impl SafetyRail for Tier1SafetyRail {
    fn verify_proposal(&self, proposal: &ActionProposal) -> SafetyVerdict {
        let start = Instant::now();
        let now_ms = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_millis() as u64;
        
        // 1. Extract facts from proposal
        let facts = match extract_facts(proposal) {
            Ok(f) => f,
            Err(e) => return SafetyVerdict::Unsafe {
                proposal_id: proposal.id,
                violation: ViolationReport {
                    violated_constraints: std::collections::HashMap::new(),
                    unsat_core: None,
                    remediation_hint: Some(format!("Failed to extract facts: {}", e)),
                    max_severity: ConstraintSeverity::Mandatory,
                },
                policy_fingerprint: self.policy_fingerprint(),
                verified_at_ms: now_ms,
                duration_ms: start.elapsed().as_millis() as u64,
            },
        };

        // 2. Run Z3 verification
        let result = match self.z3_engine.verify(&facts) {
            Ok(maybe_violation) => maybe_violation,
            Err(e) => return SafetyVerdict::Unsafe {
                proposal_id: proposal.id,
                violation: ViolationReport {
                    violated_constraints: std::collections::HashMap::new(),
                    unsat_core: None,
                    remediation_hint: Some(format!("Z3 Error: {}", e)),
                    max_severity: ConstraintSeverity::Mandatory,
                },
                policy_fingerprint: self.policy_fingerprint(),
                verified_at_ms: now_ms,
                duration_ms: start.elapsed().as_millis() as u64,
            },
        };

        let elapsed_ms = start.elapsed().as_millis() as u64;
        self.m_latency.record(start.elapsed().as_micros() as u64, &[]);

        if elapsed_ms > 100 {
            return SafetyVerdict::Timeout {
                proposal_id: proposal.id,
                elapsed_ms,
                budget_ms: 100,
            };
        }

        match result {
            None => SafetyVerdict::Safe {
                proposal_id: proposal.id,
                proof: crate::ProofCertificate {
                    bytes: Vec::new(), // Tier 1 emits dummy proof
                    digest: [0u8; 32],
                },
                policy_fingerprint: self.policy_fingerprint(),
                verified_at_ms: now_ms,
                duration_ms: elapsed_ms,
            },
            Some(violation) => {
                self.m_violations.add(1, &[]);
                SafetyVerdict::Unsafe {
                    proposal_id: proposal.id,
                    violation,
                    policy_fingerprint: self.policy_fingerprint(),
                    verified_at_ms: now_ms,
                    duration_ms: elapsed_ms,
                }
            }
        }
    }

    fn execute_sandboxed(&self, artifact: &VerifiedArtifact) -> ExecutionResult {
        self.m_sandbox_runs.add(1, &[]);
        self.sandbox.execute(artifact)
    }

    fn register_constraint(&self, constraint: &PolicyConstraint) -> crate::RegistrationResult {
        match self.z3_engine.add_constraint(constraint.clone()) {
            Ok(_) => crate::RegistrationResult::Accepted {
                id: constraint.id.clone(),
                new_fingerprint: self.policy_fingerprint(),
            },
            Err(e) => crate::RegistrationResult::Rejected {
                id: constraint.id.clone(),
                reason: e,
                violation: ViolationReport {
                    violated_constraints: std::collections::HashMap::new(),
                    unsat_core: None,
                    remediation_hint: None,
                    max_severity: ConstraintSeverity::Mandatory,
                },
            },
        }
    }

    fn policy_fingerprint(&self) -> PolicyFingerprint {
        compute_fingerprint(&self.z3_engine.constraints().lock().unwrap())
    }

    fn tier(&self) -> crate::SafetyTier {
        crate::SafetyTier::Tier1
    }
}
