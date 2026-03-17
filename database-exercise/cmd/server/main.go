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
	port := ":8082"
	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}
	log.Println("Server is running on port " + port)
	log.Fatal(srv.ListenAndServe())
}
