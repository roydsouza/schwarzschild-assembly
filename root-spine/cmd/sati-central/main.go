package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	sati_grpc "github.com/rds/sati-central/root-spine/internal/grpc"
	"github.com/rds/sati-central/root-spine/internal/grpc/pb"
	"github.com/rds/sati-central/root-spine/internal/merkle"
	"github.com/rds/sati-central/root-spine/internal/orchestrator"
	"github.com/rds/sati-central/root-spine/internal/persistence"
	"github.com/rds/sati-central/root-spine/internal/safety"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Initialize Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger.Info("Sati-Central Root Spine starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Initialize Persistence
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres@localhost:5432/sati_central?sslmode=disable"
	}
	store, err := persistence.NewStore(ctx, dbURL)
	if err != nil {
		logger.Fatal("failed to initialize persistence", zap.Error(err))
	}
	defer store.Close()
	logger.Info("PostgreSQL connection established")

	// 3. Initialize Safety Bridge
	bridge, err := safety.NewBridge(logger, "libsafety_rail.a")
	if err != nil {
		logger.Fatal("failed to initialize safety bridge", zap.Error(err))
	}
	defer bridge.Close()

	// 4. Initialize Merkle Tree
	tree := merkle.NewTree()

	// 5. Initialize Analyst Verdict Manager
	analyst := orchestrator.NewVerdictManager(logger, "analyst-verdicts")
	analyst.Start(5 * time.Second)

	// 6. Start gRPC Server
	grpcServer := grpc.NewServer()
	satiServer := sati_grpc.NewServer(logger, store, bridge, tree, analyst)
	pb.RegisterOrchestratorServer(grpcServer, satiServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	// 7. Graceful Shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	sig := <-sigCh
	logger.Info("received signal, shutting down", zap.String("signal", sig.String()))
	
	grpcServer.GracefulStop()
	logger.Info("Sati-Central Root Spine stopped cleanly")
}
