package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// HTTPTransport implements the HTTP/SSE transport for MCP.
type HTTPTransport struct {
	logger *zap.Logger
	host   *Host
}

// NewHTTPTransport creates a new HTTPTransport.
func NewHTTPTransport(logger *zap.Logger, host *Host) *HTTPTransport {
	return &HTTPTransport{
		logger: logger,
		host:   host,
	}
}

// Handler returns an http.Handler for the MCP endpoint.
// For Phase 5, we implement the core POST handler for tool execution.
func (t *HTTPTransport) Handler() http.Handler {
	mux := http.NewServeMux()

	// MCP Endpoint
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Process request via the host
		resp := t.host.HandleRequest(r.Context(), &req)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "MCP Host Active")
	})

	return mux
}
