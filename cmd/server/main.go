// Command server starts the Ice Cream Shop API.
package main

import (
	"log"

	"siri_ice_cream/internal/config"
	"siri_ice_cream/internal/database"
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

	log.Println("Database ready and seeded!")
}
