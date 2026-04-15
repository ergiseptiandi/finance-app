# Finance Backend

Go backend scaffold dengan struktur yang lebih rapi untuk mobile authentication:

- `repository` untuk akses MySQL
- `service` untuk business logic auth
- `httpapi` untuk transport HTTP
- `chi router` untuk grouping route modular
- access token JWT
- refresh token opaque yang di-hash di database

Struktur router sekarang:

- `internal/httpapi/router.go` untuk root router dan versioning `/v1`
- `internal/httpapi/auth/routes.go` untuk auth module routes
- `internal/httpapi/middleware/auth.go` untuk auth middleware

## Requirements

- Go `1.26.2`

## Run

```powershell
go run ./cmd/api
```

The server starts on `http://localhost:8080`.

## Endpoints

- `GET /`
- `GET /health`
- `GET /routes`
- `GET /openapi.json`
- `GET /docs`

## Auth Configuration

Project ini sekarang mengasumsikan auth aktif dan MySQL tersedia. Saat startup, app akan:

- konek ke MySQL
- membuat table `users` dan `refresh_tokens` jika belum ada
- upsert 1 akun bootstrap dari environment
- otomatis membaca file `.env` jika ada di root project

Isi environment seperti ini:

Example PowerShell:

```powershell
$env:DB_ENABLED="true"
$env:DB_HOST="127.0.0.1"
$env:DB_PORT="3306"
$env:DB_USER="root"
$env:DB_PASSWORD="secret"
$env:DB_NAME="finance_db"
$env:AUTH_BOOTSTRAP_NAME="Owner"
$env:AUTH_BOOTSTRAP_EMAIL="owner@example.com"
$env:AUTH_BOOTSTRAP_PASSWORD="supersecret123"
$env:AUTH_JWT_SECRET="change-this-to-a-long-random-secret"
go run ./cmd/api
```

`AUTH_BOOTSTRAP_PASSWORD` akan di-hash ke bcrypt lalu disimpan ke table `users`.

## Auth Endpoints

- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout`
- `GET /v1/auth/me`

Contoh login:

```json
{
  "email": "owner@example.com",
  "password": "supersecret123",
  "device_name": "android-phone"
}
```

Respons login berisi:

- `access_token`
- `access_token_expires_at`
- `refresh_token`
- `refresh_token_expires_at`
- data `user`

## API Explorer

- `GET /routes` menampilkan daftar route aktif dalam JSON
- `GET /openapi.json` menampilkan OpenAPI spec
- `GET /docs` membuka Swagger UI

## Test

```powershell
go test ./...
```

## Optional

Use a custom port:

```powershell
$env:PORT="9000"
go run ./cmd/api
```
