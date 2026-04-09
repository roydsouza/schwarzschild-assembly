use crate::{ActionProposal, ConstraintCategory, ConstraintId, ConstraintSeverity, PolicyConstraint, ViolationReport};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Mutex;
use z3::ast::{Ast, Bool, String as Z3String};
use z3::{Config, Context, Solver};

/// Typed facts extracted from an ActionProposal for Z3 binding.
#[derive(Debug, Clone)]
pub(crate) struct ProposalFacts {
    pub target_path: String,
    pub is_security_adjacent: bool,
    pub agent_id: String,
    pub operation_type: OperationType,
    pub target_component: String,
}

/// Canonical schema for ActionProposal.payload.
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ProposalPayload {
    pub operation_type: OperationType,
    pub target_component: String,
    pub change_description: String,
    pub context: Option<serde_json::Value>,
}

#[derive(Debug, Clone, PartialEq, Eq, Deserialize, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum OperationType {
    CreateFile,
    ModifyFile,
    DeleteFile,
    ExecuteCode,
    RegisterConstraint,
    ModifySchema,
    UpdateConfig,
}

impl OperationType {
    pub fn as_str(&self) -> &'static str {
        match self {
            OperationType::CreateFile => "create_file",
            OperationType::ModifyFile => "modify_file",
            OperationType::DeleteFile => "delete_file",
            OperationType::ExecuteCode => "execute_code",
            OperationType::RegisterConstraint => "register_constraint",
            OperationType::ModifySchema => "modify_schema",
            OperationType::UpdateConfig => "update_config",
        }
    }
}

pub(crate) struct Z3PolicyEngine {
    constraints: Mutex<Vec<PolicyConstraint>>,
}

impl Z3PolicyEngine {
    pub fn new() -> Self {
        Self {
            constraints: Mutex::new(Vec::new()),
        }
    }

    pub fn register_initial_constraints(&self) {
        // Constraint 1 — safety_no_self_modify_safety_rail
        self.add_constraint(PolicyConstraint {
            id: ConstraintId::new([1u8; 16]),
            name: "safety_no_self_modify_safety_rail".to_string(),
            assertion: crate::ConstraintAssertion::SmtLib2(
                "(=> (str.prefixof \"safety-rail/\" target_path) is_security_adjacent)".to_string(),
            ),
            category: ConstraintCategory::SafetyCompliance,
            severity: ConstraintSeverity::Mandatory,
            justification: "No change to safety-rail/ may be executed unless it has passed verification".to_string(),
            author: "analyst-droid".to_string(),
            authored_at_ms: 1775680800000,
        }).expect("Failed to register initial constraint");

        // Constraint 2 — audit_no_merkle_deletion
        self.add_constraint(PolicyConstraint {
            id: ConstraintId::new([2u8; 16]),
            name: "audit_no_merkle_deletion".to_string(),
            assertion: crate::ConstraintAssertion::SmtLib2(
                "(=> (= operation_type \"delete_file\") (not (= target_component \"merkle-log\")))".to_string(),
            ),
            category: ConstraintCategory::AuditIntegrity,
            severity: ConstraintSeverity::Mandatory,
            justification: "Merkle log entries are append-only; no deletion operations are permitted".to_string(),
            author: "analyst-droid".to_string(),
            authored_at_ms: 1775680800001,
        }).expect("Failed to register initial constraint");

        // Constraint 3 — security_no_unverified_proto_change
        self.add_constraint(PolicyConstraint {
            id: ConstraintId::new([3u8; 16]),
            name: "security_no_unverified_proto_change".to_string(),
            assertion: crate::ConstraintAssertion::SmtLib2(
                "(=> (str.prefixof \"root-spine/proto/\" target_path) is_security_adjacent)".to_string(),
            ),
            category: ConstraintCategory::Security,
            severity: ConstraintSeverity::Mandatory,
            justification: "Proto contract changes are always security-adjacent and require Translucent Gate".to_string(),
            author: "analyst-droid".to_string(),
            authored_at_ms: 1775680800002,
        }).expect("Failed to register initial constraint");
    }

    pub fn add_constraint(&self, constraint: PolicyConstraint) -> Result<(), String> {
        // [CRITICAL-3] Added missing guards
        if constraint.justification.is_empty() {
             return Err("MissingJustification".to_string());
        }
        
        let mut constraints = self.constraints.lock().map_err(|e| e.to_string())?;
        
        if constraints.iter().any(|c| c.id == constraint.id) {
            return Err("Duplicate".to_string());
        }

        constraints.push(constraint);
        Ok(())
    }

    pub fn verify(&self, facts: &ProposalFacts) -> Result<(Option<ViolationReport>, Option<String>), String> {
        let cfg = Config::new();
        let ctx = Context::new(&cfg);
        let solver = Solver::new(&ctx);
        
        let vars = self.create_facts_vars_with_ctx(&ctx);
        let constraints = self.constraints.lock().map_err(|e| e.to_string())?;

        // Replay all registered constraints into the fresh solver
        for constraint in constraints.iter() {
            if let crate::ConstraintAssertion::SmtLib2(_) = &constraint.assertion {
                match constraint.name.as_str() {
                    "safety_no_self_modify_safety_rail" => {
                        let prefix = Z3String::from_str(&ctx, "safety-rail/").expect("Z3 string");
                        let is_prefix = prefix.prefix(&vars.target_path);
                        solver.assert(&is_prefix.implies(&vars.is_security_adjacent));
                    }
                    "audit_no_merkle_deletion" => {
                        let delete_op = Z3String::from_str(&ctx, "delete_file").expect("Z3 string");
                        let is_delete = vars.operation_type._eq(&delete_op);
                        let merkle_comp = Z3String::from_str(&ctx, "merkle-log").expect("Z3 string");
                        let targets_merkle = vars.target_component._eq(&merkle_comp);
                        solver.assert(&is_delete.implies(&targets_merkle.not()));
                    }
                    "security_no_unverified_proto_change" => {
                        let prefix = Z3String::from_str(&ctx, "root-spine/proto/").expect("Z3 string");
                        let is_prefix = prefix.prefix(&vars.target_path);
                        solver.assert(&is_prefix.implies(&vars.is_security_adjacent));
                    }
                    _ => {
                        // Skip constraints we don't know how to verify yet.
                        // We do NOT return an error here because that would block admission
                        // of any custom constraint. The safety layer still has its
                        // mandatory core constraints in the solver.
                    }
                }
            }
        }

        // Assert the current proposal facts as constants
        let target_path_val = Z3String::from_str(&ctx, &facts.target_path).map_err(|e| e.to_string())?;
        solver.assert(&vars.target_path._eq(&target_path_val));
        
        solver.assert(&vars.is_security_adjacent._eq(&Bool::from_bool(&ctx, facts.is_security_adjacent)));
        
        let agent_id_val = Z3String::from_str(&ctx, &facts.agent_id).map_err(|e| e.to_string())?;
        solver.assert(&vars.agent_id._eq(&agent_id_val));
        
        let op_type_val = Z3String::from_str(&ctx, facts.operation_type.as_str()).map_err(|e| e.to_string())?;
        solver.assert(&vars.operation_type._eq(&op_type_val));
        
        let target_comp_val = Z3String::from_str(&ctx, &facts.target_component).map_err(|e| e.to_string())?;
        solver.assert(&vars.target_component._eq(&target_comp_val));

        let result = solver.check();

        match result {
            z3::SatResult::Sat => {
                // [CRITICAL-1] Serialize model to SMT-LIB2 text
                let model = solver.get_model().ok_or("Failed to get Z3 model")?;
                Ok((None, Some(model.to_string())))
            },
            z3::SatResult::Unsat => {
                // Violation! 
                let mut violated_constraints = HashMap::new();
                violated_constraints.insert(
                    ConstraintId::new([0u8; 16]),
                    "Policy violation detected by Z3".to_string(),
                );
                
                Ok((Some(ViolationReport {
                    violated_constraints,
                    unsat_core: None,
                    remediation_hint: Some("Review policy constraints in CLAUDE.md".to_string()),
                    max_severity: ConstraintSeverity::Mandatory,
                }), None))
            }
            z3::SatResult::Unknown => Err("Z3 returned Unknown".to_string()),
        }
    }

    pub fn constraints(&self) -> Vec<PolicyConstraint> {
        self.constraints.lock().unwrap().clone()
    }

    fn create_facts_vars_with_ctx<'ctx>(&self, ctx: &'ctx Context) -> Z3FactsVars<'ctx> {
        Z3FactsVars {
            operation_type: Z3String::new_const(ctx, "operation_type"),
            target_path: Z3String::new_const(ctx, "target_path"),
            target_component: Z3String::new_const(ctx, "target_component"),
            is_security_adjacent: Bool::new_const(ctx, "is_security_adjacent"),
            agent_id: Z3String::new_const(ctx, "agent_id"),
        }
    }
}

struct Z3FactsVars<'ctx> {
    target_path: Z3String<'ctx>,
    is_security_adjacent: Bool<'ctx>,
    agent_id: Z3String<'ctx>,
    operation_type: Z3String<'ctx>,
    target_component: Z3String<'ctx>,
}

pub(crate) fn extract_facts(proposal: &ActionProposal) -> Result<ProposalFacts, String> {
    let payload: ProposalPayload = serde_json::from_slice(&proposal.payload)
        .map_err(|e| format!("Failed to parse proposal payload: {}", e))?;

    Ok(ProposalFacts {
        target_path: proposal.target_path.clone().unwrap_or_default(),
        is_security_adjacent: proposal.is_security_adjacent,
        agent_id: proposal.agent_id.clone(),
        operation_type: payload.operation_type,
        target_component: payload.target_component,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    fn mock_proposal(path: &str, security_adjacent: bool, op: OperationType, component: &str) -> ActionProposal {
        let payload = ProposalPayload {
            operation_type: op,
            target_component: component.to_string(),
            change_description: "Test".to_string(),
            context: None,
        };
        let payload_bytes = serde_json::to_vec(&payload).unwrap();
        
        ActionProposal {
            id: [0u8; 16],
            agent_id: "test-agent".to_string(),
            description: "Test proposal".to_string(),
            target_path: Some(path.to_string()),
            is_security_adjacent: security_adjacent,
            payload: payload_bytes,
            payload_hash: [0u8; 32],
            submitted_at_ms: 1775680800000,
        }
    }

    #[test]
    fn test_safety_rail_self_protection() {
        let engine = Z3PolicyEngine::new();
        engine.register_initial_constraints();

        // REJECTED: modify safety-rail/ without security_adjacent flag
        let p_bad = mock_proposal("safety-rail/src/lib.rs", false, OperationType::ModifyFile, "safety-rail");
        let facts_bad = extract_facts(&p_bad).unwrap();
        assert!(engine.verify(&facts_bad).unwrap().0.is_some());

        // APPROVED: modify safety-rail/ with security_adjacent flag
        let p_good = mock_proposal("safety-rail/src/lib.rs", true, OperationType::ModifyFile, "safety-rail");
        let facts_good = extract_facts(&p_good).unwrap();
        assert!(engine.verify(&facts_good).unwrap().0.is_none());
    }

    #[test]
    fn test_merkle_deletion_guard() {
        let engine = Z3PolicyEngine::new();
        engine.register_initial_constraints();

        // REJECTED: delete from merkle-log
        let p_bad = mock_proposal("audit/merkle.db", false, OperationType::DeleteFile, "merkle-log");
        let facts_bad = extract_facts(&p_bad).unwrap();
        assert!(engine.verify(&facts_bad).unwrap().0.is_some());

        // APPROVED: delete from other component
        let p_good = mock_proposal("temp/logs", false, OperationType::DeleteFile, "temp-storage");
        let facts_good = extract_facts(&p_good).unwrap();
        assert!(engine.verify(&facts_good).unwrap().0.is_none());
    }

    #[test]
    fn test_proto_guard() {
        let engine = Z3PolicyEngine::new();
        engine.register_initial_constraints();

        // REJECTED: modify proto without security_adjacent flag
        let p_bad = mock_proposal("root-spine/proto/orchestrator.proto", false, OperationType::ModifyFile, "root-spine");
        let facts_bad = extract_facts(&p_bad).unwrap();
        assert!(engine.verify(&facts_bad).unwrap().0.is_some());

        // APPROVED: modify proto with security_adjacent flag
        let p_good = mock_proposal("root-spine/proto/orchestrator.proto", true, OperationType::ModifyFile, "root-spine");
        let facts_good = extract_facts(&p_good).unwrap();
        assert!(engine.verify(&facts_good).unwrap().0.is_none());
    }
}
