ALTER TABLE users ADD COLUMN ai_max_chats INT NOT NULL DEFAULT 10 AFTER ai_chat_count;
