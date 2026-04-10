#ifndef SAFETY_RAIL_H
#define SAFETY_RAIL_H

#include <stdint.h>
#include <stdbool.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// Opaque handle to a SafetyRail instance.
typedef struct SafetyRailHandle SafetyRailHandle;

// Safety Verdict structure.
// Mirrors rust::c_api::C_SafetyVerdict.
typedef struct {
    bool is_safe;
    uint8_t proposal_id[16];
    uint64_t duration_ms;
    char* error_message; // Must be freed with safety_rail_free_string if not NULL.
    uint8_t* proof_bytes; // Must be freed with safety_rail_free_bytes if not NULL.
    size_t proof_len;
} C_SafetyVerdict;

// Execution Result structure.
// Mirrors rust::c_api::C_ExecutionResult.
typedef struct {
    bool is_success;
    uint8_t artifact_id[16];
    uint8_t* output_bytes; // Must be freed if not NULL.
    size_t output_len;
    int32_t exit_code;
    uint64_t elapsed_ms;
    char* error_message; // Must be freed if not NULL.
} C_ExecutionResult;

// Function Prototypes

// Create a new Tier 1 Safety Rail instance.
SafetyRailHandle* safety_rail_new();

// Free a Safety Rail instance.
void safety_rail_free(SafetyRailHandle* handle);

// Verify a proposal.
C_SafetyVerdict safety_rail_verify_proposal(
    SafetyRailHandle* handle,
    const uint8_t* proposal_id, // [16]
    const char* agent_id,
    const char* description,
    const uint8_t* payload,
    size_t payload_len,
    const uint8_t* payload_hash, // [32]
    const char* target_path,
    bool is_security_adjacent,
    uint64_t submitted_at_ms
);

// Free a string returned by the API.
void safety_rail_free_string(char* s);

// Free a byte array returned by the API.
void safety_rail_free_bytes(uint8_t* p, size_t len);

#ifdef __cplusplus
}
#endif

#endif // SAFETY_RAIL_H
