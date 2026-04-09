package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rds/sati-central/root-spine/internal/grpc/pb"
	"github.com/rds/sati-central/root-spine/internal/merkle"
	"github.com/rds/sati-central/root-spine/internal/orchestrator"
	"github.com/rds/sati-central/root-spine/internal/persistence"
	"github.com/rds/sati-central/root-spine/internal/safety"
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
}

// NewServer creates a new gRPC orchestrator server.
func NewServer(logger *zap.Logger, store *persistence.Store, safety *safety.Bridge, tree *merkle.Tree, analyst *orchestrator.VerdictManager) *Server {
	return &Server{
		logger:  logger,
		store:   store,
		safety:  safety,
		merkle:  tree,
		analyst: analyst,
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
	// (Actually we need a factory_id, for now we search or skip)
	// For this phase, let's assume global or stub it.
	fID := uuid.Nil // TODO: retrieve from context or agent mapping
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
	// copy(...) 

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
				ProofTier:     pb.SafetyTierProto_SAFETY_TIER_1,
				DurationMs:    int64(result.DurationMS),
			},
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

	// Update DB
	verdictStr := "UNSAFE"
	if result.IsSafe { verdictStr = "SAFE" }
	s.store.UpdateProposalVerdict(stream.Context(), pID, verdictStr, "", result.DurationMS, result.Proof)

	return nil
}

// ReadAnalystVerdict retrieves the latest automated analysis for an artifact.
func (s *Server) ReadAnalystVerdict(ctx context.Context, req *pb.VerdictQuery) (*pb.AnalystVerdict, error) {
	v, ok := s.analyst.GetVerdict(req.ArtifactId)
	if !ok {
		return &pb.AnalystVerdict{
			Verdict: pb.AnalystVerdict_VERDICT_PENDING,
			Rationale: "Pending analyst review...",
		}, nil
	}

	state := pb.AnalystVerdict_VERDICT_PENDING
	if v.State == orchestrator.VerdictApproved {
		state = pb.AnalystVerdict_VERDICT_APPROVED
	} else if v.State == orchestrator.VerdictVetoed {
		state = pb.AnalystVerdict_VERDICT_VETOED
	}

	return &pb.AnalystVerdict{
		Verdict: state,
		Rationale: v.Rationale,
		Date: v.ModTime.Format(time.RFC3339),
	}, nil
}
