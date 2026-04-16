# Media API Documentation

This module handles generic file upload and URL generation.

**Base URL**: `/v1/media`  
**Authentication**: All endpoints require `Authorization: Bearer <access_token>`

> [!NOTE]
> Uploaded files are stored in the existing `uploads/` directory and served statically via `/uploads/*`.

---

## 1. Upload File
**POST** `/v1/media/upload`

Multipart form fields:
- `file` required
- `dir` optional

Response includes:
- `path` relative public path, for example `/uploads/media/1/...`
- `url` absolute file URL

---

## 2. Get File URL
**GET** `/v1/media/url?path=/uploads/...`

Returns the absolute URL for an existing public path.

---

## 3. Delete File
**DELETE** `/v1/media?path=/uploads/...`

Deletes the file from storage. This is optional and safe if file is already missing.
