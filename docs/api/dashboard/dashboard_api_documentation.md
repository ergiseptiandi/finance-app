# Dashboard & Analytics API Documentation

This module handles dashboard summary cards and spending analytics.

**Base URL**: `/v1/dashboard`  
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> Semua success response sekarang memakai envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> Semua error response memakai:
> `{ "Status": "...", "Message": "..." }`

---

## 1. Get Dashboard Summary
**GET** `/v1/dashboard/summary`

Returns current balance plus income and expense totals for the selected period.

Query params:
- `month=YYYY-MM` to load one month
- `start_date=YYYY-MM-DD` and `end_date=YYYY-MM-DD` for a custom range

Rules:
- If no date filter is sent, the API defaults to the current month.
- `month` cannot be combined with `start_date` or `end_date`.
- Custom ranges are limited to 3 months.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "total_balance": 15000000,
    "monthly_income": 12000000,
    "monthly_expense": 3000000
  }
}
```

---

## 2. Get Daily Spending Data
**GET** `/v1/dashboard/daily-spending`

Returns daily expense data for the selected period, including transaction expenses and debt payments.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": [
    {
      "date": "2026-04-01",
      "amount": 0
    },
    {
      "date": "2026-04-02",
      "amount": 50000
    }
  ]
}
```

---

## 3. Get Monthly Spending Data
**GET** `/v1/dashboard/monthly-spending`

Returns monthly expense data grouped by month for the selected period.

---

## 4. Get Comparison
**GET** `/v1/dashboard/comparison`

Returns expense comparison for today vs yesterday and this month vs last month.
