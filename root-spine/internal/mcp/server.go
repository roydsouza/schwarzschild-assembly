package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
				Name:        "create_spec",
				Description: "Initialize a new software service specification.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"service_name":     map[string]string{"type": "string"},
						"description":      map[string]string{"type": "string"},
						"primary_language": map[string]string{"type": "string"},
					},
					"required": []string{"service_name", "description", "primary_language"},
				},
			},
			{
				Name:        "add_requirement",
				Description: "Add a new functional or non-functional requirement to an existing spec.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"service_name": map[string]string{"type": "string"},
						"requirement":  map[string]string{"type": "string"},
					},
					"required": []string{"service_name", "requirement"},
				},
			},
			{
				Name:        "record_challenge",
				Description: "Record a technical challenge or constraint for the service.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"service_name": map[string]string{"type": "string"},
						"challenge":    map[string]string{"type": "string"},
					},
					"required": []string{"service_name", "challenge"},
				},
			},
			{
				Name:        "add_acceptance_criterion",
				Description: "Add a validation criterion for the service's final delivery.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"service_name": map[string]string{"type": "string"},
						"criterion":    map[string]string{"type": "string"},
					},
					"required": []string{"service_name", "criterion"},
				},
			},
			{
				Name:        "finalize_spec",
				Description: "Mark a specification as complete and ready for DESIGN phase.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"service_name": map[string]string{"type": "string"},
					},
					"required": []string{"service_name"},
				},
			},
			{
				Name:        "get_assembly_line_status",
				Description: "Check the current lifecycle state of an assembly line.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{"type": "string"},
					},
					"required": []string{"id"},
				},
			},
			{
				Name:        "advance_lifecycle",
				Description: "Transition an assembly line to the next lifecycle state (e.g., INTAKE -> DESIGN).",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":            map[string]string{"type": "string"},
						"target_state":  map[string]interface{}{"type": "string", "enum": []string{"DESIGN", "SCAFFOLD", "BUILD", "VERIFY", "DELIVERED"}},
						"justification": map[string]string{"type": "string"},
					},
					"required": []string{"id", "target_state", "justification"},
				},
			},
			{
				Name:        "set_deployment_target",
				Description: "Specify the target deployment environment and configuration for a service.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"service_name": map[string]string{"type": "string"},
						"target":       map[string]interface{}{"type": "string", "enum": []string{"LOCAL", "CONTAINER", "AWS", "GCP"}},
						"config_json":  map[string]string{"type": "string", "description": "JSON string of deployment configuration"},
					},
					"required": []string{"service_name", "target"},
				},
			},
			{
				Name:        "submit_skill_proposal",
				Description: "Propose a new Prolog skill clause for safety verification.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"agent_id":   map[string]string{"type": "string"},
						"skill_name": map[string]string{"type": "string"},
						"clause":     map[string]string{"type": "string"},
					},
					"required": []string{"agent_id", "skill_name", "clause"},
				},
			},
			{
				Name:        "report_metrics",
				Description: "Report telemetry metrics directly to the Antigravity fitness vector.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"factory_id":   map[string]string{"type": "string"},
						"factory_type": map[string]string{"type": "string"},
						"metrics":      map[string]interface{}{"type": "object", "additionalProperties": map[string]string{"type": "number"}},
					},
					"required": []string{"factory_id", "factory_type", "metrics"},
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
	case "create_spec":
		return h.callCreateSpec(ctx, callParams.Arguments)
	case "add_requirement", "record_challenge", "add_acceptance_criterion", "finalize_spec":
		return h.callUpdateSpec(ctx, callParams.Name, callParams.Arguments)
	case "get_assembly_line_status":
		return h.callGetAssemblyLineStatus(ctx, callParams.Arguments)
	case "advance_lifecycle":
		return h.callAdvanceLifecycle(ctx, callParams.Arguments)
	case "set_deployment_target":
		return h.callSetDeploymentTarget(ctx, callParams.Arguments)
	case "submit_skill_proposal":
		return h.callSubmitSkillProposal(ctx, callParams.Arguments)
	case "report_metrics":
		return h.callReportMetrics(ctx, callParams.Arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", callParams.Name)
	}
}

func (h *Host) callCreateSpec(ctx context.Context, args json.RawMessage) (*CallToolResult, error) {
	var input struct {
		ServiceName     string `json:"service_name"`
		Description      string `json:"description"`
		PrimaryLanguage string `json:"primary_language"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	req := &pb.SpecDocument{
		Id:              uuid.New().String(),
		ServiceName:     input.ServiceName,
		Description:     input.Description,
		PrimaryLanguage: input.PrimaryLanguage,
	}

	res, err := h.spine.CreateAssemblyLine(ctx, req)
	if err != nil {
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}}, nil
	}

	jsonRes, _ := json.Marshal(res)
	return &CallToolResult{Content: []Content{{Type: "text", Text: string(jsonRes)}}}, nil
}

func (h *Host) callUpdateSpec(ctx context.Context, toolName string, args json.RawMessage) (*CallToolResult, error) {
	var input struct {
		ServiceName string `json:"service_name"`
		Requirement string `json:"requirement"`
		Challenge   string `json:"challenge"`
		Criterion   string `json:"criterion"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	// 1. Fetch existing spec
	spec, err := h.store.GetSpecDocument(ctx, input.ServiceName)
	if err != nil {
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: spec not found for %s", input.ServiceName)}}}, nil
	}

	// 2. Unmarshal existing PB data
	var pbSpec pb.SpecDocument
	if err := json.Unmarshal(spec.DataJSON, &pbSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec data: %w", err)
	}

	// 3. Apply incremental update
	switch toolName {
	case "add_requirement":
		pbSpec.Requirements = append(pbSpec.Requirements, &pb.Requirement{
			Id:   uuid.New().String(),
			Text: input.Requirement,
		})
	case "record_challenge":
		pbSpec.Challenges = append(pbSpec.Challenges, &pb.ChallengeRecord{
			Id:        uuid.New().String(),
			Challenge: input.Requirement, // input.Requirement actually holds 'challenge' from JSON if called this way
		})
	case "add_acceptance_criterion":
		pbSpec.AcceptanceCriteria = append(pbSpec.AcceptanceCriteria, &pb.AcceptanceCriterion{
			Id:   uuid.New().String(),
			Text: input.Criterion,
		})
	case "finalize_spec":
		pbSpec.IsFinalized = true
	case "set_deployment_target":
		// This case is handled by callSetDeploymentTarget for cleaner separation
		// but I'll keep it here if called via this generic helper.
	}

	// 4. Save via Spine (CreateAssemblyLine upserts the spec)
	res, err := h.spine.CreateAssemblyLine(ctx, &pbSpec)
	if err != nil {
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: update failed: %v", err)}}}, nil
	}

	jsonRes, _ := json.Marshal(res)
	return &CallToolResult{Content: []Content{{Type: "text", Text: string(jsonRes)}}}, nil
}

func (h *Host) callGetAssemblyLineStatus(ctx context.Context, args json.RawMessage) (*CallToolResult, error) {
	var input struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	res, err := h.spine.GetAssemblyLineStatus(ctx, &pb.AssemblyLineID{Id: input.ID})
	if err != nil {
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}}, nil
	}

	return &CallToolResult{Content: []Content{{Type: "text", Text: fmt.Sprintf("Current Lifecycle State: %s", res.State.String())}}}, nil
}

func (h *Host) callAdvanceLifecycle(ctx context.Context, args json.RawMessage) (*CallToolResult, error) {
	var input struct {
		ID            string `json:"id"`
		TargetState   string `json:"target_state"`
		Justification string `json:"justification"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	var target pb.LifecycleState
	switch input.TargetState {
	case "DESIGN":
		target = pb.LifecycleState_LIFECYCLE_DESIGN
	case "SCAFFOLD":
		target = pb.LifecycleState_LIFECYCLE_SCAFFOLD
	case "BUILD":
		target = pb.LifecycleState_LIFECYCLE_BUILD
	case "VERIFY":
		target = pb.LifecycleState_LIFECYCLE_VERIFY
	case "DELIVERED":
		target = pb.LifecycleState_LIFECYCLE_DELIVERED
	default:
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: "Error: invalid target state"}}}, nil
	}

	req := &pb.LifecycleAdvance{
		AssemblyLineId: input.ID,
		TargetState:    target,
		Justification:  input.Justification,
	}

	res, err := h.spine.AdvanceLifecycle(ctx, req)
	if err != nil {
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: advance failed: %v", err)}}}, nil
	}

	return &CallToolResult{Content: []Content{{Type: "text", Text: fmt.Sprintf("Successfully advanced to %s", res.State.String())}}}, nil
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

func (h *Host) callSetDeploymentTarget(ctx context.Context, args json.RawMessage) (*CallToolResult, error) {
	var input struct {
		ServiceName string `json:"service_name"`
		Target      string `json:"target"`
		ConfigJSON  string `json:"config_json"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	// 1. Fetch existing spec
	spec, err := h.store.GetSpecDocument(ctx, input.ServiceName)
	if err != nil {
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: spec not found: %v", err)}}}, nil
	}

	// 2. Unmarshal existing PB data
	var pbSpec pb.SpecDocument
	if err := json.Unmarshal(spec.DataJSON, &pbSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec data: %w", err)
	}

	// 3. Update deployment target
	pbSpec.DeploymentTarget = &pb.DeploymentTarget{
		TargetType: input.Target,
		ConfigJson: []byte(input.ConfigJSON),
	}

	// 4. Save via Spine
	res, err := h.spine.CreateAssemblyLine(ctx, &pbSpec)
	if err != nil {
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: failed to set deployment target: %v", err)}}}, nil
	}

	jsonRes, _ := json.Marshal(res)
	return &CallToolResult{Content: []Content{{Type: "text", Text: string(jsonRes)}}}, nil
}

func (h *Host) callSubmitSkillProposal(ctx context.Context, args json.RawMessage) (*CallToolResult, error) {
	var input struct {
		AgentID   string `json:"agent_id"`
		SkillName string `json:"skill_name"`
		Clause    string `json:"clause"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	req := &pb.SkillUpdateRequest{
		AgentId:    input.AgentID,
		SkillName:  input.SkillName,
		NewContent: []byte(input.Clause),
		Rationale:  "Self-modification via Prolog escape",
	}

	res, err := h.spine.UpdateSkill(ctx, req)
	if err != nil {
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}}, nil
	}

	jsonRes, _ := json.Marshal(res)
	return &CallToolResult{Content: []Content{{Type: "text", Text: string(jsonRes)}}}, nil
}

func (h *Host) callReportMetrics(ctx context.Context, args json.RawMessage) (*CallToolResult, error) {
	var input struct {
		FactoryID   string             `json:"factory_id"`
		FactoryType string             `json:"factory_type"`
		Metrics     map[string]float64 `json:"metrics"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	req := &pb.MetricReport{
		FactoryId:    input.FactoryID,
		Metrics:      input.Metrics,
		ObservedAtMs: time.Now().UnixMilli(),
	}

	res, err := h.spine.ReportMetrics(ctx, req)
	if err != nil {
		return &CallToolResult{IsError: true, Content: []Content{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}}, nil
	}

	return &CallToolResult{Content: []Content{{Type: "text", Text: res.Message}}}, nil
}
