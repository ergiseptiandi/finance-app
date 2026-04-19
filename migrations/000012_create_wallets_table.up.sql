CREATE TABLE IF NOT EXISTS wallets (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    opening_balance DECIMAL(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uq_wallets_user_name (user_id, name),
    INDEX idx_wallets_user_created_at (user_id, created_at, id)
);

ALTER TABLE transactions
    ADD COLUMN wallet_id BIGINT NULL AFTER user_id,
    ADD INDEX idx_transactions_user_wallet_date (user_id, wallet_id, date, id);

ALTER TABLE debt_payments
    ADD COLUMN wallet_id BIGINT NULL AFTER debt_id,
    ADD INDEX idx_debt_payments_wallet_payment_date (wallet_id, payment_date, id);

CREATE TABLE IF NOT EXISTS wallet_transfers (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    from_wallet_id BIGINT NOT NULL,
    to_wallet_id BIGINT NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    note TEXT,
    transfer_date DATETIME NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (from_wallet_id) REFERENCES wallets(id) ON DELETE RESTRICT,
    FOREIGN KEY (to_wallet_id) REFERENCES wallets(id) ON DELETE RESTRICT,
    INDEX idx_wallet_transfers_user_date (user_id, transfer_date, id),
    INDEX idx_wallet_transfers_from_wallet (from_wallet_id, transfer_date, id),
    INDEX idx_wallet_transfers_to_wallet (to_wallet_id, transfer_date, id)
);

INSERT INTO wallets (user_id, name, opening_balance)
SELECT u.id, 'Main', 0
FROM users u
WHERE NOT EXISTS (
    SELECT 1
    FROM wallets w
    WHERE w.user_id = u.id AND w.name = 'Main'
);

UPDATE transactions t
JOIN wallets w ON w.user_id = t.user_id AND w.name = 'Main'
SET t.wallet_id = w.id
WHERE t.wallet_id IS NULL;

UPDATE debt_payments p
JOIN debts d ON d.id = p.debt_id
JOIN wallets w ON w.user_id = d.user_id AND w.name = 'Main'
SET p.wallet_id = w.id
WHERE p.wallet_id IS NULL;

ALTER TABLE transactions
    MODIFY COLUMN wallet_id BIGINT NOT NULL,
    ADD CONSTRAINT fk_transactions_wallet FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE RESTRICT;

ALTER TABLE debt_payments
    MODIFY COLUMN wallet_id BIGINT NOT NULL,
    ADD CONSTRAINT fk_debt_payments_wallet FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE RESTRICT;
