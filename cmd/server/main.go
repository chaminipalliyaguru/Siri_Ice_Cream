// Command server starts the Ice Cream Shop API.
package main

import (
	"log"

	"siri_ice_cream/internal/config"
	"siri_ice_cream/internal/database"
	"siri_ice_cream/internal/routes"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	if err := database.Seed(db); err != nil {
		log.Fatalf("could not seed database: %v", err)
	}

	app := routes.NewApp(db, cfg.JWTSecret)

	log.Printf("API running on http://localhost:%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
