package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	fitness "github.com/rds/sati-central/factories/scaffold-engine/domain-fitness"
	"github.com/rds/sati-central/factories/scaffold-engine/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	spineAddr := flag.String("spine", "localhost:50051", "Root Spine gRPC address")
	factoryName := flag.String("name", "scaffold-engine-alpha", "Unique name for this factory instance")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("starting scaffold engine factory", zap.String("name", *factoryName))

	// 1. Connect to Root Spine
	conn, err := grpc.Dial(*spineAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed to connect to spine", zap.Error(err))
	}
	defer conn.Close()

	client := pb.NewOrchestratorClient(conn)

	// 2. Register Factory
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	regReq := &pb.FactoryRequest{
		FactoryType: "scaffold-engine",
		FactoryName: *factoryName,
		ConfigJson:  []byte("{}"),
	}
	regRes, err := client.CreateFactory(ctx, regReq)
	cancel()

	if err != nil {
		logger.Fatal("failed to register factory", zap.Error(err))
	}
	fID := regRes.FactoryId.Id
	logger.Info("factory registered", zap.String("id", fID))

	// 3. Setup Metrics
	metrics := fitness.NewMetricsManager(logger, conn, fID)
	if err := metrics.Register(context.Background()); err != nil {
		logger.Error("failed to register domain metrics", zap.Error(err))
	}

	// 4. Main Event Loop (Simulation for Phase 7)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	logger.Info("scaffold engine worker active")

	for {
		select {
		case <-ticker.C:
			// Heartbeat / Metrics reporting
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			if err := metrics.Report(ctx, 1.0, 500.0, 100.0, 0.0); err != nil {
				logger.Warn("failed to report metrics", zap.Error(err))
			}
			cancel()
		case sig := <-sigChan:
			logger.Info("shutting down on signal", zap.String("signal", sig.String()))
			// Notify spine of shutdown
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			client.StopFactory(ctx, &pb.FactoryID{Id: fID})
			cancel()
			return
		}
	}
}
