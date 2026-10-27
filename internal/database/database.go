// Package database opens the SQLite connection and prepares the tables.
package database

import (
	"fmt"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"siri_ice_cream/internal/models"
)

// Connect opens the SQLite database at the given path and creates the tables.
// Use ":memory:" as the path for a temporary in-memory database (used in tests).
func Connect(path string) (*gorm.DB, error) {
	// foreign_keys(1) turns on foreign key checks in SQLite.
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", path)

	db, err := gorm.Open(gormlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// SQLite works best with a single open connection.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	if err := Migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

// Migrate creates or updates the tables to match the models.
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Item{},
		&models.Customer{},
	)
	if err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
}
