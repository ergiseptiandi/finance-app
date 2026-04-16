package seed

import (
	"context"
	"database/sql"

	"finance-backend/internal/category"
)

func UpsertDefaultCategories(ctx context.Context, db *sql.DB) error {
	const query = `
		INSERT INTO categories (name, type)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
			updated_at = CURRENT_TIMESTAMP
	`

	defaults := []category.Category{
		{Name: "income", Type: category.TypeIncome},
		{Name: "expense", Type: category.TypeExpense},
	}

	for _, item := range defaults {
		if _, err := db.ExecContext(ctx, query, item.Name, item.Type); err != nil {
			return err
		}
	}

	return nil
}
