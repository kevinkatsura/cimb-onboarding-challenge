package main

// @title Payment Initiation Acquiring Service API
// @version 1.0
// @description API for intrabank fund transfers (PJP)
// @BasePath /

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"payment-initiation-acquiring-service/config"
	"payment-initiation-acquiring-service/internal/server"
	"payment-initiation-acquiring-service/internal/transfer"
	"payment-initiation-acquiring-service/pkg/database"
	"payment-initiation-acquiring-service/pkg/logging"
	"payment-initiation-acquiring-service/pkg/messaging"
	"payment-initiation-acquiring-service/pkg/telemetry"

	accountpb "proto/account/v1"
	fraudpb "proto/fraud/v1"
	ledgerpb "proto/ledger/v1"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger, _, err := logging.InitLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	bgCtx := context.Background()
	shutdown, err := telemetry.InitProvider(bgCtx, "payment-initiation-service")
	if err != nil {
		logging.Logger().Fatalw("failed to init telemetry", "error", err)
	}
	defer func() { _ = shutdown(bgCtx) }()

	cfg := config.LoadConfig()
	redisCfg := config.LoadRedisConfig()

	db := database.NewPostgres(cfg)
	defer db.Close()

	redisClient := database.NewRedis(redisCfg)
	defer redisClient.Close()

	// Kafka
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	producer := messaging.NewKafkaProducer(strings.Split(kafkaBrokers, ","), logger)
	defer producer.Close()

	// gRPC Clients
	aisAddr := os.Getenv("AIS_GRPC_ADDR")
	if aisAddr == "" {
		aisAddr = "localhost:50051"
	}
	aisConn, err := grpc.NewClient(aisAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logging.Logger().Fatalw("failed to connect to account service", "error", err)
	}
	defer aisConn.Close()
	accountClient := accountpb.NewAccountServiceClient(aisConn)

	cbsAddr := os.Getenv("CBS_GRPC_ADDR")
	if cbsAddr == "" {
		cbsAddr = "localhost:50052"
	}
	cbsConn, err := grpc.NewClient(cbsAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logging.Logger().Fatalw("failed to connect to CBS", "error", err)
	}
	defer cbsConn.Close()
	ledgerClient := ledgerpb.NewLedgerServiceClient(cbsConn)

	// Fraud Detection Client (gRPC)
	fraudAddr := os.Getenv("FRAUD_GRPC_ADDR")
	if fraudAddr == "" {
		fraudAddr = "localhost:50055"
	}
	fraudConn, err := grpc.NewClient(fraudAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logging.Logger().Fatalw("failed to connect to fraud service", "error", err)
	}
	defer fraudConn.Close()
	fraudClient := fraudpb.NewFraudDetectionClient(fraudConn)
	logging.Logger().Infow("Fraud detection client configured (gRPC)", "addr", fraudAddr)

	// Domain
	repo := transfer.NewRepository(db)
	svc := transfer.NewService(repo, accountClient, ledgerClient, fraudClient, producer, redisClient)
	handler := transfer.NewHandler(svc)

	// HTTP Server
	router := server.NewRouter(handler)
	httpSrv := &http.Server{Addr: ":8080", Handler: router}

	go func() {
		logging.Logger().Infow("HTTP server starting", "port", ":8080")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger().Fatalw("HTTP server error", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(bgCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logging.Logger().Infow("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(bgCtx, 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	logging.Logger().Infow("server stopped")
}
