# Category API Documentation

This module handles user-specific transaction categories.

**Base URL**: `/v1/categories`  
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> All success responses now use this envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> All error responses use:
> `{ "Status": "...", "Message": "..." }`

---

## 1. Create Category
**POST** `/v1/categories`

Create a new category for income or expense transactions owned by the authenticated user.

**Request Body (JSON)**:
```json
{
  "name": "Salary",
  "type": "income"
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
    "name": "Salary",
    "type": "income",
    "created_at": "2026-04-16T14:00:00Z",
    "updated_at": "2026-04-16T14:00:00Z"
  }
}
```

---

## 2. List Categories
**GET** `/v1/categories`

Retrieve all categories owned by the authenticated user.

**Query Parameters**:
- `type` (optional, enum): `income` or `expense`

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": [
    {
      "id": 1,
      "name": "Salary",
      "type": "income",
      "created_at": "2026-04-16T14:00:00Z",
      "updated_at": "2026-04-16T14:00:00Z"
    },
    {
      "id": 2,
      "name": "Food",
      "type": "expense",
      "created_at": "2026-04-16T14:00:00Z",
      "updated_at": "2026-04-16T14:00:00Z"
    }
  ]
}
```

---

## 3. Update Category
**PATCH** `/v1/categories/{id}`

Update an existing category by ID. The category must belong to the authenticated user.

**Request Body (JSON)**:
*Send only the fields you want to update.*
```json
{
  "name": "Monthly Salary",
  "type": "income"
}
```

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Update",
  "Data": {
    "id": 1,
    "name": "Monthly Salary",
    "type": "income",
    "created_at": "2026-04-16T14:00:00Z",
    "updated_at": "2026-04-16T15:00:00Z"
  }
}
```

---

## 4. Delete Category
**DELETE** `/v1/categories/{id}`

Delete a category by ID. The category must belong to the authenticated user.

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

## 5. Default Categories Seed

The application seeds default category master data during initialization.

The seed command populates the bootstrap user's default categories:
- `income`
- `expense`

Run migration and seed with:
```powershell
go run ./cmd/migrate up
go run ./cmd/seed
```
