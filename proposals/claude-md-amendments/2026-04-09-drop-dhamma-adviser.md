# CLAUDE.md Amendment Proposal
**Date:** 2026-04-09 05:30:00 UTC
**Proposal ID:** amendment-2026-04-09-drop-dhamma-adviser
**Proposed by:** Roy Peter D'Souza (Human)
**Status:** APPROVED — explicit operator directive

## Problem Statement

The Dhamma-Adviser (Phase 6) and its associated `dhamma_alignment` fitness metric add significant implementation complexity — real stylometric ML classifier, BDRC dataset ingestion, Bilara segment addressing infrastructure, RAG pipeline over Buddhist canonical texts — for zero measurable benefit to the project's actual use case.

The project builds innocuous knowledge management, analysis, and similar software applications. There is no meaningful causal connection between Pāḷi ethical categories (kusala/akusala/neutral) and software correctness, security, or performance. A `MoralWeighting` score cannot tell us whether generated code is correct, whether dependencies have CVEs, or whether a RAG pipeline's retrieval quality is acceptable. The metric occupies 0.15 of fitness vector weight while delivering no signal.

Roy's stated goal: a high-quality dark factory where droids divide and conquer, review each other's work, and produce correct, secure, efficient, modular, observable software. Lights-out trajectory.

## Changes

1. **Drop Phase 6 (Dhamma-Adviser)** — remove entirely. Drop `dhamma-adviser/` from canonical directory structure.
2. **Replace Phase 6 with Code Assurance Factory** — a quality-challenger droid that reviews artifacts produced by all other factories: static analysis, dependency audit, test coverage, complexity scoring.
3. **Replace `dhamma_alignment` in fitness vector** with two software quality metrics: `artifact_correctness` (test pass rate) and `code_quality` (lint+complexity composite).
4. **Replace `DhammaReflection` UI component** with `CodeQualityPanel` — renders test pass rate, lint score, security findings, complexity score for the proposal under review.
5. **Remove `GetDhammaContext` RPC** from the proto spec (mark deprecated in existing generated code; remove from canonical spec).
6. **Remove `dhamma_ref` field** from Merkle leaf schema.
7. **Update Synthetic Analyst domain fitness metrics** — remove `pali_filter_rate`, replace with `answer_accuracy` (% of generated answers verified against ground truth sample).

## Approval

Explicitly approved by Roy Peter D'Souza, 2026-04-09.
Amendment takes effect immediately per CLAUDE.md §1.
