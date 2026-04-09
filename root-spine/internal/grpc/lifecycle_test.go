package grpc

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rds/sati-central/root-spine/internal/grpc/pb"
	"github.com/rds/sati-central/root-spine/internal/persistence"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockStore struct {
	Store
	assemblyLines map[uuid.UUID]persistence.AssemblyLine
	specs         map[string]persistence.SpecDocument
}

func (m *mockStore) GetAssemblyLine(ctx context.Context, id uuid.UUID) (persistence.AssemblyLine, error) {
	al, ok := m.assemblyLines[id]
	if !ok {
		return persistence.AssemblyLine{}, status.Error(codes.NotFound, "not found")
	}
	return al, nil
}

func (m *mockStore) GetSpecDocument(ctx context.Context, name string) (persistence.SpecDocument, error) {
	spec, ok := m.specs[name]
	if !ok {
		return persistence.SpecDocument{}, status.Error(codes.NotFound, "not found")
	}
	return spec, nil
}

func (m *mockStore) UpdateAssemblyLineState(ctx context.Context, id uuid.UUID, state, just string) (string, error) {
	al := m.assemblyLines[id]
	old := al.CurrentState
	al.CurrentState = state
	m.assemblyLines[id] = al
	return old, nil
}

func TestAdvanceLifecycle_Gates(t *testing.T) {
	alID := uuid.New()
	serviceName := "test-service"
	
	tests := []struct {
		name         string
		currentState string
		targetState  pb.LifecycleState
		isFinalized  bool
		expectedCode codes.Code
	}{
		{
			name:         "Success INTAKE to DESIGN",
			currentState: "INTAKE",
			targetState:  pb.LifecycleState_LIFECYCLE_DESIGN,
			isFinalized:  true,
			expectedCode: codes.OK,
		},
		{
			name:         "Fail INTAKE to DESIGN (not finalized)",
			currentState: "INTAKE",
			targetState:  pb.LifecycleState_LIFECYCLE_DESIGN,
			isFinalized:  false,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "Fail Backward Transition",
			currentState: "DESIGN",
			targetState:  pb.LifecycleState_LIFECYCLE_INTAKE,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "Fail State Skip",
			currentState: "INTAKE",
			targetState:  pb.LifecycleState_LIFECYCLE_SCAFFOLD,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "Fail Unimplemented Phase 8 Gate",
			currentState: "DESIGN",
			targetState:  pb.LifecycleState_LIFECYCLE_SCAFFOLD,
			expectedCode: codes.Unimplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{
				assemblyLines: map[uuid.UUID]persistence.AssemblyLine{
					alID: {ID: alID, ServiceName: serviceName, CurrentState: tt.currentState},
				},
				specs: map[string]persistence.SpecDocument{
					serviceName: {ServiceName: serviceName, IsFinalized: tt.isFinalized},
				},
			}
			server := &Server{store: store}

			req := &pb.LifecycleAdvance{
				AssemblyLineId: alID.String(),
				TargetState:    tt.targetState,
				Justification:  "test",
			}

			_, err := server.AdvanceLifecycle(context.Background(), req)
			if st, ok := status.FromError(err); ok {
				if st.Code() != tt.expectedCode {
					t.Errorf("expected code %v, got %v (msg: %v)", tt.expectedCode, st.Code(), st.Message())
				}
			} else if err != nil {
				t.Fatalf("unexpected non-gRPC error: %v", err)
			} else if tt.expectedCode != codes.OK {
				t.Errorf("expected error code %v, but got success", tt.expectedCode)
			}
		})
	}
}
