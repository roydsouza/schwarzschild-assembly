# Dhamma-Adviser

Scored evidence producer for Sati-Central. Contributes a quantitative `MoralWeighting`
to every proposal's fitness evaluation. Not a chatbot persona — an evidence system.

## What it does

- RAG over BDRC open Buddhist datasets using the Bilara data model (SuttaCentral)
- Stylometric filter: only segments scoring ≥ 0.95 on the early-style classifier
  are eligible for retrieval
- Produces `MoralWeighting`: `{ score: float, root: kusala|akusala|neutral,
  citations: [BilaraSegmentID], reasoning: string }`
- Contributes its score to the global fitness vector (weight 0.15)
- Auto-escalates to Analyst Droid when score drops below 0.6

## Depends on

- Python 3.12+ with uv
- Event Horizon Core (EHC) on port 8000 for local LLM inference
- BDRC open Buddhist datasets (acquired during Phase 6 setup)

## How to run tests

```bash
cd dhamma-adviser && uv run pytest
```

## Metrics emitted

| OTel Metric Name | Type | Fitness Vector |
|-----------------|------|----------------|
| `sati_central.dhamma.alignment_score` | gauge | dhamma_alignment |
| `sati_central.dhamma.retrievals_total` | counter | — |
| `sati_central.dhamma.stylometric_filter_pass_rate` | gauge | — |

## Implementation status

Phase 6 — not yet implemented. Implemented after Phase 5 (Synthetic Analyst Factory).
