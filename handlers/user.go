package handlers

import (
	"cuentas-claras/config"
	"cuentas-claras/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetProfile(c *fiber.Ctx) error {
	// Obtener user_id del middleware
	userID := c.Locals("user_id").(uint)

	// Buscar usuario con sus sesiones activas
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Buscar sesiones activas
	var activeSessions []models.DeviceSession
	config.DB.Where("user_id = ? AND is_active = true", userID).Find(&activeSessions)

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":                    user.ID,
			"name":                  user.Name,
			"email":                 user.Email,
			"phone_number":          user.PhoneNumber,
			"notifications_enabled": user.NotificationsEnabled,
			"push_notifications":    user.PushNotifications,
			"in_app_notifications":  user.InAppNotifications,
			"timezone":              user.Timezone,
			"quiet_hours_start":     user.QuietHoursStart,
			"quiet_hours_end":       user.QuietHoursEnd,
			"created_at":            user.CreatedAt,
		},
		"active_sessions": len(activeSessions),
		"current_session": c.Locals("device_session_id"),
	})
}

func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req struct {
		Name                 string `json:"name,omitempty"`
		PhoneNumber          string `json:"phone_number,omitempty"`
		NotificationsEnabled *bool  `json:"notifications_enabled,omitempty"`
		PushNotifications    *bool  `json:"push_notifications,omitempty"`
		Timezone             string `json:"timezone,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Buscar usuario
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Actualizar campos si están presentes
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.NotificationsEnabled != nil {
		user.NotificationsEnabled = *req.NotificationsEnabled
	}
	if req.PushNotifications != nil {
		user.PushNotifications = *req.PushNotifications
	}
	if req.Timezone != "" {
		user.Timezone = req.Timezone
	}

	// Guardar cambios
	if err := config.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update profile"})
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
		"user": fiber.Map{
			"id":                    user.ID,
			"name":                  user.Name,
			"email":                 user.Email,
			"phone_number":          user.PhoneNumber,
			"notifications_enabled": user.NotificationsEnabled,
			"push_notifications":    user.PushNotifications,
			"timezone":              user.Timezone,
		},
	})
}

func Logout(c *fiber.Ctx) error {
	deviceSessionID := c.Locals("device_session_id").(uint)

	// Marcar sesión como inactiva
	var deviceSession models.DeviceSession
	if err := config.DB.First(&deviceSession, deviceSessionID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Session not found"})
	}

	deviceSession.IsActive = false
	deviceSession.LogoutReason = "User logout"
	now := time.Now()
	deviceSession.LogoutAt = &now

	config.DB.Save(&deviceSession)

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
