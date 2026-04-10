package pipeline

import (
	"context"
	"testing"
)

func TestResult_RunTool(t *testing.T) {
	ctx := context.Background()

	t.Run("ToolFailureAppendsFindingAndPenalizes", func(t *testing.T) {
		res := Result{CorrectnessScore: 1.0, QualityScore: 1.0}
		
		// Use a command that will definitely fail (exit 1)
		// We can use 'false' or 'ls' on a non-existent file
		score := float32(1.0)
		res.runTool(ctx, ".", "ls", []string{"/non-existent-file-path-xyz"}, "TestTool", &score, 0.1)

		if score != 0.9 {
			t.Errorf("Expected score 0.9 after penalty, got %f", score)
		}

		if len(res.Findings) != 1 {
			t.Fatalf("Expected 1 finding, got %d", len(res.Findings))
		}

		f := res.Findings[0]
		if f.Tool != "TestTool" {
			t.Errorf("Expected tool TestTool, got %s", f.Tool)
		}
		if f.Severity != SeverityWarning {
			t.Errorf("Expected SeverityWarning, got %s", f.Severity)
		}
	})

	t.Run("ScoreFloorAtZero", func(t *testing.T) {
		res := Result{}
		score := float32(0.05)
		res.runTool(ctx, ".", "ls", []string{"/non-existent"}, "TestTool", &score, 0.1)

		if score != 0.0 {
			t.Errorf("Score should not drop below 0.0, got %f", score)
		}
	})
}
