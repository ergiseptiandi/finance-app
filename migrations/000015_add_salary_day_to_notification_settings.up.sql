ALTER TABLE notification_settings
    ADD COLUMN salary_day TINYINT UNSIGNED NOT NULL DEFAULT 25 AFTER salary_reminder_days_before;
