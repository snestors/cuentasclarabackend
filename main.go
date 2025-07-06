package main

import (
	"log"
	"os"

	"cuentas-claras/config"
	"cuentas-claras/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Conectar a la base de datos
	config.ConnectDatabase()

	// Ejecutar migraciones
	config.RunMigrations()

	// Crear app Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Ruta de prueba
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "CuentasClaras API - Running!",
		})
	})

	// Configurar rutas
	routes.SetupRoutes(app)

	// Puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
