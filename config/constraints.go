package config

import (
	"fmt"
)

func AddTransactionConstraints() {
	// 1. Direction solo acepta 'in' o 'out'
	err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS temp_constraint_check AS 
		SELECT 1 WHERE NOT EXISTS (
			SELECT 1 FROM pragma_table_info('transactions') 
			WHERE name = 'direction'
		);
	`).Error

	if err == nil {
		// SQLite constraint para direction
		DB.Exec(`
			CREATE TRIGGER IF NOT EXISTS check_direction_insert
			BEFORE INSERT ON transactions
			FOR EACH ROW
			WHEN NEW.direction NOT IN ('in', 'out')
			BEGIN
				SELECT RAISE(ABORT, 'Direction must be either "in" or "out"');
			END;
		`)

		DB.Exec(`
			CREATE TRIGGER IF NOT EXISTS check_direction_update
			BEFORE UPDATE ON transactions
			FOR EACH ROW
			WHEN NEW.direction NOT IN ('in', 'out')
			BEGIN
				SELECT RAISE(ABORT, 'Direction must be either "in" or "out"');
			END;
		`)

		// 2. Consistencia Amount + Direction
		DB.Exec(`
			CREATE TRIGGER IF NOT EXISTS check_amount_direction_insert
			BEFORE INSERT ON transactions
			FOR EACH ROW
			WHEN NOT (
				(NEW.direction = 'in' AND NEW.amount > 0) OR 
				(NEW.direction = 'out' AND NEW.amount < 0)
			)
			BEGIN
				SELECT RAISE(ABORT, 'Amount must be positive for "in" direction and negative for "out" direction');
			END;
		`)

		DB.Exec(`
			CREATE TRIGGER IF NOT EXISTS check_amount_direction_update
			BEFORE UPDATE ON transactions
			FOR EACH ROW
			WHEN NOT (
				(NEW.direction = 'in' AND NEW.amount > 0) OR 
				(NEW.direction = 'out' AND NEW.amount < 0)
			)
			BEGIN
				SELECT RAISE(ABORT, 'Amount must be positive for "in" direction and negative for "out" direction');
			END;
		`)

		// 3. Consistencia Type + Direction
		DB.Exec(`
			CREATE TRIGGER IF NOT EXISTS check_type_direction_insert
			BEFORE INSERT ON transactions
			FOR EACH ROW
			WHEN NOT (
				(NEW.type IN ('income', 'loan_received', 'loan_payment_received') AND NEW.direction = 'in') OR
				(NEW.type IN ('expense', 'loan_given', 'loan_payment_given') AND NEW.direction = 'out')
			)
			BEGIN
				SELECT RAISE(ABORT, 'Transaction type and direction are inconsistent');
			END;
		`)

		DB.Exec(`
			CREATE TRIGGER IF NOT EXISTS check_type_direction_update
			BEFORE UPDATE ON transactions
			FOR EACH ROW
			WHEN NOT (
				(NEW.type IN ('income', 'loan_received', 'loan_payment_received') AND NEW.direction = 'in') OR
				(NEW.type IN ('expense', 'loan_given', 'loan_payment_given') AND NEW.direction = 'out')
			)
			BEGIN
				SELECT RAISE(ABORT, 'Transaction type and direction are inconsistent');
			END;
		`)

		// 4. Validar tipos permitidos
		DB.Exec(`
			CREATE TRIGGER IF NOT EXISTS check_valid_type_insert
			BEFORE INSERT ON transactions
			FOR EACH ROW
			WHEN NEW.type NOT IN ('income', 'expense', 'loan_given', 'loan_received', 'loan_payment_given', 'loan_payment_received')
			BEGIN
				SELECT RAISE(ABORT, 'Invalid transaction type');
			END;
		`)

		DB.Exec(`
			CREATE TRIGGER IF NOT EXISTS check_valid_type_update
			BEFORE UPDATE ON transactions
			FOR EACH ROW
			WHEN NEW.type NOT IN ('income', 'expense', 'loan_given', 'loan_received', 'loan_payment_given', 'loan_payment_received')
			BEGIN
				SELECT RAISE(ABORT, 'Invalid transaction type');
			END;
		`)

		// 5. Amount no puede ser cero
		DB.Exec(`
			CREATE TRIGGER IF NOT EXISTS check_amount_not_zero_insert
			BEFORE INSERT ON transactions
			FOR EACH ROW
			WHEN NEW.amount = 0
			BEGIN
				SELECT RAISE(ABORT, 'Amount cannot be zero');
			END;
		`)

		DB.Exec(`
			CREATE TRIGGER IF NOT EXISTS check_amount_not_zero_update
			BEFORE UPDATE ON transactions
			FOR EACH ROW
			WHEN NEW.amount = 0
			BEGIN
				SELECT RAISE(ABORT, 'Amount cannot be zero');
			END;
		`)

		fmt.Println("Transaction constraints created successfully")
	}
}

// Constraints para RecurringExpense
func AddRecurringExpenseConstraints() {
	// Frecuencias válidas
	DB.Exec(`
		CREATE TRIGGER IF NOT EXISTS check_frequency_insert
		BEFORE INSERT ON recurring_expenses
		FOR EACH ROW
		WHEN NEW.frequency NOT IN ('daily', 'weekly', 'monthly', 'yearly')
		BEGIN
			SELECT RAISE(ABORT, 'Frequency must be daily, weekly, monthly, or yearly');
		END;
	`)

	DB.Exec(`
		CREATE TRIGGER IF NOT EXISTS check_frequency_update
		BEFORE UPDATE ON recurring_expenses
		FOR EACH ROW
		WHEN NEW.frequency NOT IN ('daily', 'weekly', 'monthly', 'yearly')
		BEGIN
			SELECT RAISE(ABORT, 'Frequency must be daily, weekly, monthly, or yearly');
		END;
	`)

	// Amount debe ser positivo
	DB.Exec(`
		CREATE TRIGGER IF NOT EXISTS check_amount_positive_insert
		BEFORE INSERT ON recurring_expenses
		FOR EACH ROW
		WHEN NEW.amount <= 0
		BEGIN
			SELECT RAISE(ABORT, 'Amount must be positive');
		END;
	`)

	DB.Exec(`
		CREATE TRIGGER IF NOT EXISTS check_amount_positive_update
		BEFORE UPDATE ON recurring_expenses
		FOR EACH ROW
		WHEN NEW.amount <= 0
		BEGIN
			SELECT RAISE(ABORT, 'Amount must be positive');
		END;
	`)

	fmt.Println("RecurringExpense constraints created successfully")
}
