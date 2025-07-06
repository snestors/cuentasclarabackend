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
	)

	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Agregar constraints personalizados para transacciones
	AddTransactionConstraints()

	fmt.Println("Database migrations completed successfully")
}
