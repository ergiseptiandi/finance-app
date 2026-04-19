ALTER TABLE categories
    DROP FOREIGN KEY fk_categories_user;

ALTER TABLE categories
    DROP INDEX uk_categories_user_name_type,
    DROP INDEX idx_categories_user_type_name;

ALTER TABLE categories
    DROP COLUMN user_id,
    ADD UNIQUE KEY uk_categories_name_type (name, type);
