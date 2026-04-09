package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rds/sati-central/root-spine/internal/grpc/pb"
	"github.com/rds/sati-central/root-spine/internal/persistence"
	"go.uber.org/zap"
)

// Host represents the MCP server host that wraps Root Spine capabilities.
type Host struct {
	logger *zap.Logger
	spine  pb.OrchestratorServer
	store  *persistence.Store
}

// NewHost creates a new MCP Host.
func NewHost(logger *zap.Logger, spine pb.OrchestratorServer, store *persistence.Store) *Host {
	return &Host{
		logger: logger,
		spine:  spine,
		store:  store,
	}
}

// HandleRequest routes an incoming MCP JSON-RPC request to the appropriate handler.
func (h *Host) HandleRequest(ctx context.Context, req *Request) *Response {
	h.logger.Debug("handling MCP request", zap.String("method", req.Method))

	var result interface{}
	var err error

	switch req.Method {
	case "initialize":
		result = h.handleInitialize()
	case "tools/list":
		result = h.handleListTools()
	case "tools/call":
		result, err = h.handleCallTool(ctx, req.Params)
	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}

	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32000,
				Message: err.Error(),
			},
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (h *Host) handleInitialize() interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]string{
			"name":    "sati-central-root-spine",
			"version": "1.0.0-phase5",
		},
	}
}

func (h *Host) handleListTools() *ListToolsResult {
	return &ListToolsResult{
		Tools: []Tool{
			{
				Name:        "get_analyst_verdict",
				Description: "Retrieve the latest automated analyst verdict for a given proposal ID or topic.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"proposal_id": map[string]string{"type": "string"},
						"topic":       map[string]string{"type": "string"},
					},
				},
			},
			{
				Name:        "approve_action",
				Description: "Issue a human-in-the-loop approval for a security-adjacent proposal.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"proposal_id": map[string]string{"type": "string"},
						"signature":   map[string]string{"type": "string"},
						"operator":    map[string]string{"type": "string"},
					},
					"required": []string{"proposal_id", "signature", "operator"},
				},
			},
		},
	}
}

func (h *Host) handleCallTool(ctx context.Context, params json.RawMessage) (*CallToolResult, error) {
	var callParams CallToolParams
	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, err
	}

	switch callParams.Name {
	case "get_analyst_verdict":
		return h.callGetAnalystVerdict(ctx, callParams.Arguments)
	case "approve_action":
		return h.callApproveAction(ctx, callParams.Arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", callParams.Name)
	}
}

func (h *Host) callGetAnalystVerdict(ctx context.Context, args json.RawMessage) (*CallToolResult, error) {
	var input struct {
		ProposalID string `json:"proposal_id"`
		Topic      string `json:"topic"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	query := &pb.VerdictQuery{}
	if input.ProposalID != "" {
		query.Query = &pb.VerdictQuery_ProposalId{ProposalId: input.ProposalID}
	} else if input.Topic != "" {
		query.Query = &pb.VerdictQuery_Topic{Topic: input.Topic}
	} else {
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: "Error: missing proposal_id or topic"}}}, nil
	}

	// Internal call to local gRPC handler
	res, err := h.spine.ReadAnalystVerdict(ctx, query)
	if err != nil {
		return &CallToolResult{
			IsError: true,
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
		}, nil
	}

	jsonRes, _ := json.Marshal(res)
	return &CallToolResult{
		Content: []Content{{Type: "text", Text: string(jsonRes)}},
	}, nil
}

func (h *Host) callApproveAction(ctx context.Context, args json.RawMessage) (*CallToolResult, error) {
	var input struct {
		ProposalID string `json:"proposal_id"`
		Signature  string `json:"signature"`
		Operator   string `json:"operator"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	// FIXUP-3: Block MCP approval of security-adjacent proposals.
	// These MUST go through the Translucent Gate (Control Panel) to ensure
	// human signature verification. See proposals/pending/mcp-tool-security.md
	pID, err := uuid.Parse(input.ProposalID)
	if err != nil {
		return &CallToolResult{
			IsError: true,
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: invalid proposal ID: %v", err)}},
		}, nil
	}

	isSec, err := h.store.IsProposalSecurityAdjacent(ctx, pID)
	if err != nil {
		return &CallToolResult{
			IsError: true,
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: failed to check proposal security status: %v", err)}},
		}, nil
	}
	if isSec {
		h.logger.Warn("MCP approve_action blocked: proposal is security-adjacent",
			zap.String("proposal_id", input.ProposalID),
			zap.String("operator", input.Operator),
		)
		return &CallToolResult{
			IsError: true,
			Content: []Content{{Type: "text", Text: "Error: security-adjacent proposals cannot be approved via MCP. Use the Sati-Central Control Panel for human signature verification."}},
		}, nil
	}

	req := &pb.ApprovalRequest{
		ProposalId:        input.ProposalID,
		ApprovalSignature: input.Signature,
		ApprovedBy:        input.Operator,
	}

	res, err := h.spine.ApproveAction(ctx, req)
	if err != nil {
		return &CallToolResult{
			IsError: true,
			Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
		}, nil
	}

	jsonRes, _ := json.Marshal(res)
	return &CallToolResult{
		Content: []Content{{Type: "text", Text: string(jsonRes)}},
	}, nil
}
