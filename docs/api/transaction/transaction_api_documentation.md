# Transaction API Documentation

This module handles recording and fetching financial transactions.

**Base URL**: `/v1/transactions`
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> Semua success response sekarang memakai envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> Semua error response memakai:
> `{ "Status": "...", "Message": "..." }`

---

## 1. Create Transaction
**POST** `/v1/transactions`

Create a new income or expense transaction.

**Request Body (JSON)**:
```json
{
  "type": "income",      
  "category": "Salary",
  "amount": 5500.00,
  "date": "2026-04-16T15:00:00Z",
  "description": "Monthly salary"
}
```
*Note: `type` must be either `income` or `expense`.*

**Response (201 Created)**:
```json
{
  "Status": "201",
  "Message": "Success Create",
  "Data": {
    "id": 1,
    "user_id": 1,
    "type": "income",
    "category": "Salary",
    "amount": 5500.00,
    "date": "2026-04-16T15:00:00Z",
    "description": "Monthly salary",
    "created_at": "2026-04-16T14:00:00Z",
    "updated_at": "2026-04-16T14:00:00Z"
  }
}
```

---

## 2. List Transactions (with Pagination and Filters)
**GET** `/v1/transactions`

Retrieve transactions tied to the authenticated user.

**Query Parameters**:
- `start_date` (optional, format: `YYYY-MM-DD`): Filter by starting date.
- `end_date` (optional, format: `YYYY-MM-DD`): Filter by ending date.
- `category` (optional, string): Exact match on category name.
- `type` (optional, enum): `income` or `expense`.
- `page` (optional, integer): Defaults to 1.
- `per_page` (optional, integer): Defaults to 10.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "data": [
      {
        "id": 1,
        "user_id": 1,
        "type": "income",
        "category": "Salary",
        "amount": 5500.00,
        "date": "2026-04-16T15:00:00Z",
        "description": "Monthly salary",
        "created_at": "2026-04-16T14:00:00Z",
        "updated_at": "2026-04-16T14:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "per_page": 10,
    "total_pages": 1
  }
}
```

---

## 3. Financial Summary
**GET** `/v1/transactions/summary`

Get aggregated metrics for the user's account.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "total_income": 5500.00,
    "total_expense": 200.00,
    "balance": 5300.00
  }
}
```

---

## 4. Get Transaction Detail
**GET** `/v1/transactions/{id}`

Retrieve full details of a specific transaction by its ID.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "id": 1,
    "user_id": 1,
    "type": "income",
    "category": "Salary",
    "amount": 5500.00,
    "date": "2026-04-16T15:00:00Z",
    "description": "Monthly salary",
    "created_at": "2026-04-16T14:00:00Z",
    "updated_at": "2026-04-16T14:00:00Z"
  }
}
```

**Error Cases**:
- `404 Not Found`: If the transaction doesn't exist or is owned by another user.

---

## 5. Update Transaction
**PATCH** `/v1/transactions/{id}`

Partially update a transaction matching the given ID.

**Request Body (JSON)**:
*Send only the fields you wish to update.*
```json
{
  "amount": 6000.00,
  "description": "Updated bonus salary"
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
    "type": "income",
    "category": "Salary",
    "amount": 6000.00,
    "date": "2026-04-16T15:00:00Z",
    "description": "Updated bonus salary",
    "created_at": "2026-04-16T14:00:00Z",
    "updated_at": "2026-04-16T16:00:00Z"
  }
}
```

---

## 6. Delete Transaction
**DELETE** `/v1/transactions/{id}`

Permanently delete an existing transaction.

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
