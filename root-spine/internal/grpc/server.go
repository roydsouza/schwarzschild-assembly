package grpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	factory *orchestrator.SyntheticAnalystFactory
	hub     *websocket.Hub
	gate    *gate.Gate
}

// NewServer creates a new gRPC orchestrator server.
func NewServer(logger *zap.Logger, store *persistence.Store, safety *safety.Bridge, tree *merkle.Tree, analyst *orchestrator.VerdictManager, hub *websocket.Hub, gate *gate.Gate, factory *orchestrator.SyntheticAnalystFactory) *Server {
	return &Server{
		logger:  logger,
		store:   store,
		safety:  safety,
		merkle:  tree,
		analyst: analyst,
		factory: factory,
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

	// 2b. Trigger Autonomous Analysis (Phase 5.1)
	if s.factory != nil {
		go s.factory.AnalyzeProposal(context.Background(), req)
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
			if err := s.store.UpdateProposalVerdict(stream.Context(), pID, "PENDING_APPROVAL", merkle.LeafHash(result.Proof).Hex(), result.DurationMS, result.Proof); err != nil {
				s.logger.Error("failed to update proposal to PENDING", zap.Error(err))
				return status.Errorf(codes.Internal, "persistence failure: %v", err)
			}
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
		
		if err := s.store.SaveMerkleLeaf(stream.Context(), int64(s.merkle.Size()-1), pID, leafHash.Hex(), "VERIFICATION_SAFE", fingerprint, s.merkle.Root().Hex()); err != nil {
			s.logger.Error("failed to save merkle leaf", zap.Error(err))
			return status.Errorf(codes.Internal, "audit log persistence failure: %v", err)
		}
		if err := s.store.UpdateProposalVerdict(stream.Context(), pID, verdictStr, fingerprint, result.DurationMS, result.Proof); err != nil {
			s.logger.Error("failed to update proposal verdict", zap.Error(err))
			return status.Errorf(codes.Internal, "verdict persistence failure: %v", err)
		}
	} else {
		// [PHASE-5] Audit UNSAFE verdicts
		leafData := []byte(fmt.Sprintf("%s:UNSAFE:%s", req.Id, result.Error))
		leafHash := merkle.LeafHash(leafData)
		s.merkle.Append(leafHash)
		
		if err := s.store.SaveMerkleLeaf(stream.Context(), int64(s.merkle.Size()-1), pID, leafHash.Hex(), "VERIFICATION_UNSAFE", "", s.merkle.Root().Hex()); err != nil {
			s.logger.Error("failed to save unsafe merkle leaf", zap.Error(err))
		}

		if err := s.store.UpdateProposalVerdict(stream.Context(), pID, verdictStr, "", result.DurationMS, nil); err != nil {
			s.logger.Error("failed to update UNSAFE verdict", zap.Error(err))
			return status.Errorf(codes.Internal, "verdict persistence failure: %v", err)
		}
	}

	return nil
}

// WriteAnalystBriefing writes a briefing packet to analyst-inbox/.
func (s *Server) WriteAnalystBriefing(ctx context.Context, req *pb.Briefing) (*pb.OperationStatus, error) {
	filename := fmt.Sprintf("%s-%s.md", time.Now().Format("2006-01-02-150405"), req.Topic)
	path := filepath.Join("analyst-inbox", filename)

	content := fmt.Sprintf(`# Analyst Briefing: %s
**Date:** %s
**Author:** %s
**Phase:** %s
**Briefing ID:** %s

## Summary
%s

## Artifacts
%s

## Analyst Questions
%s

---
*Submitted via Sati-Central Root Spine (%s)*
`, req.Topic, time.Now().Format("2006-01-02 15:04:05 UTC"), req.Author, req.Phase, req.BriefingId, 
   req.SummaryMarkdown, "- " + strings.Join(req.Artifacts, "\n- "), "- " + strings.Join(req.Questions, "\n- "), req.BriefingId)

	if err := os.MkdirAll("analyst-inbox", 0755); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create inbox: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write briefing: %v", err)
	}

	s.logger.Info("analyst briefing published", zap.String("topic", req.Topic), zap.String("file", filename))

	return &pb.OperationStatus{
		Success: true,
		Message: fmt.Sprintf("Briefing %s published to analyst-inbox/", req.Topic),
	}, nil
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
	if err := s.store.SaveMerkleLeaf(ctx, int64(s.merkle.Size()-1), pID, leafHash.Hex(), "GATE_APPROVED", "", s.merkle.Root().Hex()); err != nil {
		s.logger.Error("failed to save approved merkle leaf", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "audit log persistence failure: %v", err)
	}

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
	if err := s.store.SaveMerkleLeaf(ctx, int64(s.merkle.Size()-1), pID, leafHash.Hex(), "GATE_DENIED", "", s.merkle.Root().Hex()); err != nil {
		s.logger.Error("failed to save veto merkle leaf", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "audit log persistence failure: %v", err)
	}

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

// RegisterDomainMetrics persists domain-specific metric schemas.
func (s *Server) RegisterDomainMetrics(ctx context.Context, req *pb.DomainFitnessExtension) (*pb.RegistrationResult, error) {
	fID, err := uuid.Parse(req.FactoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid factory ID: %v", err)
	}

	count := 0
	for _, m := range req.Metrics {
		// Fully qualified metric ID: factory_type.metric_id
		fqID := fmt.Sprintf("%s.%s", req.FactoryType, m.MetricId)
		err := s.store.SaveMetricDeclaration(ctx, fID, fqID, m.DisplayName, m.Description, m.Unit, m.Direction.String(), m.EscalationThreshold, m.EscalationOperator.String())
		if err != nil {
			s.logger.Error("failed to save metric declaration", zap.String("metric_id", fqID), zap.Error(err))
			continue
		}
		count++
	}

	s.logger.Info("factory metrics registered", zap.String("factory_id", req.FactoryId), zap.Int("count", count))

	return &pb.RegistrationResult{
		Success:         true,
		FactoryId:       req.FactoryId,
		RegisteredCount: int32(count),
	}, nil
}

// ReportMetrics records time-series metric data and broadcasts to the UI.
func (s *Server) ReportMetrics(ctx context.Context, req *pb.MetricReport) (*pb.OperationStatus, error) {
	observedAt := time.UnixMilli(req.ObservedAtMs)
	
	for id, val := range req.Metrics {
		// We use a simple status logic here; real logic would check the threshold
		// In Phase 5/6 we just mark it GREEN unless we implement the evaluator
		err := s.store.SaveMetricValue(ctx, id, val, "GREEN", observedAt)
		if err != nil {
			s.logger.Warn("failed to save metric value", zap.String("id", id), zap.Error(err))
		}
	}

	// Broadcast update to Control Panel Hub
	// We wrap the report in a generic event for now
	metricsJSON, _ := json.Marshal(req.Metrics)
	s.hub.Broadcast(&pb.VerificationEvent{
		ProposalId: req.FactoryId,
		EventType:  pb.VerificationEventType_VERIFICATION_RECEIVED, // Re-using event type for trigger
		TimestampMs: req.ObservedAtMs,
		Result: &pb.VerificationEvent_SafeResult{
			SafeResult: &pb.SafeResult{
				ProofCertificateHex: string(metricsJSON), // Using proof_cert as payload carrier for UI
			},
		},
	})

	return &pb.OperationStatus{
		Success: true,
		Message: "Metrics reported successfully",
	}, nil
}
