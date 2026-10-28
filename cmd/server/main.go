// Command server starts the Ice Cream Shop API.
package main

import (
	"log/slog"
	"os"

	"siri_ice_cream/internal/config"
	"siri_ice_cream/internal/database"
	"siri_ice_cream/internal/logging"
	"siri_ice_cream/internal/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("could not load config", "error", err)
		os.Exit(1)
	}

	logger := logging.New(cfg.IsProduction())
	slog.SetDefault(logger)

	db, err := database.Connect(cfg.DBPath)
	if err != nil {
		logger.Error("could not connect to database", "error", err)
		os.Exit(1)
	}

	if err := database.Seed(db); err != nil {
		logger.Error("could not seed database", "error", err)
		os.Exit(1)
	}

	app := routes.NewApp(db, cfg.JWTSecret, logger)

	logger.Info("server starting",
		"env", cfg.AppEnv,
		"port", cfg.Port,
		"db_path", cfg.DBPath,
	)
	if err := app.Listen(":" + cfg.Port); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
