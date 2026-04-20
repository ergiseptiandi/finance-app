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

---

## 5. Proposed Dashboard Extensions

The following additions are recommended to make the home dashboard more informative without turning `/summary` into a heavy payload.

### A. Fields to Add to `/summary`

Recommended additions:
- `budget_summary`
  - `monthly_budget`
  - `spent`
  - `remaining`
  - `usage_rate`
  - `over_budget_amount`
  - `is_over_budget`
- `category_breakdown_preview`
  - compact list of top spending categories for the selected period
  - each item contains `category`, `amount`, and `percentage`
- `upcoming_bills`
  - `count`
  - `total_amount`
  - `next_due_date`
  - optional preview list of the nearest bills
- `top_merchants_preview`
  - compact list of top merchants or payees
  - each item contains `merchant_name`, `amount`, and `transaction_count`
- `alerts`
  - list of generated insights or warnings
  - each item contains `type`, `code`, `title`, and `message`
- `goals_progress`
  - compact list of active financial goals
  - each item contains `name`, `target_amount`, `current_amount`, `progress_percentage`, and `target_date`

Suggested `/summary` shape:
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
    "budget_summary": {
      "monthly_budget": 5000000,
      "spent": 3000000,
      "remaining": 2000000,
      "usage_rate": 60,
      "over_budget_amount": 0,
      "is_over_budget": false
    },
    "category_breakdown_preview": [
      {
        "category": "Food",
        "amount": 1200000,
        "percentage": 40
      }
    ],
    "upcoming_bills": {
      "count": 3,
      "total_amount": 1500000,
      "next_due_date": "2026-04-25"
    },
    "top_merchants_preview": [
      {
        "merchant_name": "Gojek",
        "amount": 400000,
        "transaction_count": 8
      }
    ],
    "alerts": [
      {
        "type": "warning",
        "code": "OVER_BUDGET",
        "title": "Budget hampir habis",
        "message": "Pengeluaran sudah mencapai 92% dari budget bulan ini"
      }
    ],
    "goals_progress": [
      {
        "name": "Dana Darurat",
        "target_amount": 20000000,
        "current_amount": 8000000,
        "progress_percentage": 40,
        "target_date": "2026-12-31"
      }
    ],
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

### B. Proposed New Dashboard Endpoints

These endpoints can stay small and focused. `/summary` can keep only preview data, while these endpoints return the full dataset.

#### 5.1 Budget vs Actual
**GET** `/v1/dashboard/budget-vs-actual`

Compares the configured monthly budget against actual spending for the selected period.
If `budget_amount` is not provided, the backend falls back to the selected period income as the budget baseline.

Suggested response fields:
- `budget_amount`
- `actual_spent`
- `remaining_budget`
- `usage_rate`
- `over_budget_amount`
- `status`

#### 5.2 Category Breakdown
**GET** `/v1/dashboard/category-breakdown`

Returns spending grouped by category, sorted from highest to lowest amount.
The breakdown includes both transaction expenses and debt payments.

Suggested response fields:
- `category`
- `amount`
- `percentage`
- `transaction_count`

This endpoint can reuse the same data model as `/v1/reports/expense-by-category` if you want a single source of truth.

#### 5.3 Upcoming Bills
**GET** `/v1/dashboard/upcoming-bills`

Returns unpaid or soon-to-due bills, including debt installments and recurring obligations.
Current implementation returns debt installments due within the lookahead window.

Suggested query params:
- `days=7|14|30` to control the lookahead window

Suggested response fields:
- `bill_name`
- `amount`
- `due_date`
- `status`
- `source_type` (`debt`, `subscription`, `utility`, etc.)

#### 5.4 Top Merchants
**GET** `/v1/dashboard/top-merchants`

Returns the merchants, vendors, or payees with the highest transaction volume in the selected period.
The backend derives the merchant label from transaction description or category, and uses debt name for debt payments.

Suggested response fields:
- `merchant_name`
- `amount`
- `transaction_count`
- `last_transaction_date`

#### 5.5 Alerts / Insights
**GET** `/v1/dashboard/insights`

Returns computed dashboard insights based on spending behavior, budget usage, debt health, and goals progress.
The current backend generates these as derived insights from existing transactions and debt data.

Suggested response fields:
- `type`
- `code`
- `title`
- `message`
- `severity`
- `change_value`

Example insight types:
- `budget_warning`
- `spending_spike`
- `overdue_bill`
- `low_savings_rate`
- `goal_at_risk`

#### 5.6 Goals Progress
**GET** `/v1/dashboard/goals-progress`

Returns the progress of active financial goals.
Until a dedicated goals module exists, the backend returns derived goals such as Emergency Fund and Debt Freedom.

Suggested response fields:
- `goal_name`
- `target_amount`
- `current_amount`
- `progress_percentage`
- `target_date`
- `status`

### C. Recommended Implementation Notes

- Keep `/summary` as the homepage payload and return only previews for heavy sections.
- Use separate endpoints for detail views so the frontend can lazy-load expensive data.
- Reuse shared aggregation logic between `/dashboard` and `/reports` to avoid duplicated business rules.
- If budgets and goals are not configured, return empty arrays or `null` objects instead of forcing placeholder values.
