use crate::{ActionProposal, SafetyRail, SafetyVerdict};
use crate::tier1::Tier1SafetyRail;
use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use std::slice;

/// Opaque handle to a SafetyRail implementation.
pub struct SafetyRailHandle(Box<dyn SafetyRail>);

#[repr(C)]
pub struct C_SafetyVerdict {
    pub is_safe: bool,
    pub proposal_id: [u8; 16],
    pub duration_ms: u64,
    pub error_message: *mut c_char,
    pub proof_bytes: *mut u8,
    pub proof_len: usize,
}

#[repr(C)]
pub struct C_ExecutionResult {
    pub is_success: bool,
    pub artifact_id: [u8; 16],
    pub output_bytes: *mut u8,
    pub output_len: usize,
    pub exit_code: i32,
    pub elapsed_ms: u64,
    pub error_message: *mut c_char,
}

/// Create a new Tier 1 Safety Rail instance.
/// Returns null on failure.
#[no_mangle]
pub extern "C" fn safety_rail_new() -> *mut SafetyRailHandle {
    match Tier1SafetyRail::new(None) {
        Ok(rail) => Box::into_raw(Box::new(SafetyRailHandle(Box::new(rail)))),
        Err(_) => std::ptr::null_mut(),
    }
}

/// Free a Safety Rail instance.
#[no_mangle]
pub unsafe extern "C" fn safety_rail_free(handle: *mut SafetyRailHandle) {
    if !handle.is_null() {
        let _ = Box::from_raw(handle);
    }
}

/// Verify a proposal.
/// Caller must free error_message and proof_bytes if not null.
#[no_mangle]
pub unsafe extern "C" fn safety_rail_verify_proposal(
    handle: *mut SafetyRailHandle,
    proposal_id: *const u8, // [u8; 16]
    agent_id: *const c_char,
    description: *const c_char,
    payload: *const u8,
    payload_len: usize,
    payload_hash: *const u8, // [u8; 32]
    target_path: *const c_char,
    is_security_adjacent: bool,
    submitted_at_ms: u64,
) -> C_SafetyVerdict {
    let mut verdict = C_SafetyVerdict {
        is_safe: false,
        proposal_id: [0; 16],
        duration_ms: 0,
        error_message: std::ptr::null_mut(),
        proof_bytes: std::ptr::null_mut(),
        proof_len: 0,
    };

    if handle.is_null() {
        if let Ok(msg) = CString::new("Null handle") {
            verdict.error_message = msg.into_raw();
        }
        return verdict;
    }

    let rail = &*(handle);
    
    let id_slice = slice::from_raw_parts(proposal_id, 16);
    let mut id = [0u8; 16];
    id.copy_from_slice(id_slice);
    verdict.proposal_id = id;

    let agent_id_str = CStr::from_ptr(agent_id).to_string_lossy().into_owned();
    let description_str = CStr::from_ptr(description).to_string_lossy().into_owned();
    let payload_vec = slice::from_raw_parts(payload, payload_len).to_vec();
    
    let hash_slice = slice::from_raw_parts(payload_hash, 32);
    let mut hash = [0u8; 32];
    hash.copy_from_slice(hash_slice);

    let target_path_opt = if target_path.is_null() {
        None
    } else {
        Some(CStr::from_ptr(target_path).to_string_lossy().into_owned())
    };

    let proposal = ActionProposal {
        id,
        agent_id: agent_id_str,
        description: description_str,
        payload: payload_vec,
        payload_hash: hash,
        target_path: target_path_opt,
        is_security_adjacent,
        submitted_at_ms,
    };

    let result = rail.0.verify_proposal(&proposal);
    match result {
        SafetyVerdict::Safe { duration_ms, proof, .. } => {
            verdict.is_safe = true;
            verdict.duration_ms = duration_ms;
            let mut bytes = proof.bytes.clone();
            verdict.proof_len = bytes.len();
            verdict.proof_bytes = bytes.as_mut_ptr();
            std::mem::forget(bytes); // Caller must free
        }
        SafetyVerdict::Unsafe { violation, duration_ms, .. } => {
            verdict.is_safe = false;
            verdict.duration_ms = duration_ms;
            let msg = format!("Unsafe: {:?}", violation.violated_constraints);
            if let Ok(s) = CString::new(msg) {
                verdict.error_message = s.into_raw();
            }
        }
        SafetyVerdict::Timeout { elapsed_ms, .. } => {
            verdict.is_safe = false;
            verdict.duration_ms = elapsed_ms;
            if let Ok(s) = CString::new("Verification timeout") {
                verdict.error_message = s.into_raw();
            }
        }
        SafetyVerdict::TamperedPayload { .. } => {
            verdict.is_safe = false;
            if let Ok(s) = CString::new("Tampered payload") {
                verdict.error_message = s.into_raw();
            }
        }
    }

    verdict
}

/// Free a string returned by the API.
#[no_mangle]
pub unsafe extern "C" fn safety_rail_free_string(s: *mut c_char) {
    if !s.is_null() {
        let _ = CString::from_raw(s);
    }
}

/// Free a byte array returned by the API.
#[no_mangle]
pub unsafe extern "C" fn safety_rail_free_bytes(p: *mut u8, len: usize) {
    if !p.is_null() {
        let _ = Vec::from_raw_parts(p, len, len);
    }
}
