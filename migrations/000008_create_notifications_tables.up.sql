CREATE TABLE IF NOT EXISTS notification_settings (
    user_id BIGINT NOT NULL PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    daily_expense_reminder_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    daily_expense_reminder_time VARCHAR(5) NOT NULL DEFAULT '20:00',
    debt_payment_reminder_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    debt_payment_reminder_time VARCHAR(5) NOT NULL DEFAULT '09:00',
    debt_payment_reminder_days_before INT NOT NULL DEFAULT 3,
    salary_reminder_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    salary_reminder_time VARCHAR(5) NOT NULL DEFAULT '08:00',
    salary_reminder_days_before INT NOT NULL DEFAULT 1,
    push_token TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    kind VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    delivery_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    scheduled_for DATETIME NOT NULL,
    sent_at DATETIME NULL,
    read_at DATETIME NULL,
    dedupe_key VARCHAR(120) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uq_notifications_user_dedupe (user_id, dedupe_key),
    INDEX idx_notifications_user_created_at (user_id, created_at)
);
