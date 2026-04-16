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
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger, _, err := logging.InitLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	shutdown, err := telemetry.InitProvider(context.Background(), "core-banking-system")
	if err != nil {
		logging.Logger().Fatalw("failed to init telemetry", "error", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
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
	// Register service manually (proto-generated code will replace this)
	registerLedgerService(s, ledgerServer)
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logging.Logger().Infow("shutdown signal received")
	s.GracefulStop()
	logging.Logger().Infow("gRPC server stopped")
}

// registerLedgerService registers the Ledger gRPC service.
// This is a simplified registration. Once proto stubs are generated,
// replace with: ledgerpb.RegisterLedgerServiceServer(s, ledgerServer)
func registerLedgerService(s *grpc.Server, srv *grpcserver.LedgerServiceServer) {
	desc := grpc.ServiceDesc{
		ServiceName: "ledger.v1.LedgerService",
		HandlerType: (*interface{})(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "CreateJournalEntry",
				Handler: func(s interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					req := &grpcserver.CreateJournalEntryRequest{}
					if err := dec(req); err != nil {
						return nil, err
					}
					return srv.CreateJournalEntry(ctx, req)
				},
			},
			{
				MethodName: "GetBalance",
				Handler: func(s interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					req := &grpcserver.GetBalanceRequest{}
					if err := dec(req); err != nil {
						return nil, err
					}
					return srv.GetBalance(ctx, req)
				},
			},
			{
				MethodName: "InitializeAccount",
				Handler: func(s interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					req := &grpcserver.InitializeAccountRequest{}
					if err := dec(req); err != nil {
						return nil, err
					}
					return srv.InitializeAccount(ctx, req)
				},
			},
		},
		Streams: []grpc.StreamDesc{},
	}
	_ = fmt.Sprintf("registering %s", desc.ServiceName)
	s.RegisterService(&desc, srv)
}
