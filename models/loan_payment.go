package models

import (
	"cuentas-claras/utils"
	"time"

	"gorm.io/gorm"
)

type LoanPayment struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	LoanID    uint `json:"loan_id" gorm:"not null"`
	UserID    uint `json:"user_id" gorm:"not null"`
	AccountID uint `json:"account_id" gorm:"not null"` // Cuenta donde entra/sale el dinero

	// Información del pago
	Amount      float64   `json:"amount" gorm:"not null"`
	Date        time.Time `json:"date" gorm:"not null"`
	Description string    `json:"description" gorm:"not null;column:description_encrypted"` // 🔒 ENCRIPTADO
	Notes       string    `json:"notes" gorm:"size:500;column:notes_encrypted"`             // 🔒 ENCRIPTADO

	// Control de confirmación
	TransactionID *uint `json:"transaction_id,omitempty"` // NULL = pendiente, ID = confirmado

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User        User         `json:"-" gorm:"foreignKey:UserID"`
	Loan        Loan         `json:"loan,omitempty" gorm:"foreignKey:LoanID"`
	Account     Account      `json:"account,omitempty" gorm:"foreignKey:AccountID"`
	Transaction *Transaction `json:"transaction,omitempty" gorm:"foreignKey:TransactionID"`
}

// Hook ANTES de guardar - encriptar
func (lp *LoanPayment) BeforeSave(tx *gorm.DB) error {
	if lp.Description != "" {
		lp.Description = utils.EncryptField(lp.Description)
	}
	if lp.Notes != "" {
		lp.Notes = utils.EncryptField(lp.Notes)
	}
	return nil
}

// Hook DESPUÉS de encontrar - desencriptar
func (lp *LoanPayment) AfterFind(tx *gorm.DB) error {
	if lp.Description != "" {
		lp.Description = utils.DecryptField(lp.Description)
	}
	if lp.Notes != "" {
		lp.Notes = utils.DecryptField(lp.Notes)
	}
	return nil
}

// Verificar si está confirmado
func (lp *LoanPayment) IsConfirmed() bool {
	return lp.TransactionID != nil
}
