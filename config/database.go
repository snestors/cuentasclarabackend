package config

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite" // Default para desarrollo
	}

	var database *gorm.DB
	var err error

	switch driver {
	case "sqlite":
		dbname := os.Getenv("DB_NAME_SQLITE")
		if dbname == "" {
			dbname = "cuentas_claras.db"
		}
		database, err = gorm.Open(sqlite.Open(dbname), &gorm.Config{})
		fmt.Println("Using SQLite database:", dbname)

	case "postgres":
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		fmt.Println("Using PostgreSQL database:", dbname)

	default:
		panic("Unsupported database driver: " + driver)
	}

	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	DB = database
	fmt.Println("Database connected successfully")
}
