package seed

import (
	"context"
	"database/sql"

	"finance-backend/internal/category"
)

func UpsertDefaultCategories(ctx context.Context, db *sql.DB, userID int64) error {
	const query = `
		INSERT INTO categories (user_id, name, type)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			updated_at = CURRENT_TIMESTAMP
	`

	defaults := []category.Category{
		{Name: "income", Type: category.TypeIncome},
		{Name: "expense", Type: category.TypeExpense},
	}

	for _, item := range defaults {
		if _, err := db.ExecContext(ctx, query, userID, item.Name, item.Type); err != nil {
			return err
		}
	}

	return nil
}
