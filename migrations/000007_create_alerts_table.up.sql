CREATE TABLE IF NOT EXISTS alerts (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    metric_value DECIMAL(15, 2) NOT NULL DEFAULT 0,
    threshold_value DECIMAL(15, 2) NOT NULL DEFAULT 0,
    dedupe_key VARCHAR(100) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uq_alerts_user_dedupe (user_id, dedupe_key),
    INDEX idx_alerts_user_created_at (user_id, created_at)
);
