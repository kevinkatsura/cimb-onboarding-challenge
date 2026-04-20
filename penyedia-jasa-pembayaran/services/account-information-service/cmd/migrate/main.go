package main

import (
	"fmt"
	"os"

	"github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/internal/config"
	"github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/pkg/database"
	"github.com/katsuke/cimb-onboarding-challenge/penyedia-jasa-pembayaran/account-information-service/pkg/logging"
)

func main() {
	logging.InitLogger()
	if len(os.Args) < 2 {
		fmt.Println("Usage: ais-migrate [up|down]")
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
