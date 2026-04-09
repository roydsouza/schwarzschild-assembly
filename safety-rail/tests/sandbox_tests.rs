use safety_rail::{ActionProposal, SafetyRail, VerifiedArtifact, ExecutionErrorKind, ExecutionResult};
use safety_rail::tier1::Tier1SafetyRail;
use safety_rail::tier1::z3_policy::{ProposalPayload, OperationType};
use wat;

fn setup_rail() -> Tier1SafetyRail {
    Tier1SafetyRail::new(None).expect("Failed to create Tier1SafetyRail")
}

fn create_verified_artifact(rail: &Tier1SafetyRail, wasm_bytes: Vec<u8>) -> VerifiedArtifact {
    let payload = ProposalPayload {
        operation_type: OperationType::ExecuteCode,
        target_component: "test-sandbox".to_string(),
        change_description: "Sandbox test".to_string(),
        context: None,
    };
    let _payload_bytes = serde_json::to_vec(&payload).unwrap();
    
    let proposal = ActionProposal {
        id: [0u8; 16],
        agent_id: "test-agent".to_string(),
        description: "Test description".to_string(),
        target_path: None,
        is_security_adjacent: true,
        payload: wasm_bytes, // We put the WASM directly in the payload for this test
        payload_hash: [0u8; 32],
        submitted_at_ms: 1775680800000,
    };
    
    VerifiedArtifact {
        proposal,
        proof: safety_rail::ProofCertificate {
            bytes: Vec::new(),
            digest: [0u8; 32],
            tier: safety_rail::SafetyTier::Tier1,
        },
        policy_fingerprint: rail.policy_fingerprint(),
        verified_at_ms: 1775680800000,
    }
}

#[tokio::test]
async fn test_sandbox_success() {
    let rail = setup_rail();
    // Valid minimal component WAT
    let wat = r#"
        (component
          (core module $M
            (func (export "f"))
          )
          (core instance $m (instantiate $M))
          (func (export "main") (canon lift (core func $m "f")))
        )
    "#;
    let wasm = wat::parse_str(wat).expect("Failed to parse WAT");
    let artifact = create_verified_artifact(&rail, wasm);
    
    let result = rail.execute_sandboxed(&artifact);
    match result {
        ExecutionResult::Success { .. } => {}
        _ => panic!("Expected success, got {:?}", result),
    }
}

#[tokio::test]
async fn test_sandbox_timeout_enforced() {
    let rail = setup_rail();
    // Component with infinite loop in core
    let wat = r#"
        (component
          (core module $M
            (func (export "f")
              (loop $infinite
                br $infinite
              )
            )
          )
          (core instance $m (instantiate $M))
          (func (export "main") (canon lift (core func $m "f")))
        )
    "#;
    let wasm = wat::parse_str(wat).expect("Failed to parse WAT");
    let artifact = create_verified_artifact(&rail, wasm);
    
    let result = rail.execute_sandboxed(&artifact);
    match result {
        ExecutionResult::Failure { error_kind, .. } => {
            assert_eq!(error_kind, ExecutionErrorKind::Timeout);
        }
        _ => panic!("Expected timeout failure, got {:?}", result),
    }
}

#[tokio::test]
async fn test_sandbox_memory_limit_enforced() {
    let rail = setup_rail();
    // Attempt to grow memory beyond 256MB (4096 pages)
    // 256 MiB = 4096 * 64 KiB
    let wat = r#"
        (module
            (memory 1)
            (func (export "main")
                (drop (memory.grow (i32.const 5000))) ;; Request ~312MB
            )
        )
    "#;
    let wasm = wat::parse_str(wat).expect("Failed to parse WAT");
    let artifact = create_verified_artifact(&rail, wasm);
    
    let result = rail.execute_sandboxed(&artifact);
    match result {
        ExecutionResult::Failure { .. } => {
        }
        _ => panic!("Expected failure due to memory limit, got {:?}", result),
    }
}
