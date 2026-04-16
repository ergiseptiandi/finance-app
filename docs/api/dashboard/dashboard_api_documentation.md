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

Returns current balance plus monthly income and expense totals.

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

Returns daily expense data for the current month, including transaction expenses and debt payments.

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

Returns monthly expense data for the last 12 months.

---

## 4. Get Comparison
**GET** `/v1/dashboard/comparison`

Returns expense comparison for today vs yesterday and this month vs last month.

---

## 5. Get Expense Percentage vs Salary
**GET** `/v1/dashboard/expense-vs-salary`

Returns monthly expense as a percentage of the latest salary record.
