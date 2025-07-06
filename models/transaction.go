package models

import (
	"cuentas-claras/utils"
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	UserID    uint `json:"user_id" gorm:"not null"`
	AccountID uint `json:"account_id" gorm:"not null"`

	// Información financiera
	Amount      float64   `json:"amount" gorm:"not null"`
	Direction   string    `json:"direction" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	Date        time.Time `json:"date" gorm:"not null"`
	Notes       string    `json:"notes" gorm:"size:500;column:notes_encrypted"`

	// Clasificación
	Type       string `json:"type" gorm:"not null"`
	CategoryID *uint  `json:"category_id,omitempty"`

	// Referencias a otros objetos
	ReferenceID   *uint  `json:"reference_id,omitempty"`
	ReferenceType string `json:"reference_type,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User     User      `json:"-" gorm:"foreignKey:UserID"`
	Account  Account   `json:"account,omitempty" gorm:"foreignKey:AccountID"`
	Category *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
}

// Hook ANTES de guardar - encriptar notes
func (t *Transaction) BeforeSave(tx *gorm.DB) error {
	if t.Notes != "" {
		t.Notes = utils.EncryptField(t.Notes)
	}
	return nil
}

// Hook DESPUÉS de encontrar - desencriptar notes
func (t *Transaction) AfterFind(tx *gorm.DB) error {
	if t.Notes != "" {
		t.Notes = utils.DecryptField(t.Notes)
	}
	return nil
}
