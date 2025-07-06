package handlers

import (
	"cuentas-claras/config"
	"cuentas-claras/models"
	"cuentas-claras/utils"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

type RegisterRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type LoginRequest struct {
	Email      string                 `json:"email" validate:"required,email"`
	Password   string                 `json:"password" validate:"required"`
	DeviceInfo map[string]interface{} `json:"device_info,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func Register(c *fiber.Ctx) error {
	var req RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	// Verificar si el email ya existe
	var existingUser models.User
	if err := config.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Email already exists"})
	}

	// Encriptar password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not hash password"})
	}

	// Crear usuario
	user := models.User{
		Name:        req.Name,
		Email:       req.Email,
		Password:    string(hashedPassword),
		PhoneNumber: req.PhoneNumber,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create user"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "User created successfully",
		"user": fiber.Map{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	// Buscar usuario
	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Verificar password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Marcar todas las sesiones anteriores como inactivas (login único)
	config.DB.Model(&models.DeviceSession{}).Where("user_id = ? AND is_active = true", user.ID).
		Updates(map[string]interface{}{
			"is_active":     false,
			"logout_at":     time.Now(),
			"logout_reason": "New session started",
		})

	// Generar tokens
	refreshTokenID := utils.GenerateRefreshTokenID()
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, refreshTokenID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate access token"})
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate refresh token"})
	}

	refreshTokenHash, err := utils.HashRefreshToken(refreshToken)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not hash refresh token"})
	}

	// Crear nueva sesión de dispositivo
	deviceInfo := req.DeviceInfo
	deviceSession := models.DeviceSession{
		UserID:           user.ID,
		DeviceID:         getStringFromMap(deviceInfo, "device_id", "unknown"),
		DeviceName:       getStringFromMap(deviceInfo, "device_name", "Unknown Device"),
		DeviceType:       getStringFromMap(deviceInfo, "device_type", "unknown"),
		DeviceModel:      getStringFromMap(deviceInfo, "device_model", ""),
		OSVersion:        getStringFromMap(deviceInfo, "os_version", ""),
		IPAddress:        c.IP(),
		UserAgent:        c.Get("User-Agent"),
		RefreshTokenHash: refreshTokenHash,
		RefreshTokenID:   refreshTokenID,
		FCMToken:         getStringFromMap(deviceInfo, "fcm_token", ""),
		IsActive:         true,
		LoginAt:          time.Now(),
		LastActivity:     time.Now(),
	}

	if err := config.DB.Create(&deviceSession).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create device session"})
	}

	return c.JSON(fiber.Map{
		"access_token":      accessToken,
		"refresh_token":     refreshToken,
		"device_session_id": deviceSession.ID,
		"user": fiber.Map{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func RefreshToken(c *fiber.Ctx) error {
	var req RefreshRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	// Buscar TODAS las sesiones activas y verificar el refresh token
	var deviceSessions []models.DeviceSession
	if err := config.DB.Preload("User").Where("is_active = true").Find(&deviceSessions).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	// Verificar refresh token contra todas las sesiones activas
	var validSession *models.DeviceSession
	for i := range deviceSessions {
		if utils.VerifyRefreshToken(req.RefreshToken, deviceSessions[i].RefreshTokenHash) {
			validSession = &deviceSessions[i]
			break
		}
	}

	if validSession == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	// Generar nuevo access token
	newAccessToken, err := utils.GenerateAccessToken(validSession.User.ID, validSession.User.Email, validSession.RefreshTokenID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate access token"})
	}

	// Generar nuevo refresh token (rotación)
	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate refresh token"})
	}

	newRefreshTokenHash, err := utils.HashRefreshToken(newRefreshToken)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not hash refresh token"})
	}

	// Actualizar sesión con nuevo refresh token
	validSession.RefreshTokenHash = newRefreshTokenHash
	validSession.LastActivity = time.Now()
	config.DB.Save(validSession)

	return c.JSON(fiber.Map{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

// Función auxiliar para obtener valores del map de device_info
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}
