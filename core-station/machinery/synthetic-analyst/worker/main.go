package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/rds/aethereum-spine/factories/synthetic-analyst/pb"
	"github.com/rds/aethereum-spine/factories/synthetic-analyst/domain-fitness"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	factoryID := uuid.New().String()
	logger.Info("Synthetic Analyst Factory starting...", zap.String("factory_id", factoryID))

	// 1. Connect to Root Spine
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed to connect to root spine", zap.Error(err))
	}
	defer conn.Close()

	client := pb.NewOrchestratorClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Register with Root Spine
	resp, err := client.CreateFactory(ctx, &pb.FactoryRequest{
		FactoryType: "synthetic-analyst",
		FactoryName: "analyst-droid-alpha",
		RequestId:   factoryID,
	})
	if err != nil {
		logger.Fatal("failed to register factory", zap.Error(err))
	}
	logger.Info("Factory registered", zap.String("id", resp.FactoryId.Id))

	// 3. Register Domain Metrics
	collector := fitness.NewCollector(resp.FactoryId.Id)
	regRes, err := client.RegisterDomainMetrics(ctx, collector.GetDeclarations())
	if err != nil {
		logger.Fatal("failed to register domain metrics", zap.Error(err))
	}
	logger.Info("Domain metrics registered", zap.Int32("count", regRes.RegisteredCount))

	// 4. Background Monitoring & Reporting Loop
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metrics := collector.Collect()
				logger.Info("collected domain metrics", zap.Int("count", len(metrics)))
				
				// Push metrics to Root Spine
				metricMap := make(map[string]float64)
				for id, val := range metrics {
					metricMap[id] = val.Value
				}

				_, err := client.ReportMetrics(ctx, &pb.MetricReport{
					FactoryId:    resp.FactoryId.Id,
					Metrics:      metricMap,
					ObservedAtMs: time.Now().UnixMilli(),
				})
				if err != nil {
					logger.Error("failed to report metrics", zap.Error(err))
				} else {
					logger.Info("reported metrics to root spine")
				}
			}
		}
	}()

	// Handle Graceful Shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	logger.Info("Synthetic Analyst Factory shutting down")
}
