package models

import (
	"time"

	"gorm.io/gorm"
)

type Reminder struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	UserID      uint   `json:"user_id" gorm:"not null"`
	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description" gorm:"size:255"`
	Type        string `json:"type" gorm:"not null"` // "recurring_expense", "loan", "custom"

	// Referencia al objeto relacionado
	ReferenceID   *uint  `json:"reference_id,omitempty"`
	ReferenceType string `json:"reference_type,omitempty"` // "recurring_expense", "loan"

	// Configuración de recordatorio
	RemindAt time.Time `json:"remind_at" gorm:"not null"`
	IsActive bool      `json:"is_active" gorm:"default:true"`
	IsSent   bool      `json:"is_sent" gorm:"default:false"`

	// Información adicional
	Priority string `json:"priority" gorm:"default:'normal'"` // "low", "normal", "high"

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// Verificar si el recordatorio debe enviarse
func (r *Reminder) ShouldBeSent() bool {
	return r.IsActive && !r.IsSent && time.Now().After(r.RemindAt)
}

// Marcar como enviado
func (r *Reminder) MarkAsSent(db *gorm.DB) error {
	r.IsSent = true
	return db.Save(r).Error
}
