package main

import (
	"core-banking/config"
	"core-banking/internal/repository"
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
		repository.EnsureDatabase(cfg)
		repository.RunMigrateUp(cfg)
	case "down":
		repository.RunMigrateDown(cfg)
	default:
		fmt.Printf("Unknown command: %s\nUsage: core-banking-migrate [up|down]\n", os.Args[1])
		os.Exit(1)
	}
}
