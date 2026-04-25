ALTER TABLE notification_settings
    ADD COLUMN budget_amount DECIMAL(15, 2) NOT NULL DEFAULT 0 AFTER salary_day;
