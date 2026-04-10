package fitness

import (
	"context"
	"fmt"
	"time"

	"github.com/rds/aethereum-spine/factories/scaffold-engine/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// MetricsManager handles domain-specific metrics for the Scaffold Engine.
type MetricsManager struct {
	logger    *zap.Logger
	spine     pb.OrchestratorClient
	factoryID string
}

// NewMetricsManager creates a new manager.
func NewMetricsManager(logger *zap.Logger, conn *grpc.ClientConn, factoryID string) *MetricsManager {
	return &MetricsManager{
		logger:    logger,
		spine:     pb.NewOrchestratorClient(conn),
		factoryID: factoryID,
	}
}

// Register declarations of Scaffold Engine metrics.
func (m *MetricsManager) Register(ctx context.Context) error {
	req := &pb.DomainFitnessExtension{
		FactoryId:   m.factoryID,
		FactoryType: "scaffold-engine",
		Metrics: []*pb.DomainMetricDeclaration{
			{
				MetricId:            "scaffold_success_rate",
				DisplayName:         "Spec Completion Rate",
				Description:         "Percentage of created specs that reach DELIVERED state.",
				Unit:                "percent",
				Direction:           pb.MetricDirection_METRIC_HIGHER_IS_BETTER,
				EscalationThreshold: 60.0,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_LT,
			},
			{
				MetricId:            "scaffold_latency",
				DisplayName:         "Scaffold Latency (p99)",
				Description:         "p99 ms from finalize_spec to scaffold artifacts ready.",
				Unit:                "ms",
				Direction:           pb.MetricDirection_METRIC_LOWER_IS_BETTER,
				EscalationThreshold: 30000,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_GT,
			},
			{
				MetricId:            "acceptance_criterion_coverage",
				DisplayName:         "Acceptance Criterion Coverage",
				Description:         "Percentage of spec criteria with non-stub test implementations at VERIFY phase.",
				Unit:                "percent",
				Direction:           pb.MetricDirection_METRIC_HIGHER_IS_BETTER,
				EscalationThreshold: 100.0,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_LT,
			},
			{
				MetricId:            "rework_rate",
				DisplayName:         "Rework Rate",
				Description:         "Percentage of assembly lines requiring more than 2 CONDITIONAL verdicts before BUILD.",
				Unit:                "percent",
				Direction:           pb.MetricDirection_METRIC_LOWER_IS_BETTER,
				EscalationThreshold: 30.0,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_GT,
			},
		},
	}

	res, err := m.spine.RegisterDomainMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to register metrics: %w", err)
	}

	if !res.Success {
		return fmt.Errorf("spine rejected metrics registration: %s", res.ErrorMessage)
	}

	return nil
}

// Report sends current metric values to the spine.
// criterionCoverage: % of acceptance criteria with non-stub implementations (0–100).
// reworkRate: % of assembly lines that needed >2 CONDITIONAL verdicts before BUILD (0–100).
func (m *MetricsManager) Report(ctx context.Context, successRate, latency, criterionCoverage, reworkRate float64) error {
	req := &pb.MetricReport{
		FactoryId: m.factoryID,
		Metrics: map[string]float64{
			"scaffold_success_rate":        successRate,
			"scaffold_latency":             latency,
			"acceptance_criterion_coverage": criterionCoverage,
			"rework_rate":                  reworkRate,
		},
		ObservedAtMs: time.Now().UnixMilli(),
	}

	_, err := m.spine.ReportMetrics(ctx, req)
	return err
}
