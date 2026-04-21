# Finance Backend

Backend API berbasis Go untuk autentikasi dan transaksi keuangan.

## Package-by-Feature (Vertical Slices)

Project ini memakai pendekatan **Package-by-Feature (Vertical Slices)**. Artinya, kode dipisah berdasarkan fitur, bukan dipisah global per layer.

Contohnya:

- `internal/auth/` berisi kebutuhan fitur autentikasi
- `internal/transaction/` berisi kebutuhan fitur transaksi
- `internal/server/` berisi kebutuhan global seperti router, middleware, dan dokumentasi

Di dalam tiap fitur, komponen seperti handler, service, repository, dan model disimpan berdekatan. Dengan pola ini:

- perubahan biasanya tetap di satu slice fitur
- penambahan fitur baru lebih mudah karena struktur sudah jelas
- dependensi antar fitur lebih terkontrol

Struktur ringkas:

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

## Persyaratan

- Go `1.26.2`
- MySQL aktif
- File environment `.env`

Gunakan `.env.example` sebagai template awal.

## Konfigurasi

Salin `.env.example` menjadi `.env`, lalu isi minimal value berikut:

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

Keterangan singkat:

- `DB_ENABLED` harus `true`
- `AUTH_JWT_SECRET` wajib diisi untuk menjalankan API
- `SEED_USER_*` dipakai saat menjalankan seeder
- konfigurasi SMTP bersifat opsional

## Cara Menjalankan

Urutan lokal yang disarankan:

```powershell
go run ./cmd/migrate up
go run ./cmd/seed
go run ./cmd/api
```

Server berjalan di `http://localhost:8080`.

Jika ingin mengganti port:

```powershell
$env:PORT="9000"
go run ./cmd/api
```

## Migration

Semua file migration ada di folder `migrations/`.

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

Paksa versi jika status migration dirty:

```powershell
go run ./cmd/migrate force 2
```

## Seed Data

Untuk membuat atau update user awal:

```powershell
go run ./cmd/seed
```

Seeder akan membaca:

- `SEED_USER_NAME`
- `SEED_USER_EMAIL`
- `SEED_USER_PASSWORD`

## Testing

Jalankan seluruh test:

```powershell
go test ./...
```

## FCM Debug Send

Untuk mengirim push FCM minimal langsung ke satu device token memakai kredensial backend yang sama:

```powershell
go run ./cmd/fcm-debug-send -token "<FCM_DEVICE_TOKEN>"
```

Opsional:

```powershell
go run ./cmd/fcm-debug-send -token "<FCM_DEVICE_TOKEN>" -title "Debug" -body "Tes push minimal"
```

Mode pembanding payload:

```powershell
go run ./cmd/fcm-debug-send -token "<FCM_DEVICE_TOKEN>" -mode notification-only
go run ./cmd/fcm-debug-send -token "<FCM_DEVICE_TOKEN>" -mode android-priority
go run ./cmd/fcm-debug-send -token "<FCM_DEVICE_TOKEN>" -mode android-channel
```

Arti mode:

- `notification-only`: hanya `notification.title/body`
- `android-priority`: `notification.title/body` + `android.priority=high`
- `android-channel`: `notification.title/body` + `android.priority=high` + `channel_id=finance-go-default`

Command ini akan memakai:

- `FIREBASE_PROJECT_ID`
- `FIREBASE_CREDENTIALS_PATH`
- `FIREBASE_CREDENTIALS_JSON`

## Catatan

- API akan membaca file `.env` otomatis jika tersedia di root project
- database dan tabel tidak dibuat otomatis saat API start, jadi jalankan migration terlebih dahulu
- jika butuh detail endpoint dan request/response, lihat dokumentasi di folder `docs/api/`
