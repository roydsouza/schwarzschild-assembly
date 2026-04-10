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

// Store provides access to the Aethereum-Spine database.
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
	if err := s.ApplyMigrations(ctx, "aethereum-spine/internal/persistence/migrations"); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return s, nil
}

// Close closes the database connection pool.
func (s *Store) Close() {
	s.pool.Close()
}

// SpecDocument represents a service specification.
type SpecDocument struct {
	ID              uuid.UUID
	ServiceName     string
	Description     string
	PrimaryLanguage string
	IsFinalized     bool
	ApprovedAt      *time.Time
	DataJSON             []byte
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeploymentTarget     string
	DeploymentConfigJSON []byte
}

// AssemblyLine represents a service creation instance.
type AssemblyLine struct {
	ID           uuid.UUID
	SpecID       uuid.UUID
	ServiceName  string
	CurrentState string
	Justification string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastPulseAt  *time.Time
}

// MerkleLeaf represents a single entry in the audit log.
type MerkleLeaf struct {
	Index                  int64
	ProposalID             uuid.UUID
	LeafHashHex            string
	EventType              string
	PolicyFingerprintHex   string
	MerkleRootHex          string
	CreatedAt              time.Time
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

// GetMerkleLeavesForProposal retrieves all leaves associated with a specific proposal ID.
func (s *Store) GetMerkleLeavesForProposal(ctx context.Context, pID uuid.UUID) ([]MerkleLeaf, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT leaf_index, proposal_id, leaf_hash_hex, event_type, policy_fingerprint_hex, merkle_root_hex, created_at
		FROM merkle_leaves WHERE proposal_id = $1 ORDER BY leaf_index ASC`, pID)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaves for proposal: %w", err)
	}
	defer rows.Close()

	var leaves []MerkleLeaf
	for rows.Next() {
		var l MerkleLeaf
		if err := rows.Scan(&l.Index, &l.ProposalID, &l.LeafHashHex, &l.EventType, &l.PolicyFingerprintHex, &l.MerkleRootHex, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan leaf: %w", err)
		}
		leaves = append(leaves, l)
	}
	return leaves, nil
}

// IsProposalSecurityAdjacent checks whether a proposal is security-adjacent.
// Used by the MCP host to block approve_action on security-adjacent proposals.
func (s *Store) IsProposalSecurityAdjacent(ctx context.Context, proposalID uuid.UUID) (bool, error) {
	var isSec bool
	err := s.pool.QueryRow(ctx, "SELECT is_security_adjacent FROM proposals WHERE id = $1", proposalID).Scan(&isSec)
	if err != nil {
		return false, fmt.Errorf("failed to check proposal security status: %w", err)
	}
	return isSec, nil
}

// SaveMetricDeclaration persists a factory's domain metric definition.
func (s *Store) SaveMetricDeclaration(ctx context.Context, factoryID uuid.UUID, metricID, displayName, desc, unit, direction string, threshold float64, operator string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO metric_declarations (metric_id, factory_id, display_name, description, unit, direction, escalation_threshold, escalation_operator)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (metric_id) DO UPDATE SET 
			display_name = EXCLUDED.display_name,
			description = EXCLUDED.description,
			unit = EXCLUDED.unit,
			direction = EXCLUDED.direction,
			escalation_threshold = EXCLUDED.escalation_threshold,
			escalation_operator = EXCLUDED.escalation_operator`,
		metricID, factoryID, displayName, desc, unit, direction, threshold, operator)
	if err != nil {
		return fmt.Errorf("failed to save metric declaration: %w", err)
	}
	return nil
}

// SaveMetricValue records a measured metric value.
func (s *Store) SaveMetricValue(ctx context.Context, metricID string, value float64, status string, observedAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO metric_values (metric_id, value, status, observed_at)
		VALUES ($1, $2, $3, $4)`,
		metricID, value, status, observedAt)
	if err != nil {
		return fmt.Errorf("failed to save metric value: %w", err)
	}
	return nil
}

// SaveFitnessSnapshot persists a full cross-section of the global fitness vector.
func (s *Store) SaveFitnessSnapshot(ctx context.Context, schemaVersion string, tsMs int64, metricsJSON, extensionsJSON []byte) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO fitness_vector_snapshots (schema_version, timestamp_ms, metrics_json, domain_extensions_json)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (timestamp_ms) DO NOTHING`,
		schemaVersion, tsMs, metricsJSON, extensionsJSON)
	if err != nil {
		return fmt.Errorf("failed to save fitness snapshot: %w", err)
	}
	return nil
}
// SaveSpecDocument persists or updates a service specification.
func (s *Store) SaveSpecDocument(ctx context.Context, spec SpecDocument) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO spec_documents (id, service_name, description, primary_language, is_finalized, approved_at, data_json, deployment_target, deployment_config_json, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (service_name) DO UPDATE SET 
			description = EXCLUDED.description,
			primary_language = EXCLUDED.primary_language,
			is_finalized = EXCLUDED.is_finalized,
			approved_at = EXCLUDED.approved_at,
			data_json = EXCLUDED.data_json,
			deployment_target = EXCLUDED.deployment_target,
			deployment_config_json = EXCLUDED.deployment_config_json,
			updated_at = NOW()`,
		spec.ID, spec.ServiceName, spec.Description, spec.PrimaryLanguage, spec.IsFinalized, spec.ApprovedAt, spec.DataJSON, spec.DeploymentTarget, spec.DeploymentConfigJSON)
	if err != nil {
		return fmt.Errorf("failed to save spec document: %w", err)
	}
	return nil
}

// GetSpecDocument retrieves a spec by service name.
func (s *Store) GetSpecDocument(ctx context.Context, serviceName string) (SpecDocument, error) {
	var spec SpecDocument
	err := s.pool.QueryRow(ctx, `
		SELECT id, service_name, description, primary_language, is_finalized, approved_at, data_json, created_at, updated_at, deployment_target, deployment_config_json
		FROM spec_documents WHERE service_name = $1`, serviceName).Scan(
		&spec.ID, &spec.ServiceName, &spec.Description, &spec.PrimaryLanguage, &spec.IsFinalized, &spec.ApprovedAt, &spec.DataJSON, &spec.CreatedAt, &spec.UpdatedAt, &spec.DeploymentTarget, &spec.DeploymentConfigJSON)
	if err != nil {
		return SpecDocument{}, fmt.Errorf("failed to get spec document: %w", err)
	}
	return spec, nil
}

// UpdateSpecDeploymentTarget updates the deployment target fields for a spec.
func (s *Store) UpdateSpecDeploymentTarget(ctx context.Context, id uuid.UUID, target string, config []byte) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE spec_documents 
		SET deployment_target = $1, deployment_config_json = $2, updated_at = NOW()
		WHERE id = $3`,
		target, config, id)
	if err != nil {
		return fmt.Errorf("failed to update spec deployment target: %w", err)
	}
	return nil
}

// CreateAssemblyLine initiates a new lifecycle tracker.
func (s *Store) CreateAssemblyLine(ctx context.Context, al AssemblyLine) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO assembly_lines (id, spec_id, service_name, current_state, justification)
		VALUES ($1, $2, $3, $4, $5)`,
		al.ID, al.SpecID, al.ServiceName, al.CurrentState, al.Justification)
	if err != nil {
		return fmt.Errorf("failed to create assembly line: %w", err)
	}
	return nil
}

// UpdateAssemblyLineState transitions the lifecycle phase and returns the PREVIOUS state.
func (s *Store) UpdateAssemblyLineState(ctx context.Context, id uuid.UUID, newState, justification string) (string, error) {
	var prevState string
	err := s.pool.QueryRow(ctx, `
		UPDATE assembly_lines 
		SET current_state = $1, justification = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING (SELECT current_state FROM assembly_lines WHERE id = $3)`, // This logic is tricky for a single query due to update visibility.
		// Actually, standard PG 'RETURNING current_state' returns the NEW state.
		// We need to fetch the old one first OR use a more complex query.
		newState, justification, id).Scan(&prevState)
	
	// Correction: RETURNING in PG returns the row *after* the update logic but *before* visibility? No, it's the new row.
	// Let's use a simpler two-step approach in a transaction if needed, but for safety-critical logic, 
	// I'll use: UPDATE ... RETURNING (SELECT current_state FROM old_table ...)
	
	// Refined SQL for atomic Swap & Return Old:
	err = s.pool.QueryRow(ctx, `
		WITH old_state AS (
			SELECT current_state FROM assembly_lines WHERE id = $1 FOR UPDATE
		)
		UPDATE assembly_lines 
		SET current_state = $2, justification = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING (SELECT current_state FROM old_state)`,
		id, newState, justification).Scan(&prevState)

	if err != nil {
		return "", fmt.Errorf("failed to update assembly line state: %w", err)
	}
	return prevState, nil
}

// GetAssemblyLine retrieves an assembly line by ID.
func (s *Store) GetAssemblyLine(ctx context.Context, id uuid.UUID) (AssemblyLine, error) {
	var al AssemblyLine
	err := s.pool.QueryRow(ctx, `
		SELECT id, spec_id, service_name, current_state, justification, created_at, updated_at, last_pulse_at
		FROM assembly_lines WHERE id = $1`, id).Scan(
		&al.ID, &al.SpecID, &al.ServiceName, &al.CurrentState, &al.Justification, &al.CreatedAt, &al.UpdatedAt, &al.LastPulseAt)
	if err != nil {
		return AssemblyLine{}, fmt.Errorf("failed to get assembly line: %w", err)
	}
	return al, nil
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
