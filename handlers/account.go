package handlers

import (
	"cuentas-claras/config"
	"cuentas-claras/models"

	"github.com/gofiber/fiber/v2"
)

type CreateAccountRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Type        string `json:"type" validate:"required,oneof=cash bank credit_card savings"`
	Currency    string `json:"currency" validate:"required,len=3"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

type UpdateAccountRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

func CreateAccount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req CreateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	// Verificar que no existe una cuenta con el mismo nombre para el usuario
	var existingAccount models.Account
	if err := config.DB.Where("user_id = ? AND name = ? AND deleted_at IS NULL", userID, req.Name).First(&existingAccount).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Account name already exists"})
	}

	account := models.Account{
		UserID:      userID,
		Name:        req.Name,
		Type:        req.Type,
		Currency:    req.Currency,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	}

	// Asignar valores por defecto si no se proporcionan
	if account.Color == "" {
		account.Color = "#2196F3"
	}
	if account.Icon == "" {
		account.Icon = "account_balance_wallet"
	}

	if err := config.DB.Create(&account).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create account"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Account created successfully",
		"account": account,
	})
}

func GetAccounts(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var accounts []models.Account
	if err := config.DB.Where("user_id = ? AND is_active = true", userID).Find(&accounts).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch accounts"})
	}

	// Calcular balance para cada cuenta
	accountsWithBalance := make([]fiber.Map, len(accounts))
	for i, account := range accounts {
		balance := account.GetBalance(config.DB)
		accountsWithBalance[i] = fiber.Map{
			"id":          account.ID,
			"name":        account.Name,
			"type":        account.Type,
			"currency":    account.Currency,
			"description": account.Description,
			"color":       account.Color,
			"icon":        account.Icon,
			"is_active":   account.IsActive,
			"balance":     balance,
			"created_at":  account.CreatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"accounts": accountsWithBalance,
	})
}

func GetAccount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	accountID := c.Params("id")

	var account models.Account
	if err := config.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&account).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	balance := account.GetBalance(config.DB)

	return c.JSON(fiber.Map{
		"account": fiber.Map{
			"id":          account.ID,
			"name":        account.Name,
			"type":        account.Type,
			"currency":    account.Currency,
			"description": account.Description,
			"color":       account.Color,
			"icon":        account.Icon,
			"is_active":   account.IsActive,
			"balance":     balance,
			"created_at":  account.CreatedAt,
			"updated_at":  account.UpdatedAt,
		},
	})
}

func UpdateAccount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	accountID := c.Params("id")

	var req UpdateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	var account models.Account
	if err := config.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&account).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	// Verificar nombre único si se está cambiando
	if req.Name != "" && req.Name != account.Name {
		var existingAccount models.Account
		if err := config.DB.Where("user_id = ? AND name = ? AND id != ? AND deleted_at IS NULL", userID, req.Name, accountID).First(&existingAccount).Error; err == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Account name already exists"})
		}
		account.Name = req.Name
	}

	// Actualizar campos
	if req.Description != "" {
		account.Description = req.Description
	}
	if req.Color != "" {
		account.Color = req.Color
	}
	if req.Icon != "" {
		account.Icon = req.Icon
	}
	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}

	if err := config.DB.Save(&account).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update account"})
	}

	return c.JSON(fiber.Map{
		"message": "Account updated successfully",
		"account": account,
	})
}

func DeleteAccount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	accountID := c.Params("id")

	var account models.Account
	if err := config.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&account).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	// Soft delete
	if err := config.DB.Delete(&account).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not delete account"})
	}

	return c.JSON(fiber.Map{
		"message": "Account deleted successfully",
	})
}
