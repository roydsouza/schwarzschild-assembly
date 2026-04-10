# Code Assurance Factory

The **Code Assurance Factory** provides autonomous security, quality, and correctness auditing for the Schwarzschild Assembly. It executes multi-language static analysis toolchains to calculate fitness metrics and surface actionable findings for analysts.

## Core Capabilities

### 1. Assessment Pipeline
The factory implements a pluggable `Analyzer` interface, currently supporting:
- **Go**: `go vet`, `staticcheck` (high-integrity lint), `gocyclo` (complexity), and `govulncheck` (CVEs).
- **Rust**: `cargo clippy` and `cargo audit` (security).
- **TypeScript/Frontend**: `tsc` (type safety) and `vitest` (logic correctness).

### 2. Metrics (Fitness Vector)
Emits two primary domain metrics to the Root Spine:
- `artifact_correctness`: (Threshold ≥ 0.90) Based on unit tests and security audits.
- `code_quality`: (Threshold ≥ 0.85) Based on lint hygiene and cyclomatic complexity.

## Usage

### MCP Interface
Exposes diagnostics to Analyst Droids via the following tools:
- `get_assurance_report`: Returns a typed `CodeQualityAssessment` JSON.
- `trigger_scan`: Initiates an immediate project-wide audit.

## Development

### Running Tests
To verify the aggregator and analyzer logic:
```bash
go test ./assessment-pipeline/...
```

### Build
```bash
go build ./...
```

## Maintenance
Configuration for specific linters (e.g., `.staticcheck.conf`, `clippy.toml`) should be maintained in the project root to ensure consistency between developer environments and factory audits.
