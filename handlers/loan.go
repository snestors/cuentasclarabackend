package handlers

import (
	"cuentas-claras/config"
	"cuentas-claras/models"
	"cuentas-claras/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CreateLoanRequest struct {
	AccountID    uint       `json:"account_id" validate:"required"`
	Amount       float64    `json:"amount" validate:"required,gt=0"`
	Description  string     `json:"description" validate:"required,min=1,max=255"`
	PersonName   string     `json:"person_name" validate:"required,min=1,max=100"`
	Type         string     `json:"type" validate:"required,oneof=given received"`
	LoanDate     time.Time  `json:"loan_date" validate:"required"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	InterestRate float64    `json:"interest_rate,omitempty"`
	Notes        string     `json:"notes,omitempty"`
}

type UpdateLoanRequest struct {
	Description  string     `json:"description,omitempty" validate:"omitempty,min=1,max=255"`
	PersonName   string     `json:"person_name,omitempty" validate:"omitempty,min=1,max=100"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	InterestRate *float64   `json:"interest_rate,omitempty"`
	Notes        string     `json:"notes,omitempty"`
}

type CreateLoanPaymentRequest struct {
	AccountID   uint      `json:"account_id" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Date        time.Time `json:"date" validate:"required"`
	Description string    `json:"description" validate:"required,min=1,max=255"`
	Notes       string    `json:"notes,omitempty"`
}

func CreateLoan(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req CreateLoanRequest
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

	// Crear préstamo
	loan := models.Loan{
		UserID:       userID,
		AccountID:    req.AccountID,
		Amount:       req.Amount,
		Description:  req.Description,
		PersonName:   req.PersonName,
		Type:         req.Type,
		LoanDate:     req.LoanDate,
		DueDate:      req.DueDate,
		InterestRate: req.InterestRate,
		Notes:        req.Notes,
		Status:       "pending",
	}

	if err := config.DB.Create(&loan).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create loan"})
	}

	// Crear transacción automática del préstamo inicial
	var transactionType string
	var direction string
	var amount float64

	if req.Type == "given" {
		// Préstamo dado: dinero sale de mi cuenta
		transactionType = "loan_given"
		direction = "out"
		amount = -abs(req.Amount)
	} else {
		// Préstamo recibido: dinero entra a mi cuenta
		transactionType = "loan_received"
		direction = "in"
		amount = abs(req.Amount)
	}

	transaction := models.Transaction{
		UserID:        userID,
		AccountID:     req.AccountID,
		Amount:        amount,
		Direction:     direction,
		Description:   "Préstamo: " + req.Description,
		Date:          req.LoanDate,
		Type:          transactionType,
		ReferenceID:   &loan.ID,
		ReferenceType: "loan",
		Notes:         req.Notes,
	}

	if err := config.DB.Create(&transaction).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create transaction"})
	}

	// Cargar relaciones para la respuesta
	config.DB.Preload("Account").First(&loan, loan.ID)

	return c.Status(201).JSON(fiber.Map{
		"message": "Loan created successfully",
		"loan":    loan,
	})
}

func GetLoans(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var loans []models.Loan
	if err := config.DB.Preload("Account").Where("user_id = ?", userID).Find(&loans).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch loans"})
	}

	// Crear response manualmente
	loansWithBalance := make([]fiber.Map, len(loans))
	for i, loan := range loans {
		// PRIMERO actualizar status
		loan.UpdateStatus(config.DB)

		// DESPUÉS desencriptar para el response (sin modificar el struct original)
		description := loan.Description
		personName := loan.PersonName
		notes := loan.Notes

		if description != "" {
			description = utils.DecryptField(description)
		}
		if personName != "" {
			personName = utils.DecryptField(personName)
		}
		if notes != "" {
			notes = utils.DecryptField(notes)
		}

		totalPaid := loan.GetTotalPaid(config.DB)
		balance := loan.GetBalance(config.DB)

		loansWithBalance[i] = fiber.Map{
			"id":            loan.ID,
			"user_id":       loan.UserID,
			"account_id":    loan.AccountID,
			"amount":        loan.Amount,
			"description":   description, // Desencriptado solo para response
			"person_name":   personName,  // Desencriptado solo para response
			"type":          loan.Type,
			"status":        loan.Status,
			"loan_date":     loan.LoanDate,
			"due_date":      loan.DueDate,
			"interest_rate": loan.InterestRate,
			"notes":         notes, // Desencriptado solo para response
			"created_at":    loan.CreatedAt,
			"updated_at":    loan.UpdatedAt,
			"account":       loan.Account,
			"total_paid":    totalPaid,
			"balance":       balance,
		}
	}

	return c.JSON(fiber.Map{
		"loans": loansWithBalance,
	})
}

func GetLoan(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	loanID := c.Params("id")

	var loan models.Loan
	if err := config.DB.Preload("Account").Preload("Payments.Account").Where("id = ? AND user_id = ?", loanID, userID).First(&loan).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Loan not found"})
	}

	totalPaid := loan.GetTotalPaid(config.DB)
	balance := loan.GetBalance(config.DB)
	loan.UpdateStatus(config.DB)

	return c.JSON(fiber.Map{
		"loan": fiber.Map{
			"id":            loan.ID,
			"account_id":    loan.AccountID,
			"amount":        loan.Amount,
			"description":   loan.Description,
			"person_name":   loan.PersonName,
			"type":          loan.Type,
			"status":        loan.Status,
			"loan_date":     loan.LoanDate,
			"due_date":      loan.DueDate,
			"interest_rate": loan.InterestRate,
			"notes":         loan.Notes,
			"total_paid":    totalPaid,
			"balance":       balance,
			"created_at":    loan.CreatedAt,
			"updated_at":    loan.UpdatedAt,
			"account":       loan.Account,
			"payments":      loan.Payments,
		},
	})
}

func UpdateLoan(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	loanID := c.Params("id")

	var req UpdateLoanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	var loan models.Loan
	if err := config.DB.Where("id = ? AND user_id = ?", loanID, userID).First(&loan).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Loan not found"})
	}

	// Actualizar campos
	if req.Description != "" {
		loan.Description = req.Description
	}
	if req.PersonName != "" {
		loan.PersonName = req.PersonName
	}
	if req.DueDate != nil {
		loan.DueDate = req.DueDate
	}
	if req.InterestRate != nil {
		loan.InterestRate = *req.InterestRate
	}
	if req.Notes != "" {
		loan.Notes = req.Notes
	}

	if err := config.DB.Save(&loan).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update loan"})
	}

	return c.JSON(fiber.Map{
		"message": "Loan updated successfully",
		"loan":    loan,
	})
}

func DeleteLoan(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	loanID := c.Params("id")

	var loan models.Loan
	if err := config.DB.Where("id = ? AND user_id = ?", loanID, userID).First(&loan).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Loan not found"})
	}

	// Verificar que no tenga pagos confirmados
	var confirmedPayments int64
	config.DB.Model(&models.LoanPayment{}).Where("loan_id = ? AND transaction_id IS NOT NULL", loanID).Count(&confirmedPayments)

	if confirmedPayments > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot delete loan with confirmed payments"})
	}

	// Soft delete
	if err := config.DB.Delete(&loan).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not delete loan"})
	}

	return c.JSON(fiber.Map{
		"message": "Loan deleted successfully",
	})
}

func CreateLoanPayment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	loanID := c.Params("id")

	var req CreateLoanPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	// Verificar que el préstamo existe y pertenece al usuario
	var loan models.Loan
	if err := config.DB.Where("id = ? AND user_id = ?", loanID, userID).First(&loan).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Loan not found"})
	}

	// Verificar que la cuenta pertenece al usuario
	var account models.Account
	if err := config.DB.Where("id = ? AND user_id = ?", req.AccountID, userID).First(&account).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	// Verificar que el monto no exceda el balance pendiente
	balance := loan.GetBalance(config.DB)
	if req.Amount > balance {
		return c.Status(400).JSON(fiber.Map{"error": "Payment amount exceeds loan balance"})
	}

	// Crear pago (sin confirmar)
	payment := models.LoanPayment{
		LoanID:      loan.ID,
		UserID:      userID,
		AccountID:   req.AccountID,
		Amount:      req.Amount,
		Date:        req.Date,
		Description: req.Description,
		Notes:       req.Notes,
		// TransactionID permanece nil (no confirmado)
	}

	if err := config.DB.Create(&payment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create payment"})
	}

	// Cargar relaciones para la respuesta
	config.DB.Preload("Account").Preload("Loan").First(&payment, payment.ID)

	return c.Status(201).JSON(fiber.Map{
		"message": "Payment created successfully (pending confirmation)",
		"payment": payment,
	})
}

func ConfirmLoanPayment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	paymentID := c.Params("id")

	var payment models.LoanPayment
	if err := config.DB.Preload("Loan").Preload("Account").Where("id = ? AND user_id = ?", paymentID, userID).First(&payment).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Payment not found"})
	}

	// Verificar que no esté ya confirmado
	if payment.IsConfirmed() {
		return c.Status(400).JSON(fiber.Map{"error": "Payment already confirmed"})
	}

	// Determinar tipo de transacción basado en el tipo de préstamo
	var transactionType string
	var direction string
	var amount float64

	if payment.Loan.Type == "given" {
		// Préstamo dado: me están pagando (dinero entra)
		transactionType = "loan_payment_received"
		direction = "in"
		amount = abs(payment.Amount)
	} else {
		// Préstamo recibido: estoy pagando (dinero sale)
		transactionType = "loan_payment_given"
		direction = "out"
		amount = -abs(payment.Amount)
	}

	// Crear transacción
	transaction := models.Transaction{
		UserID:        userID,
		AccountID:     payment.AccountID,
		Amount:        amount,
		Direction:     direction,
		Description:   "Pago préstamo: " + payment.Description,
		Date:          payment.Date,
		Type:          transactionType,
		ReferenceID:   &payment.Loan.ID,
		ReferenceType: "loan",
		Notes:         payment.Notes,
	}

	if err := config.DB.Create(&transaction).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create transaction"})
	}

	// Actualizar payment con transaction_id
	payment.TransactionID = &transaction.ID
	if err := config.DB.Save(&payment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update payment"})
	}

	// Actualizar status del préstamo
	payment.Loan.UpdateStatus(config.DB)

	// RECARGAR TODO con las relaciones correctas
	var finalPayment models.LoanPayment
	config.DB.Preload("Loan.Account").Preload("Account").First(&finalPayment, payment.ID)

	var finalTransaction models.Transaction
	config.DB.Preload("Account").First(&finalTransaction, transaction.ID)

	return c.JSON(fiber.Map{
		"message":     "Payment confirmed successfully",
		"payment":     finalPayment,
		"transaction": finalTransaction,
	})
}
