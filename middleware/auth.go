package middleware

import (
	"cuentas-claras/config"
	"cuentas-claras/models"
	"cuentas-claras/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(c *fiber.Ctx) error {
	// Obtener header Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "Authorization header required",
		})
	}

	// Verificar formato "Bearer token"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid authorization format. Use: Bearer <token>",
		})
	}

	tokenString := parts[1]

	// Validar JWT token
	token, err := utils.ValidateAccessToken(tokenString)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// Verificar que el token sea válido y extraer claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Verificar que sea un access token
		if tokenType, exists := claims["type"]; !exists || tokenType != "access" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid token type",
			})
		}

		// Extraer user_id y refresh_token_id
		userIDFloat, exists := claims["user_id"]
		if !exists {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		refreshTokenID, exists := claims["refresh_token_id"]
		if !exists {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		userID := uint(userIDFloat.(float64))

		// Verificar que el refresh_token_id existe y está activo
		var deviceSession models.DeviceSession
		if err := config.DB.Where("refresh_token_id = ? AND is_active = true", refreshTokenID).First(&deviceSession).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Token has been revoked",
			})
		}

		// Verificar que pertenece al usuario correcto
		if deviceSession.UserID != userID {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Verificar que el usuario existe
		var user models.User
		if err := config.DB.First(&user, userID).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		// Actualizar última actividad
		deviceSession.LastActivity = time.Now()
		config.DB.Save(&deviceSession)

		// Guardar información del usuario en el contexto
		c.Locals("user_id", userID)
		c.Locals("user", user)
		c.Locals("device_session_id", deviceSession.ID)

		return c.Next()
	}

	return c.Status(401).JSON(fiber.Map{
		"error": "Invalid token",
	})
}

// Middleware opcional - no requiere autenticación pero extrae info si existe
func OptionalAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Next() // Continuar sin autenticación
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Next() // Continuar sin autenticación
	}

	tokenString := parts[1]
	token, err := utils.ValidateAccessToken(tokenString)
	if err != nil {
		return c.Next() // Continuar sin autenticación
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userIDFloat, exists := claims["user_id"]; exists {
			userID := uint(userIDFloat.(float64))
			c.Locals("user_id", userID)
		}
	}

	return c.Next()
}
