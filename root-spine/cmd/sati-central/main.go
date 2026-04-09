package main

import (
	"context"
	"encoding/hex"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	sati_grpc "github.com/rds/sati-central/root-spine/internal/grpc"
	"github.com/rds/sati-central/root-spine/internal/grpc/pb"
	"github.com/rds/sati-central/root-spine/internal/gate"
	"github.com/rds/sati-central/root-spine/internal/merkle"
	"github.com/rds/sati-central/root-spine/internal/orchestrator"
	"github.com/rds/sati-central/root-spine/internal/persistence"
	"github.com/rds/sati-central/root-spine/internal/safety"
	"github.com/rds/sati-central/root-spine/internal/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
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

	// 1a. Initialize OTel
	exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint("localhost:4317"), otlpmetricgrpc.WithInsecure())
	if err != nil {
		logger.Fatal("failed to create OTel exporter", zap.Error(err))
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)))
	otel.SetMeterProvider(provider)
	defer provider.Shutdown(ctx)
	logger.Info("OTel metrics provider initialized (target: localhost:4317)")

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
	
	// Restore Merkle state from DB (CRITICAL-4)
	leaves, err := store.GetMerkleLeaves(ctx)
	if err != nil {
		logger.Fatal("failed to load merkle leaves", zap.Error(err))
	}
	for _, leafHex := range leaves {
		leafBytes, err := hex.DecodeString(leafHex)
		if err != nil {
			logger.Fatal("corrupt merkle leaf in DB", zap.String("hash", leafHex))
		}
		var h merkle.Hash
		copy(h[:], leafBytes)
		tree.Append(h)
	}
	logger.Info("Merkle tree restored", zap.Int("leaves", tree.Size()), zap.String("root", tree.Root().Hex()))

	analyst := orchestrator.NewVerdictManager(logger, "analyst-verdicts")
	analyst.Start(5 * time.Second)

	// 5a. Initialize Hub and Gate (Phase 3 CRITICAL-A/B)
	hub := websocket.NewHub(logger)
	gateHandler := gate.NewGate()

	// 6. Start gRPC Server
	grpcServer := grpc.NewServer()
	satiServer := sati_grpc.NewServer(logger, store, bridge, tree, analyst, hub, gateHandler)
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
