package main

import (
	"context"
	"core-banking/internal/config"
	"core-banking/internal/database"
	"core-banking/internal/modules/account"
	"core-banking/internal/modules/transaction"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	wd, _ := os.Getwd()
	log.Println(wd)
	cfg := config.LoadConfig()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode,
	)
	// ---- DB Bootstrap ----
	database.EnsureDatabase(cfg)
	db := database.NewPostgres(cfg)

	// ---- Migrations ----
	database.RunMigrateUp(dsn)

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
	mux.HandleFunc("POST /transfer", txHandler.Transfer)
	mux.HandleFunc("GET /transactions", txHandler.ListAll)
	mux.HandleFunc("GET /accounts/{id}/transactions", txHandler.ListByAccount)

	mux.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			accountHandler.List(w, r)
		case http.MethodPost:
			accountHandler.Create(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/accounts/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			accountHandler.Get(w, r)
		case http.MethodPatch:
			accountHandler.UpdateStatus(w, r)
		case http.MethodDelete:
			accountHandler.Delete(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	port := os.Getenv("PORT")
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

	// ---- Optional: DB Cleanup ----
	// ⚠ only for dev/test
	database.RunMigrateDown(dsn)

	log.Println("Application exited")
}
