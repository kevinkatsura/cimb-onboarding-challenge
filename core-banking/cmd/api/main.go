package main

import (
	"core-banking/internal/database"
	"core-banking/internal/modules/transaction"
	"log"
	"net/http"
	"os"
)

func main() {
	db := database.NewDB(os.Getenv("DB_URL"))

	repo := transaction.NewRepository(db)
	service := transaction.NewService(repo)
	handler := transaction.NewHandler(service)

	http.HandleFunc("/transfer", handler.Transfer)

	port := os.Getenv("PORT")
	log.Println("server running on " + port)
	http.ListenAndServe(port, nil)
}
