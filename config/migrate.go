package config

import (
	"cuentas-claras/models"
	"fmt"
)

func RunMigrations() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.DeviceSession{},
		&models.Account{},
		&models.Category{},
		&models.Transaction{},
		&models.Loan{},
		&models.LoanPayment{},
		&models.RecurringExpense{},
		&models.Reminder{}, // ✨ NUEVO
	)

	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Agregar constraints personalizados para transacciones
	AddTransactionConstraints()
	AddRecurringExpenseConstraints()

	fmt.Println("Database migrations completed successfully")
}
