module github.com/rds/sati-central/root-spine

go 1.26

// Dependencies are added here as AntiGravity implements the root-spine.
// Every dependency requires a one-line justification comment.
// Pinned versions are mandatory — no floating major versions.

require (

	// UUID v7 — time-ordered identifiers for all events
	github.com/google/uuid v1.6.0

	// PostgreSQL driver — state persistence for non-critical assets
	github.com/jackc/pgx/v5 v5.6.0

	// Structured logging — JSON format per architecture requirements
	go.uber.org/zap v1.27.0
	// gRPC and Protobuf runtime — core transport layer
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240701130421-f6361c86f094 // indirect
)
