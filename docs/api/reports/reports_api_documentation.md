# Reports API Documentation

This module handles reporting endpoints for charts and key financial insights.

**Base URL**: `/v1/reports`  
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> Semua success response sekarang memakai envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> Semua error response memakai:
> `{ "Status": "...", "Message": "..." }`

---

## 1. Expense by Category
**GET** `/v1/reports/expense-by-category`

Pie chart data for expenses in the current month.

**Response (200 OK)**:
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": [
    {
      "category": "Food",
      "amount": 1500000,
      "percentage": 50
    },
    {
      "category": "Debt Payment",
      "amount": 1000000,
      "percentage": 33.33
    }
  ]
}
```

---

## 2. Spending Trends
**GET** `/v1/reports/spending-trends`

Line chart data for the last 12 months.

---

## 3. Highest Spending Category
**GET** `/v1/reports/highest-spending-category`

Returns the highest spending category for the current month.

---

## 4. Average Daily Spending
**GET** `/v1/reports/average-daily-spending`

Returns current month total expense, elapsed day count, and average daily spending.

---

## 5. Remaining Balance
**GET** `/v1/reports/remaining-balance`

Returns total income, total expense, and remaining balance.
