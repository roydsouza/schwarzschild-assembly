package fitness

import (
	"sync"
	"time"

	"github.com/rds/aethereum-spine/factories/code-assurance/pb"
)

// Metric constants
const (
	MetricArtifactCorrectness = "artifact_correctness" // (0.25) Tests & Security
	MetricCodeQuality         = "code_quality"        // (0.15) Lint & Complexity
)

// Collector manages the accumulation of metrics for the refinery factory.
type Collector struct {
	mu      sync.RWMutex
	metrics map[string]*pb.DomainMetricValue
}

func NewCollector() *Collector {
	return &Collector{
		metrics: make(map[string]*pb.DomainMetricValue),
	}
}

// UpdateMetric sets the current value for a given metric ID.
func (c *Collector) UpdateMetric(id string, value float32, status pb.MetricStatus, unit string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics[id] = &pb.DomainMetricValue{
		MetricName:    id,
		Value:         float64(value),
		Unit:          unit,
		Status:        status,
		LastUpdatedMs: time.Now().UnixMilli(),
	}
}

// Collect returns the latest snapshot of all metrics.
func (c *Collector) Collect() map[string]*pb.DomainMetricValue {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snapshot := make(map[string]*pb.DomainMetricValue)
	for k, v := range c.metrics {
		snapshot[k] = v
	}
	return snapshot
}

// GetDeclarations returns the schema for the Code Assurance domain metrics.
func GetDeclarations(factoryID, factoryType string) *pb.DomainFitnessExtension {
	return &pb.DomainFitnessExtension{
		FactoryId:   factoryID,
		FactoryType: factoryType,
		Metrics: []*pb.DomainMetricDeclaration{
			{
				MetricId:            MetricArtifactCorrectness,
				DisplayName:         "Artifact Correctness",
				Description:         "Aggregate score for unit tests (Go/Rust/Vitest) and security audits (govulncheck/cargo-audit).",
				Unit:                "Ratio",
				Direction:           pb.MetricDirection_METRIC_HIGHER_IS_BETTER,
				EscalationThreshold: 0.90,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_LT,
			},
			{
				MetricId:            MetricCodeQuality,
				DisplayName:         "Code Quality",
				Description:         "Aggregate score for linter hygiene (staticcheck/clippy) and cyclomatic complexity (threshold: 10).",
				Unit:                "Ratio",
				Direction:           pb.MetricDirection_METRIC_HIGHER_IS_BETTER,
				EscalationThreshold: 0.85,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_LT,
			},
		},
	}
}
