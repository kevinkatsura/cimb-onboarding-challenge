package main

import (
	"core-banking-system/config"
	"core-banking-system/pkg/database"
	"core-banking-system/pkg/logging"
	"fmt"
	"os"
)

func main() {
	logging.InitLogger()
	if len(os.Args) < 2 {
		fmt.Println("Usage: cbs-migrate [up|down]")
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
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
