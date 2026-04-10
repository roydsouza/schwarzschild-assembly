package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rds/sati-central/factories/code-assurance/assessment-pipeline"
)

// AssuranceMCPServer serves the factory's assurance tools.
type AssuranceMCPServer struct {
	mu            sync.RWMutex
	lastReport    *pipeline.Result
	aggregator    *pipeline.Aggregator
	projectRoot   string
}

func NewAssuranceMCPServer(aggregator *pipeline.Aggregator, projectRoot string) *AssuranceMCPServer {
	return &AssuranceMCPServer{
		aggregator:  aggregator,
		projectRoot: projectRoot,
	}
}

// SetLastReport updates the cached report.
func (s *AssuranceMCPServer) SetLastReport(res *pipeline.Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastReport = res
}

// HandleToolCall processes MCP tool requests for code assurance.
func (s *AssuranceMCPServer) HandleToolCall(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	switch toolName {
	case "get_assurance_report":
		s.mu.RLock()
		report := s.lastReport
		s.mu.RUnlock()

		if report == nil {
			return "No report available. A scan may be in progress.", nil
		}
		
		// Map to typed assessment for the control panel
		assessment := report.MapToAssessment("schwarzschild-assembly-root")
		assessment.AssessedAtMs = time.Now().UnixMilli()

		data, _ := json.MarshalIndent(assessment, "", "  ")
		return string(data), nil

	case "trigger_scan":
		res, err := s.aggregator.Run(ctx, s.projectRoot)
		if err != nil {
			return "", fmt.Errorf("scan failed: %w", err)
		}
		s.SetLastReport(&res)
		return fmt.Sprintf("Scan completed. Correctness: %.2f, Quality: %.2f. Findings: %d", 
			res.CorrectnessScore, res.QualityScore, len(res.Findings)), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

// ToolDefinitions returns the JSON-schema for the factory tools.
func (s *AssuranceMCPServer) ToolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "get_assurance_report",
			"description": "Returns the most recent detailed report from the Code Assurance Assessment Pipeline.",
		},
		{
			"name":        "trigger_scan",
			"description": "Triggers an immediate full-repository assessment for security, quality, and correctness.",
		},
	}
}
