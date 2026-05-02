ALTER TABLE users ADD COLUMN ai_chat_count INT NOT NULL DEFAULT 0 AFTER password_hash;
