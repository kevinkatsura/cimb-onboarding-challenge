package main

import (
	"context"
	"core-banking/config"
	"core-banking/internal/database/seeder"
	"core-banking/pkg/database"
	"core-banking/pkg/logging"
)

func main() {
	logging.InitLogger()

	cfg := config.LoadConfig()
	db := database.NewPostgres(cfg)
	defer db.Close()

	s := seeder.New(db)
	if err := s.Seed(context.Background()); err != nil {
		logging.Logger().Warnw("seed bypassed due to collisions", "error", err)
	} else {
		logging.Logger().Infow("seeding completed")
	}
}
