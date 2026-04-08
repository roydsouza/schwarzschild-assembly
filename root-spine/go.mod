module github.com/rds/sati-central/root-spine

go 1.26

// Dependencies are added here as AntiGravity implements the root-spine.
// Every dependency requires a one-line justification comment.
// Pinned versions are mandatory — no floating major versions.

require (
	// gRPC and Protobuf runtime — core transport layer
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.2

	// OpenTelemetry Go SDK — all instrumentation paths use this
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/trace v1.28.0
	go.opentelemetry.io/otel/metric v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.28.0

	// PostgreSQL driver — state persistence for non-critical assets
	github.com/jackc/pgx/v5 v5.6.0

	// WebSocket — signaling plane for control-panel
	github.com/gorilla/websocket v1.5.3

	// Structured logging — JSON format per architecture requirements
	go.uber.org/zap v1.27.0

	// UUID v7 — time-ordered identifiers for all events
	github.com/google/uuid v1.6.0

	// RFC 8785 canonical JSON — required for Merkle leaf serialization
	github.com/cyberphone/json-canonicalization v0.0.0-20231217082505-2617cb25e073
)
