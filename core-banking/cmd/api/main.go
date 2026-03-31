package main

import (
	"context"
	"core-banking/internal/config"
	"core-banking/internal/database"
	"core-banking/internal/database/seeder"
	"core-banking/internal/modules/account"
	"core-banking/internal/modules/transaction"
	"core-banking/internal/pkg/logging"
	"core-banking/internal/service"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger, _, err := logging.InitLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	cfg := config.LoadConfig()
	redisCfg := config.LoadRedisConfig()

	// ---- DB Bootstrap ----
	database.EnsureDatabase(cfg)
	db := database.NewPostgres(cfg)

	// ---- Redis Connection ----
	redisClient := database.NewRedis(redisCfg)
	defer database.CloseRedis(redisClient)

	// ---- Migrations ----
	database.RunMigrateUp(cfg)

	// ---- Seeder ----
	s := seeder.New(db)
	if err := s.Seed(context.Background()); err != nil {
		logging.Logger().Fatalw("seed failed", "error", err)
	}
	logging.Logger().Infow("seeding completed")

	lock := service.NewAccountLockManager()

	// Transaction
	txRepo := transaction.NewRepository(db)
	txService := transaction.NewService(txRepo, lock)
	txHandler := transaction.NewHandler(txService)

	// Account
	accountRepo := account.NewRepository(db)
	accountService := account.NewService(accountRepo, &account.RandomAccountNumberGenerator{})
	accountService.SetRedisClient(redisClient) // Inject Redis client
	accountHandler := account.NewHandler(accountService)

	// ---- HTTP Handler ----
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/transfer", txHandler.Transfer)
	mux.HandleFunc("POST /v2/transfer", txHandler.TransferWithLock)
	mux.HandleFunc("GET /transactions", txHandler.ListAll)
	mux.HandleFunc("GET /accounts/{id}/transactions", txHandler.ListByAccount)

	mux.HandleFunc("GET /accounts", accountHandler.List)
	mux.HandleFunc("GET /accounts/{id}", accountHandler.Get)
	mux.HandleFunc("POST /accounts", accountHandler.Create)
	mux.HandleFunc("PATCH /accounts/{id}", accountHandler.UpdateStatus)
	mux.HandleFunc("DELETE /accounts/{id}", accountHandler.Delete)

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

	database.RunMigrateDown(cfg)

	logging.Logger().Infow("application exited")
}
