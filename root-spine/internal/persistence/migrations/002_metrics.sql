-- 002_metrics.sql
-- Persistence for factory domain metrics and the global fitness vector.

-- metric_declarations stores the registry of metrics emitted by each factory.
CREATE TABLE IF NOT EXISTS metric_declarations (
    metric_id TEXT PRIMARY KEY, -- fully qualified, e.g., "synthetic-analyst.defi_coverage"
    factory_id UUID REFERENCES factories(id),
    display_name TEXT NOT NULL,
    description TEXT NOT NULL,
    unit TEXT NOT NULL,
    direction TEXT NOT NULL, -- higher_is_better, lower_is_better
    escalation_threshold DOUBLE PRECISION NOT NULL,
    escalation_operator TEXT NOT NULL, -- gt, gte, lt, lte
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- metric_values stores the time-series history of reported metric values.
CREATE TABLE IF NOT EXISTS metric_values (
    id BIGSERIAL PRIMARY KEY,
    metric_id TEXT REFERENCES metric_declarations(metric_id),
    value DOUBLE PRECISION NOT NULL,
    status TEXT NOT NULL, -- GREEN, AMBER, RED, UNINITIALIZED
    observed_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- fitness_vector_snapshots stores the computed global fitness vector state.
CREATE TABLE IF NOT EXISTS fitness_vector_snapshots (
    schema_version TEXT NOT NULL,
    timestamp_ms BIGINT PRIMARY KEY,
    metrics_json JSONB NOT NULL, -- Stores the 6 global metrics
    domain_extensions_json JSONB NOT NULL, -- Stores current domain values
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metric_values_metric_id ON metric_values(metric_id);
CREATE INDEX IF NOT EXISTS idx_metric_values_observed_at ON metric_values(observed_at);
