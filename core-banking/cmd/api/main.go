package main

// @title Core Banking API
// @version 1.0
// @description API documentation for the Core Banking Application.
// @BasePath /

import (
	"context"
	"core-banking/config"
	accountHandler "core-banking/internal/handler/account"
	txHandler "core-banking/internal/handler/transaction"
	"core-banking/internal/repository"
	"core-banking/internal/server"
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

	_ "core-banking/docs"
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

	// ---- DB Connection ----
	db := repository.NewPostgres(cfg)
	defer db.Close()

	// ---- Redis Connection ----
	redisClient := repository.NewRedis(redisCfg)
	defer repository.CloseRedis(redisClient)

	// ---- Dependency Injection ----
	lock := svc.NewAccountLockManager()

	txRepo := repository.NewTransactionRepository(db)
	txService := transactionSvc.NewService(txRepo, lock)
	txH := txHandler.NewHandler(txService)

	accountRepo := repository.NewAccountRepository(db)
	accountService := accountSvc.NewService(accountRepo, &idgen.RandomAccountNumberGenerator{})
	accountService.SetRedisClient(redisClient)
	accountH := accountHandler.NewHandler(accountService)

	// ---- HTTP Server ----
	handler := server.NewRouter(accountH, txH)
	port := ":8080"
	srv := &http.Server{
		Addr:    port,
		Handler: handler,
	}

	// ---- Graceful Shutdown ----
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		logging.Logger().Infow("secure server starting", "port", port)
		if err := srv.ListenAndServeTLS("certs/server.crt", "certs/server.key"); err != nil && err != http.ErrServerClosed {
			logging.Logger().Fatalw("secure server error", "error", err)
		}
	}()

	<-ctx.Done()
	logging.Logger().Infow("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logging.Logger().Errorw("http shutdown error", "error", err)
	} else {
		logging.Logger().Infow("http server stopped gracefully")
	}

	logging.Logger().Infow("application exited")
}
