# Synthetic Analyst Factory: Metrics Reporting RPC

## Goal
Enable Go-based (and future) analyst factories to push their collected domain fitness metrics to the Root Spine's fitness vector.

## Proposed RPC
Add the following method to the `Orchestrator` service in `proto/orchestrator.proto`:

```proto
// ReportMetrics submits domain-specific metric values to the global fitness vector.
// Called periodically by factories after their internal collection loops.
rpc ReportMetrics(MetricReport) returns (OperationStatus);

message MetricReport {
    // factory_id of the submitting factory.
    string factory_id = 1;
    // metrics maps metric ID to current value.
    // Example: {"defi-coverage": 0.82, "macro-precision": 0.94}
    map<string, double> metrics = 2;
    // observed_at_ms is when the metrics were sampled.
    int64 observed_at_ms = 3;
}
```

## Impact
- **Root Spine:** The `Server` will update the global fitness vector and broadcast the change via Socket.IO to the Control Panel.
- **OTel:** Updates Prometheus gauges for the corresponding domain metrics.
- **Safety:** Allows the system to react to degrading domain performance (e.g., if DeFi coverage drops below threshold).

## Verification
- Protobuf re-generation.
- Synthetic Analyst worker updated to call `ReportMetrics` in its main loop.
- Control Panel "Fitness Vector" visualizes the reported values.
