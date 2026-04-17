package main

// @title Account Issuance Service API
// @version 1.0
// @description API for account registration and management (PJP)
// @BasePath /

import (
	"account-issuance-service/config"
	"account-issuance-service/internal/account"
	grpcserver "account-issuance-service/internal/grpc"
	"account-issuance-service/internal/server"
	"account-issuance-service/pkg/database"
	"account-issuance-service/pkg/logging"
	"account-issuance-service/pkg/messaging"
	"account-issuance-service/pkg/telemetry"
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	accountpb "proto/account/v1"
)

func main() {
	logger, _, err := logging.InitLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	bgCtx := context.Background()
	shutdown, err := telemetry.InitProvider(bgCtx, "account-issuance-service")
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

	// Domain
	repo := account.NewRepository(db)
	svc := account.NewService(repo, producer, redisClient)
	httpHandler := account.NewHandler(svc)
	grpcSvc := grpcserver.NewAccountServiceServer(svc)

	// HTTP Server
	handler := server.NewRouter(httpHandler)
	httpSrv := &http.Server{Addr: ":8080", Handler: handler}

	go func() {
		logging.Logger().Infow("HTTP server starting", "port", ":8080")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger().Fatalw("HTTP server error", "error", err)
		}
	}()

	// gRPC Server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = ":50051"
	}
	s := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))

	// Import the generated PB package
	accountpb.RegisterAccountServiceServer(s, grpcSvc)
	reflection.Register(s)

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		logging.Logger().Fatalw("gRPC listen failed", "port", grpcPort, "error", err)
	}
	go func() {
		logging.Logger().Infow("gRPC server starting", "port", grpcPort)
		if err := s.Serve(lis); err != nil {
			logging.Logger().Fatalw("gRPC server error", "error", err)
		}
	}()

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(bgCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logging.Logger().Infow("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(bgCtx, 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	s.GracefulStop()
	logging.Logger().Infow("servers stopped")
}
