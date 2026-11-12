# Tiket.in — Backend REST API

Sistem Backend Manajemen Event dan E-Ticketing Kampus berbasis **Clean Architecture** menggunakan Go (Golang), GoFiber, PostgreSQL, dan MongoDB.

## Tim Pengembang
- **Fahmi (Git Master):** Domain Event & Ticket Types, Setup Fondasi & DevOps
- **Fatin:** Domain Auth, Kategori, & Event Reviews (MongoDB)
- **Naura:** Domain Registrasi Tiket, Concurrency Locking (`TxManager`), & Check-in

---

## Struktur Clean Architecture (Modul 4 & Modul 12)
```text
tiket-in-backend/
├── main.go                       # Entry point aplikasi & graceful shutdown
├── app/
│   ├── model/                    # Layer 1: Entities & DTO
│   ├── repository/               # Layer 3: Database Gateway (PostgreSQL & MongoDB)
│   └── service/                  # Layer 2 (Use Cases) & Layer 3 (Controller)
├── config/                       # Layer 4: Konfigurasi app, env, logger
├── database/                     # Layer 4: Database Connection Pools
├── helper/                       # Presenter & Utilities lintas layer
├── middleware/                   # Layer 4: HTTP Interceptors
├── migrations/                   # Skrip DDL Database Migration
└── route/                        # Layer 4: Routing HTTP
```

---

## Cara Menjalankan Proyek

### 1. Prasyarat
- Go 1.22+
- PostgreSQL 15+
- MongoDB 6+

### 2. Konfigurasi Lingkungan
Salin file `.env.example` menjadi `.env` lalu sesuaikan kredensial:
```bash
cp .env.example .env
```

### 3. Menjalankan Aplikasi
```bash
go run main.go
```

Akses endpoint uji kesehatan:
```bash
curl http://localhost:3000/api/v1/health
```
