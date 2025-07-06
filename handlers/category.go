package handlers

import (
	"cuentas-claras/config"
	"cuentas-claras/models"

	"github.com/gofiber/fiber/v2"
)

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

func CreateCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	// Verificar que no existe una categoría con el mismo nombre para el usuario
	var existingCategory models.Category
	if err := config.DB.Where("user_id = ? AND name = ? AND deleted_at IS NULL", userID, req.Name).First(&existingCategory).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Category name already exists"})
	}

	category := models.Category{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	}

	// Asignar valores por defecto si no se proporcionan
	if category.Color == "" {
		category.Color = "#4CAF50"
	}
	if category.Icon == "" {
		category.Icon = "category"
	}

	if err := config.DB.Create(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create category"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":  "Category created successfully",
		"category": category,
	})
}

func GetCategories(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var categories []models.Category
	if err := config.DB.Where("user_id = ? AND is_active = true", userID).Find(&categories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch categories"})
	}

	return c.JSON(fiber.Map{
		"categories": categories,
	})
}

func GetCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	categoryID := c.Params("id")

	var category models.Category
	if err := config.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
	}

	return c.JSON(fiber.Map{
		"category": category,
	})
}

func UpdateCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	categoryID := c.Params("id")

	var req UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
	}

	var category models.Category
	if err := config.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
	}

	// Verificar nombre único si se está cambiando
	if req.Name != "" && req.Name != category.Name {
		var existingCategory models.Category
		if err := config.DB.Where("user_id = ? AND name = ? AND id != ? AND deleted_at IS NULL", userID, req.Name, categoryID).First(&existingCategory).Error; err == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Category name already exists"})
		}
		category.Name = req.Name
	}

	// Actualizar campos
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.Color != "" {
		category.Color = req.Color
	}
	if req.Icon != "" {
		category.Icon = req.Icon
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := config.DB.Save(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update category"})
	}

	return c.JSON(fiber.Map{
		"message":  "Category updated successfully",
		"category": category,
	})
}

func DeleteCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	categoryID := c.Params("id")

	var category models.Category
	if err := config.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
	}

	// Soft delete
	if err := config.DB.Delete(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not delete category"})
	}

	return c.JSON(fiber.Map{
		"message": "Category deleted successfully",
	})
}
