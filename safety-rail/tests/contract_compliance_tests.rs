use safety_rail::{
    ActionProposal, SafetyRail, VerifiedArtifact, SafetyVerdict, 
    ProofCertificate, ExecutionErrorKind, ExecutionResult,
    PolicyConstraint, ConstraintId, ConstraintAssertion, 
    ConstraintCategory, ConstraintSeverity
};
use safety_rail::tier1::Tier1SafetyRail;
use safety_rail::contract_tests;
use sha2::{Digest, Sha256};

#[tokio::test]
async fn test_stale_proof_rejected() {
    let rail = Tier1SafetyRail::new().unwrap();
    
    // 1. Create a safe artifact
    let payload = b"{\"operation_type\":\"modify_file\",\"target_component\":\"safety-rail\",\"change_description\":\"test\",\"context\":null}";
    let mut hasher = Sha256::new();
    hasher.update(payload);
    let payload_hash: [u8; 32] = hasher.finalize().into();

    let proposal = ActionProposal {
        id: [1u8; 16],
        agent_id: "test".to_string(),
        description: "test".to_string(),
        payload: payload.to_vec(),
        payload_hash,
        target_path: Some("safety-rail/src/lib.rs".to_string()),
        is_security_adjacent: true,
        submitted_at_ms: 0,
    };

    let verdict = rail.verify_proposal(&proposal);
    let artifact = VerifiedArtifact::from_verdict(proposal, verdict).unwrap();

    // 2. Change policy (adds a new constraint)
    let mut id_bytes = [0u8; 16];
    id_bytes[0] = 0xFF;
    rail.register_constraint(&PolicyConstraint {
        id: ConstraintId::new(id_bytes),
        name: "extra".to_string(),
        assertion: ConstraintAssertion::SmtLib2("(assert true)".to_string()),
        category: ConstraintCategory::Security,
        severity: ConstraintSeverity::Mandatory,
        justification: "extra justification".to_string(),
        author: "test".to_string(),
        authored_at_ms: 0,
    });

    // 3. Verify rejection
    contract_tests::assert_stale_proof_rejected(&rail, &artifact);
}

#[tokio::test]
async fn test_empty_justification_rejected() {
    let rail = Tier1SafetyRail::new().unwrap();
    let mut id_bytes = [0u8; 16];
    id_bytes[0] = 0xFE;
    let constraint = PolicyConstraint {
        id: ConstraintId::new(id_bytes),
        name: "test_empty".to_string(),
        assertion: ConstraintAssertion::SmtLib2("(assert true)".to_string()),
        category: ConstraintCategory::Security,
        severity: ConstraintSeverity::Mandatory,
        justification: "test".to_string(),
        author: "test".to_string(),
        authored_at_ms: 0,
    };
    contract_tests::assert_empty_justification_rejected(&rail, constraint);
}

#[tokio::test]
async fn test_duplicate_constraint_rejected() {
    let rail = Tier1SafetyRail::new().unwrap();
    let mut id_bytes = [0u8; 16];
    id_bytes[0] = 0xFD;
    let constraint = PolicyConstraint {
        id: ConstraintId::new(id_bytes),
        name: "test_dup".to_string(),
        assertion: ConstraintAssertion::SmtLib2("(assert true)".to_string()),
        category: ConstraintCategory::Security,
        severity: ConstraintSeverity::Mandatory,
        justification: "test".to_string(),
        author: "test".to_string(),
        authored_at_ms: 0,
    };
    
    rail.register_constraint(&constraint);
    let result = rail.register_constraint(&constraint);
    println!("Duplicate Result: {:?}", result);
    assert!(
        matches!(result, safety_rail::RegistrationResult::Duplicate { .. }),
        "register_constraint must return Duplicate for already-registered IDs, got {:?}", result
    );
}

#[tokio::test]
async fn test_verify_timing_under_100ms() {
    let rail = Tier1SafetyRail::new().unwrap();
    let payload = b"{\"operation_type\":\"modify_file\",\"target_component\":\"safety-rail\",\"change_description\":\"test\",\"context\":null}";
    let mut hasher = Sha256::new();
    hasher.update(payload);
    let payload_hash: [u8; 32] = hasher.finalize().into();

    let proposal = ActionProposal {
        id: [1u8; 16],
        agent_id: "test".to_string(),
        description: "test".to_string(),
        payload: payload.to_vec(),
        payload_hash,
        target_path: Some("safety-rail/src/lib.rs".to_string()),
        is_security_adjacent: true,
        submitted_at_ms: 0,
    };

    let mut durations = Vec::new();
    for _ in 0..100 {
        let start = std::time::Instant::now();
        rail.verify_proposal(&proposal);
        durations.push(start.elapsed());
    }
    
    durations.sort();
    let p99 = durations[98];
    assert!(p99.as_millis() < 100, "p99 verification time must be < 100ms, got {:?}", p99);
}

#[tokio::test]
async fn test_mandatory_constraint_violation_produces_unsafe_verdict() {
    let rail = Tier1SafetyRail::new().unwrap();
    
    // Violates Constraint 1: safety-rail modify without is_security_adjacent=true
    let payload = b"{\"operation_type\":\"modify_file\",\"target_component\":\"safety-rail\",\"change_description\":\"test\",\"context\":null}";
    let mut hasher = Sha256::new();
    hasher.update(payload);
    let payload_hash: [u8; 32] = hasher.finalize().into();

    let proposal = ActionProposal {
        id: [1u8; 16],
        agent_id: "test".to_string(),
        description: "test".to_string(),
        payload: payload.to_vec(),
        payload_hash,
        target_path: Some("safety-rail/src/lib.rs".to_string()),
        is_security_adjacent: false, // VIOLATION
        submitted_at_ms: 0,
    };

    let verdict = rail.verify_proposal(&proposal);
    assert!(matches!(verdict, SafetyVerdict::Unsafe { .. }), "Expected Unsafe verdict, got {:?}", verdict);
}

#[tokio::test]
async fn test_tampered_payload_rejected() {
    let rail = Tier1SafetyRail::new().unwrap();
    let payload = b"{\"operation_type\":\"modify_file\",\"target_component\":\"safety-rail\",\"change_description\":\"test\",\"context\":null}";
    let payload_hash = [0u8; 32]; // WRONG HASH

    let proposal = ActionProposal {
        id: [1u8; 16],
        agent_id: "test".to_string(),
        description: "test".to_string(),
        payload: payload.to_vec(),
        payload_hash,
        target_path: Some("safety-rail/src/lib.rs".to_string()),
        is_security_adjacent: true,
        submitted_at_ms: 0,
    };

    let verdict = rail.verify_proposal(&proposal);
    assert!(matches!(verdict, SafetyVerdict::TamperedPayload { .. }), "Expected TamperedPayload, got {:?}", verdict);
}
