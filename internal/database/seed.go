package database

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"siri_ice_cream/internal/models"
)

// Seed inserts the starting data (users, categories and a few items).
func Seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedUsers(tx); err != nil {
			return err
		}
		return seedCategoriesAndItems(tx)
	})
}

func seedUsers(tx *gorm.DB) error {
	users := []struct {
		name, username, password string
		role                     models.Role
	}{
		{"Shop Owner", "owner", "owner123", models.RoleOwner},
		{"Shop Manager", "manager", "manager123", models.RoleManager},
		{"Shop Cashier", "cashier", "cashier123", models.RoleCashier},
	}

	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		user := models.User{Name: u.name, Username: u.username, PasswordHash: string(hash), Role: u.role}
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("seed user %s: %w", u.username, err)
		}
	}
	return nil
}

func seedCategoriesAndItems(tx *gorm.DB) error {
	chocolate := models.Category{Name: "Chocolate"}
	vanilla := models.Category{Name: "Vanilla"}
	strawberry := models.Category{Name: "Strawberry"}
	for _, c := range []*models.Category{&chocolate, &vanilla, &strawberry} {
		if err := tx.Create(c).Error; err != nil {
			return fmt.Errorf("seed category: %w", err)
		}
	}

	items := []models.Item{
		{Name: "Chocolate Cone", CategoryID: chocolate.ID, Price: 250, Quantity: 40},
		{Name: "Double Chocolate Cup", CategoryID: chocolate.ID, Price: 350, Quantity: 25},
		{Name: "Vanilla Cone", CategoryID: vanilla.ID, Price: 200, Quantity: 50},
		{Name: "Strawberry Sundae", CategoryID: strawberry.ID, Price: 450, Quantity: 15},
	}
	if err := tx.Create(&items).Error; err != nil {
		return fmt.Errorf("seed items: %w", err)
	}
	return nil
}
