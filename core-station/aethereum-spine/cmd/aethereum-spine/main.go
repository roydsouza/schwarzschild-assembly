package main

import (
	"context"
	"encoding/hex"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sati_grpc "github.com/rds/aethereum-spine/aethereum-spine/internal/grpc"
	"github.com/rds/aethereum-spine/aethereum-spine/internal/grpc/pb"
	"github.com/rds/aethereum-spine/aethereum-spine/internal/gate"
	"github.com/rds/aethereum-spine/aethereum-spine/internal/merkle"
	"github.com/rds/aethereum-spine/aethereum-spine/internal/orchestrator"
	"github.com/rds/aethereum-spine/aethereum-spine/internal/persistence"
	"github.com/rds/aethereum-spine/aethereum-spine/internal/safety"
	"github.com/rds/aethereum-spine/aethereum-spine/internal/websocket"
	"github.com/rds/aethereum-spine/aethereum-spine/internal/mcp"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/rs/cors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/zap"
	google_grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Initialize Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger.Info("Aethereum-Spine Root Spine starting...")

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
		dbURL = "postgres://postgres@localhost:5432/aethereum_spine?sslmode=disable"
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

	analystManager := orchestrator.NewVerdictManager(logger, "analyst-verdicts")
	analystManager.Start(1 * time.Second)

	// Phase 5.1: Initialize Go-based Synthetic Analyst Factory
	analystFactory := orchestrator.NewSyntheticAnalystFactory(logger, "analyst-verdicts", nil) // spine will be set shortly
	analystFactory.Start(ctx)

	// 5a. Initialize Hub and Gate (Phase 3 CRITICAL-A/B)
	hub := websocket.NewHub(logger)
	gateController := gate.NewGate()

	// 6. Start gRPC Server
	satiServer := sati_grpc.NewServer(logger, store, bridge, tree, analystManager, hub, gateController, analystFactory)
	grpcServer := google_grpc.NewServer()
	pb.RegisterOrchestratorServer(grpcServer, satiServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	// 7. Graceful Shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// 7. Start Socket.IO Server (Phase 4 wiring)
	hub.Start()
	defer hub.Stop()

	go func() {
		logger.Info("Socket.IO server listening on :8080")
		if err := http.ListenAndServe(":8080", hub.Handler()); err != nil {
			logger.Error("Socket.IO server failed", zap.Error(err))
		}
	}()

	// 8. Start gRPC Server
	go func() {
		logger.Info("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	// 9. Start gRPC-Web Proxy (Phase 4.1 remediation)
	wrappedGrpc := grpcweb.WrapServer(grpcServer, grpcweb.WithOriginFunc(func(origin string) bool { return true }))
	corsWrapper := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "GET", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "x-grpc-web", "x-user-agent"},
	})
	
	grpcWebHandler := corsWrapper.Handler(wrappedGrpc)
	go func() {
		logger.Info("gRPC-Web proxy listening on :8081")
		if err := http.ListenAndServe(":8081", grpcWebHandler); err != nil {
			logger.Error("gRPC-Web server failed", zap.Error(err))
		}
	}()

	// 10. Start MCP Host (Phase 5)
	mcpHost := mcp.NewHost(logger, satiServer, store)
	mcpTransport := mcp.NewHTTPTransport(logger, mcpHost)
	go func() {
		logger.Info("MCP Host (HTTP) listening on :8082")
		if err := http.ListenAndServe(":8082", mcpTransport.Handler()); err != nil {
			logger.Error("MCP Host server failed", zap.Error(err))
		}
	}()

	sig := <-sigCh
	logger.Info("received signal, shutting down", zap.String("signal", sig.String()))
	
	grpcServer.GracefulStop()
	logger.Info("Aethereum-Spine Root Spine stopped cleanly")
}
