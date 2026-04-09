package mcp

import (
	"context"
	"fmt"

	fitness "github.com/rds/sati-central/factories/synthetic-analyst/domain-fitness"
)

// DomainMCPServer serves the factory's domain-specific tools.
type DomainMCPServer struct {
	collector *fitness.Collector
}

// NewDomainMCPServer creates a new domain MCP server.
func NewDomainMCPServer(collector *fitness.Collector) *DomainMCPServer {
	return &DomainMCPServer{collector: collector}
}

// HandleToolCall processes MCP tool requests for the synthetic analyst.
func (s *DomainMCPServer) HandleToolCall(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	metrics := s.collector.Collect()
	switch toolName {
	case "query_defi_coverage":
		v := metrics["defi_coverage"]
		if v == nil {
			return "", fmt.Errorf("metric 'defi_coverage' not available from collector")
		}
		return fmt.Sprintf("Current DeFi Coverage: %.2f (Threshold: 0.80)", v.Value), nil
	case "get_pali_filter_status":
		v := metrics["pali_filter_rate"]
		if v == nil {
			return "", fmt.Errorf("metric 'pali_filter_rate' not available from collector")
		}
		return fmt.Sprintf("Pāḷi Filter Precision: %.2f", v.Value), nil
	case "get_macro_precision":
		v := metrics["macro_precision"]
		if v == nil {
			return "", fmt.Errorf("metric 'macro_precision' not available")
		}
		return fmt.Sprintf("Macro Precision: %.2f", v.Value), nil
	case "get_rag_quality_score":
		v := metrics["rag_quality"]
		if v == nil {
			return "", fmt.Errorf("metric 'rag_quality' not available")
		}
		return fmt.Sprintf("RAG Quality Score: %.2f", v.Value), nil
	case "get_alert_latency":
		v := metrics["alert_latency"]
		if v == nil {
			return "", fmt.Errorf("metric 'alert_latency' not available")
		}
		return fmt.Sprintf("Average Alert Latency: %.2f ms", v.Value), nil
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

// ToolDefinitions returns the JSON-schema for the factory tools.
func (s *DomainMCPServer) ToolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "query_defi_coverage",
			"description": "Returns the current DeFi protocol coverage metric from the synthetic analyst factory.",
		},
		{
			"name":        "get_pali_filter_status",
			"description": "Returns the current precision of the Pāḷi content filter.",
		},
		{
			"name":        "get_macro_precision",
			"description": "Returns the current macro-precision of the analyst droid model.",
		},
		{
			"name":        "get_rag_quality_score",
			"description": "Returns the current quality score of the analyst's RAG pipeline.",
		},
		{
			"name":        "get_alert_latency",
			"description": "Returns the average latency of the factory's safety alert pipeline in milliseconds.",
		},
	}
}
