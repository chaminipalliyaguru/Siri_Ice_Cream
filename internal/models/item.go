package models

import "time"

// Item is a product in the shop's inventory.
type Item struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"not null;index"`
	CategoryID uint      `json:"category_id" gorm:"not null"`
	Category   *Category `json:"category,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Price      float64   `json:"price" gorm:"not null"`
	Quantity   int       `json:"quantity" gorm:"not null;default:0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
