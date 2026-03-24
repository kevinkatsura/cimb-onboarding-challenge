package main

import (
	"core-banking/internal/database"
	"core-banking/internal/modules/account"
	"core-banking/internal/modules/transaction"
	"log"
	"net/http"
	"os"
)

func main() {
	db := database.NewDB(os.Getenv("DB_URL"))

	// Transaction
	txRepo := transaction.NewRepository(db)
	txService := transaction.NewService(txRepo)
	txHandler := transaction.NewHandler(txService)

	http.HandleFunc("POST /transfer", txHandler.Transfer)

	// Account
	accountRepo := account.NewRepository(db)
	accountService := account.NewService(accountRepo)
	accountHandler := account.NewHandler(accountService)

	http.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			accountHandler.ListAll(w, r)
		case http.MethodPost:
			accountHandler.Create(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	http.HandleFunc("/accounts/", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("GET /customers/{id}/accounts", accountHandler.List)

	port := os.Getenv("PORT")
	log.Println("server running on " + port)
	http.ListenAndServe(port, nil)
}
