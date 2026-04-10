-- Migration: Phase 7 — Assembly Line Manager
-- Author: AntiGravity
-- Date: 2026-04-09

-- 1. Spec Documents
-- Stores the authoritative software requirements spec.
CREATE TABLE IF NOT EXISTS spec_documents (
    id UUID PRIMARY KEY,
    service_name TEXT NOT NULL UNIQUE,
    description TEXT,
    primary_language TEXT,
    is_finalized BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at TIMESTAMPTZ,
    data_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Assembly Lines
-- Tracks the lifecycle state of a service creation instance.
CREATE TABLE IF NOT EXISTS assembly_lines (
    id UUID PRIMARY KEY,
    spec_id UUID NOT NULL REFERENCES spec_documents(id),
    service_name TEXT NOT NULL,
    current_state TEXT NOT NULL DEFAULT 'INTAKE',
    justification TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_pulse_at TIMESTAMPTZ
);

-- 3. Skills (Phase 7/8 skeletal)
-- Stores agent skill versions for self-modification.
CREATE TABLE IF NOT EXISTS agent_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL,
    skill_name TEXT NOT NULL,
    version_hash_hex TEXT NOT NULL,
    content_bytes BYTEA NOT NULL,
    rationale TEXT,
    committed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id, skill_name, version_hash_hex)
);

-- Indexing for state machine lookups
CREATE INDEX IF NOT EXISTS idx_assembly_lines_state ON assembly_lines(current_state);
CREATE INDEX IF NOT EXISTS idx_spec_docs_name ON spec_documents(service_name);
