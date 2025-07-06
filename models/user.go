package models

import (
	"cuentas-claras/utils"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"not null;column:name_encrypted"`
	Email       string `json:"email" gorm:"unique;not null"`
	Password    string `json:"-" gorm:"not null"`
	PhoneNumber string `json:"phone_number" gorm:"column:phone_encrypted"`

	// Configuraciones
	NotificationsEnabled bool       `json:"notifications_enabled" gorm:"default:true"`
	PushNotifications    bool       `json:"push_notifications" gorm:"default:true"`
	InAppNotifications   bool       `json:"in_app_notifications" gorm:"default:true"`
	Timezone             string     `json:"timezone" gorm:"default:'America/Lima'"`
	QuietHoursStart      *time.Time `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd        *time.Time `json:"quiet_hours_end,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Hook ANTES de guardar - encriptar
func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Name != "" {
		u.Name = utils.EncryptField(u.Name)
	}
	if u.PhoneNumber != "" {
		u.PhoneNumber = utils.EncryptField(u.PhoneNumber)
	}
	return nil
}

// Hook DESPUÉS de encontrar - desencriptar
func (u *User) AfterFind(tx *gorm.DB) error {
	if u.Name != "" {
		u.Name = utils.DecryptField(u.Name)
	}
	if u.PhoneNumber != "" {
		u.PhoneNumber = utils.DecryptField(u.PhoneNumber)
	}
	return nil
}
