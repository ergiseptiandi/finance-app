ALTER TABLE debt_payments
    DROP FOREIGN KEY fk_debt_payments_wallet;

ALTER TABLE debt_payments
    DROP INDEX idx_debt_payments_wallet_payment_date,
    DROP COLUMN wallet_id;

ALTER TABLE transactions
    DROP FOREIGN KEY fk_transactions_wallet;

ALTER TABLE transactions
    DROP INDEX idx_transactions_user_wallet_date,
    DROP COLUMN wallet_id;

DROP TABLE IF EXISTS wallet_transfers;
DROP TABLE IF EXISTS wallets;
