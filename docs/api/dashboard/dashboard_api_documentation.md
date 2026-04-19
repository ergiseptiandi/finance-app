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
The summary response also includes core debt health metrics so you can show a more useful home dashboard without making extra requests.

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
    "monthly_expense": 3000000,
    "net_cashflow": 9000000,
    "savings_rate": 75,
    "expense_ratio": 25,
    "debt": {
      "total_debt": 10000000,
      "paid_debt": 4000000,
      "remaining_debt": 6000000,
      "total_debt_count": 2,
      "active_debt_count": 1,
      "overdue_debt_count": 1,
      "paid_installments": 3,
      "overdue_installments": 1,
      "upcoming_due_amount": 1500000,
      "upcoming_due_installments": 2,
      "debt_to_income_ratio": 50,
      "debt_to_balance_ratio": 40,
      "completion_rate": 40
    }
  }
}
```

Notes:
- `net_cashflow` = `monthly_income - monthly_expense`
- `savings_rate` = `net_cashflow / monthly_income * 100`
- `expense_ratio` = `monthly_expense / monthly_income * 100`
- `debt_to_income_ratio` = `remaining_debt / monthly_income * 100`
- `debt_to_balance_ratio` = `remaining_debt / total_balance * 100` when total balance is positive
- `upcoming_due_amount` is the sum of unpaid installments inside the selected period

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

This endpoint is still focused on expense movement. Debt health is exposed in `/summary` so the home dashboard can stay compact.
