package main

import (
	"database-exercise/config"
	"database-exercise/database"
	"database-exercise/internal/handler"
	"database-exercise/internal/repository"
	"database-exercise/internal/service"

	"github.com/gin-gonic/gin"
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

	app := gin.Default()
	app.POST("/users", userHandler.CreateUser)
	app.GET("/users", userHandler.ListUsers)
	app.GET("/users/:id", userHandler.GetUserByID)
	app.DELETE("/users/:id", userHandler.DeleteUser)

	app.Run(":8081")
}
