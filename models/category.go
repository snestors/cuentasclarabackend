package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	UserID      uint   `json:"user_id" gorm:"not null"`
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description" gorm:"size:255"`

	// Personalización visual
	Color    string `json:"color" gorm:"default:'#4CAF50'"`
	Icon     string `json:"icon" gorm:"default:'category'"`
	IsActive bool   `json:"is_active" gorm:"default:true"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User User `json:"-" gorm:"foreignKey:UserID"`
}
