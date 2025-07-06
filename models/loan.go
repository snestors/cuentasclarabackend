package models

import (
	"cuentas-claras/utils"
	"time"

	"gorm.io/gorm"
)

type Loan struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	UserID    uint `json:"user_id" gorm:"not null"`
	AccountID uint `json:"account_id" gorm:"not null"` // Cuenta inicial del préstamo

	// Información básica
	Amount      float64 `json:"amount" gorm:"not null"`
	Description string  `json:"description" gorm:"not null;column:description_encrypted"` // 🔒 ENCRIPTADO
	PersonName  string  `json:"person_name" gorm:"column:person_name_encrypted"`          // 🔒 ENCRIPTADO
	Type        string  `json:"type" gorm:"not null"`                                     // 'given' o 'received'

	// Control y fechas
	Status       string     `json:"status" gorm:"default:'pending'"` // pending, partial_paid, paid
	LoanDate     time.Time  `json:"loan_date" gorm:"not null"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	InterestRate float64    `json:"interest_rate" gorm:"default:0"`
	Notes        string     `json:"notes" gorm:"size:500;column:notes_encrypted"` // 🔒 ENCRIPTADO

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User     User          `json:"-" gorm:"foreignKey:UserID"`
	Account  Account       `json:"account,omitempty" gorm:"foreignKey:AccountID"`
	Payments []LoanPayment `json:"payments,omitempty" gorm:"foreignKey:LoanID"`
}

// Hook ANTES de guardar - encriptar
func (l *Loan) BeforeSave(tx *gorm.DB) error {
	if l.Description != "" {
		l.Description = utils.EncryptField(l.Description)
	}
	if l.PersonName != "" {
		l.PersonName = utils.EncryptField(l.PersonName)
	}
	if l.Notes != "" {
		l.Notes = utils.EncryptField(l.Notes)
	}
	return nil
}

// Hook DESPUÉS de encontrar - desencriptar
func (l *Loan) AfterFind(tx *gorm.DB) error {
	if l.Description != "" {
		l.Description = utils.DecryptField(l.Description)
	}
	if l.PersonName != "" {
		l.PersonName = utils.DecryptField(l.PersonName)
	}
	if l.Notes != "" {
		l.Notes = utils.DecryptField(l.Notes)
	}
	return nil
}

// Calcular total pagado
func (l *Loan) GetTotalPaid(db *gorm.DB) float64 {
	var totalPaid float64
	db.Table("loan_payments").
		Where("loan_id = ? AND transaction_id IS NOT NULL AND deleted_at IS NULL", l.ID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalPaid)
	return totalPaid
}

// Calcular balance pendiente
func (l *Loan) GetBalance(db *gorm.DB) float64 {
	totalPaid := l.GetTotalPaid(db)
	return l.Amount - totalPaid
}

// Actualizar status automáticamente SIN tocar campos encriptados
func (l *Loan) UpdateStatus(db *gorm.DB) {
	totalPaid := l.GetTotalPaid(db)

	var newStatus string
	if totalPaid == 0 {
		newStatus = "pending"
	} else if totalPaid < l.Amount {
		newStatus = "partial_paid"
	} else {
		newStatus = "paid"
	}

	// Solo actualizar el status SIN pasar por hooks
	if l.Status != newStatus {
		db.Model(l).Update("status", newStatus)
		l.Status = newStatus // Actualizar el campo local también
	}
}
