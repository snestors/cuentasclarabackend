package routes

import (
	"cuentas-claras/handlers"
	"cuentas-claras/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// API Group
	api := app.Group("/api/v1")

	// Auth routes (públicas)
	auth := api.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Post("/refresh", handlers.RefreshToken)

	// Rutas protegidas (requieren autenticación)
	auth.Get("/profile", middleware.RequireAuth, handlers.GetProfile)
	auth.Put("/profile", middleware.RequireAuth, handlers.UpdateProfile)
	auth.Post("/logout", middleware.RequireAuth, handlers.Logout)

	// Account routes (protegidas)
	accounts := api.Group("/accounts", middleware.RequireAuth)
	accounts.Post("/", handlers.CreateAccount)
	accounts.Get("/", handlers.GetAccounts)
	accounts.Get("/:id", handlers.GetAccount)
	accounts.Put("/:id", handlers.UpdateAccount)
	accounts.Delete("/:id", handlers.DeleteAccount)

	// Category routes (protegidas)
	categories := api.Group("/categories", middleware.RequireAuth)
	categories.Post("/", handlers.CreateCategory)
	categories.Get("/", handlers.GetCategories)
	categories.Get("/:id", handlers.GetCategory)
	categories.Put("/:id", handlers.UpdateCategory)
	categories.Delete("/:id", handlers.DeleteCategory)

	// Transaction routes (protegidas)
	transactions := api.Group("/transactions", middleware.RequireAuth)
	transactions.Post("/", handlers.CreateTransaction)
	transactions.Get("/", handlers.GetTransactions)
	transactions.Get("/:id", handlers.GetTransaction)
	transactions.Put("/:id", handlers.UpdateTransaction)
	transactions.Delete("/:id", handlers.DeleteTransaction)
}
