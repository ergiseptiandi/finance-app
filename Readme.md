# Finance Backend

Go backend API for financial authentication and transaction workflows.

## Package-by-Feature (Vertical Slices)

This project uses a **Package-by-Feature (Vertical Slices)** structure. Code is organized by feature instead of by global layer.

Examples:

- `internal/auth/` contains authentication-related code
- `internal/transaction/` contains transaction-related code
- `internal/server/` contains shared concerns such as the router, middleware, and documentation

Within each feature, handlers, services, repositories, and models live close together. This approach helps because:

- changes usually stay within one feature slice
- new features are easier to add because the structure is clear
- dependencies between features are easier to control

Project layout:

```text
cmd/
  api/
  migrate/
  seed/
internal/
  auth/
  transaction/
  server/
  config/
  database/
  seed/
migrations/
```

## Requirements

- Go `1.26.2`
- Running MySQL instance
- `.env` file

Use `.env.example` as the starting template.

## Configuration

Copy `.env.example` to `.env`, then fill in at least the following values:

```env
PORT=8080

DB_ENABLED=true
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=secret
DB_NAME=finance_db

AUTH_JWT_SECRET=change-this-to-a-long-random-secret

SEED_USER_NAME=Owner
SEED_USER_EMAIL=owner@example.com
SEED_USER_PASSWORD=supersecret123
```

Quick notes:

- `DB_ENABLED` must be `true`
- `AUTH_JWT_SECRET` is required to run the API
- `SEED_USER_*` values are used by the seeder
- SMTP configuration is optional

## Running the App

Recommended local startup order:

```powershell
go run ./cmd/migrate up
go run ./cmd/seed
go run ./cmd/api
```

The server runs at `http://localhost:8080`.

To change the port:

```powershell
$env:PORT="9000"
go run ./cmd/api
```

## Migrations

All migration files are stored in the `migrations/` folder.

Run all migrations:

```powershell
go run ./cmd/migrate up
```

Rollback one migration:

```powershell
go run ./cmd/migrate down 1
```

Check the migration version:

```powershell
go run ./cmd/migrate version
```

Force a version when migration state is dirty:

```powershell
go run ./cmd/migrate force 2
```

## Seed Data

To create or update the initial user:

```powershell
go run ./cmd/seed
```

The seeder reads:

- `SEED_USER_NAME`
- `SEED_USER_EMAIL`
- `SEED_USER_PASSWORD`

## Testing

Run the full test suite:

```powershell
go test ./...
```

## FCM Debug Send

To send a minimal FCM push directly to a single device token using the same backend credentials:

```powershell
go run ./cmd/fcm-debug-send -token "<FCM_DEVICE_TOKEN>"
```

Optional example:

```powershell
go run ./cmd/fcm-debug-send -token "<FCM_DEVICE_TOKEN>" -title "Debug" -body "Test minimal push"
```

Payload comparison modes:

```powershell
go run ./cmd/fcm-debug-send -token "<FCM_DEVICE_TOKEN>" -mode notification-only
go run ./cmd/fcm-debug-send -token "<FCM_DEVICE_TOKEN>" -mode android-priority
go run ./cmd/fcm-debug-send -token "<FCM_DEVICE_TOKEN>" -mode android-channel
```

Mode meanings:

- `notification-only`: only `notification.title/body`
- `android-priority`: `notification.title/body` + `android.priority=high`
- `android-channel`: `notification.title/body` + `android.priority=high` + `channel_id=finance-go-default`

This command uses:

- `FIREBASE_PROJECT_ID`
- `FIREBASE_CREDENTIALS_PATH`
- `FIREBASE_CREDENTIALS_JSON`

## Notes

- The API automatically reads `.env` from the project root when available
- the database and tables are not created automatically on API startup, so run migrations first
- for endpoint-level request/response details, see the documentation in `docs/api/`
