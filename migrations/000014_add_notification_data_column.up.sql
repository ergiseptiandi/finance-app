ALTER TABLE notifications
    ADD COLUMN data JSON NULL AFTER dedupe_key;
