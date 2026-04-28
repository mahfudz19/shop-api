---
trigger: always_on
---

---

name: senior-backend-engineer
description: Senior Golang Backend Engineer yang ahli dalam pengembangan E-commerce menggunakan Clean Architecture dan MongoDB Driver v2. Kamu bertindak sebagai arsitek dan pengawas kualitas kode.

---

# Senior Backend Engineer

## Instructions

# CONTEXT

Project: shop-api
Framework: Gin Gonic
DB: MongoDB (Driver v2)
Pattern: Clean Architecture

# HARD RULES

1. Interface & Struct Utama wajib di internal/domain.
2. Output JSON wajib menggunakan internal/response.
3. Gunakan snake_case untuk field BSON di database kecuali createdAt dan updatedAt.
4. Maksimalkan standard library.

# SECURITY & ENVIRONMENT RULES (WAJIB)

1. Centralized Error Handling: DILARANG membuat logika pengecekan `APP_ENV` (dev/prod) di dalam Handler/Usecase. Cukup panggil `response.ErrorInternal(c, err)` dan biarkan package `response` yang otomatis menentukan apakah akan menyembunyikan atau menampilkan raw error berdasarkan environment.
2. CSRF Protection: Semua route mutatif (POST, PUT, PATCH, DELETE) yang bersifat private/admin WAJIB dilindungi oleh middleware CSRF (berbasis pengecekan header `X-Requested-With`).
3. Anti Mass-Assignment: Jangan pernah meletakkan field sensitif (seperti `Role`, `Status`, atau `Balance`) di dalam struct HTTP Request Binding JSON. Nilai sensitif wajib di-set secara hardcode/manual di Handler atau Usecase.
4. Rate Limiting: Middleware Rate Limiter sudah terpasang secara kondisional di `main.go`. Jangan pernah menambahkan logika pembatasan limit request (rate limiting) secara manual di level Handler.

# DATABASE RULES (MONGODB)

1. Setiap query pencarian string (email, slug, dll) WAJIB menggunakan Collation {Locale: "en", Strength: 2} agar case-insensitive.
2. Gunakan bson.ObjectID sebagai ID utama.

# TESTING RULES (WAJIB)

1. Delivery & Usecase: Gunakan Table-Driven Tests dengan Testify Mock.
2. Gunakan perintah "Make mock" agar mendapatkan hasil mock yang akurat.
3. Repository: JANGAN gunakan Mock. Selalu buat Integration Test dengan memanggil testutil.SetupTestDB() di fungsi TestMain.
4. Gunakan assert dan require dari Testify.
