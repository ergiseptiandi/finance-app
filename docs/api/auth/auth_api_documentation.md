# Authentication API Documentation

This document outlines the API endpoints available under `/v1/auth` for user management and authentication.

> [!IMPORTANT]
> All endpoints protected by authentication require the `Authorization` header formatted as:
> `Authorization: Bearer <access_token>`

> [!NOTE]
> All success responses now use this envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> All error responses use:
> `{ "Status": "...", "Message": "..." }`

## Authentication Model

This API does not use server-side HTTP sessions or cookies. Authentication is token-based:

- `access_token` is sent on every protected request using `Authorization: Bearer <access_token>`
- `refresh_token` is used only to call `/v1/auth/refresh`

## Token Lifetime and Session Behavior

Current default runtime values:

- `access_token` TTL: `15 minutes`
- `refresh_token` TTL: `30 days` (`720h`)

Important behavior:

- every successful login or register returns a new `access_token` and `refresh_token`
- every successful refresh returns a new `access_token` and also rotates the `refresh_token`
- after refresh succeeds, the previous `refresh_token` is revoked and can no longer be used
- clients must always replace the stored refresh token with the latest value returned by the API
- as long as the client keeps refreshing with the latest valid refresh token, the login session can continue beyond the original 30-day window because a new refresh token with a new expiry is issued on each refresh
- logout revokes the submitted refresh token; any already-issued access token may still work until its own expiry time unless the client discards it immediately

Recommended client behavior:

- use the `access_token` for normal API calls
- when the API returns `401` because the access token expired, call `/v1/auth/refresh`
- if refresh succeeds, store both the new `access_token` and the new `refresh_token`
- if refresh fails with `401`, force the user to log in again

---

## 1. Register a New User

Creates a new user account and immediately logs them in, returning an access token and refresh token bundle.

- **URL:** `/v1/auth/register`
- **Method:** `POST`
- **Auth Required:** No

### Request Body
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "supersecretpassword",
  "device_name": "iPhone 15 Pro" 
}
```
*(Note: `device_name` is optional but highly recommended to track refresh tokens in the database).*

### Success Response (201 Created)
```json
{
  "Status": "201",
  "Message": "Success Register",
  "Data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2026-04-16T10:00:00Z",
      "updated_at": "2026-04-16T10:00:00Z"
    },
    "token": {
      "access_token": "eyJhbG...",
      "access_token_expires_at": "2026-04-16T10:15:00Z",
      "refresh_token": "abc123hex...",
      "refresh_token_expires_at": "2026-05-16T10:00:00Z",
      "token_type": "Bearer"
    }
  }
}
```

---

## 2. Login

Authenticates an existing user and returns a token bundle.

- **URL:** `/v1/auth/login`
- **Method:** `POST`
- **Auth Required:** No

### Request Body
```json
{
  "email": "john@example.com",
  "password": "supersecretpassword",
  "device_name": "iPhone 15 Pro"
}
```

### Success Response (200 OK)
```json
{
  "Status": "200",
  "Message": "Success Login",
  "Data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2026-04-16T10:00:00Z",
      "updated_at": "2026-04-16T10:00:00Z"
    },
    "token": {
      "access_token": "eyJhbG...",
      "access_token_expires_at": "2026-04-16T10:15:00Z",
      "refresh_token": "abc123hex...",
      "refresh_token_expires_at": "2026-05-16T10:00:00Z",
      "token_type": "Bearer"
    }
  }
}
```

---

## 3. Refresh Token

Issues a new `access_token` and rotates the `refresh_token`. The old refresh token becomes invalid immediately after a successful refresh. Mobile apps should call this silently when the current access token expires before forcing a logout.

- **URL:** `/v1/auth/refresh`
- **Method:** `POST`
- **Auth Required:** No

### Request Body
```json
{
  "refresh_token": "abc123hex...",
  "device_name": "iPhone 15 Pro"
}
```

`device_name` is optional. If omitted, the API keeps the previous device name associated with the refresh token record.

### Success Response (200 OK)
```json
{
  "Status": "200",
  "Message": "Success Refresh",
  "Data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2026-04-16T10:00:00Z",
      "updated_at": "2026-04-16T10:00:00Z"
    },
    "token": {
      "access_token": "eyJhbG...",
      "access_token_expires_at": "2026-04-16T10:15:00Z",
      "refresh_token": "def456hex...",
      "refresh_token_expires_at": "2026-05-16T10:15:00Z",
      "token_type": "Bearer"
    }
  }
}
```

> [!IMPORTANT]
> After a successful refresh, the client must overwrite the previously stored `refresh_token`.

---

## 4. Logout

Revokes the submitted refresh token, effectively logging out that device/session chain.

- **URL:** `/v1/auth/logout`
- **Method:** `POST`
- **Auth Required:** No

### Request Body
```json
{
  "refresh_token": "abc123hex..."
}
```

> [!NOTE]
> Logout does not invalidate other refresh tokens belonging to other devices, and it does not retroactively revoke an access token that was already issued.

### Success Response (200 OK)
```json
{
  "Status": "200",
  "Message": "Success Logout",
  "Data": {
    "status": "logged_out"
  }
}
```

---

## 5. Get Current User Profile

Fetches the currently authenticated user's profile information.

- **URL:** `/v1/auth/me`
- **Method:** `GET`
- **Auth Required:** Yes (`Bearer <access_token>`)

### Success Response (200 OK)
```json
{
  "Status": "200",
  "Message": "Success Get",
  "Data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-04-16T10:00:00Z",
    "updated_at": "2026-04-16T10:00:00Z"
  }
}
```

---

## 6. Update Profile

Updates the logged-in user's name or email.

- **URL:** `/v1/auth/profile`
- **Method:** `PATCH`
- **Auth Required:** Yes (`Bearer <access_token>`)

### Request Body
```json
{
  "name": "Johnathan Doe",
  "email": "johnathan@example.com"
}
```
*(Both fields are optional; send only what you want to update).*

### Success Response (200 OK)
```json
{
  "Status": "200",
  "Message": "Success Update",
  "Data": {
    "id": 1,
    "name": "Johnathan Doe",
    "email": "johnathan@example.com",
    "created_at": "2026-04-16T10:00:00Z",
    "updated_at": "2026-04-16T10:30:00Z"
  }
}
```

---

## 7. Change Password

Updates the currently logged-in user's password.

- **URL:** `/v1/auth/password`
- **Method:** `PATCH`
- **Auth Required:** Yes (`Bearer <access_token>`)

### Request Body
```json
{
  "old_password": "supersecretpassword",
  "new_password": "new_supersecretpassword"
}
```

### Success Response (200 OK)
```json
{
  "Status": "200",
  "Message": "Success Change Password",
  "Data": {
    "status": "password_changed"
  }
}
```

---

## 8. Forgot Password

Requests a password reset token via email.

- **URL:** `/v1/auth/forgot-password`
- **Method:** `POST`
- **Auth Required:** No

### Request Body
```json
{
  "email": "john@example.com"
}
```

### Success Response (200 OK)
```json
{
  "Status": "200",
  "Message": "Success Forgot Password",
  "Data": {
    "status": "email_sent",
    "reset_token": "random_token_string..."
  }
}
```
> [!NOTE]
> During active development, `reset_token` is intentionally returned in this JSON payload so frontend/mobile developers don't have to check their inboxes every time they want to test the flow. In production, this field may be stripped out!

---

## 9. Reset Password

Resets a user's password using the token received via email (or from the `forgot-password` response).

- **URL:** `/v1/auth/reset-password`
- **Method:** `POST`
- **Auth Required:** No

### Request Body
```json
{
  "token": "random_token_string...",
  "new_password": "recovered_supersecretpassword"
}
```

### Success Response (200 OK)
```json
{
  "Status": "200",
  "Message": "Success Reset Password",
  "Data": {
    "status": "password_reset"
  }
}
```
