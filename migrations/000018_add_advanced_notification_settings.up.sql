ALTER TABLE notification_settings
    ADD COLUMN budget_warning_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN budget_warning_threshold INT NOT NULL DEFAULT 80,
    ADD COLUMN weekly_summary_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN weekly_summary_day INT NOT NULL DEFAULT 0,
    ADD COLUMN large_transaction_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN large_transaction_threshold DECIMAL(18,2) NOT NULL DEFAULT 1000000,
    ADD COLUMN goal_reminder_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN goal_reminder_days_before INT NOT NULL DEFAULT 7;
