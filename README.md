# Tiket.in — Backend REST API

Sistem Backend Manajemen Event dan E-Ticketing Kampus berbasis **Clean Architecture** menggunakan Go (Golang), GoFiber, dan PostgreSQL.

---

## Tim Pengembang & Pembagian Tugas (Sesuai Spesifikasi Proyek)
- **Fahmi (Git Master):** Domain Event & Ticket Types, Setup Fondasi, Routing, & Integrasi
- **Fatin:** Domain Autentikasi (JWT/Bcrypt), User Management, & Kategori Event
- **Naura:** Domain Registrasi Tiket, Concurrency Locking (`TxManager`), & Check-in Presensi

---

## Struktur Clean Architecture (Modul 4 & Modul 12)
Aplikasi mematuhi aturan Clean Architecture baku tanpa folder `controller/` atau `handler/` (role controller menyatu pada method service di `*_service.go`, aturan bisnis murni terisolasi di `*_rules.go`):

```text
tiket-in-backend/
├── main.go                       # Entry point aplikasi & graceful shutdown
├── app/
│   ├── model/                    # Layer 1: Entities & DTO Structs
│   ├── repository/               # Layer 3: Database Gateway (PostgreSQL pgxpool & TxManager)
│   └── service/                  # Layer 2 (Use Cases/Rules) & Layer 3 (Controller Handlers)
├── config/                       # Layer 4: Konfigurasi app, env, logger
├── database/                     # Layer 4: PostgreSQL Connection Pool
├── helper/                       # Presenter (WebResponse) & Utilities lintas layer
├── middleware/                   # Layer 4: HTTP Interceptors (JWT Auth, Role RBAC)
├── migrations/                   # Skrip DDL Database Migration (001 - 006)
└── route/                        # Layer 4: Routing HTTP Fiber
```

---

## Skema Basis Data PostgreSQL (6 Tabel Utama)
1. `001_create_users.sql` : Tabel `users` (Role: Peserta, Panitia, Admin)
2. `002_create_categories.sql` : Tabel `categories`
3. `003_create_events.sql` : Tabel `events`
4. `004_create_ticket_types.sql` : Tabel `ticket_types` (Constraints `sold <= quota`, unique name per event)
5. `005_create_registrations.sql` : Tabel `registrations` (Unique `user_id, event_id`, Ticket Code)
6. `006_create_attendance.sql` : Tabel `attendances` (Check-in relasi 1:1 dengan registrasi)

---

## Daftar Endpoint RESTful

| No | Method | Endpoint | Hak Akses | Deskripsi |
| :---: | :--- | :--- | :--- | :--- |
| 1 | POST | `/api/v1/auth/register` | Publik | Pendaftaran akun baru |
| 2 | POST | `/api/v1/auth/login` | Publik | Login & perolehan token JWT |
| 3 | GET | `/api/v1/auth/me` | Peserta / Panitia / Admin | Profil pengguna yang sedang login |
| 4 | GET | `/api/v1/categories` | Publik | Daftar seluruh kategori event |
| 5 | POST | `/api/v1/categories` | Admin | Tambah kategori event baru |
| 6 | PUT | `/api/v1/categories/:id` | Admin | Ubah data kategori |
| 7 | DELETE | `/api/v1/categories/:id` | Admin | Hapus kategori |
| 8 | GET | `/api/v1/events` | Publik | Katalog event dengan filter & pagination |
| 9 | GET | `/api/v1/events/:id` | Publik | Detail lengkap event & kuota tiket |
| 10 | POST | `/api/v1/events` | Panitia / Admin | Buat event baru |
| 11 | PUT | `/api/v1/events/:id` | Panitia (Penyelenggara) / Admin | Perbarui data event |
| 12 | POST | `/api/v1/events/:id/tickets` | Panitia (Penyelenggara) / Admin | Tambah jenis tiket (Early Bird, VIP) |
| 13 | GET | `/api/v1/events/:id/tickets` | Publik | Daftar tiket & sisa kuota kursi |
| 14 | POST | `/api/v1/registrations` | Peserta | Booking tiket (pessimistic row locking) |
| 15 | GET | `/api/v1/registrations/me` | Peserta | Riwayat tiket milik peserta yang login |
| 16 | POST | `/api/v1/attendances/check-in` | Panitia / Admin | Verifikasi kehadiran & check-in peserta |

---

## Cara Menjalankan Proyek

### 1. Prasyarat
- Go 1.22+
- PostgreSQL 15+

### 2. Konfigurasi Lingkungan
Salin file `.env.example` menjadi `.env` lalu sesuaikan kredensial PostgreSQL Anda:
```bash
cp .env.example .env
```

### 3. Migrasi Basis Data
Jalankan skrip SQL berurutan di folder `migrations/` pada database PostgreSQL:
- `001_create_users.sql`
- `002_create_categories.sql`
- `003_create_events.sql`
- `004_create_ticket_types.sql`
- `005_create_registrations.sql`
- `006_create_attendance.sql`

### 4. Menjalankan Server Backend
```bash
go run main.go
```

Endpoint pengecekan kesehatan server:
```bash
curl http://localhost:3000/api/v1/health
```

### 5. Menjalankan Unit Testing Aturan Bisnis Murni
```bash
go test -v -shuffle=on -count=3 ./app/service/...
```
