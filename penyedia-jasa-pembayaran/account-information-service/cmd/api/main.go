package main

// @title Account Information Service API
// @version 1.0
// @description API for retrieving read-only account information and aggregated transactions for fraud services
// @BasePath /

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	account_informationpb "proto/account_information/v1"

	_ "github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/docs"
	"github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/internal/config"
	grpcserver "github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/internal/grpc"
	httpapi "github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/internal/http"
	"github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/internal/kafka"
	"github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/internal/repository"
)

func main() {
	zLog, _ := zap.NewProduction()
	logger := zLog.Sugar()
	defer logger.Sync()

	cfg := config.LoadConfig()

	// Init DB
	db := repository.NewPostgres(cfg)
	defer db.Close()

	// Init Kafka Consumer
	consumer := kafka.NewConsumer(cfg, db, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer.Start(ctx)
	defer consumer.Close()

	// Init HTTP & Swagger
	mux := http.NewServeMux()
	httpHandler := httpapi.NewHandler(db)
	httpapi.RegisterRoutes(mux, httpHandler)

	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	httpSrv := &http.Server{Addr: cfg.HTTPPort, Handler: mux}
	go func() {
		logger.Infow("HTTP server starting", "port", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalw("HTTP server error", "error", err)
		}
	}()

	// Init gRPC Server
	s := grpc.NewServer()
	grpcSvc := grpcserver.NewAccountInformationServer(db)
	account_informationpb.RegisterAccountInformationServiceServer(s, grpcSvc)
	reflection.Register(s)

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Fatalw("gRPC listen failed", "error", err)
	}

	go func() {
		logger.Infow("gRPC server starting", "port", cfg.GRPCPort)
		if err := s.Serve(lis); err != nil {
			logger.Fatalw("gRPC server error", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infow("shutting down servers...")
	cancel()
	httpSrv.Shutdown(context.Background())
	s.GracefulStop()
	logger.Infow("servers stopped")
}
