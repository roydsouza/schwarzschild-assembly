# Observability

OTel collector configuration and schemas for Sati-Central.

## What it does

- Receives OTLP metrics, traces, and logs from all components
- Exports fitness vector metrics to `otel-snapshots/latest.json` (cold-readable by agents)
- Exports all telemetry to `otel-snapshots/telemetry.jsonl`
- Exposes Prometheus scrape endpoint at `:8888` for future dashboard work

## Depends on

- `otelcol-contrib` (OpenTelemetry Collector Contrib) — install via `brew install opentelemetry-collector`

## How to start

```bash
otelcol-contrib --config observability/otel-collector-config.yaml
```

## How to run tests

```bash
./observability/tests/smoke_test.sh
```

## Schemas

- `fitness-vector-schema.json` — 5 global fitness metrics with evaluation direction,
  weight, and auto-escalation threshold
- `schemas/log-schema.json` — Typed structured log entry schema

## Key files

| File | Purpose |
|------|---------|
| `otel-collector-config.yaml` | OTel collector config (hybrid exporter) |
| `fitness-vector-schema.json` | Global fitness metric definitions |
| `schemas/log-schema.json` | Structured log entry schema |
| `tests/smoke_test.sh` | Phase 1 end-to-end smoke test |

## Metrics exported

See `fitness-vector-schema.json` for the 5 global metrics.
Each component adds its own metrics via the `sati_central.*` namespace.
