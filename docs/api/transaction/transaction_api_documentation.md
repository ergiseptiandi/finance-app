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
- `month` (optional, format: `YYYY-MM`): Filter 1 full month. Example: `2026-04`.
- `start_date` (optional, format: `YYYY-MM-DD`): Custom range start date. Must be sent together with `end_date`.
- `end_date` (optional, format: `YYYY-MM-DD`): Custom range end date. Must be sent together with `start_date`.
- `category` (optional, string): Exact match on category name.
- `type` (optional, enum): `income` or `expense`.
- `page` (optional, integer): Defaults to 1.
- `per_page` (optional, integer): Defaults to 10.

**Filter Rules**:
- Use **either** `month` **or** `start_date` + `end_date`.
- `month` cannot be combined with `start_date` or `end_date`.
- If `month`, `start_date`, and `end_date` are not sent, the API defaults to the current month.
- Custom date range is inclusive and can span a maximum of 2 months.
- `end_date` must be greater than or equal to `start_date`.

**Examples**:

Filter by month:
```http
GET /v1/transactions?month=2026-04&page=1&per_page=10
```

Without date filter, defaults to current month:
```http
GET /v1/transactions?type=expense
```

Filter by custom range:
```http
GET /v1/transactions?start_date=2026-04-01&end_date=2026-05-31&type=expense
```

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

**Error Cases**:
- `400 Bad Request`: If `month` format is invalid.
- `400 Bad Request`: If `start_date` or `end_date` format is invalid.
- `400 Bad Request`: If `month` is combined with `start_date` or `end_date`.
- `400 Bad Request`: If only one of `start_date` or `end_date` is sent.
- `400 Bad Request`: If custom date range is more than 2 months.
- `400 Bad Request`: If `end_date` is earlier than `start_date`.

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
