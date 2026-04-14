package main

// @title Core Banking API
// @version 1.0
// @description API documentation for the Core Banking Application.
// @BasePath /

import (
	"context"
	"core-banking/config"
	account "core-banking/internal/account"
	"core-banking/internal/server"
	"core-banking/internal/transaction"
	"core-banking/pkg/database"
	"core-banking/pkg/idgen"
	"core-banking/pkg/lock"
	"core-banking/pkg/logging"
	"core-banking/pkg/messaging"
	"core-banking/pkg/telemetry"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	db := database.NewPostgres(cfg)
	defer db.Close()

	// ---- Redis Connection ----
	redisClient := database.NewRedis(redisCfg)
	defer database.CloseRedis(redisClient)

	// ---- Messaging ----
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	producer := messaging.NewKafkaProducer(strings.Split(kafkaBrokers, ","), logger)
	defer producer.Close()

	// ---- Dependency Injection ----
	lockManager := lock.NewAccountLockManager()

	txRepo := transaction.NewRepository(db)
	auditSvc := transaction.NewAuditService(txRepo)
	txService := transaction.NewService(txRepo, lockManager, auditSvc, producer)
	txH := transaction.NewHandler(txService)

	accountRepo := account.NewRepository(db)
	accountService := account.NewService(accountRepo, &idgen.RandomAccountNumberGenerator{}, producer)
	accountService.SetRedisClient(redisClient)
	accountH := account.NewHandler(accountService)

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
