package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	fitness "github.com/rds/sati-central/factories/code-assurance/domain-fitness"
	"github.com/rds/sati-central/factories/code-assurance/assessment-pipeline"
	factorymcp "github.com/rds/sati-central/factories/code-assurance/mcp-server"
	"github.com/rds/sati-central/factories/code-assurance/pb"
)

const (
	FactoryID   = "factory-code-assurance-01"
	FactoryType = "code-assurance"
	SpineAddr   = "localhost:50051"
)

func main() {
	log.Printf("[CODE-ASSURANCE] Initializing Factory %s...", FactoryID)

	// 1. Setup Assessment Pipeline
	aggregator := pipeline.NewAggregator(
		&pipeline.GoAnalyzer{},
		&pipeline.RustAnalyzer{},
		&pipeline.TSAnalyzer{},
	)
	projectRoot, _ := os.Getwd()
	// Navigate up to workspace root if needed
	// Assuming worker runs from factories/code-assurance/worker
	if strings.HasSuffix(projectRoot, "worker") {
		projectRoot = filepath.Join(projectRoot, "../../..")
	}

	// 2. Setup MCP Server
	mcpServer := factorymcp.NewAssuranceMCPServer(aggregator, projectRoot)
	collector := fitness.NewCollector()

	// 3. Connect to Root Spine
	conn, err := grpc.NewClient(SpineAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Root Spine: %v", err)
	}
	defer conn.Close()
	client := pb.NewOrchestratorClient(conn)

	// 4. Register Domain Metrics
	ctx := context.Background()
	decls := fitness.GetDeclarations(FactoryID, FactoryType)
	_, err = client.RegisterDomainMetrics(ctx, decls)
	if err != nil {
		log.Printf("Warning: Failed to register metrics: %v", err)
	}

	// 5. Main Assessment Loop
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Initial scan
	go runScan(ctx, client, aggregator, mcpServer, collector, projectRoot)

	for {
		select {
		case <-ticker.C:
			runScan(ctx, client, aggregator, mcpServer, collector, projectRoot)
		}
	}
}

func runScan(ctx context.Context, client pb.OrchestratorClient, agg *pipeline.Aggregator, mcp *factorymcp.AssuranceMCPServer, coll *fitness.Collector, projectRoot string) {
	log.Println("[CODE-ASSURANCE] Starting periodic assessment scan...")
	
	res, err := agg.Run(ctx, projectRoot)
	if err != nil {
		log.Printf("Scan failed: %v", err)
		return
	}

	mcp.SetLastReport(&res)
	
	// Map to domain metrics with threshold-based status (PRE-7-1)
	correctnessStatus := pb.MetricStatus_METRIC_GREEN
	if res.CorrectnessScore < 0.90 {
		correctnessStatus = pb.MetricStatus_METRIC_RED
	}

	qualityStatus := pb.MetricStatus_METRIC_GREEN
	if res.QualityScore < 0.85 {
		qualityStatus = pb.MetricStatus_METRIC_AMBER
	}

	coll.UpdateMetric(fitness.MetricArtifactCorrectness, res.CorrectnessScore, correctnessStatus, "Ratio")
	coll.UpdateMetric(fitness.MetricCodeQuality, res.QualityScore, qualityStatus, "Ratio")

	// Transform map to scalar float64 for reporting
	scalarMetrics := make(map[string]float64)
	for k, v := range coll.Collect() {
		scalarMetrics[k] = v.Value
	}

	// Report to Spine
	report := &pb.MetricReport{
		FactoryId:    FactoryID,
		Metrics:      scalarMetrics,
		ObservedAtMs: time.Now().UnixMilli(),
	}
	_, err = client.ReportMetrics(ctx, report)
	if err != nil {
		log.Printf("Failed to report metrics: %v", err)
	}

	log.Printf("[CODE-ASSURANCE] Scan complete. Correctness: %.2f, Quality: %.2f", res.CorrectnessScore, res.QualityScore)
}
