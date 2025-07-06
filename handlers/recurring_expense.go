package handlers

import (
	"cuentas-claras/config"
	"cuentas-claras/models"
	"cuentas-claras/services"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CreateRecurringExpenseRequest struct {
	AccountID   uint       `json:"account_id" validate:"required"`
	CategoryID  uint       `json:"category_id" validate:"required"`
	Amount      float64    `json:"amount" validate:"required,gt=0"`
	Description string     `json:"description" validate:"required,min=1,max=255"`
	Frequency   string     `json:"frequency" validate:"required,oneof=daily weekly monthly yearly"`
	StartDate   time.Time  `json:"start_date" validate:"required"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Notes       string     `json:"notes,omitempty"`
}

type UpdateRecurringExpenseRequest struct {
	AccountID   *uint      `json:"account_id,omitempty"`
	CategoryID  *uint      `json:"category_id,omitempty"`
	Amount      *float64   `json:"amount,omitempty" validate:"omitempty,gt=0"`
	Description string     `json:"description,omitempty" validate:"omitempty,min=1,max=255"`
	Frequency   string     `json:"frequency,omitempty" validate:"omitempty,oneof=daily weekly monthly yearly"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Notes       string     `json:"notes,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
}

func CreateRecurringExpense(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req CreateRecurringExpenseRequest
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

	// Verificar que la categoría pertenece al usuario
	var category models.Category
	if err := config.DB.Where("id = ? AND user_id = ?", req.CategoryID, userID).First(&category).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
	}

	// Crear gasto recurrente
	recurringExpense := models.RecurringExpense{
		UserID:      userID,
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		Description: req.Description,
		Frequency:   req.Frequency,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		NextDueDate: req.StartDate, // Primera vez vence en start_date
		Notes:       req.Notes,
	}

	if err := config.DB.Create(&recurringExpense).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create recurring expense"})
	}

	// Crear recordatorios automáticamente
	reminderService := &services.ReminderService{}
	reminderService.CreateRemindersForRecurringExpense(&recurringExpense)

	// Cargar relaciones para la respuesta
	config.DB.Preload("Account").Preload("Category").First(&recurringExpense, recurringExpense.ID)

	return c.Status(201).JSON(fiber.Map{
		"message":           "Recurring expense created successfully",
		"recurring_expense": recurringExpense,
	})
}

func GetRecurringExpenses(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	// Filtros opcionales
	isActive := c.Query("is_active", "true") // Por defecto solo activos
	frequency := c.Query("frequency")

	query := config.DB.Where("user_id = ?", userID)

	// Aplicar filtros
	if isActive == "true" {
		query = query.Where("is_active = true")
	} else if isActive == "false" {
		query = query.Where("is_active = false")
	}
	// Si es "all", no filtrar por is_active

	if frequency != "" {
		query = query.Where("frequency = ?", frequency)
	}

	var recurringExpenses []models.RecurringExpense
	if err := query.Preload("Account").Preload("Category").Order("next_due_date asc").Find(&recurringExpenses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch recurring expenses"})
	}

	// Agregar información útil
	result := make([]fiber.Map, len(recurringExpenses))
	for i, re := range recurringExpenses {
		result[i] = fiber.Map{
			"id":            re.ID,
			"account":       re.Account,
			"category":      re.Category,
			"amount":        re.Amount,
			"description":   re.Description,
			"frequency":     re.Frequency,
			"start_date":    re.StartDate,
			"end_date":      re.EndDate,
			"next_due_date": re.NextDueDate,
			"is_active":     re.IsActive,
			"notes":         re.Notes,
			"created_at":    re.CreatedAt,
			"is_overdue":    re.IsOverdue(),
			"is_due_today":  re.IsDueToday(),
		}
	}

	return c.JSON(fiber.Map{
		"recurring_expenses": result,
	})
}

func GetRecurringExpense(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	expenseID := c.Params("id")

	var recurringExpense models.RecurringExpense
	if err := config.DB.Preload("Account").Preload("Category").Where("id = ? AND user_id = ?", expenseID, userID).First(&recurringExpense).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Recurring expense not found"})
	}

	return c.JSON(fiber.Map{
		"recurring_expense": fiber.Map{
			"id":            recurringExpense.ID,
			"account":       recurringExpense.Account,
			"category":      recurringExpense.Category,
			"amount":        recurringExpense.Amount,
			"description":   recurringExpense.Description,
			"frequency":     recurringExpense.Frequency,
			"start_date":    recurringExpense.StartDate,
			"end_date":      recurringExpense.EndDate,
			"next_due_date": recurringExpense.NextDueDate,
			"is_active":     recurringExpense.IsActive,
			"notes":         recurringExpense.Notes,
			"created_at":    recurringExpense.CreatedAt,
			"updated_at":    recurringExpense.UpdatedAt,
			"is_overdue":    recurringExpense.IsOverdue(),
			"is_due_today":  recurringExpense.IsDueToday(),
		},
	})
}

func UpdateRecurringExpense(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	expenseID := c.Params("id")

	var req UpdateRecurringExpenseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	var recurringExpense models.RecurringExpense
	if err := config.DB.Where("id = ? AND user_id = ?", expenseID, userID).First(&recurringExpense).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Recurring expense not found"})
	}

	// Verificar cuenta si se está cambiando
	if req.AccountID != nil {
		var account models.Account
		if err := config.DB.Where("id = ? AND user_id = ?", *req.AccountID, userID).First(&account).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
		}
		recurringExpense.AccountID = *req.AccountID
	}

	// Verificar categoría si se está cambiando
	if req.CategoryID != nil {
		var category models.Category
		if err := config.DB.Where("id = ? AND user_id = ?", *req.CategoryID, userID).First(&category).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
		}
		recurringExpense.CategoryID = *req.CategoryID
	}

	// Actualizar campos
	if req.Amount != nil {
		recurringExpense.Amount = *req.Amount
	}
	if req.Description != "" {
		recurringExpense.Description = req.Description
	}
	if req.Frequency != "" {
		recurringExpense.Frequency = req.Frequency
	}
	if req.EndDate != nil {
		recurringExpense.EndDate = req.EndDate
	}
	if req.Notes != "" {
		recurringExpense.Notes = req.Notes
	}
	if req.IsActive != nil {
		recurringExpense.IsActive = *req.IsActive
	}

	if err := config.DB.Save(&recurringExpense).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update recurring expense"})
	}

	// Cargar relaciones para la respuesta
	config.DB.Preload("Account").Preload("Category").First(&recurringExpense, recurringExpense.ID)

	return c.JSON(fiber.Map{
		"message":           "Recurring expense updated successfully",
		"recurring_expense": recurringExpense,
	})
}

func DeleteRecurringExpense(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	expenseID := c.Params("id")

	var recurringExpense models.RecurringExpense
	if err := config.DB.Where("id = ? AND user_id = ?", expenseID, userID).First(&recurringExpense).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Recurring expense not found"})
	}

	// Soft delete
	if err := config.DB.Delete(&recurringExpense).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not delete recurring expense"})
	}

	return c.JSON(fiber.Map{
		"message": "Recurring expense deleted successfully",
	})
}

// Endpoint para ejecutar (pagar) un gasto recurrente
func ExecuteRecurringExpense(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	expenseID := c.Params("id")

	type ExecuteRequest struct {
		AccountID *uint      `json:"account_id,omitempty"` // Puede ser diferente al original
		Amount    *float64   `json:"amount,omitempty"`     // Puede ser diferente al original
		Date      *time.Time `json:"date,omitempty"`       // Fecha real del pago
		Notes     string     `json:"notes,omitempty"`      // Notas del pago específico
	}

	var req ExecuteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Buscar el gasto recurrente
	var recurringExpense models.RecurringExpense
	if err := config.DB.Where("id = ? AND user_id = ?", expenseID, userID).First(&recurringExpense).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Recurring expense not found"})
	}

	if !recurringExpense.IsActive {
		return c.Status(400).JSON(fiber.Map{"error": "Recurring expense is not active"})
	}

	// Usar valores por defecto si no se proporcionan
	accountID := recurringExpense.AccountID
	if req.AccountID != nil {
		accountID = *req.AccountID
	}

	amount := recurringExpense.Amount
	if req.Amount != nil {
		amount = *req.Amount
	}

	paymentDate := time.Now()
	if req.Date != nil {
		paymentDate = *req.Date
	}

	// Verificar que la cuenta pertenece al usuario
	var account models.Account
	if err := config.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&account).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	// Crear la transacción (gasto)
	transaction := models.Transaction{
		UserID:        userID,
		AccountID:     accountID,
		Amount:        -amount, // Negativo porque es un gasto
		Direction:     "out",
		Description:   recurringExpense.Description,
		Date:          paymentDate,
		Notes:         req.Notes,
		Type:          "expense",
		CategoryID:    &recurringExpense.CategoryID,
		ReferenceID:   &recurringExpense.ID,
		ReferenceType: "recurring_expense",
	}

	if err := config.DB.Create(&transaction).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create transaction"})
	}

	// Actualizar next_due_date al siguiente vencimiento
	recurringExpense.NextDueDate = recurringExpense.CalculateNextDueDate()
	if err := config.DB.Save(&recurringExpense).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update next due date"})
	}

	// Crear recordatorios para el próximo vencimiento
	reminderService := &services.ReminderService{}
	reminderService.CreateRemindersForRecurringExpense(&recurringExpense)

	// Cargar relaciones para la respuesta
	config.DB.Preload("Account").Preload("Category").First(&transaction, transaction.ID)

	return c.JSON(fiber.Map{
		"message":           "Recurring expense executed successfully",
		"transaction":       transaction,
		"next_due_date":     recurringExpense.NextDueDate,
		"recurring_expense": recurringExpense.ID,
	})
}
