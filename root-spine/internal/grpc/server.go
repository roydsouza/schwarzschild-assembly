package grpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rds/sati-central/root-spine/internal/gate"
	"github.com/rds/sati-central/root-spine/internal/grpc/pb"
	"github.com/rds/sati-central/root-spine/internal/merkle"
	"github.com/rds/sati-central/root-spine/internal/orchestrator"
	"github.com/rds/sati-central/root-spine/internal/persistence"
	"github.com/rds/sati-central/root-spine/internal/safety"
	"github.com/rds/sati-central/root-spine/internal/websocket"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the Orchestrator service.
type Server struct {
	pb.UnimplementedOrchestratorServer
	logger  *zap.Logger
	store   *persistence.Store
	safety  *safety.Bridge
	merkle  *merkle.Tree
	analyst *orchestrator.VerdictManager
	hub     *websocket.Hub
	gate    *gate.Gate
}

// NewServer creates a new gRPC orchestrator server.
func NewServer(logger *zap.Logger, store *persistence.Store, safety *safety.Bridge, tree *merkle.Tree, analyst *orchestrator.VerdictManager, hub *websocket.Hub, gate *gate.Gate) *Server {
	return &Server{
		logger:  logger,
		store:   store,
		safety:  safety,
		merkle:  tree,
		analyst: analyst,
		hub:     hub,
		gate:    gate,
	}
}

// CreateFactory registers a new agent factory.
func (s *Server) CreateFactory(ctx context.Context, req *pb.FactoryRequest) (*pb.FactoryResponse, error) {
	id := uuid.New()
	if req.RequestId != "" {
		if rid, err := uuid.Parse(req.RequestId); err == nil {
			id = rid
		}
	}

	f := persistence.Factory{
		ID:         id,
		Name:       req.FactoryName,
		Type:       req.FactoryType,
		ConfigJSON: req.ConfigJson,
		State:      "RUNNING",
	}

	actualID, err := s.store.GetOrCreateFactory(ctx, f)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register factory: %v", err)
	}

	return &pb.FactoryResponse{
		FactoryId: &pb.FactoryID{Id: actualID.String(), Name: req.FactoryName},
		Status:    &pb.FactoryStatus{State: pb.FactoryState_FACTORY_RUNNING},
	}, nil
}

// SubmitProposal handles the safety verification pipeline.
func (s *Server) SubmitProposal(req *pb.ActionProposal, stream pb.Orchestrator_SubmitProposalServer) error {
	pID, err := uuid.Parse(req.Id)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid proposal ID: %v", err)
	}

	// 1. Initial acknowledgment
	if err := stream.Send(&pb.VerificationEvent{
		ProposalId: req.Id,
		EventType:  pb.VerificationEventType_VERIFICATION_RECEIVED,
		TimestampMs: time.Now().UnixMilli(),
	}); err != nil {
		return err
	}

	// 2. Persist proposal
	fID, err := s.store.GetDefaultFactoryID(stream.Context())
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get default factory: %v", err)
	}

	if err := s.store.SaveProposal(stream.Context(), pID, fID, req.AgentId, req.Description, req.PayloadHash, req.IsSecurityAdjacent, time.UnixMilli(req.SubmittedAtMs)); err != nil {
		return status.Errorf(codes.Internal, "failed to persist proposal: %v", err)
	}

	// 3. Verify via Safety Rail
	if err := stream.Send(&pb.VerificationEvent{
		ProposalId: req.Id,
		EventType:  pb.VerificationEventType_VERIFICATION_CHECKING,
		TimestampMs: time.Now().UnixMilli(),
	}); err != nil {
		return err
	}

	var pIDBytes [16]byte
	copy(pIDBytes[:], pID[:])
	
	// Parse payload hash
	// For now, assume it's valid hex or dummy
	var hashBytes [32]byte
	if req.PayloadHash != "" {
		decoded, err := hex.DecodeString(req.PayloadHash)
		if err == nil && len(decoded) == 32 {
			copy(hashBytes[:], decoded)
		}
	}
	result, err := s.safety.VerifyProposal(
		pIDBytes,
		req.AgentId,
		req.Description,
		req.Payload,
		hashBytes,
		req.TargetPath,
		req.IsSecurityAdjacent,
		uint64(req.SubmittedAtMs),
	)

	if err != nil {
		return status.Errorf(codes.Internal, "safety verification failed: %v", err)
	}

	// 4. Handle verdict
	event := &pb.VerificationEvent{
		ProposalId: req.Id,
		TimestampMs: time.Now().UnixMilli(),
	}

	if result.IsSafe {
		event.EventType = pb.VerificationEventType_VERIFICATION_SAFE
		event.Result = &pb.VerificationEvent_SafeResult{
			SafeResult: &pb.SafeResult{
				ProofTier:           pb.SafetyTierProto_SAFETY_TIER_1,
				DurationMs:          int64(result.DurationMS),
				ProofCertificateHex: merkle.LeafHash(result.Proof).Hex(),
			},
		}

		// [CRITICAL-B] Wire Translucent Gate for security-adjacent proposals
		if s.gate.Route(req) {
			event.EventType = pb.VerificationEventType_VERIFICATION_GATE_PENDING
			if err := stream.Send(event); err != nil {
				return err
			}
			s.hub.Broadcast(event)
			
			// Update DB to PENDING
			s.store.UpdateProposalVerdict(stream.Context(), pID, "PENDING_APPROVAL", merkle.LeafHash(result.Proof).Hex(), result.DurationMS, result.Proof)
			return nil
		}
	} else {
		event.EventType = pb.VerificationEventType_VERIFICATION_UNSAFE
		event.Result = &pb.VerificationEvent_UnsafeResult{
			UnsafeResult: &pb.UnsafeResult{
				ViolationDetails: map[string]string{"policy": result.Error},
				DurationMs:       int64(result.DurationMS),
			},
		}
	}

	if err := stream.Send(event); err != nil {
		return err
	}
	s.hub.Broadcast(event)

	// Update DB
	verdictStr := "UNSAFE"
	if result.IsSafe {
		verdictStr = "SAFE"

		// [CRITICAL-D] Persist Merkle Leaf for Safe & Gate-Not-Required proposals
		// Note: Using proof hash as a surrogate for fingerprint in Phase 3
		fingerprint := merkle.LeafHash(result.Proof).Hex()
		leafData := []byte(fmt.Sprintf("%s:%s:%s", req.Id, verdictStr, fingerprint))
		leafHash := merkle.LeafHash(leafData)
		s.merkle.Append(leafHash)
		
		s.store.SaveMerkleLeaf(stream.Context(), int64(s.merkle.Size()-1), pID, leafHash.Hex(), "VERIFICATION_SAFE", fingerprint, s.merkle.Root().Hex())
		s.store.UpdateProposalVerdict(stream.Context(), pID, verdictStr, fingerprint, result.DurationMS, result.Proof)
	} else {
		s.store.UpdateProposalVerdict(stream.Context(), pID, verdictStr, "", result.DurationMS, nil)
	}

	return nil
}

// ReadAnalystVerdict retrieves the latest automated analysis for an artifact.
func (s *Server) ReadAnalystVerdict(ctx context.Context, req *pb.VerdictQuery) (*pb.AnalystVerdict, error) {
	var artifactID string
	switch q := req.Query.(type) {
	case *pb.VerdictQuery_ProposalId:
		artifactID = q.ProposalId
	case *pb.VerdictQuery_Topic:
		artifactID = q.Topic
	case *pb.VerdictQuery_BriefingId:
		artifactID = q.BriefingId
	default:
		return nil, status.Error(codes.InvalidArgument, "missing query field")
	}

	v, ok := s.analyst.GetVerdict(artifactID)
	if !ok {
		return &pb.AnalystVerdict{
			Verdict:   pb.VerdictDecision_VERDICT_UNSPECIFIED,
			Rationale: "Pending analyst review...",
		}, nil
	}

	var decision pb.VerdictDecision = pb.VerdictDecision_VERDICT_UNSPECIFIED
	if v.State == orchestrator.VerdictApproved {
		decision = pb.VerdictDecision_VERDICT_APPROVED
	} else if v.State == orchestrator.VerdictVetoed {
		decision = pb.VerdictDecision_VERDICT_VETOED
	}

	return &pb.AnalystVerdict{
		Verdict:    decision,
		Rationale:  v.Rationale,
		IssuedAtMs: v.ModTime.UnixMilli(),
	}, nil
}

// ApproveAction manual approval for is_security_adjacent proposals.
func (s *Server) ApproveAction(ctx context.Context, req *pb.ApprovalRequest) (*pb.MerkleProof, error) {
	pID, err := uuid.Parse(req.ProposalId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid proposal ID: %v", err)
	}

	// 1. Update DB to APPROVED
	err = s.store.UpdateProposalVerdict(ctx, pID, "APPROVED", "", 0, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to approve proposal: %v", err)
	}

	// 2. Commit Merkle Leaf
	leafData := []byte(fmt.Sprintf("%s:APPROVED", req.ProposalId))
	leafHash := merkle.LeafHash(leafData)
	s.merkle.Append(leafHash)
	s.store.SaveMerkleLeaf(ctx, int64(s.merkle.Size()-1), pID, leafHash.Hex(), "GATE_APPROVED", "", s.merkle.Root().Hex())

	// 3. Notify Hub
	s.hub.Broadcast(&pb.VerificationEvent{
		ProposalId:  req.ProposalId,
		EventType:   pb.VerificationEventType_VERIFICATION_GATE_APPROVED,
		TimestampMs: time.Now().UnixMilli(),
	})

	return &pb.MerkleProof{
		LeafHashHex: leafHash.Hex(),
		TreeSize:    int64(s.merkle.Size()),
		RootHashHex: s.merkle.Root().Hex(),
	}, nil
}

// VetoAction manual rejection.
func (s *Server) VetoAction(ctx context.Context, req *pb.VetoRequest) (*pb.OperationStatus, error) {
	pID, err := uuid.Parse(req.ProposalId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid proposal ID: %v", err)
	}

	err = s.store.UpdateProposalVerdict(ctx, pID, "VETOED", "", 0, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to veto proposal: %v", err)
	}

	// Commit Merkle Leaf
	leafData := []byte(fmt.Sprintf("%s:VETOED", req.ProposalId))
	leafHash := merkle.LeafHash(leafData)
	s.merkle.Append(leafHash)
	s.store.SaveMerkleLeaf(ctx, int64(s.merkle.Size()-1), pID, leafHash.Hex(), "GATE_DENIED", "", s.merkle.Root().Hex())

	// Notify Hub
	s.hub.Broadcast(&pb.VerificationEvent{
		ProposalId:  req.ProposalId,
		EventType:   pb.VerificationEventType_VERIFICATION_GATE_DENIED,
		TimestampMs: time.Now().UnixMilli(),
	})

	return &pb.OperationStatus{
		Success: true,
		Message: "Proposal vetoed at Translucent Gate",
	}, nil
}
