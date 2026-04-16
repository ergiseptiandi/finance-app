# Alerts API Documentation

This module handles smart alert generation and alert history.

**Base URL**: `/v1/alerts`  
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> Semua success response memakai envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> Semua error response memakai:
> `{ "Status": "...", "Message": "..." }`

---

## 1. List Alert History
**GET** `/v1/alerts`

Returns alert history ordered from newest to oldest.

Optional query params:
- `type=expense_threshold`
- `read=true|false`

---

## 2. Evaluate Smart Alerts
**POST** `/v1/alerts/evaluate`

Generates alerts based on current financial data and stores them in history.

Request body:
```json
{
  "expense_threshold_percent": 80,
  "daily_spike_multiplier": 1.5
}
```

Default values:
- `expense_threshold_percent` = `80`
- `daily_spike_multiplier` = `1.5`

---

## 3. Mark Alert As Read
**PATCH** `/v1/alerts/{id}/read`

Marks one alert as read.
