package server

import (
	"database-exercise/internal/handler"
	"net/http"
)

func NewAPIPath(method, path string) string {
	return method + " " + path
}

func NewRouter(userHandler *handler.UserHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(NewAPIPath(http.MethodGet, "/users"), userHandler.ListUsers)
	mux.HandleFunc(NewAPIPath(http.MethodGet, "/users/{id}"), userHandler.GetUserByID)
	mux.HandleFunc(NewAPIPath(http.MethodPost, "/users"), userHandler.CreateUser)
	mux.HandleFunc(NewAPIPath(http.MethodDelete, "/users/{id}"), userHandler.DeleteUser)

	return Chain(
		mux,
		JSONMiddleware,
		RecoveryMiddleware,
		LoggingMiddleware,
	)
}
