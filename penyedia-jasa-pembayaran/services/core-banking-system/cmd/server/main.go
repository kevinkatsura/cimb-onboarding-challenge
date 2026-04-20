package main

// @title Core Banking System API
// @version 1.0
// @description Internal ledger and balance materialization service (PJP)
// @BasePath /

import (
	"context"
	"core-banking-system/config"
	_ "core-banking-system/docs"
	grpcserver "core-banking-system/internal/grpc"
	"core-banking-system/internal/journal"
	"core-banking-system/pkg/database"
	"core-banking-system/pkg/logging"
	"core-banking-system/pkg/telemetry"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	ledgerpb "proto/ledger/v1"
)

func main() {
	logger, _, err := logging.InitLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	bgCtx := context.Background()
	shutdown, err := telemetry.InitProvider(bgCtx, "core-banking-system")
	if err != nil {
		logging.Logger().Fatalw("failed to init telemetry", "error", err)
	}
	defer func() {
		if err := shutdown(bgCtx); err != nil {
			logging.Logger().Errorw("telemetry shutdown error", "error", err)
		}
	}()

	cfg := config.LoadConfig()
	db := database.NewPostgres(cfg)
	defer db.Close()

	// Domain
	repo := journal.NewRepository(db)
	svc := journal.NewService(repo)
	ledgerServer := grpcserver.NewLedgerServiceServer(svc)

	// gRPC Server with OTel interceptors
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = ":50052"
	}

	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	ledgerpb.RegisterLedgerServiceServer(s, ledgerServer)
	reflection.Register(s)

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		logging.Logger().Fatalw("failed to listen", "port", grpcPort, "error", err)
	}

	go func() {
		logging.Logger().Infow("gRPC server starting", "port", grpcPort)
		if err := s.Serve(lis); err != nil {
			logging.Logger().Fatalw("gRPC server error", "error", err)
		}
	}()

	// HTTP Server for Health + Swagger
	go func() {
		mux := http.NewServeMux()
		// HealthCheck godoc
		// @Summary      Service Health Check
		// @Description  Returns the health status of the CBS service
		// @Tags         System
		// @Produce      json
		// @Success      200 {object} map[string]string
		// @Router       /health [get]
		mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})
		mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
		logging.Logger().Infow("HTTP server starting (health+swagger)", "port", ":8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			logging.Logger().Fatalw("HTTP server error", "error", err)
		}
	}()

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(bgCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logging.Logger().Infow("shutdown signal received")
	s.GracefulStop()
	logging.Logger().Infow("gRPC server stopped")
}
