# Postman Collection

Folder ini berisi file Postman untuk semua fitur API di project ini.

## File

- `finance-backend.postman_collection.json`: Collection semua endpoint.
- `finance-backend.local.postman_environment.json`: Environment lokal dengan variabel dasar.

## Cara Pakai

1. Import kedua file ke Postman.
2. Pilih environment `Finance Backend Local`.
3. Jalankan request `Auth > Login`.

## Auto Save Token

Request `Auth > Login` sudah punya script test yang otomatis menyimpan:

- `access_token`
- `refresh_token`
- `token_type`
- `user_id`

Token disimpan ke **collection variables** dan **environment variables**, jadi request protected bisa langsung dipakai.
