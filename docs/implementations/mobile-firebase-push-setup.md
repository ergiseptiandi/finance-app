# Mobile Firebase Push Setup

This document explains the mobile-side setup required for push notifications from the backend.

## Goal

- register the mobile app with Firebase
- obtain a device token from Firebase
- send the token to the backend
- receive push notifications from the backend

## High-Level Flow

Current flow:

1. The mobile app asks the user for notification permission.
2. The mobile app receives an **FCM device token** from Firebase.
3. The mobile app sends the token to the backend through `PATCH /v1/notifications/settings`.
4. The backend stores the token in the `push_token` field.
5. The backend notification worker runs the reminder schedule.
6. When notification conditions are met, the backend sends a push message to the stored token.

## Final Contract To Align On

### Backend fields used by the app

- `enabled` for the global notification toggle
- `push_token` for the FCM device token
- `daily_expense_reminder_enabled`
- `daily_expense_reminder_time`
- `debt_payment_reminder_enabled`
- `debt_payment_reminder_time`
- `debt_payment_reminder_days_before`
- `salary_reminder_enabled`
- `salary_reminder_time`
- `salary_reminder_days_before`
- `salary_day`

Note: in the mobile app, a local field may be named `notification_enabled`, but the backend still expects `enabled`.

### Routes sent by the backend

- `daily_expense_input` -> `/activity`
- `debt_payment` -> `/debts`
- `salary_reminder` -> `/transactions?type=income`

The mobile app should read `data.kind`, `data.type`, and `data.route` from the payload or backend inbox.

## What Needs To Be Prepared In Firebase

Firebase project setup alone is not enough. The mobile app must also be registered.

### 1. Add a New App

In Firebase Console:

- click `Add app`
- choose the target platform:
  - `Android`
  - `iOS`

### 2. Enter App Identity

Use the correct identifier for each platform:

- Android: `package name`
- iOS: `bundle identifier`

The identifier must match the mobile app configuration.

### 3. Download Configuration Files

After registration, Firebase provides configuration files:

- Android: `google-services.json`
- iOS: `GoogleService-Info.plist`

### 4. Add The Files To The Mobile Project

Place the configuration files in the mobile project according to the platform documentation.

## Expo Notes

If the mobile project uses Expo:

- push notification support usually requires a **development build** or an **EAS build**
- Expo Go is generally not enough for native production push setup
- if this backend is used as-is, the mobile app must use an **FCM device token**
- this backend does **not** use Expo Push Token as the primary format

If the team wants to keep Expo Push Token, the backend architecture must be changed to send through the Expo Push API instead of Firebase directly.

## What The Mobile App Sends To The Backend

The mobile app should send the token to the following endpoint:

```http
PATCH /v1/notifications/settings
```

Minimal body:

```json
{
  "push_token": "fcm-device-token"
}
```

Full example:

```json
{
  "enabled": true,
  "daily_expense_reminder_enabled": true,
  "daily_expense_reminder_time": "20:00",
  "debt_payment_reminder_enabled": true,
  "debt_payment_reminder_time": "09:00",
  "debt_payment_reminder_days_before": 3,
  "salary_reminder_enabled": true,
  "salary_reminder_time": "08:00",
  "salary_reminder_days_before": 1,
  "salary_day": 25,
  "push_token": "fcm-device-token"
}
```

## Mobile Checklist

- request notification permission from the user
- get the FCM token from Firebase
- send the token to the backend after login or when the app is ready
- refresh the token if Firebase changes the device token
- handle notifications while the app is in the foreground
- handle notification taps while the app is in the background
- open the correct screen based on the notification type
- route taps using `data.route`
- show an unread badge or notification inbox from the backend

## Notification Types Supported By The Backend

- `daily_expense_input`
- `debt_payment`
- `salary_reminder`

Recommended UI behavior:

- `daily_expense_input` -> open the expense input screen
- `debt_payment` -> open the debt detail or payment screen
- `salary_reminder` -> open the income / salary input screen

The backend sends FCM payloads with `notification` + `data`, plus the Android channel `finance-go-default` and high priority.

`salary_day` is the monthly salary day selected by the user. If the month is shorter, the backend uses the last day of that month.

## Related Backend Environment Variables

The backend reads the following configuration values:

- `APP_MODE`
- `NOTIFICATION_CRON_SCHEDULE`
- `FIREBASE_PROJECT_ID`
- `FIREBASE_CREDENTIALS_PATH`
- `FIREBASE_CREDENTIALS_JSON`

If Firebase credentials are not provided, the backend can still store reminder history, but push notifications will not be sent.

The `google-services.json` file belongs in the mobile Android project. The backend requires a **service account key JSON** from Firebase Admin SDK.

## Implementation Notes

- the backend uses Firebase Admin SDK / ADC
- the backend runs notification scheduling through a worker service
- the scheduler can also run as a separate service in Docker Compose

## Summary

In short:

- register the mobile app with Firebase
- obtain the FCM token from the mobile app
- send the token to the backend
- let the backend store the token and send pushes
