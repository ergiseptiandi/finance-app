# Reports API Documentation

This module handles financial report endpoints for charts, summaries, and deeper spending insights.

**Base URL**: `/v1/reports`  
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> All success responses use this envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> All error responses use:
> `{ "Status": "...", "Message": "..." }`

---

## 1. Available Endpoints

- `GET /v1/reports/expense-by-category`
- `GET /v1/reports/spending-trends`
- `GET /v1/reports/highest-spending-category`
- `GET /v1/reports/average-daily-spending`
- `GET /v1/reports/remaining-balance`

---

## 2. Shared Filter Standard

All report endpoints accept the same period filter shape.

### Filter Modes

Use one of these modes only:

| Mode | Query Params | Description |
| --- | --- | --- |
| `month` | `month=YYYY-MM` | One full month |
| `year` | `year=YYYY` | One full year |
| `custom` | `start_date=YYYY-MM-DD&end_date=YYYY-MM-DD` | Inclusive custom date range |

### Filter Rules

- Use **only one** mode per request.
- `month` cannot be combined with `year`, `start_date`, or `end_date`.
- `year` cannot be combined with `month`, `start_date`, or `end_date`.
- `start_date` and `end_date` must be sent together.
- Custom range is inclusive.
- Custom range can span a maximum of **1 year**.
- `end_date` must be greater than or equal to `start_date`.

### Default Behavior

- `expense-by-category`, `highest-spending-category`, and `average-daily-spending` default to the current month if no filter is sent.
- `spending-trends` defaults to the latest 12 months if no filter is sent.
- `remaining-balance` defaults to all-time totals if no filter is sent.

---

## 3. Expense by Category
**GET** `/v1/reports/expense-by-category`

Returns spending grouped by category for the selected period.

### Response

```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "period": {
      "mode": "month",
      "month": "2026-04",
      "start_date": "2026-04-01",
      "end_date": "2026-04-30"
    },
    "summary": {
      "total_expense": 4000000,
      "category_count": 2,
      "top_category": "Food"
    },
    "items": [
      {
        "category": "Food",
        "amount": 3000000,
        "percentage": 75,
        "transaction_count": 6
      },
      {
        "category": "Transport",
        "amount": 1000000,
        "percentage": 25,
        "transaction_count": 4
      }
    ]
  }
}
```

### Notes

- `percentage` is calculated against the total expense in the selected period.
- `transaction_count` includes transaction rows plus debt payment rows that are aggregated into the category.

---

## 4. Spending Trends
**GET** `/v1/reports/spending-trends`

Returns spending movement over time.

### Grouping Rules

- `month` filter -> daily points
- `year` filter -> monthly points
- `custom` filter -> daily points for shorter ranges, monthly points for longer ranges
- no filter -> latest 12 months grouped by month

### Response

```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "period": {
      "mode": "year",
      "year": 2026,
      "start_date": "2026-01-01",
      "end_date": "2026-12-31"
    },
    "group_by": "month",
    "items": [
      {
        "period": "2026-01",
        "income": 5000000,
        "expense": 1000000,
        "net_cashflow": 4000000
      },
      {
        "period": "2026-02",
        "income": 0,
        "expense": 0,
        "net_cashflow": 0
      }
    ]
  }
}
```

### Notes

- `income` comes from income transactions only.
- `expense` combines expense transactions and debt payments.
- `net_cashflow` = `income - expense`

---

## 5. Highest Spending Category
**GET** `/v1/reports/highest-spending-category`

Returns the highest spending category for the selected period.

### Response

```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "period": {
      "mode": "month",
      "month": "2026-04",
      "start_date": "2026-04-01",
      "end_date": "2026-04-30"
    },
    "category": "Food",
    "amount": 3000000,
    "percentage": 75,
    "transaction_count": 6
  }
}
```

---

## 6. Average Daily Spending
**GET** `/v1/reports/average-daily-spending`

Returns average spending per day for the selected period.

### Response

```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "period": {
      "mode": "month",
      "month": "2026-04",
      "start_date": "2026-04-01",
      "end_date": "2026-04-30"
    },
    "total_expense": 200000,
    "days_count": 30,
    "average_daily_spending": 6666.67,
    "highest_daily_spending": 150000,
    "lowest_daily_spending": 0
  }
}
```

---

## 7. Remaining Balance
**GET** `/v1/reports/remaining-balance`

Returns income, expense, and remaining balance.

### Default Behavior

- Without filters, this endpoint returns all-time totals and uses the balance provider when available.
- With filters, it returns totals for the selected period and computes `remaining_balance` from that period only.

### Response

```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "period": {
      "mode": "all-time"
    },
    "total_income": 12000000,
    "total_expense": 3000000,
    "remaining_balance": 20000000,
    "savings_rate": 75,
    "expense_ratio": 25
  }
}
```

---

## 8. Error Cases

The API should return `400 Bad Request` for:

- invalid `month` format
- invalid `year` format
- invalid `start_date` or `end_date` format
- `month` combined with `year` or custom range
- `year` combined with `month` or custom range
- only one of `start_date` or `end_date` is sent
- `end_date` earlier than `start_date`
- custom date range longer than 1 year
