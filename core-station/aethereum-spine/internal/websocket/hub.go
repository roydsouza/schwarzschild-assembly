package websocket

import (
	"encoding/json"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
	"github.com/rds/aethereum-spine/aethereum-spine/internal/grpc/pb"
	"go.uber.org/zap"
)

// Hub maintains the set of active Socket.IO connections and broadcasts events.
type Hub struct {
	logger  *zap.Logger
	server  *socketio.Server
}

// NewHub creates a new Socket.IO hub.
func NewHub(logger *zap.Logger) *Hub {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		logger.Info("client connected", zap.String("id", s.ID()))
		return nil
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		logger.Info("client disconnected", zap.String("id", s.ID()), zap.String("reason", reason))
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		logger.Error("socket.io error", zap.Error(e))
	})

	return &Hub{
		logger: logger,
		server: server,
	}
}

// Start initiates the Socket.IO server loop.
func (h *Hub) Start() {
	go h.server.Serve()
}

// Stop shuts down the Socket.IO server.
func (h *Hub) Stop() {
	h.server.Close()
}

// Handler returns the HTTP handler for Socket.IO.
func (h *Hub) Handler() http.Handler {
	return h.server
}

// Broadcast sends a pb.VerificationEvent to all connected clients.
func (h *Hub) Broadcast(event *pb.VerificationEvent) {
	// Convert proto message to JSON for the browser
	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("failed to marshal event for broadcast", zap.Error(err))
		return
	}

	h.logger.Info("broadcasting event", 
		zap.String("proposal_id", event.ProposalId),
		zap.String("type", event.EventType.String()))
		
	h.server.BroadcastToRoom("/", "", "verification_event", string(data))
}
