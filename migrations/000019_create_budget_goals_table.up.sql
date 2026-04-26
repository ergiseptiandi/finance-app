CREATE TABLE IF NOT EXISTS budget_goals (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    monthly_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_budget_goals_user_category (user_id, category_id),
    INDEX idx_budget_goals_user_created_at (user_id, created_at, id),
    CONSTRAINT fk_budget_goals_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_budget_goals_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);
