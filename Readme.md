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

## Workflow

Jalankan perintah ini dari root project:

```powershell
go run ./cmd/migrate up
go run ./cmd/seed
go run ./cmd/api
```

## Endpoints

- `GET /`
- `GET /health`
- `GET /routes`
- `GET /openapi.json`
- `GET /docs`

## Auth Configuration

Project ini sekarang mengasumsikan auth aktif dan MySQL tersedia. Server API tidak lagi membuat table atau data awal saat startup. Alurnya:

- konek ke MySQL
- otomatis membaca file `.env` jika ada di root project
- migration membuat table `users` dan `refresh_tokens`
- seed membuat atau update 1 akun awal

Isi environment seperti ini:

Example PowerShell:

```powershell
$env:DB_ENABLED="true"
$env:DB_HOST="127.0.0.1"
$env:DB_PORT="3306"
$env:DB_USER="root"
$env:DB_PASSWORD="secret"
$env:DB_NAME="finance_db"
$env:AUTH_JWT_SECRET="change-this-to-a-long-random-secret"
$env:SEED_USER_NAME="Owner"
$env:SEED_USER_EMAIL="owner@example.com"
$env:SEED_USER_PASSWORD="supersecret123"
go run ./cmd/api
```

`SEED_USER_PASSWORD` akan di-hash ke bcrypt lalu disimpan ke table `users`.

## Database Migration

Schema database sekarang ada di folder [migrations](d:/freelance/finance-backend/migrations:1).

Naikkan semua migration:

```powershell
go run ./cmd/migrate up
```

Turunkan 1 migration:

```powershell
go run ./cmd/migrate down 1
```

Cek versi migration:

```powershell
go run ./cmd/migrate version
```

Paksa versi jika database dirty:

```powershell
go run ./cmd/migrate force 2
```

## Seed User

Buat atau update akun awal:

```powershell
go run ./cmd/seed
```

Seeder menggunakan env:

- `SEED_USER_NAME`
- `SEED_USER_EMAIL`
- `SEED_USER_PASSWORD`

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
