use crate::{PolicyConstraint, PolicyFingerprint};
use sha2::{Digest, Sha256};
use std::time::{SystemTime, UNIX_EPOCH};

/// Compute a deterministic fingerprint for a set of policy constraints.
/// 
/// The constraints are sorted by their `ConstraintId` before hashing to ensure
/// that insertion order does not affect the fingerprint.
pub(crate) fn compute_fingerprint(constraints: &[PolicyConstraint]) -> PolicyFingerprint {
    if constraints.is_empty() {
        return PolicyFingerprint::empty();
    }

    // Sort constraints by ID to ensure deterministic hashing
    let mut sorted_constraints: Vec<&PolicyConstraint> = constraints.iter().collect();
    sorted_constraints.sort_by_key(|c| &c.id);

    let mut hasher = Sha256::new();
    
    // Hash each constraint
    // We hash: [ConstraintId (16 bytes) || NameLength (u32) || Name (bytes) || Severity (u8)]
    // This is a minimal set for Tier 1 fingerprinting. 
    // Tier 2 will require exhaustive serialization.
    for constraint in sorted_constraints {
        hasher.update(&constraint.id.0);
        hasher.update(&(constraint.name.len() as u32).to_le_bytes());
        hasher.update(constraint.name.as_bytes());
        
        let severity_byte = match constraint.severity {
            crate::ConstraintSeverity::Advisory => 0u8,
            crate::ConstraintSeverity::Mandatory => 1u8,
        };
        hasher.update(&[severity_byte]);
        
        // Add justification hash to ensure it's part of the fingerprint
        hasher.update(&(constraint.justification.len() as u32).to_le_bytes());
        hasher.update(constraint.justification.as_bytes());
    }

    let digest = hasher.finalize();
    let mut hash_bytes = [0u8; 32];
    hash_bytes.copy_from_slice(&digest);

    let now_ms = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_millis() as u64;

    PolicyFingerprint {
        digest: hash_bytes,
        constraint_count: constraints.len() as u32,
        computed_at_ms: now_ms,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{ConstraintId, ConstraintCategory, ConstraintSeverity, ConstraintAssertion};

    fn mock_constraint(id_byte: u8, name: &str) -> PolicyConstraint {
        let mut id_bytes = [0u8; 16];
        id_bytes[0] = id_byte;
        PolicyConstraint {
            id: ConstraintId::new(id_bytes),
            name: name.to_string(),
            assertion: ConstraintAssertion::SmtLib2("(assert true)".to_string()),
            category: ConstraintCategory::SafetyCompliance,
            severity: ConstraintSeverity::Mandatory,
            justification: "Test justification".to_string(),
            author: "test-agent".to_string(),
            authored_at_ms: 123456789,
        }
    }

    #[test]
    fn test_empty_fingerprint() {
        let fp = compute_fingerprint(&[]);
        assert_eq!(fp.digest, PolicyFingerprint::empty().digest);
        assert_eq!(fp.constraint_count, 0);
    }

    #[test]
    fn test_determinism() {
        let c1 = mock_constraint(1, "Alpha");
        let c2 = mock_constraint(2, "Beta");

        let fp1 = compute_fingerprint(&[c1.clone(), c2.clone()]);
        let fp2 = compute_fingerprint(&[c2.clone(), c1.clone()]);

        assert_eq!(fp1.digest, fp2.digest);
        assert_eq!(fp1.constraint_count, 2);
    }

    #[test]
    fn test_change_detection() {
        let c1 = mock_constraint(1, "Alpha");
        let mut c2 = mock_constraint(2, "Beta");
        
        let fp_base = compute_fingerprint(&[c1.clone(), c2.clone()]);
        
        // Change name
        c2.name = "Beta Modified".to_string();
        let fp_modified = compute_fingerprint(&[c1.clone(), c2.clone()]);
        
        assert_ne!(fp_base.digest, fp_modified.digest);
    }
}
