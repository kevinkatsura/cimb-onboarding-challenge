package main

import (
	"database-exercise/config"
	"database-exercise/database"
	"database-exercise/internal/handler"
	"database-exercise/internal/repository"
	"database-exercise/internal/server"
	"database-exercise/internal/service"
	"log"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()
	db, err := database.NewGormConnection(cfg.DBSource)
	if err != nil {
		panic(err)
	}
	userRepo := repository.NewUserGormRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	router := server.NewRouter(userHandler)
	srv := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	log.Println("Server is running on port 8081")
	log.Fatal(srv.ListenAndServe())
}
