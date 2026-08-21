# Sistem Manajemen Klinik

Aplikasi manajemen klinik berbasis web untuk klinik di Indonesia.

## Tech Stack

- **Backend**: Go (Golang)
- **Database**: PostgreSQL
- **Frontend**: HTML, CSS, JavaScript
- **Authentication**: JWT + RBAC

## Struktur Project

```
klinik-app/
├── AGENTS.md
├── cmd/
│   ├── server/main.go     # Entry point server
│   ├── migrate/main.go    # CLI untuk migration & seed
│   └── seed_meds/main.go  # Seed data obat contoh
├── internal/
│   ├── audit/             # Audit log
│   ├── auth/              # Autentikasi & autorisasi
│   ├── clinic/            # Pengaturan klinik
│   ├── config/            # Konfigurasi environment
│   ├── database/          # Koneksi DB & migrasi
│   ├── doctors/           # Data dokter
│   ├── finance/           # Invoice & pembayaran
│   ├── handler/           # HTTP handlers (halaman web)
│   ├── logger/            # Sistem logging
│   ├── medical_records/   # Rekam medis, diagnosis, tindakan
│   ├── medicines/         # Obat & stok apotek
│   ├── middleware/        # HTTP middleware (auth, RBAC, rate limit)
│   ├── patients/          # Data pasien
│   ├── prescriptions/     # Resep obat
│   ├── queues/            # Antrian pasien
│   ├── registrations/     # Pendaftaran pasien
│   ├── reports/           # Laporan
│   ├── server/            # HTTP server & routes
│   ├── staff/             # Data staf
│   └── users/             # Manajemen pengguna
├── migrations/            # File migrasi database (.sql)
├── web/
│   ├── templates/         # HTML templates
│   └── static/            # CSS, JS, gambar
└── tests/                 # Test files
```

## Prerequisites

- Go 1.21+
- PostgreSQL 14+

## Setup

1. Clone repository

```bash
git clone <repository-url>
cd klinik-app
```

2. Setup environment

```bash
cp .env.example .env
```

3. Edit `.env` sesuai konfigurasi PostgreSQL Anda:

```env
DATABASE_URL=postgres://username:password@localhost:5432/klinik_db?sslmode=disable
JWT_SECRET=your-random-secret-key
APP_ENV=development
APP_PORT=8080
LOG_LEVEL=info
```

4. Jalankan migration

```bash
# Lihat status migration
go run ./cmd/migrate status

# Jalankan semua migration
go run ./cmd/migrate up

# Buat admin user default
go run ./cmd/migrate seed
```

5. Jalankan server

```bash
go run ./cmd/server
```

Server berjalan di http://localhost:8080

## Perintah Database

```bash
# Melihat status semua migration
go run ./cmd/migrate status

# Menjalankan migration yang belum dijalankan
go run ./cmd/migrate up

# Membuat admin user default
go run ./cmd/migrate seed
```

Migration dijalankan secara eksplisit. Server TIDAK menjalankan migration otomatis.

## API Endpoints

### Public (Tanpa Autentikasi)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/health | Health check |
| POST | /api/auth/login | Login |

### Protected (Butuh Bearer Token)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/auth/me | Data user yang login |
| POST | /api/auth/logout | Logout |
| GET | /api/users | Daftar semua user |
| POST | /api/users | Buat user baru |
| GET | /api/users/:id | Detail user |
| PUT | /api/users/:id | Update user |
| DELETE | /api/users/:id | Hapus user |

## Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

Response:

```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": 1724227200,
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@klinik.com",
      "full_name": "Administrator",
      "is_active": true
    }
  }
}
```

## Menggunakan Token

```bash
# Ambil data user yang login
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."

# Daftar user
curl http://localhost:8080/api/users \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

## Default Roles

| Role | Deskripsi |
|------|-----------|
| ADMIN | Akses penuh ke seluruh sistem |
| DOCTOR | Data pasien, antrian, pemeriksaan, rekam medis, resep |
| NURSE | Pendaftaran pasien, antrian, pemeriksaan dasar |
| PHARMACIST | Resep, obat, stok, transaksi apotek |
| CASHIER | Invoice, pembayaran, transaksi |
| OWNER | Dashboard, laporan, laporan keuangan |

## Development

### Build

```bash
go build -o bin/clinic-app.exe ./cmd/server
```

### Build Migrate Tool

```bash
go build -o bin/migrate.exe ./cmd/migrate
```

### Test

```bash
go test ./...
```

### Lint

```bash
go vet ./...
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| DATABASE_URL | PostgreSQL connection string | Required |
| JWT_SECRET | Secret key for JWT tokens | Required |
| APP_ENV | Environment (development/production) | development |
| APP_PORT | Server port | 8080 |
| LOG_LEVEL | Log level (debug/info/error) | info |

## Catatan Penting

- Migration harus dijalankan secara manual menggunakan `go run ./cmd/migrate up`
- Server membutuhkan database yang sudah di-migrate sebelum bisa dijalankan
- File `.env` tidak boleh di-commit ke repository
- Gunakan `.env.example` sebagai template
- Default admin: username `admin`, password `admin123`
- Segera ubah password default setelah login pertama kali

## Module Development

Development dilakukan secara bertahap sesuai urutan:

1. **Phase 1 - Foundation** ✅
   - Project structure
   - Configuration
   - Database connection
   - Migration system
   - Logging
   - Basic REST API
   - Health check

2. **Phase 2 - Authentication** ✅
   - Users & Roles
   - Login/Logout
   - JWT
   - Middleware
   - RBAC
   - Seed admin user

3. **Phase 3 - Clinic Core** ✅
   - Pengaturan klinik
   - Dokter
   - Staf
   - Pasien & pencarian pasien

4. **Phase 4 - Registration** ✅
   - Pendaftaran pasien
   - Antrian (WAITING, CALLED, IN_EXAMINATION, COMPLETED, CANCELLED)

5. **Phase 5 - Medical** ✅
   - Pemeriksaan
   - Rekam medis
   - Diagnosis
   - Tindakan
   - Resep

6. **Phase 6 - Pharmacy** ✅
   - Master obat
   - Stok obat (masuk/keluar dengan catatan transaksi)
   - Proses resep apotek

7. **Phase 7 - Finance** ✅
   - Invoice & item invoice
   - Pembayaran multi metode (CASH, BANK_TRANSFER, QRIS, DEBIT, CREDIT_CARD, OTHER)

8. **Phase 8 - Reports** ✅
   - Laporan pasien, pendaftaran, aktivitas dokter
   - Laporan pendapatan & stok obat

9. **Phase 9 - Security & Production** ✅
   - Audit log
   - Rate limiter pada login
   - RBAC di backend
