# Go Gin Layered Architecture

Template REST API berbasis **Go + Gin** dengan arsitektur berlapis (handler → service → repository → model), PostgreSQL tanpa ORM, dan sejumlah fitur siap pakai:

- CRUD lengkap dengan pagination, search, filter, dan sort
- Autentikasi JWT (access + refresh token) dan akses third-party via API key
- Scheduler berbasis cron
- WebSocket broadcast
- Logging per modul ke file JSON, lengkap dengan endpoint pembacanya
- Integrasi Bunny Storage (upload/hapus file ke CDN)
- Integrasi Google Gemini (chat + generate image)
- Migration & seeder CLI buatan sendiri
- Dokumentasi API (Swagger/OpenAPI) yang di-render dengan Scalar

---

## Daftar Isi

1. [Arsitektur](#arsitektur)
2. [Instalasi & Menjalankan](#instalasi--menjalankan)
3. [Konfigurasi Environment](#konfigurasi-environment)
4. [Migration & Seeder](#migration--seeder)
5. [Format Response](#format-response)
6. [Autentikasi](#autentikasi)
7. [CRUD](#crud)
8. [Scheduler](#scheduler)
9. [WebSocket](#websocket)
10. [Sistem Logging per File](#sistem-logging-per-file)
11. [Integrasi Bunny Storage](#integrasi-bunny-storage)
12. [Integrasi Gemini](#integrasi-gemini)
13. [Dokumentasi API](#dokumentasi-api)
14. [Ringkasan Endpoint](#ringkasan-endpoint)

---

## Arsitektur

```
cmd/
  api/          entrypoint HTTP server, tempat semua dependency di-wiring manual
  migration/    CLI migration
  seed/         CLI seeder
internal/
  config/       loader environment (app, database, docs, bunny, gemini)
  database/     helper CLI migration/seeder + folder migrations & seeders
  dto/          struct request/response beserta tag validasi
  handler/      layer HTTP: bind DTO, mapping error → status code
  helper/       logger, response envelope, JWT, validator
  middleware/   auth middleware (JWT & third-party key)
  model/        struct domain + sentinel error
  repository/   query database/sql mentah ke PostgreSQL
  scheduler/    job cron
  service/      business logic
  websocket/    handler koneksi & broadcast websocket
docs/           output swag (swagger.json, swagger.yaml) + scalar.js
logs/           file log per modul (di-gitignore)
```

Aturan antar layer:

| Layer | Tanggung jawab |
|---|---|
| `handler` | Hanya urusan HTTP: binding DTO, validasi, mapping error domain ke status code, menyusun response envelope |
| `service` | Business rule: cek keunikan email, hashing bcrypt, default & clamping filter, penerbitan JWT |
| `repository` | Query `database/sql` ke PostgreSQL dengan placeholder `$1` (tanpa ORM) |
| `model` | Struct domain + sentinel error yang lintas layer |

**Alur error**: repository menerjemahkan `sql.ErrNoRows` menjadi sentinel error di `internal/model` (contoh `model.ErrUserNotFound`) dan membungkus error lain dengan `fmt.Errorf("...: %w", err)`. Service meneruskan sentinel apa adanya. Handler melakukan `switch` dengan `errors.Is`, lalu jatuh ke `default` yang mencatat error asli ke log dan mengembalikan pesan generik `helper.ErrInternalServerError` — detail error internal tidak pernah bocor ke klien.

Dependency memakai tipe pointer konkret (bukan interface) dan di-wiring manual di [cmd/api/main.go](cmd/api/main.go). Menambah fitur berarti menyentuh tiap layer lalu mendaftarkan route di `main`.

---

## Instalasi & Menjalankan

**Prasyarat**: Go 1.25+, PostgreSQL 13+ (butuh `gen_random_uuid()`).

```bash
git clone <repo-url>
cd go-gin-layered-architecture
go mod download

cp .env.example .env      # lalu isi nilainya
```

Jalankan migration dan seeder:

```bash
go run cmd/migration/main.go migrate
go run cmd/seed/main.go seed
```

Jalankan server:

```bash
go run ./cmd/api                    # tanpa hot reload
air                                 # dengan hot reload (config di .air.toml)
go build -o ./tmp/main ./cmd/api    # build binary
```

> **Penting**: semua perintah harus dijalankan dari root repo. CLI migration/seeder memakai path relatif (`internal/database/migrations`, `internal/database/seeders`), dan `DocsHandler` membaca `docs/swagger.json` serta `docs/static/scalar.js` langsung dari disk saat request.

Cek server hidup:

```bash
curl http://localhost:8084/api/health
# {"message":"OK"}
```

---

## Konfigurasi Environment

Salin `.env.example` ke `.env`:

| Variable | Wajib | Default | Keterangan |
|---|---|---|---|
| `PORT` | tidak | `8080` | Port HTTP server |
| `DB_DSN` | **ya** | — | DSN PostgreSQL, contoh `postgres://root:root@localhost:5432/db?sslmode=disable` |
| `JWT_SECRET` | **ya** | — | Secret penandatangan JWT (HS256) |
| `ALLOWED_ORIGINS` | tidak | `*` | Daftar origin dipisah koma |
| `DOCS_USER` | tidak | `root` | Username basic auth halaman `/docs` |
| `DOCS_PASSWORD` | tidak | `root` | Password basic auth halaman `/docs` |
| `GEMINI_API_KEY` | **ya** | — | API key Google Gemini |
| `GEMINI_DEFAULT_CHAT_MODEL` | tidak | `gemini-3.1-flash-lite` | Model default untuk chat |
| `GEMINI_DEFAULT_IMAGE_MODEL` | tidak | `gemini-3.1-flash-lite-image` | Model default untuk generate image |
| `BUNNY_STORAGE_ACCESS_KEY` | **ya** | — | Access key Bunny Storage Zone |
| `BUNNY_STORAGE_ZONE_NAME` | tidak | `cms-byb` | Nama storage zone |
| `BUNNY_STORAGE_LOCATION` | **ya** | — | Kode region storage (`sg`, `ny`, `de`, dst.) |
| `BUNNY_STORAGE_CDN_BASE_URL` | **ya** | — | Base URL CDN, contoh `https://cms-byb.b-cdn.net` |

Variabel bertanda **ya** membuat aplikasi berhenti dengan `log.Fatal` saat startup bila kosong.

---

## Migration & Seeder

CLI migration dan seeder ditulis sendiri (bukan `golang-migrate`), memakai tabel `migrations` dan `seeders` untuk mencatat apa yang sudah dijalankan. Kedua tabel dibuat otomatis bila belum ada.

### Migration

```bash
go run cmd/migration/main.go create create_products_table   # scaffold file up & down
go run cmd/migration/main.go migrate                        # jalankan semua .up.sql yang pending
go run cmd/migration/main.go down                           # rollback migration terakhir
go run cmd/migration/main.go drop                           # DROP semua tabel di schema public (CASCADE)
```

`create` menghasilkan dua file di `internal/database/migrations/` dengan nomor urut otomatis:

```
000006_create_products_table.up.sql
000006_create_products_table.down.sql
```

`migrate` mengeksekusi setiap `.up.sql` yang belum tercatat di tabel `migrations`, lalu merekam namanya. Migration yang sudah pernah jalan otomatis di-skip.

`down` menjalankan `.down.sql` dari entri terakhir lalu menghapus catatannya — hanya satu langkah per perintah.

`drop` menghapus **seluruh** tabel pada schema `public` dengan `CASCADE`. Berguna untuk reset total; setelahnya jalankan `migrate` lagi.

Migration bawaan:

| File | Isi |
|---|---|
| `000001_create_migrations_table` | Tabel pencatat migration |
| `000002_create_seeders_table` | Tabel pencatat seeder |
| `000003_create_users_table` | `users` (UUID, role `superadmin`/`admin`/`user`, soft delete) |
| `000004_create_revoked_refresh_tokens_table` | Blacklist refresh token (hash SHA-256) |
| `000005_create_third_party_keys_table` | API key untuk akses third-party |

### Seeder

```bash
go run cmd/seed/main.go create product_seeder   # scaffold file Go
go run cmd/seed/main.go seed                    # jalankan semua seeder pending
go run cmd/seed/main.go down                    # rollback seeder terakhir
```

`create` menghasilkan `internal/database/seeders/000003_product_seeder.go` yang sudah memanggil `Register()` di dalam `init()`, sehingga **tidak perlu didaftarkan manual** — cukup isi body fungsi `Seed...` dan `Rollback...`:

```go
func init() {
	Register("000003_product_seeder", Seed000003ProductSeeder, Rollback000003ProductSeeder)
}

func Seed000003ProductSeeder(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO products (name) VALUES ($1)", "Contoh")
	return err
}

func Rollback000003ProductSeeder(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM products WHERE name = $1", "Contoh")
	return err
}
```

Seeder bawaan:

- `000001_user_seeder` — 3 user contoh (password semua: `12345678`)
  - `fardanm.azizi11@gmail.com` (superadmin)
  - `superadmin@gmail.com` (superadmin)
  - `admin@gmail.com` (admin)
- `000002_third_party_keys_seeder` — key third-party bernama `smartchat`

---

## Format Response

Semua response memakai envelope dari `internal/helper/response.go`.

**Sukses tanpa data**

```json
{ "message": "User berhasil dihapus" }
```

**Sukses dengan data**

```json
{
  "message": "User berhasil diambil",
  "data": { "id": "uuid", "name": "Fardan", "email": "fardan@mail.com" }
}
```

**Sukses dengan pagination**

```json
{
  "message": "Daftar user berhasil diambil",
  "data": [ /* ... */ ],
  "meta": {
    "from": 1,
    "to": 10,
    "page": 1,
    "per_page": 10,
    "total": 42,
    "total_pages": 5
  }
}
```

**Error validasi**

```json
{
  "message": "Validasi gagal",
  "errors": {
    "email": "Kolom email harus berupa email yang valid",
    "password": "Kolom password wajib diisi"
  }
}
```

Pesan validasi dihasilkan `helper.ParseValidationError` dan sudah diterjemahkan ke Bahasa Indonesia untuk tag `required`, `min`, `max`, `email`, `oneof`, `numeric`, `gte`, `lte`, `eqfield`, dan `required_without`. Nama kolom pada pesan mengikuti tag `json` (bukan nama field Go), berkat `RegisterTagNameFunc` di `internal/helper/validator.go`.

---

## Autentikasi

### JWT

- **Access token** — TTL 1 jam, berisi claim `user_id`, `name`, `email`, `role`.
- **Refresh token** — TTL 7 hari, hanya berisi `user_id`.
- Algoritma HS256, issuer `go-gin-layered-architecture`.
- Saat logout, hash SHA-256 dari refresh token disimpan ke tabel `revoked_refresh_tokens` sehingga token tersebut tidak bisa dipakai lagi.

Kirim token pada header:

```
Authorization: Bearer <access_token>
```

### Middleware

`middleware.NewAuthMiddleware(...).RequiredAuth(methods ...string)` menentukan metode autentikasi yang diterima sebuah route group:

```go
// Hanya JWT (default saat tanpa argumen)
authFiles.Use(authMiddleware.RequiredAuth())

// JWT ATAU third-party key milik "smartchat"/"tes"
authUsers.Use(authMiddleware.RequiredAuth("jwt", "third_party:smartchat,tes"))
```

Alur pengecekan: token diambil dari header `Authorization`, dicoba sebagai JWT dulu (bila `jwt` diizinkan), lalu dicocokkan ke tabel `third_party_keys` untuk nama-nama yang terdaftar pada argumen. Gagal keduanya → `401`.

Di dalam handler, payload diambil lewat:

```go
payload, err := middleware.GetAuthPayload(c)   // helper.JWTClaims
name, err := middleware.GetThirdPartyName(c)   // nama third party
```

### Third-Party Key

Untuk memberi akses sistem luar tanpa akun user:

1. Tambahkan baris ke tabel `third_party_keys` (`name`, `key`) — lewat seeder atau `INSERT` manual.
2. Daftarkan namanya di route: `RequiredAuth("third_party:namaKlien")`.
3. Klien mengirim key-nya pada header `Authorization: Bearer <key>`.

Setiap akses third-party yang berhasil dicatat ke `logs/third_party.log` beserta method, path, dan status.

---

## CRUD

Contoh CRUD lengkap ada pada resource **Users** (`/api/users`), yang menunjukkan pola untuk seluruh layer.

### List — `GET /api/users`

Mendukung search, filter, sort, dan pagination lewat query param:

| Param | Tipe | Default | Keterangan |
|---|---|---|---|
| `search_global` | string | — | Cari di kolom `id`, `name`, dan `email` sekaligus (ILIKE) |
| `name` | string | — | Filter `name` (ILIKE partial) |
| `email` | string | — | Filter `email` (ILIKE partial) |
| `page` | int | `1` | Halaman |
| `per_page` | int | `10` | Jumlah per halaman, di-clamp maksimum `100` |
| `sort_by` | string | `name` | Salah satu dari `id`, `name`, `email` |
| `sort_dir` | string | `desc` | `asc` atau `desc` |

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8084/api/users?search_global=fardan&page=1&per_page=10&sort_by=name&sort_dir=asc"
```

Kolom sort divalidasi lewat whitelist di repository, bukan langsung dari input — mencegah SQL injection pada klausa `ORDER BY`. Nilai filter selalu dikirim sebagai parameter `$n`.

Nilai `page`/`per_page` yang tidak valid dinormalisasi di service (`page < 1` → `1`, `per_page < 1` → `10`, `per_page > 100` → `100`), lalu filter yang sudah dinormalisasi itu yang dipakai membangun `meta`.

### Detail — `GET /api/users/:id`

Mengembalikan `404` dengan sentinel `model.ErrUserNotFound` jika tidak ada atau sudah dihapus.

### Create — `POST /api/users`

```json
{ "name": "Fardan", "email": "fardan@mail.com", "password": "rahasia123" }
```

Service mengecek keunikan email (`409` bila sudah dipakai), hashing password dengan bcrypt, dan memaksa `role = "user"`.

### Update — `PUT /api/users/:id`

```json
{ "name": "Nama Baru", "email": "baru@mail.com", "password": "passwordbaru" }
```

Semua field opsional — repository membangun klausa `SET` hanya dari field yang terisi. Email yang sudah dipakai user lain menghasilkan `409`. Password otomatis di-hash ulang bila diisi.

### Delete — `DELETE /api/users/:id`

**Soft delete**: mengisi `deleted_at` dan menambahkan sufiks pada email (`email || '--trashed--' || id`) supaya alamat yang sama bisa dipakai mendaftar lagi. Semua query lain selalu menyertakan `deleted_at IS NULL`.

### Menambah resource baru

1. `internal/model/<resource>.go` — struct domain + sentinel error.
2. `internal/dto/<resource>_dto.go` — request/response beserta tag `binding`.
3. `internal/repository/<resource>_repository.go` — query SQL.
4. `internal/service/<resource>_service.go` — business rule.
5. `internal/handler/<resource>_handler.go` — binding, mapping error, anotasi swagger.
6. Wiring + registrasi route di [cmd/api/main.go](cmd/api/main.go).
7. Migration untuk tabelnya, lalu `swag init` untuk memperbarui dokumentasi.

---

## Scheduler

Menggunakan [`robfig/cron/v3`](https://github.com/robfig/cron). Job hidup di `internal/scheduler/`, satu file per job, dan didaftarkan di [cmd/api/main.go](cmd/api/main.go):

```go
testingJob := scheduler.NewTestingJob()

scheduler := cron.New()
scheduler.AddFunc("* * * * *", testingJob.Run)
scheduler.Start()
defer scheduler.Stop()
```

Job berupa struct dengan method `Run()` tanpa argumen — lihat [internal/scheduler/testing.go](internal/scheduler/testing.go):

```go
type TestingJob struct{}

func NewTestingJob() *TestingJob {
	return &TestingJob{}
}

func (j *TestingJob) Run() {
	log.Printf("testing")
}
```

### Menambah job baru

1. Buat `internal/scheduler/<nama>.go` dengan konstruktor `New<Nama>Job(...)` dan method `Run()`.
2. Bila job butuh akses data, kirimkan service lewat konstruktor: `scheduler.NewCleanupJob(userService, logger)`.
3. Daftarkan di `main.go`: `scheduler.AddFunc("0 2 * * *", cleanupJob.Run)`.

### Format cron

Format 5 field standar (menit, jam, tanggal, bulan, hari):

| Ekspresi | Arti |
|---|---|
| `* * * * *` | Setiap menit |
| `*/5 * * * *` | Setiap 5 menit |
| `0 * * * *` | Setiap jam tepat |
| `0 2 * * *` | Setiap hari pukul 02:00 |
| `0 0 * * 1` | Setiap Senin tengah malam |

Descriptor bawaan `robfig/cron` seperti `@every 30s`, `@hourly`, dan `@daily` juga bisa dipakai.

Scheduler berjalan pada goroutine terpisah dari server HTTP; job yang berjalan lama tidak memblokir request. Karena `cron.New()` tanpa opsi memakai zona waktu lokal server, jalankan container/host pada zona waktu yang diinginkan atau tambahkan `cron.WithLocation(...)`.

---

## WebSocket

Memakai [`gorilla/websocket`](https://github.com/gorilla/websocket). Implementasi contoh (chat broadcast) ada di [internal/websocket/testing.go](internal/websocket/testing.go), didaftarkan di `main.go`:

```go
r.GET("/ws/testing", websocket.HandleTestingConnections)
go websocket.HandleTestingBroadcast()
```

Cara kerjanya:

1. `HandleTestingConnections` meng-upgrade koneksi HTTP menjadi WebSocket dan mendaftarkan koneksi ke map `clients`.
2. Setiap pesan yang diterima dari klien dikirim ke channel `broadcast`.
3. `HandleTestingBroadcast` — berjalan sebagai goroutine sepanjang umur aplikasi — membaca channel tersebut dan meneruskan pesan ke seluruh klien yang terhubung.
4. Klien yang terputus atau gagal dikirimi pesan dihapus dari map.

`CheckOrigin` saat ini mengembalikan `true` untuk semua origin agar mudah dites dari file HTML lokal. **Batasi ini sebelum dipakai di produksi.**

### Mencoba

Buka [index.html](index.html) di browser (sudah menunjuk ke `ws://localhost:8084/ws/testing`), atau dari JavaScript:

```javascript
const socket = new WebSocket("ws://localhost:8084/ws/testing");

socket.onopen    = () => console.log("Terhubung ke server!");
socket.onmessage = (event) => console.log("Pesan:", event.data);
socket.onclose   = () => console.log("Koneksi terputus.");

socket.send("halo");   // akan di-broadcast ke semua klien
```

### Menambah channel WebSocket baru

Buat file baru di `internal/websocket/` dengan pola yang sama (upgrader, map client, channel broadcast, sepasang handler), lalu daftarkan di `main.go`:

```go
r.GET("/ws/notifikasi", websocket.HandleNotifikasiConnections)
go websocket.HandleNotifikasiBroadcast()
```

Perlu diketahui: map `clients` pada implementasi contoh belum dilindungi mutex. Untuk beban produksi dengan banyak koneksi bersamaan, tambahkan `sync.RWMutex` atau ganti dengan `sync.Map`.

---

## Sistem Logging per File

Logging memakai `log/slog` dengan **satu file log per modul**, dibuat lewat `helper.Logger(module)` di [internal/helper/logger.go](internal/helper/logger.go).

```go
userLogger := helper.Logger("user")     // menulis ke logs/user.log
authLogger := helper.Logger("auth")     // menulis ke logs/auth.log
```

Karakteristik:

- Output ditulis ke **file dan stdout sekaligus** (`io.MultiWriter`), sehingga tetap terlihat saat development.
- Format **JSON** (`slog.NewJSONHandler`), satu entry per baris.
- Level minimum `Info`.
- Setiap entry otomatis diberi atribut `module`.
- Folder `logs/` dibuat otomatis; bila gagal, logger fallback ke stdout saja.
- Folder `logs/` masuk `.gitignore`.

Logger di-inject ke handler lewat konstruktor, dan dipakai pada cabang `default` saat mapping error:

```go
h.logger.Error("error create user", "error", err.Error())
```

Contoh satu baris di `logs/user.log`:

```json
{"time":"2026-07-30T13:07:11.482Z","level":"ERROR","msg":"error create user","module":"user","error":"insert user: pq: duplicate key"}
```

### Modul yang tersedia

| Logger | File | Isi |
|---|---|---|
| `user` | `logs/user.log` | Operasi CRUD user |
| `auth` | `logs/auth.log` | Login, register, refresh, logout, ubah profil/password |
| `third_party` | `logs/third_party.log` | Akses via API key third-party (nama, method, path, status) |
| `chatbot` | `logs/chatbot.log` | Request ke Gemini |
| `bunny` | `logs/bunny.log` | Upload & hapus file |
| `logs` | `logs/logs.log` | Endpoint pembaca log itu sendiri |

### Membaca log lewat API

Log bisa dibaca lewat HTTP tanpa masuk ke server:

**`GET /api/logs`** — daftar file log beserta ukurannya

```json
{
  "message": "Daftar file log berhasil diambil",
  "data": [
    { "file_name": "auth", "size": 20480 },
    { "file_name": "user", "size": 8192 }
  ]
}
```

**`GET /api/logs/:file_name`** — isi satu file log (tanpa ekstensi `.log`)

```bash
curl http://localhost:8084/api/logs/auth
```

```json
{
  "message": "Log berhasil diambil",
  "data": [
    {
      "time": "2026-07-30T13:07:11.482Z",
      "level": "ERROR",
      "module": "auth",
      "msg": "error login",
      "error": "query user: connection refused"
    }
  ]
}
```

Repository membaca file baris demi baris dengan `bufio.Scanner`, mem-parse tiap baris sebagai JSON, dan melewati baris yang tidak valid. Struktur entry-nya ada di `model.LogEntry` (`level`, `module`, `msg`, `time`, `user_id`, `error`, `path`).

> Endpoint `/api/logs` saat ini **tidak dilindungi middleware auth**. Tambahkan `logs.Use(authMiddleware.RequiredAuth())` di `main.go` sebelum dipakai di lingkungan yang bisa diakses publik, karena isi log dapat memuat informasi sensitif.

### Menambah modul log baru

Cukup panggil `helper.Logger("<nama_modul>")` saat wiring di `main.go` dan kirimkan ke konstruktor handler. File `logs/<nama_modul>.log` akan dibuat otomatis dan langsung ikut muncul di `GET /api/logs`.

---

## Integrasi Bunny Storage

Upload dan hapus file di [Bunny Storage](https://bunny.net/storage/) lewat REST API-nya, diimplementasikan di [internal/service/bunny_service.go](internal/service/bunny_service.go) tanpa SDK tambahan.

### Konfigurasi

```env
BUNNY_STORAGE_ACCESS_KEY=xxxxxxxx        # Password/AccessKey dari Storage Zone
BUNNY_STORAGE_ZONE_NAME=cms-byb          # Nama storage zone
BUNNY_STORAGE_LOCATION=sg                # Kode region (kosongkan/"de" untuk region utama Falkenstein)
BUNNY_STORAGE_CDN_BASE_URL=https://cms-byb.b-cdn.net
```

Host endpoint disusun dari `BUNNY_STORAGE_LOCATION`: nilai kosong atau `de` memakai `storage.bunnycdn.com`, selain itu menjadi `<location>.storage.bunnycdn.com` (misal `sg.storage.bunnycdn.com`).

### Upload — `POST /api/files`

Butuh JWT. Body `multipart/form-data`:

| Field | Wajib | Keterangan |
|---|---|---|
| `file` | ya | File yang diunggah |
| `folder` | tidak | Folder tujuan, boleh bertingkat (`users/avatars`) |
| `file_name` | tidak | Nama file tujuan; default memakai nama file asli |

```bash
curl -X POST http://localhost:8084/api/files \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@foto.png" \
  -F "folder=users/avatars" \
  -F "file_name=avatar-1.png"
```

```json
{
  "message": "File berhasil diunggah",
  "data": { "file_url": "https://cms-byb.b-cdn.net/users/avatars/avatar-1.png" }
}
```

Service menggabungkan `folder` + `file_name`, membuang segmen kosong serta `.` dan `..` (mencegah path traversal), lalu meng-escape tiap segmen sebelum dikirim sebagai `PUT` dengan header `AccessKey`. URL yang dikembalikan adalah URL CDN publik yang siap disimpan ke database.

### Hapus — `DELETE /api/files`

Butuh JWT. Body JSON:

```json
{ "folder": "users/avatars", "file_name": "avatar-1.png" }
```

`file_name` juga menerima path lengkap atau **URL CDN penuh** hasil upload. Bila berupa URL penuh, `folder` diabaikan karena path-nya sudah lengkap:

```json
{ "file_name": "https://cms-byb.b-cdn.net/users/avatars/avatar-1.png" }
```

File yang tidak ditemukan menghasilkan `404` lewat sentinel `model.ErrFileNotFound`.

### Memakai dari service lain

```go
fileURL, err := bunnyService.UploadFile(ctx, "products", "gambar.jpg", fileBytes, "image/jpeg")
err = bunnyService.DeleteFile(ctx, "products", "gambar.jpg")
```

---

## Integrasi Gemini

Chat dan generate image lewat [Google Gemini](https://ai.google.dev/) memakai SDK resmi `google.golang.org/genai`, diimplementasikan di [internal/service/gemini_service.go](internal/service/gemini_service.go).

### Konfigurasi

```env
GEMINI_API_KEY=xxxxxxxx
GEMINI_DEFAULT_CHAT_MODEL="gemini-3.1-flash-lite"
GEMINI_DEFAULT_IMAGE_MODEL="gemini-3.1-flash-lite-image"
```

Model bisa dioverride per-request lewat field `model`; bila kosong, dipakai default dari environment.

### Chat — `POST /api/chatbot/ask`

Menerima **dua** content type.

**JSON** (teks saja):

```bash
curl -X POST http://localhost:8084/api/chatbot/ask \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Jelaskan arsitektur layered dalam 3 kalimat"}'
```

**Multipart** (teks + lampiran dokumen/gambar):

```bash
curl -X POST http://localhost:8084/api/chatbot/ask \
  -F "prompt=Ringkas isi dokumen ini" \
  -F "model=gemini-3.1-flash-lite" \
  -F "file=@laporan.pdf"
```

```json
{
  "message": "Respons berhasil dibuat",
  "data": "Arsitektur layered memisahkan tanggung jawab ..."
}
```

Lampiran dikirim sebagai `genai.NewPartFromBytes` dengan MIME type yang diambil dari header file, digabung bersama prompt teks dalam satu `Content` beroleh `RoleUser`.

> Field `model` hanya dibaca pada jalur multipart. Request JSON ke `/ask` selalu memakai `GEMINI_DEFAULT_CHAT_MODEL`.

### Generate Image — `POST /api/chatbot/generate-image`

**JSON**:

```bash
curl -X POST http://localhost:8084/api/chatbot/generate-image \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Kucing oranye memakai topi koki", "model": "gemini-3.1-flash-lite-image"}'
```

**Multipart** (dengan gambar referensi):

```bash
curl -X POST http://localhost:8084/api/chatbot/generate-image \
  -F "prompt=Ubah gaya gambar ini menjadi cat air" \
  -F "file=@referensi.png"
```

```json
{
  "message": "Gambar berhasil dibuat",
  "data": { "image_url": "data:image/png;base64,iVBORw0KGgoAAA..." }
}
```

Gambar dikembalikan sebagai **data URI base64**, bukan file. Untuk menyimpannya secara permanen, decode base64-nya lalu unggah lewat `BunnyService.UploadFile` dan simpan URL CDN-nya.

### Catatan

- Kedua endpoint chatbot saat ini **belum dilindungi middleware auth** (route-nya terdaftar tanpa `RequiredAuth`). Tambahkan `authChatbot.Use(authMiddleware.RequiredAuth())` di `main.go` bila endpoint ini tidak boleh diakses bebas — request ke Gemini dikenakan biaya per pemakaian.
- Client `genai` dibuat baru pada setiap request. Untuk trafik tinggi, pindahkan pembuatan client ke konstruktor `NewGeminiService` agar bisa dipakai ulang.
- Error dari Gemini dicatat ke `logs/chatbot.log` dan klien hanya menerima pesan `500` generik.

---

## Dokumentasi API

Anotasi Swagger ditulis sebagai komentar di atas tiap handler dan di-generate dengan [`swag`](https://github.com/swaggo/swag):

```bash
swag init -g cmd/api/main.go -o docs
```

Perintah tersebut memperbarui `docs/docs.go`, `docs/swagger.json`, dan `docs/swagger.yaml`. **Jalankan ulang setiap kali anotasi berubah.**

Dokumentasi di-render dengan [Scalar](https://scalar.com/) pada:

```
http://localhost:8084/docs
```

Halaman ini dilindungi HTTP Basic Auth memakai `DOCS_USER` dan `DOCS_PASSWORD` (default `root` / `root`). `DocsHandler` membaca `docs/swagger.json` dan `docs/static/scalar.js` dari disk saat request, jadi aplikasi harus dijalankan dari root repo.

Untuk endpoint yang butuh token, tambahkan `// @Security BearerAuth` pada anotasi handler.

---

## Ringkasan Endpoint

| Method | Endpoint | Auth | Keterangan |
|---|---|---|---|
| `GET` | `/api/health` | — | Health check |
| `GET` | `/docs` | Basic Auth | Dokumentasi API (Scalar) |
| `POST` | `/api/auth/register` | — | Daftar akun baru |
| `POST` | `/api/auth/login` | — | Login, mendapat access + refresh token |
| `POST` | `/api/auth/refresh` | — | Tukar refresh token dengan access token baru |
| `GET` | `/api/auth/me` | JWT | Profil user saat ini |
| `POST` | `/api/auth/logout` | JWT | Revoke refresh token |
| `PUT` | `/api/auth/profile` | JWT | Ubah nama dan/atau email |
| `PUT` | `/api/auth/password` | JWT | Ubah password |
| `GET` | `/api/users` | JWT / third-party | List user (search, filter, sort, pagination) |
| `GET` | `/api/users/:id` | JWT / third-party | Detail user |
| `POST` | `/api/users` | JWT / third-party | Buat user |
| `PUT` | `/api/users/:id` | JWT / third-party | Ubah user |
| `DELETE` | `/api/users/:id` | JWT / third-party | Soft delete user |
| `POST` | `/api/chatbot/ask` | — | Chat dengan Gemini (JSON / multipart + file) |
| `POST` | `/api/chatbot/generate-image` | — | Generate gambar dengan Gemini |
| `POST` | `/api/files` | JWT | Upload file ke Bunny Storage |
| `DELETE` | `/api/files` | JWT | Hapus file dari Bunny Storage |
| `GET` | `/api/logs` | — | Daftar file log |
| `GET` | `/api/logs/:file_name` | — | Isi satu file log |
| `WS` | `/ws/testing` | — | WebSocket chat broadcast |

---

## Lisensi

Belum ditentukan.

---

Dibuat oleh **Fardan Maula Azizi**.
