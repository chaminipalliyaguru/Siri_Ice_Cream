package database

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"siri_ice_cream/internal/models"
)

func TestConnectCreatesTables(t *testing.T) {
	db, err := Connect(":memory:")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	for _, table := range []string{"users", "categories", "items", "customers"} {
		if !db.Migrator().HasTable(table) {
			t.Errorf("table %q was not created", table)
		}
	}
}

func TestSeedAddsStartingData(t *testing.T) {
	db, err := Connect(":memory:")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	assertCount(t, db.Model(&models.User{}), 3, "users")
	assertCount(t, db.Model(&models.Category{}), 3, "categories")
	assertCount(t, db.Model(&models.Item{}), 4, "items")

	// Each role has exactly one seeded user.
	for _, role := range []models.Role{models.RoleOwner, models.RoleManager, models.RoleCashier} {
		assertCount(t, db.Model(&models.User{}).Where("role = ?", role), 1, string(role)+" users")
	}
}

func TestSeedTwiceDoesNotDuplicate(t *testing.T) {
	db, err := Connect(":memory:")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	assertCount(t, db.Model(&models.User{}), 3, "users")
	assertCount(t, db.Model(&models.Item{}), 4, "items")
}

func TestSeedStoresHashedPasswords(t *testing.T) {
	db, err := Connect(":memory:")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var owner models.User
	if err := db.Where("username = ?", "owner").First(&owner).Error; err != nil {
		t.Fatalf("find owner: %v", err)
	}

	if owner.PasswordHash == "owner123" {
		t.Fatal("password is stored as plain text")
	}
	if bcrypt.CompareHashAndPassword([]byte(owner.PasswordHash), []byte("owner123")) != nil {
		t.Error("stored hash does not match the seeded password")
	}
}

// assertCount checks how many rows a query matches.
func assertCount(t *testing.T, query *gorm.DB, want int64, what string) {
	t.Helper()
	var got int64
	if err := query.Count(&got).Error; err != nil {
		t.Fatalf("count %s: %v", what, err)
	}
	if got != want {
		t.Errorf("got %d %s, want %d", got, what, want)
	}
}
