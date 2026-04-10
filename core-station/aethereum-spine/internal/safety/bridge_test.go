package safety

import (
	"crypto/sha256"
	"testing"
)

func TestVerifyProposal_Safe(t *testing.T) {
	bridge, err := NewBridge(nil, "")
	if err != nil {
		t.Fatalf("failed to create bridge: %v", err)
	}
	defer bridge.Close()

	proposalID := [16]byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
	// Valid JSON payload for Fact extraction
	payload := []byte(`{"operation_type": "read", "resource": "aethereum-spine/internal/safety/bridge.go"}`)
	
	h := sha256.Sum256(payload)
	payloadHash := [32]byte(h)

	result, err := bridge.VerifyProposal(
		proposalID,
		"test-agent",
		"Test valid proposal",
		payload,
		payloadHash,
		"aethereum-spine/internal/safety/bridge.go",
		false,
		1712580000000,
	)

	if err != nil {
		t.Fatalf("VerifyProposal failed: %v", err)
	}

	// We don't assert IsSafe is true because it depends on the policy loaded in the safety-rail,
	// but we assert that it did NOT return a TamperedPayload error.
	if result.Error == "Tampered payload" {
		t.Errorf("expected no tampering error, got: %s", result.Error)
	}
}

func TestVerifyProposal_Tampered(t *testing.T) {
	bridge, err := NewBridge(nil, "")
	if err != nil {
		t.Fatalf("failed to create bridge: %v", err)
	}
	defer bridge.Close()

	proposalID := [16]byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
	payload := []byte(`{"operation_type": "tamper"}`)
	
	// Intentionally wrong hash
	payloadHash := [32]byte{0xDE, 0xAD, 0xBE, 0xEF}

	result, err := bridge.VerifyProposal(
		proposalID,
		"test-agent",
		"Test tampered proposal",
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
		t.Fatal("expected proposal to be rejected as tampered, but it was approved")
	}

	if result.Error != "Tampered payload" {
		t.Errorf("expected Error 'Tampered payload', got: '%s'", result.Error)
	}
}
