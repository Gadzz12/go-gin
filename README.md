# Go Todo API

REST API sederhana untuk manajemen Todo (CRUD) dibangun dengan **Go** dan **Gin Framework**, menggunakan **PostgreSQL** sebagai basis data.

## Fitur

- `GET /ping` — Health check
- `GET /hello` — Contoh response sederhana
- `GET /hello/:name` — Salam dengan parameter nama
- `GET /search` — Pencarian dengan query parameter `q` dan `limit`
- `GET /api/v1/todos` — Daftar semua todo (dukungan filter `?done=true/false`)
- `GET /api/v1/todos/:id` — Detail todo berdasarkan ID
- `POST /api/v1/todos` — Membuat todo baru
- `PUT /api/v1/todos/:id` — Memperbarui todo berdasarkan ID
- `PATCH /api/v1/todos/:id/done` — Mengubah status `is_done` menjadi true
- `DELETE /api/v1/todos/:id` — Menghapus todo berdasarkan ID

## Teknologi

- [Go](https://go.dev/) (versi 1.27+)
- [Gin](https://gin-gonic.com/) — Web framework
- [PostgreSQL](https://www.postgresql.org/) — Basis data
- [pgx](https://github.com/jackc/pgx) — Driver PostgreSQL untuk Go
- [godotenv](https://github.com/joho/godotenv) — Memuat variabel environment dari file `.env`

## Struktur Proyek

```
go_hari15/
├── main.go          # Titik masuk aplikasi & definisi semua handler
├── go.mod           # Modul dan dependensi Go
├── go.sum           # Checksum dependensi
├── .env             # Variabel environment (tidak di-commit)
├── .env.example     # Contoh variabel environment
├── .gitignore       # File yang diabaikan Git
└── README.md        # Dokumentasi proyek
```

## Instalasi & Menjalankan

### 1. Clone repositori

```bash
git clone https://github.com/username/go_hari15.git
cd go_hari15
```

### 2. Siapkan environment

Salin file contoh environment dan sesuaikan dengan konfigurasi Anda:

```bash
cp .env.example .env
```

Isi file `.env`:

```env
PORT=8000
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
```

### 3. Instal dependensi

```bash
go mod tidy
```

### 4. Jalankan aplikasi

```bash
go run main.go
```

Aplikasi akan berjalan di `http://localhost:8000`.

## Contoh Penggunaan API

### Health Check

```bash
curl http://localhost:8000/ping
```

Response:
```json
{"message":"pong"}
```

### Membuat Todo Baru

```bash
curl -X POST http://localhost:8000/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Belajar Go","is_done":false}'
```

### Mengubah Status Todo menjadi Selesai

```bash
curl -X PATCH http://localhost:8000/api/v1/todos/1/done
```

## Kontribusi

Pull request sangat diterima. Untuk perubahan besar, silakan buka issue terlebih dahulu untuk membahas apa yang ingin Anda ubah.

## Lisensi

[MIT](LICENSE)