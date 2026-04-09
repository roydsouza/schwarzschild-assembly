package websocket

import (
	"sync"

	"github.com/rds/sati-central/root-spine/internal/grpc/pb"
	"go.uber.org/zap"
)

// Client represents a connected WebSocket client (e.g., Control Panel).
type Client struct {
	ID   string
	Send chan *pb.VerificationEvent
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	logger  *zap.Logger
	clients map[string]*Client
	mu      sync.RWMutex
}

// NewHub creates a new Hub.
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		logger:  logger,
		clients: make(map[string]*Client),
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.ID] = c
	h.logger.Info("client registered", zap.String("client_id", c.ID))
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[id]; ok {
		delete(h.clients, id)
		h.logger.Info("client unregistered", zap.String("client_id", id))
	}
}

// Broadcast sends a message to all registered clients.
func (h *Hub) Broadcast(event *pb.VerificationEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, c := range h.clients {
		select {
		case c.Send <- event:
		default:
			h.logger.Warn("client send buffer full, skipping", zap.String("client_id", c.ID))
		}
	}
}
