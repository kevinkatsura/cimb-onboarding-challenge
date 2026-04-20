package main

import (
	"notification-service/config"
	"notification-service/pkg/database"
	"notification-service/pkg/logging"
	"os"
)

func main() {
	logging.InitLogger()
	cfg := config.LoadConfig()
	database.EnsureSchema(cfg)

	if len(os.Args) < 2 {
		logging.Logger().Fatalw("usage: ns-migrate [up|down]")
	}
	switch os.Args[1] {
	case "up":
		database.RunMigrateUp(cfg)
	case "down":
		database.RunMigrateDown(cfg)
	default:
		logging.Logger().Fatalw("unknown command", "cmd", os.Args[1])
	}
}
