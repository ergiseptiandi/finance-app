# Budget Goals API Documentation

This module handles monthly category budget goals and progress tracking.

**Base URL**: `/v1/budgets/category-goals`  
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> All success responses now use this envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> All error responses use:
> `{ "Status": "...", "Message": "..." }`

Budget goals can only be created for `expense` categories.

---

## 1. List Budget Goals
**GET** `/v1/budgets/category-goals`

Returns all budget goals owned by the authenticated user. The response also includes monthly spending summary for the selected period.

Query params:
- `month=YYYY-MM` to load one month

If `month` is omitted, the backend defaults to the current month.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "summary": {
      "monthly_budget": 5000000,
      "spent": 3200000,
      "remaining": 1800000,
      "usage_rate": 64,
      "over_budget_amount": 0,
      "is_over_budget": false
    },
    "items": [
      {
        "id": 1,
        "user_id": 1,
        "category_id": 12,
        "category_name": "Food",
        "category_type": "expense",
        "monthly_amount": 2000000,
        "created_at": "2026-04-16T14:00:00Z",
        "updated_at": "2026-04-16T14:00:00Z",
        "current_amount": 1200000,
        "remaining_amount": 800000,
        "progress_percentage": 60,
        "status": "under_budget"
      }
    ]
  }
}
```

Notes:
- `summary` is the aggregate of all active budget goals.
- `status` can be `under_budget`, `on_track`, `over_budget`, or `inactive`.
- `progress_percentage` is computed from current spending vs monthly target.

---

## 2. Create Budget Goal
**POST** `/v1/budgets/category-goals`

Create a new monthly budget goal for an expense category.

**Request Body (JSON)**:
```json
{
  "category_id": 12,
  "monthly_amount": 2000000
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
    "category_id": 12,
    "category_name": "Food",
    "category_type": "expense",
    "monthly_amount": 2000000,
    "created_at": "2026-04-16T14:00:00Z",
    "updated_at": "2026-04-16T14:00:00Z"
  }
}
```

---

## 3. Update Budget Goal
**PATCH** `/v1/budgets/category-goals/{id}`

Update the category or monthly amount of an existing budget goal.

**Request Body (JSON)**:
```json
{
  "category_id": 13,
  "monthly_amount": 2500000
}
```

Both fields are optional.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Update",
  "Data": {
    "id": 1,
    "user_id": 1,
    "category_id": 13,
    "category_name": "Transport",
    "category_type": "expense",
    "monthly_amount": 2500000,
    "created_at": "2026-04-16T14:00:00Z",
    "updated_at": "2026-04-16T15:00:00Z"
  }
}
```

---

## 4. Delete Budget Goal
**DELETE** `/v1/budgets/category-goals/{id}`

Delete a budget goal by ID.

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
