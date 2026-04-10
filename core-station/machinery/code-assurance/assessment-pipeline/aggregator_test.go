package pipeline

import (
	"context"
	"errors"
	"testing"
)

type MockAnalyzer struct {
	name   string
	result Result
	err    error
}

func (m *MockAnalyzer) Name() string { return m.name }
func (m *MockAnalyzer) Analyze(ctx context.Context, projectRoot string) (Result, error) {
	return m.result, m.err
}

func TestAggregator_Run(t *testing.T) {
	ctx := context.Background()

	t.Run("ScoreAggregationMinimum", func(t *testing.T) {
		agg := NewAggregator(
			&MockAnalyzer{name: "A", result: Result{CorrectnessScore: 0.9, QualityScore: 0.8}},
			&MockAnalyzer{name: "B", result: Result{CorrectnessScore: 0.7, QualityScore: 0.9}},
		)

		res, err := agg.Run(ctx, ".")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if res.CorrectnessScore != 0.7 {
			t.Errorf("Expected CorrectnessScore 0.7, got %f", res.CorrectnessScore)
		}
		if res.QualityScore != 0.8 {
			t.Errorf("Expected QualityScore 0.8, got %f", res.QualityScore)
		}
	})

	t.Run("AnalyzerExecutionFailureCreatesFinding", func(t *testing.T) {
		agg := NewAggregator(
			&MockAnalyzer{name: "Buggy", err: errors.New("boom")},
			&MockAnalyzer{name: "Healthy", result: Result{CorrectnessScore: 1.0, QualityScore: 1.0}},
		)

		res, err := agg.Run(ctx, ".")
		if err != nil {
			t.Fatalf("Aggregator should not return error on individual analyzer failure")
		}

		found := false
		for _, f := range res.Findings {
			if f.Tool == "Buggy" && f.Severity == SeverityError {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected SeverityError finding for failed analyzer")
		}

		if res.CorrectnessScore != 1.0 {
			t.Errorf("Healthy analyzer score should be preserved")
		}
	})
}
