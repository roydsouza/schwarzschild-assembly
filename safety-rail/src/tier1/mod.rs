use crate::{
    ActionProposal, ExecutionResult, PolicyConstraint, PolicyFingerprint, SafetyRail, 
    SafetyVerdict, VerifiedArtifact, ViolationReport, ConstraintSeverity, ConstraintId
};
use crate::tier1::z3_policy::{Z3PolicyEngine, extract_facts, OperationType};
use crate::tier1::sandbox::WasmSandbox;
use crate::tier1::fingerprint::compute_fingerprint;
use opentelemetry::{global, metrics::{Counter, Histogram, UpDownCounter}};
use sha2::{Digest, Sha256};
use std::time::Instant;
use opentelemetry_sdk::metrics::MeterProvider;

mod fingerprint;
mod sandbox;
pub mod z3_policy;
pub mod c_api;

/// Tier 1 Safety Rail implementation using Z3 SMT and Wasmtime.
pub struct Tier1SafetyRail {
    z3_engine: Z3PolicyEngine,
    sandbox: WasmSandbox,
    // Metrics
    m_latency: Histogram<u64>,
    m_violations: Counter<u64>,
    m_verifications: Counter<u64>,
    m_constraints: UpDownCounter<i64>,
    m_sandbox_runs: Counter<u64>,
}

impl Tier1SafetyRail {
    pub fn new(meter_provider: Option<MeterProvider>) -> Result<Self, String> {
        if let Some(provider) = meter_provider {
            global::set_meter_provider(provider);
        }

        let z3_engine = Z3PolicyEngine::new();
        let initial_rail = Self {
            z3_engine,
            sandbox: WasmSandbox::new()?,
            m_latency: global::meter("sati-central-safety-rail").u64_histogram("dummy").init(),
            m_violations: global::meter("sati-central-safety-rail").u64_counter("dummy").init(),
            m_verifications: global::meter("sati-central-safety-rail").u64_counter("dummy").init(),
            m_constraints: global::meter("sati-central-safety-rail").i64_up_down_counter("dummy").init(),
            m_sandbox_runs: global::meter("sati-central-safety-rail").u64_counter("dummy").init(),
        };
        
        initial_rail.register_initial_constraints();
        
        let meter = global::meter("sati-central-safety-rail");
        Ok(Self {
            z3_engine: initial_rail.z3_engine,
            sandbox: initial_rail.sandbox,
            m_latency: meter
                .u64_histogram("sati_central.safety.verification_latency")
                .with_description("Latency of Z3 policy verification in microseconds")
                .init(),
            m_violations: meter
                .u64_counter("sati_central.safety.violations_total")
                .with_description("Total number of policy violations detected")
                .init(),
            m_verifications: meter
                .u64_counter("sati_central.safety.verifications_total")
                .with_description("Total number of verification attempts")
                .init(),
            m_constraints: meter
                .i64_up_down_counter("sati_central.safety.constraints_total")
                .with_description("Current number of active constraints")
                .init(),
            m_sandbox_runs: meter
                .u64_counter("sati_central.safety.sandbox_executions_total")
                .with_description("Total number of sandboxed executions")
                .init(),
        })
    }

    fn register_initial_constraints(&self) {
        let _ = self.z3_engine.add_constraint(PolicyConstraint {
            id: ConstraintId::new([1u8; 16]),
            name: "safety_no_self_modify_safety_rail".to_string(),
            assertion: crate::ConstraintAssertion::SmtLib2(
                "(=> (str.prefixof \"safety-rail/\" target_path) is_security_adjacent)".to_string(),
            ),
            category: crate::ConstraintCategory::SafetyCompliance,
            severity: crate::ConstraintSeverity::Mandatory,
            justification: "No change to safety-rail/ may be executed unless it has passed verification".to_string(),
            author: "analyst-droid".to_string(),
            authored_at_ms: 1775680800000,
        });

        let _ = self.z3_engine.add_constraint(PolicyConstraint {
            id: ConstraintId::new([2u8; 16]),
            name: "audit_no_merkle_deletion".to_string(),
            assertion: crate::ConstraintAssertion::SmtLib2(
                "(=> (= operation_type \"delete_file\") (not (= target_component \"merkle-log\")))".to_string(),
            ),
            category: crate::ConstraintCategory::AuditIntegrity,
            severity: crate::ConstraintSeverity::Mandatory,
            justification: "Merkle log entries are append-only; no deletion operations are permitted".to_string(),
            author: "analyst-droid".to_string(),
            authored_at_ms: 1775680800001,
        });

        let _ = self.z3_engine.add_constraint(PolicyConstraint {
            id: ConstraintId::new([3u8; 16]),
            name: "security_no_unverified_proto_change".to_string(),
            assertion: crate::ConstraintAssertion::SmtLib2(
                "(=> (str.prefixof \"root-spine/proto/\" target_path) is_security_adjacent)".to_string(),
            ),
            category: crate::ConstraintCategory::Security,
            severity: crate::ConstraintSeverity::Mandatory,
            justification: "Proto contract changes are always security-adjacent and require Translucent Gate".to_string(),
            author: "analyst-droid".to_string(),
            authored_at_ms: 1775680800002,
        });

        let _ = self.z3_engine.add_constraint(PolicyConstraint {
            id: ConstraintId::new([4u8; 16]),
            name: "safety_no_prolog_injection".to_string(),
            assertion: crate::ConstraintAssertion::SmtLib2(
                "(=> (= target_component \"prolog-substrate\") (and (not (str.contains payload_content \"shell(\")) (not (str.contains payload_content \"system(\"))))".to_string(),
            ),
            category: crate::ConstraintCategory::SafetyCompliance,
            severity: crate::ConstraintSeverity::Mandatory,
            justification: "Autonomous knowledge evolution must not use OS-level shell or unauthorized mutation predicates".to_string(),
            author: "analyst-droid".to_string(),
            authored_at_ms: 1775754600000,
        });
    }
}

impl SafetyRail for Tier1SafetyRail {
    fn verify_proposal(&self, proposal: &ActionProposal) -> SafetyVerdict {
        self.m_verifications.add(1, &[]);
        let start = Instant::now();
        let now_ms = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_millis() as u64;
        
        let mut hasher = Sha256::new();
        hasher.update(&proposal.payload);
        let computed_hash = hasher.finalize();
        if computed_hash.as_slice() != proposal.payload_hash {
            return SafetyVerdict::TamperedPayload {
                proposal_id: proposal.id,
                claimed_hash: proposal.payload_hash,
                actual_hash: computed_hash.into(),
            };
        }

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
            (None, Some(model_str)) => {
                let proof_bytes = model_str.into_bytes();
                let mut hasher = Sha256::new();
                hasher.update(&proof_bytes);
                let digest: [u8; 32] = hasher.finalize().into();

                SafetyVerdict::Safe {
                    proposal_id: proposal.id,
                    proof: crate::ProofCertificate {
                        bytes: proof_bytes,
                        digest,
                        tier: crate::SafetyTier::Tier1,
                    },
                    policy_fingerprint: self.policy_fingerprint(),
                    verified_at_ms: now_ms,
                    duration_ms: elapsed_ms,
                }
            },
            (Some(violation), _) => {
                self.m_violations.add(1, &[]);
                SafetyVerdict::Unsafe {
                    proposal_id: proposal.id,
                    violation,
                    policy_fingerprint: self.policy_fingerprint(),
                    verified_at_ms: now_ms,
                    duration_ms: elapsed_ms,
                }
            },
            _ => {
                 SafetyVerdict::Unsafe {
                    proposal_id: proposal.id,
                    violation: ViolationReport {
                        violated_constraints: std::collections::HashMap::new(),
                        unsat_core: None,
                        remediation_hint: Some("Z3 logic error: Sat without model or Unsat without report".to_string()),
                        max_severity: ConstraintSeverity::Mandatory,
                    },
                    policy_fingerprint: self.policy_fingerprint(),
                    verified_at_ms: now_ms,
                    duration_ms: elapsed_ms,
                }
            }
        }
    }

    fn execute_sandboxed(&self, artifact: &VerifiedArtifact) -> ExecutionResult {
        if artifact.policy_fingerprint.digest != self.policy_fingerprint().digest {
            return ExecutionResult::Failure {
                artifact_id: artifact.proposal.id,
                error_kind: crate::ExecutionErrorKind::StaleProof,
                message: "Policy fingerprint has changed since verification; artifact is stale".to_string(),
                exit_code: None,
                elapsed_ms: 0,
            };
        }

        self.m_sandbox_runs.add(1, &[]);
        self.sandbox.execute(artifact)
    }

    fn register_constraint(&self, constraint: &PolicyConstraint) -> crate::RegistrationResult {
        // [FIX] REGRESSION-5: Guard against DoS vector by rejecting unknown constraint names
        // at the registration gate. This keeps verify() clean and invariant-safe.
        const SUPPORTED_CONSTRAINT_NAMES: &[&str] = &[
            "safety_no_self_modify_safety_rail",
            "audit_no_merkle_deletion",
            "security_no_unverified_proto_change",
            "safety_no_prolog_injection",
        ];
        if !SUPPORTED_CONSTRAINT_NAMES.contains(&constraint.name.as_str()) {
            return crate::RegistrationResult::UnsupportedAssertionKind {
                id: constraint.id.clone(),
                supported: "safety_no_self_modify_safety_rail, audit_no_merkle_deletion, security_no_unverified_proto_change".to_string(),
            };
        }

        let payload = crate::tier1::z3_policy::ProposalPayload {
            operation_type: OperationType::RegisterConstraint,
            target_component: match constraint.category {
                crate::ConstraintCategory::SafetyCompliance => "safety-rail",
                crate::ConstraintCategory::AuditIntegrity => "merkle-log",
                crate::ConstraintCategory::DhammaAlignment => "dhamma-adviser",
                crate::ConstraintCategory::SystemPerformance => "root-spine",
                crate::ConstraintCategory::OperationalCost => "factories",
                crate::ConstraintCategory::Security => "safety-rail",
            }.to_string(),
            change_description: format!("Registering constraint: {}", constraint.name),
            context: None,
        };
        let payload_bytes = serde_json::to_vec(&payload).unwrap();
        let mut hasher = Sha256::new();
        hasher.update(&payload_bytes);
        let payload_hash = hasher.finalize().into();

        let proposal = ActionProposal {
            id: [0u8; 16],
            agent_id: "safety-rail-internal".to_string(),
            description: "Self-verification of new constraint".to_string(),
            payload: payload_bytes,
            payload_hash,
            target_path: None,
            is_security_adjacent: true,
            submitted_at_ms: 0,
        };

        let current_fp = self.policy_fingerprint();
        {
            let constraints = self.z3_engine.constraints();
            if constraints.iter().any(|c| c.id == constraint.id) {
                return crate::RegistrationResult::Duplicate { 
                    id: constraint.id.clone(),
                    existing_fingerprint: current_fp,
                };
            }
        }

        let verdict = self.verify_proposal(&proposal);
        if !verdict.is_safe() {
            return crate::RegistrationResult::Rejected {
                id: constraint.id.clone(),
                reason: "Constraint failed self-verification pass".to_string(),
                violation: match verdict {
                    SafetyVerdict::Unsafe { violation, .. } => violation,
                    _ => ViolationReport {
                        violated_constraints: std::collections::HashMap::new(),
                        unsat_core: None,
                        remediation_hint: Some("Constraint registration timed out or failed hash check".to_string()),
                        max_severity: ConstraintSeverity::Mandatory,
                    },
                },
            };
        }

        match self.z3_engine.add_constraint(constraint.clone()) {
            Ok(_) => {
                let fp = self.policy_fingerprint();
                self.m_constraints.add(1, &[]);
                crate::RegistrationResult::Accepted {
                    id: constraint.id.clone(),
                    new_fingerprint: fp,
                }
            },
            Err(e) => {
                if e == "MissingJustification" {
                    crate::RegistrationResult::MissingJustification { id: constraint.id.clone() }
                } else {
                    crate::RegistrationResult::Rejected {
                        id: constraint.id.clone(),
                        reason: e.to_string(),
                        violation: ViolationReport {
                            violated_constraints: std::collections::HashMap::new(),
                            unsat_core: None,
                            remediation_hint: None,
                            max_severity: ConstraintSeverity::Mandatory,
                        },
                    }
                }
            }
        }
    }

    fn policy_fingerprint(&self) -> PolicyFingerprint {
        compute_fingerprint(&self.z3_engine.constraints())
    }

    fn tier(&self) -> crate::SafetyTier {
        crate::SafetyTier::Tier1
    }
}
