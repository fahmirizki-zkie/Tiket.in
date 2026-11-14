# Panduan Pengujian API: Domain Event & Jenis Tiket (Fahmi)

Dokumen ini menyediakan kumpulan skrip pengujian cURL dan payload JSON siap pakai untuk menguji seluruh endpoint domain **Event** dan **Jenis Tiket**.

---

## 1. Persiapan Awal

1. **Jalankan Basis Data PostgreSQL & MongoDB** (jika menggunakan Docker/Lokal):
   Pastikan tabel `events` dan `ticket_types` sudah termigrasi sesuai file di folder `migrations/`.
2. **Jalankan Server Backend:**
   ```powershell
   go run main.go
   ```
3. **Dapatkan Token JWT Panitia:**
   Lakukan login dengan akun ber-role `panitia` (atau registrasi akun panitia baru jika belum ada):
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
       "email": "panitia@unair.ac.id",
       "password": "Password123!"
     }'
   ```
   *Salin token dari field `data.token` untuk digunakan pada header `Authorization: Bearer <TOKEN>` di bawah.*

---

## 2. Kumpulan Endpoint & Payload Siap Uji

### A. Buat Event Baru (`POST /api/v1/events`)
- **Akses:** Khusus Panitia / Admin (`Authorization: Bearer <TOKEN>`)
- **cURL:**
  ```bash
  curl -X POST http://localhost:8080/api/v1/events \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <TOKEN_PANITIA>" \
    -d '{
      "category_id": 1,
      "title": "Airlangga Tech Festival 2026",
      "description": "Konferensi tahunan inovasi teknologi dan backend engineering.",
      "location": "Auditorium Candradimuka Kampus C UNAIR",
      "start_at": "2026-12-01T08:00:00Z",
      "end_at": "2026-12-01T17:00:00Z"
    }'
  ```
- **Ekspektasi Hasil:** `201 Created`

#### Skenario Negatif (Jadwal Terbalik):
```bash
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN_PANITIA>" \
  -d '{
    "category_id": 1,
    "title": "Event Jadwal Rusak",
    "location": "Online",
    "start_at": "2026-12-05T10:00:00Z",
    "end_at": "2026-12-01T08:00:00Z"
  }'
```
- **Ekspektasi Hasil:** `400 Bad Request` (`"waktu selesai harus setelah waktu mulai"`)

---

### B. Ambil Katalog Event Publik (`GET /api/v1/events`)
- **Akses:** Publik (Bebas tanpa token)
- **cURL:**
  ```bash
  curl -X GET "http://localhost:8080/api/v1/events?page=1&limit=10&search=Tech"
  ```
- **Ekspektasi Hasil:** `200 OK` (dengan pagination meta `total`, `page`, `limit`, `total_pages`)

---

### C. Ambil Detail Event & Penyelenggara (`GET /api/v1/events/:id`)
- **Akses:** Publik
- **cURL:**
  ```bash
  curl -X GET http://localhost:8080/api/v1/events/1
  ```
- **Ekspektasi Hasil:** `200 OK` (berisi data event, nama kategori, nama panitia, dan array `ticket_types`)

---

### D. Perbarui Data Event (`PUT /api/v1/events/:id`)
- **Akses:** Khusus Panitia Pemilik Event / Admin
- **cURL:**
  ```bash
  curl -X PUT http://localhost:8080/api/v1/events/1 \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <TOKEN_PANITIA_PEMILIK>" \
    -d '{
      "title": "Airlangga Tech Festival 2026 (Updated Venue)",
      "description": "Venue dialihkan ke Hall Garuda Mukti lantai 2.",
      "location": "Hall Garuda Mukti Kampus MERR C UNAIR",
      "start_at": "2026-12-01T08:30:00Z",
      "end_at": "2026-12-01T17:30:00Z",
      "status": "published"
    }'
  ```
- **Ekspektasi Hasil:** `200 OK`

#### Skenario Negatif (Panitia Lain Mencoba Mengubah):
Gunakan token milik akun panitia yang berbeda.
- **Ekspektasi Hasil:** `403 Forbidden` (`"anda tidak memiliki hak akses untuk mengubah event ini"`)

---

### E. Tambah Jenis Tiket pada Event (`POST /api/v1/events/:id/tickets`)
- **Akses:** Khusus Panitia Pemilik Event / Admin

#### 1. Tiket Early Bird:
```bash
curl -X POST http://localhost:8080/api/v1/events/1/tickets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN_PANITIA_PEMILIK>" \
  -d '{
    "name": "Early Bird",
    "price": 25000,
    "quota": 50
  }'
```
- **Ekspektasi Hasil:** `201 Created`

#### 2. Tiket VIP:
```bash
curl -X POST http://localhost:8080/api/v1/events/1/tickets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN_PANITIA_PEMILIK>" \
  -d '{
    "name": "VIP Ticket",
    "price": 75000,
    "quota": 20
  }'
```
- **Ekspektasi Hasil:** `201 Created`

#### Skenario Negatif (Duplikasi Nama Tiket):
Jalankan pembuatan tiket bernama `"Early Bird"` sekali lagi pada event yang sama.
- **Ekspektasi Hasil:** `409 Conflict` (`"jenis tiket dengan nama tersebut sudah ada pada event ini"`)

#### Skenario Negatif (Harga Tiket Negatif):
```bash
curl -X POST http://localhost:8080/api/v1/events/1/tickets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN_PANITIA_PEMILIK>" \
  -d '{
    "name": "Tiket Invalid",
    "price": -10000,
    "quota": 10
  }'
```
- **Ekspektasi Hasil:** `400 Bad Request` (`"harga tiket tidak boleh negatif"`)

---

### F. Ambil Daftar Tiket Event (`GET /api/v1/events/:id/tickets`)
- **Akses:** Publik
- **cURL:**
  ```bash
  curl -X GET http://localhost:8080/api/v1/events/1/tickets
  ```
- **Ekspektasi Hasil:** `200 OK` (daftar tiket beserta kalkulasi `available_seats = quota - sold`)
