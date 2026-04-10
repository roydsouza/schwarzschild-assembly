package fitness

import (
	"context"
	"testing"

	"github.com/rds/aethereum-spine/factories/scaffold-engine/pb"
	"google.golang.org/grpc"
)

type mockOrchestratorClient struct {
	pb.OrchestratorClient
	latestExtension *pb.DomainFitnessExtension
}

func (m *mockOrchestratorClient) RegisterDomainMetrics(ctx context.Context, in *pb.DomainFitnessExtension, opts ...grpc.CallOption) (*pb.RegistrationResult, error) {
	m.latestExtension = in
	return &pb.RegistrationResult{Success: true}, nil
}

func TestMetricsManager_Register(t *testing.T) {
	mock := &mockOrchestratorClient{}
	mm := &MetricsManager{
		logger:    nil,
		spine:     mock,
		factoryID: "test-factory",
	}

	err := mm.Register(context.Background())
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if mock.latestExtension == nil {
		t.Fatal("Extension was not sent to spine")
	}

	expectedMetrics := map[string]float64{
		"scaffold_success_rate":        60.0,
		"scaffold_latency":             30000.0,
		"acceptance_criterion_coverage": 100.0,
		"rework_rate":                  30.0,
	}

	if len(mock.latestExtension.Metrics) != len(expectedMetrics) {
		t.Errorf("expected %d metrics, got %d", len(expectedMetrics), len(mock.latestExtension.Metrics))
	}

	for _, m := range mock.latestExtension.Metrics {
		expected, ok := expectedMetrics[m.MetricId]
		if !ok {
			t.Errorf("unexpected metric ID: %s", m.MetricId)
			continue
		}
		if m.EscalationThreshold != expected {
			t.Errorf("metric %s: expected threshold %f, got %f", m.MetricId, expected, m.EscalationThreshold)
		}
	}
}
