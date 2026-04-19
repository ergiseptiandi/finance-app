ALTER TABLE categories
    ADD COLUMN user_id BIGINT NULL AFTER id;

ALTER TABLE categories
    ADD CONSTRAINT fk_categories_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE categories
    DROP INDEX uk_categories_name_type,
    ADD UNIQUE KEY uk_categories_user_name_type (user_id, name, type),
    ADD INDEX idx_categories_user_type_name (user_id, type, name);
