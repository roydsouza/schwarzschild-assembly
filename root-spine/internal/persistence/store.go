package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides access to the Sati-Central database.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a new persistence store.
func NewStore(ctx context.Context, connStr string) (*Store, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Close closes the database connection pool.
func (s *Store) Close() {
	s.pool.Close()
}

// Factory represents a registered factory.
type Factory struct {
	ID              uuid.UUID
	Name            string
	Type            string
	ConfigJSON      []byte
	State           string
	LastHeartbeatAt *time.Time
}

// GetOrCreateFactory returns a factory by name, creating it if it doesn't exist.
func (s *Store) GetOrCreateFactory(ctx context.Context, f Factory) (uuid.UUID, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO factories (id, name, factory_type, config_json, state)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (name) DO UPDATE SET last_heartbeat_at = NOW()
		RETURNING id`,
		f.ID, f.Name, f.Type, f.ConfigJSON, f.State)

	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("failed to get/create factory: %w", err)
	}
	return id, nil
}

// SaveProposal persists an action proposal.
func (s *Store) SaveProposal(ctx context.Context, p_id uuid.UUID, f_id uuid.UUID, agentID, desc, hash string, isSec bool, subAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO proposals (id, factory_id, agent_id, description, payload_hash_hex, is_security_adjacent, submitted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		p_id, f_id, agentID, desc, hash, isSec, subAt)
	if err != nil {
		return fmt.Errorf("failed to save proposal: %w", err)
	}
	return nil
}

// UpdateProposalVerdict updates the verdict for a proposal.
func (s *Store) UpdateProposalVerdict(ctx context.Context, id uuid.UUID, verdict, fingerprint string, duration uint64, proof []byte) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE proposals 
		SET verdict = $1, policy_fingerprint_hex = $2, verdict_duration_ms = $3, proof_bytes = $4, verified_at = NOW()
		WHERE id = $5`,
		verdict, fingerprint, duration, proof, id)
	if err != nil {
		return fmt.Errorf("failed to update proposal verdict: %w", err)
	}
	return nil
}
