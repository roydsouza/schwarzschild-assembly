package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

	s := &Store{pool: pool}

	// Apply migrations
	if err := s.ApplyMigrations(ctx, "root-spine/internal/persistence/migrations"); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return s, nil
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

// GetDefaultFactoryID retrieves or creates a "global" factory instance.
func (s *Store) GetDefaultFactoryID(ctx context.Context) (uuid.UUID, error) {
	return s.GetOrCreateFactory(ctx, Factory{
		ID:         uuid.Nil,
		Name:       "global-factory",
		Type:       "system",
		ConfigJSON: []byte("{}"),
		State:      "RUNNING",
	})
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

// SaveMerkleLeaf persists a leaf in the audit log.
func (s *Store) SaveMerkleLeaf(ctx context.Context, index int64, pID uuid.UUID, hash, eventType, fingerprint, root string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO merkle_leaves (leaf_index, proposal_id, leaf_hash_hex, event_type, policy_fingerprint_hex, merkle_root_hex)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		index, pID, hash, eventType, fingerprint, root)
	if err != nil {
		return fmt.Errorf("failed to save merkle leaf: %w", err)
	}
	return nil
}

// GetMerkleLeaves retrieves all committed leaves in order.
func (s *Store) GetMerkleLeaves(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, "SELECT leaf_hash_hex FROM merkle_leaves ORDER BY leaf_index ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query merkle leaves: %w", err)
	}
	defer rows.Close()

	var leaves []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, fmt.Errorf("failed to scan leaf: %w", err)
		}
		leaves = append(leaves, h)
	}
	return leaves, nil
}
// ApplyMigrations runs any pending SQL migrations in the specified directory.
func (s *Store) ApplyMigrations(ctx context.Context, migrationsDir string) error {
	// 1. Create migrations tracking table
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 2. Read migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to glob migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		version := filepath.Base(file)

		// 3. Check if already applied
		var exists bool
		err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration state: %w", err)
		}

		if exists {
			continue
		}

		// 4. Apply migration
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", version, err)
		}

		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", version, err)
		}

		fmt.Printf("Applied migration: %s\n", version)
	}

	return nil
}
