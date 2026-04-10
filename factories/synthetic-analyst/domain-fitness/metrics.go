package fitness

import (
	"math/rand"
	"time"

	"github.com/rds/sati-central/factories/synthetic-analyst/pb"
)

// Collector manages domain-specific fitness metrics for the Synthetic Analyst.
type Collector struct {
	factoryID string
}

func NewCollector(factoryID string) *Collector {
	return &Collector{factoryID: factoryID}
}

// GetDeclarations returns the domain metric declarations for Phase 6.
func (c *Collector) GetDeclarations() *pb.DomainFitnessExtension {
	return &pb.DomainFitnessExtension{
		FactoryId:   c.factoryID,
		FactoryType: "synthetic-analyst",
		Metrics: []*pb.DomainMetricDeclaration{
			{
				MetricId:            "defi_coverage",
				DisplayName:         "DeFi Protocol Coverage",
				Description:         "% of tracked TVL with live data feeds",
				Unit:                "percent",
				Direction:           pb.MetricDirection_METRIC_HIGHER_IS_BETTER,
				EscalationThreshold: 80.0,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_LT,
			},
			{
				MetricId:            "macro_precision",
				DisplayName:         "Macroeconomic Model Precision",
				Description:         "MRR (rolling 30d backtest)",
				Unit:                "score [0,1]",
				Direction:           pb.MetricDirection_METRIC_HIGHER_IS_BETTER,
				EscalationThreshold: 0.7,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_LT,
			},
			{
				MetricId:            "rag_quality",
				DisplayName:         "RAG Retrieval Quality",
				Description:         "Mean Reciprocal Rank (MRR)",
				Unit:                "score [0,1]",
				Direction:           pb.MetricDirection_METRIC_HIGHER_IS_BETTER,
				EscalationThreshold: 0.6,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_LT,
			},
			{
				MetricId:            "answer_accuracy",
				DisplayName:         "Agent Answer Accuracy",
				Description:         "Semantic correctness of generated analyst reports",
				Unit:                "Ratio",
				Direction:           pb.MetricDirection_METRIC_HIGHER_IS_BETTER,
				EscalationThreshold: 0.85,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_LT,
			},
			{
				MetricId:            "knowledge_coverage",
				DisplayName:         "Knowledge Coverage",
				Description:         "Effective range of the RAG index vs domain corpus",
				Unit:                "Ratio",
				Direction:           pb.MetricDirection_METRIC_HIGHER_IS_BETTER,
				EscalationThreshold: 0.80,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_LT,
			},
			{
				MetricId:            "query_latency",
				DisplayName:         "RAG Query Latency",
				Description:         "Internal retrieval latency (ms)",
				Unit:                "ms",
				Direction:           pb.MetricDirection_METRIC_LOWER_IS_BETTER,
				EscalationThreshold: 500.0,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_GT,
			},
			{
				MetricId:            "alert_latency",
				DisplayName:         "Alert Latency (p99)",
				Description:         "Signal to Translucent Gate latency",
				Unit:                "ms",
				Direction:           pb.MetricDirection_METRIC_LOWER_IS_BETTER,
				EscalationThreshold: 5000.0,
				EscalationOperator:  pb.ThresholdOperator_THRESHOLD_GT,
			},
		},
	}
}

// Collect simulates metric collection.
func (c *Collector) Collect() map[string]*pb.DomainMetricValue {
	now := time.Now().UnixMilli()
	return map[string]*pb.DomainMetricValue{
		"defi_coverage": {
			MetricName:    "defi_coverage",
			Value:         85.0 + rand.Float64()*5.0,
			Unit:          "percent",
			Status:        pb.MetricStatus_METRIC_GREEN,
			LastUpdatedMs: now,
		},
		"macro_precision": {
			MetricName:    "macro_precision",
			Value:         0.75 + rand.Float64()*0.1,
			Unit:          "score",
			Status:        pb.MetricStatus_METRIC_GREEN,
			LastUpdatedMs: now,
		},
		"rag_quality": {
			MetricName:    "rag_quality",
			Value:         0.65 + rand.Float64()*0.1,
			Unit:          "score",
			Status:        pb.MetricStatus_METRIC_GREEN,
			LastUpdatedMs: now,
		},
		"answer_accuracy": {
			MetricName:    "answer_accuracy",
			Value:         0.88 + rand.Float64()*0.05,
			Unit:          "Ratio",
			Status:        pb.MetricStatus_METRIC_GREEN,
			LastUpdatedMs: now,
		},
		"knowledge_coverage": {
			MetricName:    "knowledge_coverage",
			Value:         0.82 + rand.Float64()*0.1,
			Unit:          "Ratio",
			Status:        pb.MetricStatus_METRIC_GREEN,
			LastUpdatedMs: now,
		},
		"query_latency": {
			MetricName:    "query_latency",
			Value:         150.0 + rand.Float64()*100.0,
			Unit:          "ms",
			Status:        pb.MetricStatus_METRIC_GREEN,
			LastUpdatedMs: now,
		},
		"alert_latency": {
			MetricName:    "alert_latency",
			Value:         1200.0 + rand.Float64()*500.0,
			Unit:          "ms",
			Status:        pb.MetricStatus_METRIC_GREEN,
			LastUpdatedMs: now,
		},
	}
}
