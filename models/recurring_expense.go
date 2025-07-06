package models

import (
	"time"

	"gorm.io/gorm"
)

type RecurringExpense struct {
	ID         uint `json:"id" gorm:"primaryKey"`
	UserID     uint `json:"user_id" gorm:"not null"`
	AccountID  uint `json:"account_id" gorm:"not null"`
	CategoryID uint `json:"category_id" gorm:"not null"`

	// Configuración del gasto
	Amount      float64 `json:"amount" gorm:"not null"`
	Description string  `json:"description" gorm:"not null"`

	// Configuración de recurrencia
	Frequency   string     `json:"frequency" gorm:"not null"` // daily, weekly, monthly, yearly
	StartDate   time.Time  `json:"start_date" gorm:"not null"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	NextDueDate time.Time  `json:"next_due_date" gorm:"not null"`

	// Control
	IsActive     bool   `json:"is_active" gorm:"default:true"`
	AutoGenerate bool   `json:"auto_generate" gorm:"default:false"` // Solo recordatorio, NO transacción automática
	Notes        string `json:"notes" gorm:"size:500"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User     User     `json:"-" gorm:"foreignKey:UserID"`
	Account  Account  `json:"account,omitempty" gorm:"foreignKey:AccountID"`
	Category Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
}

// Método para calcular próxima fecha de vencimiento
func (re *RecurringExpense) CalculateNextDueDate() time.Time {
	switch re.Frequency {
	case "daily":
		return re.NextDueDate.AddDate(0, 0, 1)
	case "weekly":
		return re.NextDueDate.AddDate(0, 0, 7)
	case "monthly":
		return re.NextDueDate.AddDate(0, 1, 0)
	case "yearly":
		return re.NextDueDate.AddDate(1, 0, 0)
	default:
		return re.NextDueDate
	}
}

// Método para verificar si está vencido
func (re *RecurringExpense) IsOverdue() bool {
	return time.Now().After(re.NextDueDate) && re.IsActive
}

// Método para verificar si vence hoy
func (re *RecurringExpense) IsDueToday() bool {
	today := time.Now().Format("2006-01-02")
	dueDate := re.NextDueDate.Format("2006-01-02")
	return today == dueDate && re.IsActive
}
