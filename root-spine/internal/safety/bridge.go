package safety

/*
#cgo LDFLAGS: -L${SRCDIR}/../../../safety-rail/target/aarch64-apple-darwin/release -lsafety_rail
#cgo LDFLAGS: -L/opt/homebrew/lib -lz3 -lc++
#cgo LDFLAGS: -framework Security -framework CoreFoundation -framework SystemConfiguration

#include "../../../safety-rail/include/safety_rail.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"unsafe"
	"go.uber.org/zap"
)

// Bridge wraps the C-API to the Safety Rail.
type Bridge struct {
	handle *C.SafetyRailHandle
}

// NewBridge initializes the Safety Rail bridge.
func NewBridge(logger *zap.Logger, libPath string) (*Bridge, error) {
    // Note: libPath is currently ignored as the path is hardcoded in cgo LDFLAGS
    // but we keep the signature for main.go compatibility and future flexibility.
	handle := C.safety_rail_new()
	if handle == nil {
		return nil, errors.New("failed to initialize safety rail")
	}
	return &Bridge{handle: handle}, nil
}

// Close releases the safety rail handle.
func (b *Bridge) Close() {
	if b.handle != nil {
		C.safety_rail_free(b.handle)
		b.handle = nil
	}
}

// VerifyResult contains the outcome of a verification.
type VerifyResult struct {
	IsSafe     bool
	ProposalID [16]byte
	DurationMS uint64
	Error      string
	Proof      []byte
}

// VerifyProposal calls the Rust safety rail to verify an action proposal.
func (b *Bridge) VerifyProposal(
	proposalID [16]byte,
	agentID string,
	description string,
	payload []byte,
	payloadHash [32]byte,
	targetPath string,
	isSecurityAdjacent bool,
	submittedAtMS uint64,
) (*VerifyResult, error) {
	cAgentID := C.CString(agentID)
	defer C.free(unsafe.Pointer(cAgentID))

	cDescription := C.CString(description)
	defer C.free(unsafe.Pointer(cDescription))

	var cTargetPath *C.char
	if targetPath != "" {
		cTargetPath = C.CString(targetPath)
		defer C.free(unsafe.Pointer(cTargetPath))
	}

	payloadPtr := (*C.uint8_t)(unsafe.Pointer(&payload[0]))
	if len(payload) == 0 {
		payloadPtr = nil
	}

	cVerdict := C.safety_rail_verify_proposal(
		b.handle,
		(*C.uint8_t)(unsafe.Pointer(&proposalID[0])),
		cAgentID,
		cDescription,
		payloadPtr,
		C.size_t(len(payload)),
		(*C.uint8_t)(unsafe.Pointer(&payloadHash[0])),
		cTargetPath,
		C.bool(isSecurityAdjacent),
		C.uint64_t(submittedAtMS),
	)

	result := &VerifyResult{
		IsSafe:     bool(cVerdict.is_safe),
		ProposalID: proposalID,
		DurationMS: uint64(cVerdict.duration_ms),
	}

	if cVerdict.error_message != nil {
		result.Error = C.GoString(cVerdict.error_message)
		C.safety_rail_free_string(cVerdict.error_message)
	}

	if cVerdict.proof_bytes != nil {
		result.Proof = C.GoBytes(unsafe.Pointer(cVerdict.proof_bytes), C.int(cVerdict.proof_len))
		C.safety_rail_free_bytes(cVerdict.proof_bytes, cVerdict.proof_len)
	}

	return result, nil
}
