package main

import (
	"context"
	"core-banking/internal/config"
	"core-banking/internal/database"
	"core-banking/internal/database/seeder"
	"core-banking/internal/modules/account"
	"core-banking/internal/modules/transaction"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	// ---- DB Bootstrap ----
	database.EnsureDatabase(cfg)
	db := database.NewPostgres(cfg)

	// ---- Migrations ----
	database.RunMigrateUp(cfg)

	// ---- Seeder ----
	s := seeder.New(db)
	if err := s.Seed(context.Background()); err != nil {
		log.Fatal(err)
	}
	log.Println("Seeding completed")

	// Transaction
	txRepo := transaction.NewRepository(db)
	txService := transaction.NewService(txRepo)
	txHandler := transaction.NewHandler(txService)

	// Account
	accountRepo := account.NewRepository(db)
	accountService := account.NewService(accountRepo)
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

	port := ":8120"
	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	log.Println("Server is running on port " + port)
	log.Fatal(srv.ListenAndServe())

	// ---- Graceful Shutdown Handling ----
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()
	log.Println("Shutdown signal received")

	// ---- Graceful HTTP Shutdown ----
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("HTTP shutdown error:", err)
	} else {
		log.Println("HTTP server stopped gracefully")
	}

	database.RunMigrateDown(cfg)

	log.Println("Application exited")
}
