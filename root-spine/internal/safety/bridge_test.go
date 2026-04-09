package safety

import (
	"testing"
)

func TestBridge_Sanity(t *testing.T) {
	bridge, err := NewBridge()
	if err != nil {
		t.Fatalf("failed to create bridge: %v", err)
	}
	defer bridge.Close()

	proposalID := [16]byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
	payload := []byte(`{"action": "test"}`)
	
	// SHA-256 of payload
	payloadHash := [32]byte{}
	// (skipping actual hash for sanity check, safety rail should catch it)

	result, err := bridge.VerifyProposal(
		proposalID,
		"test-agent",
		"Test proposal",
		payload,
		payloadHash,
		"",
		false,
		1712580000000,
	)

	if err != nil {
		t.Fatalf("VerifyProposal failed: %v", err)
	}

	if result.IsSafe {
		t.Logf("Proposal approved! Proof len: %d", len(result.Proof))
	} else {
		t.Logf("Proposal rejected (expectedly): %s", result.Error)
	}
}
