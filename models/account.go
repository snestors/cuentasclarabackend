package models

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	UserID      uint   `json:"user_id" gorm:"not null"`
	Name        string `json:"name" gorm:"not null"`
	Type        string `json:"type" gorm:"not null"`
	Currency    string `json:"currency" gorm:"not null"`
	Description string `json:"description" gorm:"size:255"`

	// Personalización visual
	Color    string `json:"color" gorm:"default:'#2196F3'"`
	Icon     string `json:"icon" gorm:"default:'account_balance_wallet'"`
	IsActive bool   `json:"is_active" gorm:"default:true"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User User `json:"-" gorm:"foreignKey:UserID"` // Excluir del JSON
}

// Método para calcular balance dinámicamente
func (a *Account) GetBalance(db *gorm.DB) float64 {
	var balance float64
	db.Table("transactions").
		Where("account_id = ? AND user_id = ? AND deleted_at IS NULL", a.ID, a.UserID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&balance)
	return balance
}
