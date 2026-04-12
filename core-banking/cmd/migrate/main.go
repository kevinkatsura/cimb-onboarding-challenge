package main

import (
	"core-banking/config"
	"core-banking/pkg/database"
	"core-banking/pkg/logging"
	"fmt"
	"os"
)

func main() {
	logging.InitLogger()

	if len(os.Args) < 2 {
		fmt.Println("Usage: core-banking-migrate [up|down]")
		os.Exit(1)
	}

	cfg := config.LoadConfig()

	switch os.Args[1] {
	case "up":
		database.EnsureDatabase(cfg)
		database.RunMigrateUp(cfg)
	case "down":
		database.RunMigrateDown(cfg)
	default:
		fmt.Printf("Unknown command: %s\nUsage: core-banking-migrate [up|down]\n", os.Args[1])
		os.Exit(1)
	}
}
