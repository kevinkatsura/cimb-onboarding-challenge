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

	http.HandleFunc("POST /accounts", accountHandler.Create)
	http.HandleFunc("GET /accounts", accountHandler.ListAll)
	http.HandleFunc("GET /accounts/{id}", accountHandler.Get)
	http.HandleFunc("PATCH /accounts/{id}", accountHandler.UpdateStatus)
	http.HandleFunc("DELETE /accounts/{id}", accountHandler.Delete)

	http.HandleFunc("GET /customers/{id}/accounts", accountHandler.List)

	port := os.Getenv("PORT")
	log.Println("server running on " + port)
	http.ListenAndServe(port, nil)
}
