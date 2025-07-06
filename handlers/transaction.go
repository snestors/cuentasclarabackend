package handlers

import (
	"cuentas-claras/config"
	"cuentas-claras/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CreateTransactionRequest struct {
	AccountID   uint      `json:"account_id" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,ne=0"`
	Description string    `json:"description" validate:"required,min=1,max=255"`
	Date        time.Time `json:"date" validate:"required"`
	Notes       string    `json:"notes,omitempty"`
	Type        string    `json:"type" validate:"required,oneof=income expense loan_given loan_received loan_payment_given loan_payment_received"`
	CategoryID  *uint     `json:"category_id,omitempty"`
}

type UpdateTransactionRequest struct {
	AccountID   *uint      `json:"account_id,omitempty"`
	Amount      *float64   `json:"amount,omitempty" validate:"omitempty,ne=0"`
	Description string     `json:"description,omitempty" validate:"omitempty,min=1,max=255"`
	Date        *time.Time `json:"date,omitempty"`
	Notes       string     `json:"notes,omitempty"`
	CategoryID  *uint      `json:"category_id,omitempty"`
}

func CreateTransaction(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	// Verificar que la cuenta pertenece al usuario
	var account models.Account
	if err := config.DB.Where("id = ? AND user_id = ?", req.AccountID, userID).First(&account).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	// Verificar que la categoría pertenece al usuario (si se proporciona)
	if req.CategoryID != nil {
		var category models.Category
		if err := config.DB.Where("id = ? AND user_id = ?", *req.CategoryID, userID).First(&category).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
		}
	}

	// Determinar direction basado en el type y amount
	var direction string
	var amount float64

	switch req.Type {
	case "income", "loan_received", "loan_payment_received":
		direction = "in"
		amount = abs(req.Amount) // Asegurar que sea positivo
	case "expense", "loan_given", "loan_payment_given":
		direction = "out"
		amount = -abs(req.Amount) // Asegurar que sea negativo
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Invalid transaction type"})
	}

	transaction := models.Transaction{
		UserID:      userID,
		AccountID:   req.AccountID,
		Amount:      amount,
		Direction:   direction,
		Description: req.Description,
		Date:        req.Date,
		Notes:       req.Notes,
		Type:        req.Type,
		CategoryID:  req.CategoryID,
	}

	if err := config.DB.Create(&transaction).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create transaction"})
	}

	// Cargar relaciones para la respuesta
	config.DB.Preload("Account").Preload("Category").First(&transaction, transaction.ID)

	return c.Status(201).JSON(fiber.Map{
		"message":     "Transaction created successfully",
		"transaction": transaction,
	})
}

func GetTransactions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	// Filtros opcionales
	accountID := c.Query("account_id")
	categoryID := c.Query("category_id")
	transactionType := c.Query("type")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	query := config.DB.Where("user_id = ?", userID)

	// Aplicar filtros
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}
	if dateFrom != "" {
		query = query.Where("date >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("date <= ?", dateTo)
	}

	var transactions []models.Transaction
	if err := query.Preload("Account").Preload("Category").Order("date desc").Find(&transactions).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch transactions"})
	}

	return c.JSON(fiber.Map{
		"transactions": transactions,
	})
}

func GetTransaction(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	transactionID := c.Params("id")

	var transaction models.Transaction
	if err := config.DB.Preload("Account").Preload("Category").Where("id = ? AND user_id = ?", transactionID, userID).First(&transaction).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
	}

	return c.JSON(fiber.Map{
		"transaction": transaction,
	})
}

func UpdateTransaction(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	transactionID := c.Params("id")

	var req UpdateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	var transaction models.Transaction
	if err := config.DB.Where("id = ? AND user_id = ?", transactionID, userID).First(&transaction).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
	}

	// Verificar cuenta si se está cambiando
	if req.AccountID != nil {
		var account models.Account
		if err := config.DB.Where("id = ? AND user_id = ?", *req.AccountID, userID).First(&account).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
		}
		transaction.AccountID = *req.AccountID
	}

	// Verificar categoría si se está cambiando
	if req.CategoryID != nil {
		var category models.Category
		if err := config.DB.Where("id = ? AND user_id = ?", *req.CategoryID, userID).First(&category).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
		}
		transaction.CategoryID = req.CategoryID
	}

	// Actualizar campos
	if req.Amount != nil {
		// Mantener consistencia direction/amount basado en el type existente
		switch transaction.Type {
		case "income", "loan_received", "loan_payment_received":
			transaction.Amount = abs(*req.Amount)
		case "expense", "loan_given", "loan_payment_given":
			transaction.Amount = -abs(*req.Amount)
		}
	}
	if req.Description != "" {
		transaction.Description = req.Description
	}
	if req.Date != nil {
		transaction.Date = *req.Date
	}
	if req.Notes != "" {
		transaction.Notes = req.Notes
	}

	if err := config.DB.Save(&transaction).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update transaction"})
	}

	// Cargar relaciones para la respuesta
	config.DB.Preload("Account").Preload("Category").First(&transaction, transaction.ID)

	return c.JSON(fiber.Map{
		"message":     "Transaction updated successfully",
		"transaction": transaction,
	})
}

func DeleteTransaction(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	transactionID := c.Params("id")

	var transaction models.Transaction
	if err := config.DB.Where("id = ? AND user_id = ?", transactionID, userID).First(&transaction).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
	}

	// Soft delete
	if err := config.DB.Delete(&transaction).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not delete transaction"})
	}

	return c.JSON(fiber.Map{
		"message": "Transaction deleted successfully",
	})
}

// Función auxiliar para valor absoluto
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
