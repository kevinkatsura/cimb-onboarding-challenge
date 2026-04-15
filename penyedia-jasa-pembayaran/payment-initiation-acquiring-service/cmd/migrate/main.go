package main

import (
	"payment-initiation-acquiring-service/config"
	"payment-initiation-acquiring-service/pkg/database"
	"payment-initiation-acquiring-service/pkg/logging"
	"fmt"
	"os"
)

func main() {
	logging.InitLogger()
	if len(os.Args) < 2 { fmt.Println("Usage: pias-migrate [up|down]"); os.Exit(1) }
	cfg := config.LoadConfig()
	switch os.Args[1] {
	case "up": database.EnsureDatabase(cfg); database.RunMigrateUp(cfg)
	case "down": database.RunMigrateDown(cfg)
	default: fmt.Printf("Unknown: %s\n", os.Args[1]); os.Exit(1)
	}
}
