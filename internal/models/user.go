// Package models contains the database tables (entities) of the application.
package models

import "time"

// Role is the job role of a user. It decides what the user is allowed to do.
type Role string

const (
	RoleOwner   Role = "owner"
	RoleManager Role = "manager"
	RoleCashier Role = "cashier"
)

// User is a staff member who can log in to the system.
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"` // "-" means: never send in JSON
	Role         Role      `json:"role" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
