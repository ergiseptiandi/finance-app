# Debt API Documentation

This module handles debt management, payment proof uploads, and recurring installment tracking.

**Base URL**: `/v1/debts`  
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> Semua success response sekarang memakai envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> Semua error response memakai:
> `{ "Status": "...", "Message": "..." }`
>
> Untuk endpoint list/history (`GET`), jika data kosong tetap mengembalikan HTTP `200` dengan `Data: []`.

Uploaded proof images are served from the local upload directory under `/uploads/...`.

---

## 1. Create Debt
**POST** `/v1/debts`

Create a new debt and auto-generate installment schedule.

**Request Body (JSON)**:
```json
{
  "name": "Motorcycle Loan",
  "total_amount": 12000000,
  "monthly_installment": 1000000,
  "due_date": "2026-04-16T00:00:00Z"
}
```

**Response (201 Created)**:
```json
{
  "Status": "201",
  "Message": "Success Create",
  "Data": {
    "id": 1,
    "user_id": 1,
    "name": "Motorcycle Loan",
    "total_amount": 12000000,
    "monthly_installment": 1000000,
    "due_date": "2026-04-16T00:00:00Z",
    "paid_amount": 0,
    "remaining_amount": 12000000,
    "status": "pending",
    "paid_installments": 0,
    "unpaid_installments": 12,
    "overdue_installments": 0,
    "installments": [],
    "payments": []
  }
}
```

---

## 2. Get Debt List
**GET** `/v1/debts`

Return all debts for the current user.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": [
    {
      "id": 1,
      "user_id": 1,
      "name": "Motorcycle Loan",
      "total_amount": 12000000,
      "monthly_installment": 1000000,
      "due_date": "2026-04-16T00:00:00Z",
      "paid_amount": 1000000,
      "remaining_amount": 11000000,
      "status": "pending",
      "paid_installments": 1,
      "unpaid_installments": 11,
      "overdue_installments": 0
    }
  ]
}
```

**Response saat data kosong (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": []
}
```

---

## 3. Get Debt Detail
**GET** `/v1/debts/{id}`

Return one debt with installment and payment history.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "id": 1,
    "user_id": 1,
    "name": "Motorcycle Loan",
    "total_amount": 12000000,
    "monthly_installment": 1000000,
    "due_date": "2026-04-16T00:00:00Z",
    "paid_amount": 1000000,
    "remaining_amount": 11000000,
    "status": "pending",
    "paid_installments": 1,
    "unpaid_installments": 11,
    "overdue_installments": 0,
    "installments": [],
    "payments": []
  }
}
```

---

## 4. Update Debt
**PATCH** `/v1/debts/{id}`

Update debt metadata. If total amount, monthly installment, or due date changes before any payment exists, the installment schedule will be regenerated.

**Request Body (JSON)**:
```json
{
  "name": "Motorcycle Loan Revised",
  "due_date": "2026-05-16T00:00:00Z"
}
```

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Update",
  "Data": {
    "id": 1,
    "user_id": 1,
    "name": "Motorcycle Loan Revised",
    "total_amount": 12000000,
    "monthly_installment": 1000000,
    "due_date": "2026-05-16T00:00:00Z",
    "paid_amount": 1000000,
    "remaining_amount": 11000000,
    "status": "pending",
    "paid_installments": 1,
    "unpaid_installments": 11,
    "overdue_installments": 0,
    "installments": [],
    "payments": []
  }
}
```

---

## 5. Delete Debt
**DELETE** `/v1/debts/{id}`

Delete a debt and its related installments/payments.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Delete",
  "Data": {
    "status": "deleted"
  }
}
```

---

## 6. Create Payment
**POST** `/v1/debts/{id}/payments`

Create a payment record and upload proof image.

**Request Type**: `multipart/form-data`

**Form Fields**:
- `wallet_id` (optional, number): Wallet used to pay the debt. If omitted, the API uses the user's default wallet (`Main`).
- `amount` (required, number)
- `payment_date` (required, RFC3339 datetime)
- `proof_image` (required, file)

**Response (201 Created)**:
```json
{
  "Status": "201",
  "Message": "Success Create",
  "Data": {
    "id": 10,
    "debt_id": 1,
    "wallet_id": 2,
    "installment_id": 3,
    "amount": 1000000,
    "payment_date": "2026-04-16T00:00:00Z",
    "proof_image": "/uploads/debt-payments/1/1/123456789_receipt.jpg",
    "created_at": "2026-04-16T10:00:00Z",
    "updated_at": "2026-04-16T10:00:00Z"
  }
}
```

---

## 7. Update Payment
**PATCH** `/v1/debts/{id}/payments/{paymentId}`

Update payment metadata and optionally replace the proof image.

**Request Type**: `multipart/form-data`

**Form Fields**:
- `wallet_id` (optional, number)
- `amount` (optional, number)
- `payment_date` (optional, RFC3339 datetime)
- `proof_image` (optional, file)

---

## 8. Get Payment History
**GET** `/v1/debts/{id}/payments`

Return payment records for the debt.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": []
}
```

---

## 9. Get Installments
**GET** `/v1/debts/{id}/installments`

Return paid, pending, and overdue installments.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": []
}
```

---

## 10. Mark Installment as Paid
**PATCH** `/v1/debts/{id}/installments/{installmentId}/paid`

Mark one installment as paid manually.

**Request Body (JSON)**:
```json
{
  "paid_at": "2026-04-16T00:00:00Z"
}
```

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Update",
  "Data": {
    "id": 1,
    "debt_id": 1,
    "installment_no": 1,
    "due_date": "2026-04-16T00:00:00Z",
    "amount": 1000000,
    "status": "paid",
    "paid_at": "2026-04-16T00:00:00Z",
    "created_at": "2026-04-16T10:00:00Z",
    "updated_at": "2026-04-16T10:00:00Z"
  }
}
```
