ALTER TABLE wallets
    ADD COLUMN is_locked BOOLEAN NOT NULL DEFAULT 0 AFTER opening_balance;

UPDATE wallets
SET is_locked = 1
WHERE name = 'Main';
