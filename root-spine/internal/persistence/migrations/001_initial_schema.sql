-- 001_initial_schema.sql
-- Sati-Central / Schwarzschild Assembly
-- Authoritative state persistence schema

-- factories tracks registered agent factory instances.
CREATE TABLE IF NOT EXISTS factories (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    factory_type TEXT NOT NULL,
    config_json JSONB NOT NULL,
    state TEXT NOT NULL, -- STARTING, RUNNING, STOPPING, STOPPED, ERROR
    last_heartbeat_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- proposals tracks action proposals submitted by agents.
CREATE TABLE IF NOT EXISTS proposals (
    id UUID PRIMARY KEY,
    factory_id UUID REFERENCES factories(id),
    agent_id TEXT NOT NULL,
    description TEXT NOT NULL,
    payload_hash_hex TEXT NOT NULL,
    target_path TEXT,
    is_security_adjacent BOOLEAN NOT NULL DEFAULT FALSE,
    verdict TEXT, -- SAFE, UNSAFE, TIMEOUT, TAMPERED, VETOED
    policy_fingerprint_hex TEXT,
    submitted_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,
    verdict_duration_ms BIGINT,
    proof_bytes BYTEA,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- merkle_leaves tracks the committed audit log.
CREATE TABLE IF NOT EXISTS merkle_leaves (
    leaf_index BIGINT PRIMARY KEY, -- 0-based index in the Merkle log
    proposal_id UUID REFERENCES proposals(id),
    leaf_hash_hex TEXT NOT NULL,
    event_type TEXT NOT NULL,
    policy_fingerprint_hex TEXT NOT NULL,
    merkle_root_hex TEXT NOT NULL,
    sth_signature_hex TEXT,
    committed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_proposals_factory_id ON proposals(factory_id);
CREATE INDEX IF NOT EXISTS idx_merkle_leaves_proposal_id ON merkle_leaves(proposal_id);
CREATE INDEX IF NOT EXISTS idx_merkle_leaves_leaf_hash ON merkle_leaves(leaf_hash_hex);
