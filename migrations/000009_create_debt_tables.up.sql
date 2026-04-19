CREATE TABLE IF NOT EXISTS debts (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    total_amount DECIMAL(15, 2) NOT NULL,
    monthly_installment DECIMAL(15, 2) NOT NULL,
    due_date DATETIME NOT NULL,
    paid_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    remaining_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    status ENUM('pending', 'paid', 'overdue') NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_debts_user_due_date (user_id, due_date),
    INDEX idx_debts_user_status (user_id, status)
);

CREATE TABLE IF NOT EXISTS debt_installments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    debt_id BIGINT NOT NULL,
    installment_no INT NOT NULL,
    due_date DATETIME NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    status ENUM('pending', 'paid', 'overdue') NOT NULL DEFAULT 'pending',
    paid_at DATETIME NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE CASCADE,
    UNIQUE KEY uq_debt_installments_no (debt_id, installment_no),
    INDEX idx_debt_installments_status_due_date (status, due_date),
    INDEX idx_debt_installments_debt_status_no (debt_id, status, installment_no)
);

CREATE TABLE IF NOT EXISTS debt_payments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    debt_id BIGINT NOT NULL,
    installment_id BIGINT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    payment_date DATETIME NOT NULL,
    proof_image TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE CASCADE,
    FOREIGN KEY (installment_id) REFERENCES debt_installments(id) ON DELETE SET NULL,
    INDEX idx_debt_payments_debt_payment_date (debt_id, payment_date),
    INDEX idx_debt_payments_installment (installment_id)
);
