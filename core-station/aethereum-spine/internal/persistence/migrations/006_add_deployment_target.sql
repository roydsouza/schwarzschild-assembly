-- Migration: Phase 7 — Add Deployment Target
-- Author: AntiGravity
-- Date: 2026-04-09
-- Rationale: SIGNIFICANT-2 from Phase 7 Analyst Verdict. Enables Scaffold Engine target selection.

-- Add deployment target and config to spec_documents
ALTER TABLE spec_documents 
ADD COLUMN IF NOT EXISTS deployment_target TEXT,
ADD COLUMN IF NOT EXISTS deployment_config_json JSONB NOT NULL DEFAULT '{}';

-- Migration audit entry
-- [PHASE-7-REMEDIATION]
