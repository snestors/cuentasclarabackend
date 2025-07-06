package models

import (
	"time"

	"gorm.io/gorm"
)

type DeviceSession struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"not null"`

	// Información del dispositivo
	DeviceID    string `json:"device_id" gorm:"not null"`
	DeviceName  string `json:"device_name" gorm:"not null"`
	DeviceType  string `json:"device_type" gorm:"not null"`
	DeviceModel string `json:"device_model"`
	OSVersion   string `json:"os_version"`

	// Información de red
	IPAddress string `json:"ip_address"`
	Country   string `json:"country"`
	City      string `json:"city"`
	UserAgent string `json:"user_agent"`

	// Tokens y control
	RefreshTokenHash string `json:"-" gorm:"not null"`
	RefreshTokenID   string `json:"-" gorm:"unique;not null;index"` // UUID único
	FCMToken         string `json:"-" gorm:"column:fcm_token_encrypted"`
	IsFCMActive      bool   `json:"is_fcm_active" gorm:"default:true"`
	IsActive         bool   `json:"is_active" gorm:"default:true"`

	// Actividad
	LoginAt      time.Time  `json:"login_at"`
	LogoutAt     *time.Time `json:"logout_at,omitempty"`
	LogoutReason string     `json:"logout_reason,omitempty"`
	LastActivity time.Time  `json:"last_activity"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
