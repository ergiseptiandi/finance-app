# Authentication API Documentation

This document outlines the API endpoints available under `/v1/auth` for user management and authentication.

> [!IMPORTANT]
> All endpoints protected by authentication require the `Authorization` header formatted as:
> `Authorization: Bearer <access_token>`

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
*(Same as Register response)*

---

## 3. Refresh Token

Issues a new `access_token` and rolls the `refresh_token` before the background session expires. Mobile apps should call this silently in the background when getting a 401 Unauthorized before forcing a logout.

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

### Success Response (200 OK)
*(Same as Register/Login response containing the new rotated token pair)*

---

## 4. Logout

Revokes the refresh token effectively logging out the device.

- **URL:** `/v1/auth/logout`
- **Method:** `POST`
- **Auth Required:** No

### Request Body
```json
{
  "refresh_token": "abc123hex..."
}
```

### Success Response (200 OK)
```json
{
  "status": "logged_out"
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
  "user": {
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
  "user": {
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
  "status": "password_changed"
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
  "status": "email_sent", 
  "reset_token": "random_token_string..." 
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
  "status": "password_reset"
}
```
