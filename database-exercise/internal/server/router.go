package server

import (
	"database-exercise/internal/handler"
	"net/http"
)

func NewRouter(userHandler *handler.UserHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", userHandler.ListUsers)
	mux.HandleFunc("GET /users/:id", userHandler.GetUserByID)
	mux.HandleFunc("POST /users", userHandler.CreateUser)
	mux.HandleFunc("DELETE /users/:id", userHandler.DeleteUser)

	return Chain(
		mux,
		JSONMiddleware,
		RecoveryMiddleware,
		LoggingMiddleware,
	)
}
