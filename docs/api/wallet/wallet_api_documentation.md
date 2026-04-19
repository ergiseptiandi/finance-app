# Wallet API Documentation

This module handles user wallets/tabungan and internal wallet transfers.

**Base URL**: `/v1/wallets` and `/v1/wallet-transfers`  
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> Semua success response memakai envelope:
> `{ "Status": "...", "Message": "...", "Data": ... }`
>
> Semua error response memakai:
> `{ "Status": "...", "Message": "..." }`

---

## 1. Create Wallet
**POST** `/v1/wallets`

Create a new wallet for the current user.

**Request Body (JSON)**:
```json
{
  "name": "Seabank",
  "opening_balance": 200000
}
```

**Response (201 Created)**:
```json
{
  "Status": "201",
  "Message": "Success Create",
  "Data": {
    "id": 1,
    "name": "Seabank",
    "opening_balance": 200000,
    "balance": 200000,
    "created_at": "2026-04-19T09:00:00Z",
    "updated_at": "2026-04-19T09:00:00Z"
  }
}
```

---

## 2. List Wallets
**GET** `/v1/wallets`

Return all wallets owned by the current user. Each wallet includes its computed balance.

---

## 3. Wallet Summary
**GET** `/v1/wallets/summary`

Return total balance across all wallets and the wallet list.

---

## 4. Update Wallet
**PATCH** `/v1/wallets/{id}`

Update wallet name or opening balance.

---

## 5. Delete Wallet
**DELETE** `/v1/wallets/{id}`

Delete a wallet if it is not referenced by transaction history.

---

## 6. Create Transfer
**POST** `/v1/wallet-transfers`

Move money between wallets. This does not change total balance.

**Request Body (JSON)**:
```json
{
  "from_wallet_id": 1,
  "to_wallet_id": 2,
  "amount": 50000,
  "note": "Top up GoPay",
  "transfer_date": "2026-04-19T09:00:00Z"
}
```

---

## 7. List Transfers
**GET** `/v1/wallet-transfers`

Return transfer history for the current user.

---

## Default Wallet Behavior
- If a transaction or debt payment is created without `wallet_id`, the API uses the user's default wallet named `Main`.
- If `Main` does not exist yet, it is created automatically.
