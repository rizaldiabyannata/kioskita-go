# KiosKita Go API

Sistem backend Go untuk KiosKita dengan skema database yang dinamis sesuai tipe bisnis (clothing atau makanan/kafe) tanpa JSON/NoSQL di database.

## Fitur Utama

- Setup pertama kali (first-run): pilih tipe bisnis (`clothing` atau `cafe`) dan buat admin.
- Skema database otomatis menyesuaikan tipe bisnis (tabel turunan untuk atribut produk).
- Autentikasi berbasis JWT (cookie).
- Produk, Media, Keranjang/Order.

## Prasyarat

- Go 1.20+ (atau sesuai `go.mod`)
- Docker & Docker Compose (untuk Postgres)
- PowerShell (Windows)

## Konfigurasi Environment

Buat file `.env` di root (sudah di-`gitignore`). Contoh:

```
APP_PORT=8080
APP_DOMAIN=localhost

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=kioskita
DB_SSL_MODE=disable
```

> Pastikan port `5432` tidak bentrok. Ubah nilai sesuai kebutuhan.

## Menjalankan Database (Docker)

Jalankan Postgres via Docker Compose:

```powershell
docker-compose up -d
```

Cek kontainer:

```powershell
docker ps
```

## Build & Run Aplikasi

- Build seluruh modul:

```powershell
go build ./...
```

- Jalankan server API:

```powershell
go run .\cmd\api
```

Server akan berjalan di `http://localhost:8080` (atau sesuai `APP_PORT`).

## Alur Setup Pertama Kali

Saat pertama kali dijalankan (database kosong):

1. Dapatkan Setup Token (lihat log aplikasi). Lalu buat admin:

- Endpoint: `POST /api/v1/setup/admin`
- Header: `X-Setup-Token: <token-dari-log>`
- Body (JSON):

```json
{
  "adminEmail": "owner@contoh.com",
  "adminPassword": "passwordku"
}
```

2. Login untuk mendapatkan cookie token:

- Endpoint: `POST /api/v1/login`
- Body (JSON):

```json
{
  "email": "owner@contoh.com",
  "password": "passwordku"
}
```

3. Pilih tipe bisnis dan simpan konfigurasi (wajib sudah login):

- Lihat tipe tersedia: `GET /api/v1/setup/business-types`
- Simpan konfigurasi: `POST /api/v1/setup/config`
- Body (JSON):

```json
{
  "storeName": "Kiosk Kita",
  "businessTypeID": "clothing"
}
```

4. Restart aplikasi.

- Pada startup, sistem akan: migrasi tabel inti, membentuk tabel subtype sesuai `businessTypeID`, dan menghapus kolom/tabel lama yang tidak dipakai.

## Endpoints Utama (Ringkas)

- Health check: `GET /health`
- Auth:
  - `POST /api/v1/login`
  - `POST /api/v1/logout`
- Setup:
  - `POST /api/v1/setup/admin` (hanya saat belum ada admin; butuh header `X-Setup-Token`)
  - `GET /api/v1/setup/business-types`
  - `POST /api/v1/setup/config` (harus login)
- Produk (harus login untuk write):
  - `GET /api/v1/products/`
  - `GET /api/v1/products/:id`
  - `POST /api/v1/products/` (multipart form; lihat contoh di bawah)
  - `PUT /api/v1/products/:id` (JSON partial update)
  - `DELETE /api/v1/products/:id`
  - `POST /api/v1/products/:id/upload` (tambahkan media)
- Cart/Order (harus login):
  - `POST /api/v1/cart/items`
  - `GET /api/v1/cart/`
  - `PUT /api/v1/cart/items/:item_id`
  - `DELETE /api/v1/cart/items/:item_id`

## Contoh Request (PowerShell)

Login (simpan cookie):

```powershell
$body = @{ email = "owner@contoh.com"; password = "passwordku" } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/login -ContentType "application/json" -Body $body -SessionVariable sess
```

Buat produk (clothing) dengan media:

```powershell
$multipart = @{
  name = "Kaos Polos"
  description = "Kaos bahan katun"
  price = "75000"
  stock = "25"
  warna = "Hitam"
  ukuran = "L"
  bahan = "Katun"
  images = Get-Item .\foto1.jpg
}
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/products/ -WebSession $sess -Form $multipart
```

Buat produk (cafe/food) dengan media:

```powershell
$multipart = @{
  name = "Kopi Arabica"
  description = "Roast medium"
  price = "35000"
  stock = "100"
  asal_biji = "Gayo"
  level_giling = "Medium"
  images = Get-Item .\bean.jpg
}
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/products/ -WebSession $sess -Form $multipart
```

Update produk (JSON partial, clothing):

```powershell
$payload = @{
  name = "Kaos Polos Premium"
  price = 95000
  clothing = @{ warna = "Merah"; ukuran = "M" }
} | ConvertTo-Json
Invoke-RestMethod -Method Put -Uri http://localhost:8080/api/v1/products/<product-id> -ContentType "application/json" -Body $payload -WebSession $sess
```

## Struktur Tabel (Ringkas)

- products(id, name, description, price, stock, created_at)
- clothing_products(id, product_id UNIQUE, warna, ukuran, bahan, created_at)
- food_products(id, product_id UNIQUE, asal_biji, level_giling, created_at)
- media(id, product_id, url, type, created_at)
- users(id, email, password_hash, role, created_at)
- orders(id, user*id, total_amount, status, ship*\*, created_at)
- order_items(id, order_id, product_id, quantity, price_at_purchase, created_at)

## Troubleshooting

- Port 5432/8080 bentrok
  - Ubah port di `.env` atau `docker-compose.yml`.
- Gagal login karena cookie
  - Pastikan `APP_DOMAIN` di `.env` sesuai host (contoh `localhost`).
- Migrasi kolom JSON lama
  - Sistem akan menghapus kolom JSON lama pada startup. Jika perlu migrasi data lama, hubungi maintainer untuk skrip migrasi satu kali.

## Lisensi

MIT (atau sesuai preferensi proyek).
