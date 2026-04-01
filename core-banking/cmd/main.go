package main

import (
	"context"
	"core-banking/config"
	"core-banking/internal/database/seeder"
	accountHandler "core-banking/internal/handler/account"
	txHandler "core-banking/internal/handler/transaction"
	"core-banking/internal/repository"
	svc "core-banking/internal/service"
	accountSvc "core-banking/internal/service/account"
	transactionSvc "core-banking/internal/service/transaction"
	"core-banking/pkg/idgen"
	"core-banking/pkg/logging"
	"core-banking/pkg/telemetry"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger, _, err := logging.InitLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	shutdown, err := telemetry.InitProvider(context.Background(), "core-banking-api")
	if err != nil {
		logging.Logger().Fatalw("failed to initialize telemetry", "error", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			logging.Logger().Errorw("failed to shutdown telemetry gracefully", "error", err)
		}
	}()

	cfg := config.LoadConfig()
	redisCfg := config.LoadRedisConfig()

	// ---- DB Bootstrap ----
	repository.EnsureDatabase(cfg)
	db := repository.NewPostgres(cfg)

	// ---- Redis Connection ----
	redisClient := repository.NewRedis(redisCfg)
	defer repository.CloseRedis(redisClient)

	// ---- Migrations ----
	repository.RunMigrateUp(cfg)

	// ---- Seeder ----
	s := seeder.New(db)
	if err := s.Seed(context.Background()); err != nil {
		logging.Logger().Fatalw("seed failed", "error", err)
	}
	logging.Logger().Infow("seeding completed")

	lock := svc.NewAccountLockManager()

	// Transaction
	txRepo := repository.NewTransactionRepository(db)
	txService := transactionSvc.NewService(txRepo, lock)
	txHandler := txHandler.NewHandler(txService)

	// Account
	accountRepo := repository.NewAccountRepository(db)
	accountService := accountSvc.NewService(accountRepo, &idgen.RandomAccountNumberGenerator{})
	accountService.SetRedisClient(redisClient) // Inject Redis client
	accountHandler := accountHandler.NewHandler(accountService)

	// ---- HTTP Handler ----
	mux := http.NewServeMux()

	// ---- Prometheus Scrape Endpoint ----
	mux.Handle("GET /metrics", promhttp.Handler())

	// ---- Transaction Endpoints ----
	mux.Handle("POST /v1/transfer", otelhttp.NewHandler(http.HandlerFunc(txHandler.Transfer), "TransactionHandler.Transfer"))
	mux.Handle("POST /v2/transfer", otelhttp.NewHandler(http.HandlerFunc(txHandler.TransferWithLock), "TransactionHandler.TransferWithLock"))
	mux.Handle("GET /transactions", otelhttp.NewHandler(http.HandlerFunc(txHandler.ListAll), "TransactionHandler.ListAll"))
	mux.Handle("GET /accounts/{id}/transactions", otelhttp.NewHandler(http.HandlerFunc(txHandler.ListByAccount), "TransactionHandler.ListByAccount"))

	// ---- Account Endpoints ----
	mux.Handle("GET /accounts", otelhttp.NewHandler(http.HandlerFunc(accountHandler.List), "AccountHandler.ListAccounts"))
	mux.Handle("GET /accounts/{id}", otelhttp.NewHandler(http.HandlerFunc(accountHandler.Get), "AccountHandler.GetAccountByID"))
	mux.Handle("POST /accounts", otelhttp.NewHandler(http.HandlerFunc(accountHandler.Create), "AccountHandler.CreateAccount"))
	mux.Handle("PATCH /accounts/{id}", otelhttp.NewHandler(http.HandlerFunc(accountHandler.UpdateStatus), "AccountHandler.UpdateStatus"))
	mux.Handle("DELETE /accounts/{id}", otelhttp.NewHandler(http.HandlerFunc(accountHandler.Delete), "AccountHandler.DeleteAccount"))

	port := ":8080"
	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	// ---- Graceful Shutdown Handling ----
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// ---- Start Server in Goroutine ----
	go func() {
		logging.Logger().Infow("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger().Fatalw("server error", "error", err)
		}
	}()

	<-ctx.Done()
	logging.Logger().Infow("shutdown signal received")

	// ---- Graceful HTTP Shutdown ----
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logging.Logger().Errorw("http shutdown error", "error", err)
	} else {
		logging.Logger().Infow("http server stopped gracefully")
	}

	repository.RunMigrateDown(cfg)

	logging.Logger().Infow("application exited")
}
