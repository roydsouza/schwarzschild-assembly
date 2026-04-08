# Factory: Synthetic Analyst

Reference implementation for all Sati-Central factories. Every factory built after
this one follows the pattern established here.

## What it does

- Tracks DeFi protocol TVL with live data feeds
- Monitors macroeconomic indicators with 30-day backtested model precision
- RAG over financial and macro data sources
- Applies Pāḷi stylometric filter to all Dhamma-Adviser retrievals (pass rate ≥ 0.95)
- Emits domain fitness metrics to the Root Spine

## Domain fitness metrics

| Metric | Unit | Direction | Escalation Threshold |
|--------|------|-----------|---------------------|
| DeFi protocol coverage | % of tracked TVL with live feeds | higher | < 80% |
| Macroeconomic model precision | MRR (rolling 30d) | higher | < 0.7 |
| RAG retrieval quality | mean reciprocal rank | higher | < 0.6 |
| Pāḷi filter pass rate | % segments scoring ≥ 0.95 | higher | < 90% |
| Alert latency | p99 ms signal→Gate | lower | > 5000ms |

## Factory anatomy

```
synthetic-analyst/
├── mcp-server/       Domain-specific MCP server
├── worker/           AntiGravity integration layer
├── domain-fitness/   Domain metric collectors + RegisterDomainMetrics call
├── analyst-briefing/ Templates for AntiGravity → Analyst Droid packets
└── README.md         This file
```

## Depends on

- Root Spine (gRPC, Phase 3)
- Safety Rail (Tier 1, Phase 2)
- Dhamma-Adviser (Phase 6) for ethical weighting

## How to run tests

```bash
# Per sub-component (tbd in Phase 5)
cd factories/synthetic-analyst && ./run_tests.sh
```

## Implementation status

Phase 5 — not yet implemented. Implemented after Phase 4 (Translucent Gate) is approved.
