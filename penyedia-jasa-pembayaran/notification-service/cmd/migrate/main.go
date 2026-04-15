package main

import (
	"notification-service/config"
	"notification-service/pkg/database"
	"notification-service/pkg/logging"
	"fmt"
	"os"
)

func main() {
	logging.InitLogger()
	if len(os.Args) < 2 { fmt.Println("Usage: ns-migrate [up|down]"); os.Exit(1) }
	cfg := config.LoadConfig()
	switch os.Args[1] {
	case "up": database.EnsureDatabase(cfg); database.RunMigrateUp(cfg)
	case "down": database.RunMigrateDown(cfg)
	default: fmt.Printf("Unknown: %s\n", os.Args[1]); os.Exit(1)
	}
}
