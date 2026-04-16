# Salary API Documentation

This module handles salary records and recurring salary schedule settings.

**Base URL**: `/v1/salaries`  
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> Semua success response sekarang memakai envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> Semua error response memakai:
> `{ "Status": "...", "Message": "..." }`

---

## 1. Create Salary Record
**POST** `/v1/salaries`

Create a salary history record.

**Request Body (JSON)**:
```json
{
  "amount": 5000000,
  "paid_at": "2026-04-16T00:00:00Z",
  "note": "April salary"
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
    "amount": 5000000,
    "paid_at": "2026-04-16T00:00:00Z",
    "note": "April salary",
    "created_at": "2026-04-16T10:00:00Z",
    "updated_at": "2026-04-16T10:00:00Z"
  }
}
```

---

## 2. Get Salary History
**GET** `/v1/salaries`

Retrieve all salary records ordered from newest to oldest.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": [
    {
      "id": 2,
      "user_id": 1,
      "amount": 5200000,
      "paid_at": "2026-05-16T00:00:00Z",
      "note": "May salary",
      "created_at": "2026-05-16T10:00:00Z",
      "updated_at": "2026-05-16T10:00:00Z"
    },
    {
      "id": 1,
      "user_id": 1,
      "amount": 5000000,
      "paid_at": "2026-04-16T00:00:00Z",
      "note": "April salary",
      "created_at": "2026-04-16T10:00:00Z",
      "updated_at": "2026-04-16T10:00:00Z"
    }
  ]
}
```

---

## 3. Get Current Salary
**GET** `/v1/salaries/current`

Returns the latest salary record plus the recurring salary day if it has been set.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "id": 2,
    "user_id": 1,
    "amount": 5200000,
    "paid_at": "2026-05-16T00:00:00Z",
    "note": "May salary",
    "salary_day": 25,
    "created_at": "2026-05-16T10:00:00Z",
    "updated_at": "2026-05-16T10:00:00Z"
  }
}
```

---

## 4. Update Salary
**PATCH** `/v1/salaries/{id}`

Update an existing salary record.

**Request Body (JSON)**:
```json
{
  "amount": 5500000,
  "paid_at": "2026-05-20T00:00:00Z",
  "note": "Updated salary"
}
```

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Update",
  "Data": {
    "id": 2,
    "user_id": 1,
    "amount": 5500000,
    "paid_at": "2026-05-20T00:00:00Z",
    "note": "Updated salary",
    "created_at": "2026-05-16T10:00:00Z",
    "updated_at": "2026-05-16T12:00:00Z"
  }
}
```

---

## 5. Delete Salary
**DELETE** `/v1/salaries/{id}`

Delete a salary record by ID.

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

## 6. Set Salary Date
**PATCH** `/v1/salaries/schedule`

Set the recurring monthly salary day.

**Request Body (JSON)**:
```json
{
  "salary_day": 25
}
```

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Set Salary Date",
  "Data": {
    "salary_day": 25,
    "created_at": "2026-04-16T10:00:00Z",
    "updated_at": "2026-04-16T10:00:00Z"
  }
}
```
