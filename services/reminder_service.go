package services

import (
	"cuentas-claras/config"
	"cuentas-claras/models"
	"fmt"
	"log"
	"time"
)

type ReminderService struct{}

// Crear recordatorios para un gasto recurrente
func (rs *ReminderService) CreateRemindersForRecurringExpense(recurringExpense *models.RecurringExpense) error {
	// Eliminar recordatorios anteriores pendientes
	config.DB.Where("reference_id = ? AND reference_type = ? AND is_sent = false",
		recurringExpense.ID, "recurring_expense").Delete(&models.Reminder{})

	// Recordatorio 2 días antes
	reminder2Days := models.Reminder{
		UserID:        recurringExpense.UserID,
		Title:         fmt.Sprintf("Próximo: %s", recurringExpense.Description),
		Description:   fmt.Sprintf("Vence en 2 días - %.2f %s", recurringExpense.Amount, "PEN"),
		Type:          "recurring_expense",
		ReferenceID:   &recurringExpense.ID,
		ReferenceType: "recurring_expense",
		RemindAt:      recurringExpense.NextDueDate.AddDate(0, 0, -2),
		Priority:      "normal",
	}

	// Recordatorio 1 día antes
	reminder1Day := models.Reminder{
		UserID:        recurringExpense.UserID,
		Title:         fmt.Sprintf("Mañana vence: %s", recurringExpense.Description),
		Description:   fmt.Sprintf("Vence mañana - %.2f %s", recurringExpense.Amount, "PEN"),
		Type:          "recurring_expense",
		ReferenceID:   &recurringExpense.ID,
		ReferenceType: "recurring_expense",
		RemindAt:      recurringExpense.NextDueDate.AddDate(0, 0, -1),
		Priority:      "high",
	}

	// Recordatorio el día mismo
	reminderToday := models.Reminder{
		UserID:        recurringExpense.UserID,
		Title:         fmt.Sprintf("¡Vence hoy! %s", recurringExpense.Description),
		Description:   fmt.Sprintf("Vence hoy - %.2f %s", recurringExpense.Amount, "PEN"),
		Type:          "recurring_expense",
		ReferenceID:   &recurringExpense.ID,
		ReferenceType: "recurring_expense",
		RemindAt:      recurringExpense.NextDueDate,
		Priority:      "high",
	}

	// Solo crear recordatorios si la fecha es futura
	now := time.Now()
	if reminder2Days.RemindAt.After(now) {
		config.DB.Create(&reminder2Days)
	}
	if reminder1Day.RemindAt.After(now) {
		config.DB.Create(&reminder1Day)
	}
	if reminderToday.RemindAt.After(now) {
		config.DB.Create(&reminderToday)
	}

	return nil
}

// Procesar recordatorios pendientes (Job diario)
func (rs *ReminderService) ProcessPendingReminders() {
	var reminders []models.Reminder

	// Buscar recordatorios que deben enviarse
	config.DB.Where("remind_at <= ? AND is_sent = false AND is_active = true", time.Now()).
		Find(&reminders)

	for _, reminder := range reminders {
		// Enviar notificación
		rs.SendNotification(&reminder)

		// Marcar como enviado
		reminder.MarkAsSent(config.DB)

		log.Printf("Reminder sent: %s to user %d", reminder.Title, reminder.UserID)
	}

	// También verificar gastos recurrentes vencidos
	rs.CheckOverdueRecurringExpenses()
}

// Verificar gastos recurrentes vencidos
func (rs *ReminderService) CheckOverdueRecurringExpenses() {
	var overdueExpenses []models.RecurringExpense

	// Buscar gastos recurrentes vencidos (más de 1 día de retraso)
	yesterday := time.Now().AddDate(0, 0, -1)
	config.DB.Where("next_due_date < ? AND is_active = true", yesterday).
		Find(&overdueExpenses)

	for _, expense := range overdueExpenses {
		// Crear recordatorio de vencido si no existe uno reciente
		var existingReminder models.Reminder
		err := config.DB.Where("reference_id = ? AND reference_type = ? AND title LIKE ?",
			expense.ID, "recurring_expense", "¡VENCIDO!%").
			Where("created_at > ?", time.Now().AddDate(0, 0, -1)).
			First(&existingReminder).Error

		if err != nil { // No existe, crear nuevo
			overdueReminder := models.Reminder{
				UserID: expense.UserID,
				Title:  fmt.Sprintf("¡VENCIDO! %s", expense.Description),
				Description: fmt.Sprintf("Lleva %d días vencido - %.2f %s",
					int(time.Since(expense.NextDueDate).Hours()/24), expense.Amount, "PEN"),
				Type:          "recurring_expense",
				ReferenceID:   &expense.ID,
				ReferenceType: "recurring_expense",
				RemindAt:      time.Now(),
				Priority:      "high",
			}

			config.DB.Create(&overdueReminder)
			rs.SendNotification(&overdueReminder)
			overdueReminder.MarkAsSent(config.DB)

			log.Printf("Overdue reminder sent: %s to user %d", overdueReminder.Title, overdueReminder.UserID)
		}
	}
}

// Enviar notificación (placeholder por ahora)
func (rs *ReminderService) SendNotification(reminder *models.Reminder) {
	// Por ahora solo log, después implementaremos FCM/WebSocket
	log.Printf("NOTIFICATION: %s - %s (User: %d)", reminder.Title, reminder.Description, reminder.UserID)

	// TODO: Implementar FCM push notification
	// TODO: Implementar WebSocket notification
	// TODO: Implementar in-app notification
}

// Iniciar el job scheduler (corre una vez al día)
func (rs *ReminderService) StartDailyJob() {
	go func() {
		// Ejecutar inmediatamente al iniciar
		rs.ProcessPendingReminders()

		// Luego cada 24 horas
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			rs.ProcessPendingReminders()
		}
	}()

	log.Println("Daily reminder job started")
}
