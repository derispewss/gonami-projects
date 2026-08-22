# finwa 💰

> **Personal finance assistant di WhatsApp** — catat pemasukan, pengeluaran, dan transfer
> cukup dengan ngobrol biasa. Kirim teks, voice note, foto struk, bahkan PDF — finwa
> yang menguruskannya.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![WhatsApp](https://img.shields.io/badge/WhatsApp-whatsmeow-25D366?logo=whatsapp&logoColor=white)
![AI](https://img.shields.io/badge/AI-Gemini-4285F4?logo=google&logoColor=white)
![DB](https://img.shields.io/badge/PostgreSQL-17-4169E1?logo=postgresql&logoColor=white)
[![CI](https://github.com/derispewss/finwa-projects/actions/workflows/ci.yml/badge.svg)](https://github.com/derispewss/finwa-projects/actions/workflows/ci.yml)

---

## ✨ Fitur

| Fitur | Contoh | Hasil |
|---|---|---|
| Catat transaksi natural | `beli kopi 15k` | Expense Rp15.000 tersimpan |
| **Multi-item auto-sum** | `aku beli ketoprak 25k adn esteh 3k` | Satu transaksi Rp28.000 |
| Angka terbilang & typo | `bayar listrik dua juta kemarin` | Expense Rp2.000.000, tanggal kemarin |
| Transfer + tujuan | `transfer 50k ke Budi` | Transfer → Budi |
| Voice note | 🎤 "jajan bakso sepuluh ribu" | STT Gemini → dicatat |
| Foto struk | 📷 kirim struk Indomaret | Vision Gemini → draft konfirmasi |
| PDF invoice/statement | 📄 kirim file PDF | Teks native (hemat token) / vision (scan) |
| Reaksi status | 👀 diproses → ✅ sukses / ❌ gagal | Umpan balik instan |
| Draft + konfirmasi | balas `iya` / `tidak` | Simpan atau batal |
| Laporan bebas gaya | `berapa pengeluaran bulan ini?` | Rekap otomatis |

> **Cakupan:** bot hanya merespons **chat pribadi (DM)**. Pesan di grup,
> list broadcast, newsletter, dan status diabaikan diam-diam — dukungan
> grup *coming soon*.

**Prinsip desain:** tanpa perintah kaku — satu-satunya perintah fix adalah `help`.
Pesan tidak dikenali **diabaikan diam-diam** (aturan silence), bot tidak pernah spam.

---

## 🧠 Cara Kerja Parsing (Arsitektur Dua Lapis)

```mermaid
flowchart TD
    A["Pesan masuk"] --> B{"Layer 1<br/>Deterministik (0 token)<br/>regex + terbilang + tanggal"}
    B -- "confidence ≥ 0.80" --> S["💾 Auto-save"]
    B -- "0.50 – 0.79" --> D["📝 Draft konfirmasi<br/>(iya / tidak)"]
    B -- "gagal / < 0.50" --> P{"Prefilter lokal<br/>sinyal finansial?"}
    P -- "tidak ada sinyal" --> X["🔇 Diamkan (0 token)"]
    P -- "lolos" --> L["Layer 2 — LLM fallback<br/>model murah + prompt khusus<br/>(1 token-budget)"]
    L -- "valid" --> D
    L -- "tidak valid / budget habis" --> X

    style S fill:#dcfce7,stroke:#16a34a,color:#000
    style D fill:#fef9c3,stroke:#ca8a04,color:#000
    style X fill:#f1f5f9,stroke:#94a3b8,color:#000
```

**Token saver:** circuit breaker kuota harian (`LLM_DAILY_BUDGET`, reset tiap ganti hari WIB).
Kuota habis → panggilan LLM ditolak → pesan diamkan. Bot tak akan makan kuota tak terkendali.

**Multi-item safety:** hanya nominal berformat aman (`25k`, `10rb`, `2jt`, `25.000`, `Rp…`)
yang dijumlahkan — nomor rekening/telepon dikecualikan; wajib ada kata sambung
(`dan/adn/dn/plus/&/+`) antar item.

**Pipeline media** (voice note / foto struk / PDF):

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant B as 🤖 Bot
    participant M as MinIO (arsip)
    participant G as ✨ Gemini AI

    U->>B: kirim media
    B-->>U: react 👀 (terbaca, diproses)
    B->>M: arsip salinan mentah
    alt Voice note
        B->>G: STT audio → teks
        B->>B: parser layer 1 & 2
    else Foto struk / PDF scan
        B->>G: Vision → JSON transaksi
    else PDF native
        B->>B: ekstraksi teks lokal (murah)
        B->>G: teks statement → JSON
    end
    alt berhasil
        B-->>U: react ✅ + detail transaksi / draft
    else gagal / bukan transaksi
        B-->>U: react ❌ (tanpa balasan teks)
    end
```

---

## 🚀 Quick Start (Docker — Production)

Prasyarat: Docker + Docker Compose v2. **Tidak perlu build** — image resmi
diambil dari GitHub Container Registry (`linux/amd64` & `linux/arm64`).

```bash
# 0. Ambil repo (untuk docker-compose.yml & .env.example)
git clone https://github.com/derispewss/finwa-projects.git && cd finwa-projects

# 1. Siapkan konfigurasi
cp .env.example .env
# edit .env → isi minimal GEMINI_API_KEY dan POSTGRES_PASSWORD

# 2. Nyalakan seluruh stack (pull image otomatis)
docker compose up -d

# 3. Pairing pertama kali — scan QR dari log
docker compose logs -f bot
# WhatsApp > Perangkat Tertaut > Scan (QR muncul di log)

# 4. Selesai! Kirim "help" ke bot dari WhatsApp
```

> **Package GHCR** awalnya privat setelah push pertama. Agar bisa di-pull publik:
> GitHub → Packages → `finwa-bot` → Package settings → *Change visibility* → Public.
> (Atau `docker login ghcr.io` dengan PAT yang punya scope `read:packages`.)

Session WhatsApp, arsip media, dan database tersimpan di named volumes
(`wa_data`, `pgdata`) — **aman di-restart/rebuild**, tidak perlu scan ulang QR.

### Perintah berguna

```bash
make up             # pull image terbaru + start background
make down           # stop
make logs           # ikuti log bot
```

### Topologi deployment

```mermaid
flowchart LR
    subgraph docker["docker compose"]
        direction TB
        B["🤖 bot<br/>ghcr.io/derispewss/finwa-bot"] --> DB[("🐘 PostgreSQL 17<br/>volume: pgdata")]
        B --> MIO[("📦 MinIO<br/>volume: minio_data")]
        B --- W[("📱 WhatsApp session<br/>volume: wa_data")]
    end
    U["👤 User WhatsApp"] <-->|"DM teks/media + reaksi"| B
    B <-.->|"STT / Vision / fallback (HTTPS)"| G["✨ Gemini API"]
    C["🖥️ Console MinIO<br/>localhost:9001"] -.->|inspeksi arsip| MIO

    style B fill:#dcfce7,stroke:#16a34a,color:#000
    style DB fill:#dbeafe,stroke:#2563eb,color:#000
    style MIO fill:#fef3c7,stroke:#d97706,color:#000
```

Urutan start dijamin compose: **PostgreSQL & MinIO sehat dulu** → baru bot jalan.
Bucket `finwa` dibuat otomatis saat bot start. Semua state persisten ada di
named volumes — rebuild container tanpa kehilangan data.

---

## 🛠️ Development (Full Docker)

Seluruh siklus development berjalan di Docker — **tidak perlu install Go,
PostgreSQL, atau goose di mesin lokal**. Stack dev identik dengan produksi
(PostgreSQL 17 + MinIO + bot), bedanya image bot dibuild dari kode lokal.

```bash
# 1. Setup awal
make setup                 # buat .env dari contoh
# edit .env → isi GEMINI_API_KEY

# 2. Build + start stack lengkap
make dev-up                # = compose overlay build lokal

# 3. Log & QR pairing pertama kali
make dev-logs              # scan QR → sesi tersimpan di volume wa_data

# 4. Edit kode → rebuild cepat (layer cache)
make dev-up

# 5. Unit test & dependency tooling (di container Go)
make test                  # go test ./... di golang:1.26-alpine
make tidy                  # go mod tidy (hasil chown balik ke user)
```

### Cheat sheet Makefile

| Target | Fungsi |
|---|---|
| `make setup` | Buat `.env` pertama kali |
| `make dev-up` / `dev-down` / `dev-logs` | Stack development (build lokal) |
| `make test` | Unit test di container Go |
| `make tidy` | `go mod tidy` di container Go |
| `make up` / `down` / `logs` | Production — pull image GHCR, tanpa build |

> Migrasi database **otomatis** dijalankan saat bot start (goose Up) di kedua mode.
> Data dev/prod menumpang volume yang sama (`pgdata`, `wa_data`, `minio_data`) —
> pindah mode tanpa kehilangan session WhatsApp maupun transaksi.
> Hapus semuanya: `docker compose down -v`.

### CI/CD (GitHub Actions)

Setiap push & PR otomatis melalui `.github/workflows/ci.yml`:

```mermaid
flowchart LR
    A["git push / PR"] --> T["🧪 test<br/>build · vet · go test"]
    T -->|"lolos & push main"| D["🐳 docker<br/>buildx amd64+arm64"]
    T -->|"lolos & tag v*"| D
    D --> P["📦 push ke ghcr.io<br/>latest · semver · sha"]
    D -->|"hanya tag v*"| R["🏷️ GitHub Release<br/>changelog otomatis + digest image"]
```

**Upgrade produksi / pinning versi:** default compose memakai `:latest`
(`pull_policy: always` → `make up` selalu dapat build terbaru). Untuk
mengunci versi tertentu, edit `image:` di `docker-compose.yml`
misalnya `ghcr.io/derispewss/finwa-bot:1.0.0`.

---

## 🤝 Berkontribusi

Proyek ini open source — kontribusi sangat terbuka! Panduan lengkap mulai dari
fork, setup development full-Docker, konvensi commit, aturan desain, sampai
checklist PR ada di **[CONTRIBUTING.md](CONTRIBUTING.md)**.

Ringkasnya:

```bash
git fork → clone → make setup → isi GEMINI_API_KEY di .env
make dev-up          # stack dev lengkap + scan QR sekali
make test            # unit test di container Go
```

Buka Pull Request ke `main` dengan commit mengikuti konvensi
(`feat:`, `fix:`, `chore:`, `docs:`).

---

## ⚙️ Konfigurasi (.env)

### Wajib

| Variable | Default | Keterangan |
|---|---|---|
| `DATABASE_URL` | — | PostgreSQL DSN, mis. `postgres://user:pass@host:5432/finwa?sslmode=disable` |
| `GEMINI_API_KEY` | *(kosong)* | Tanpa ini: fitur audio/foto/PDF & fallback LLM nonaktif (bot tetap jalan untuk teks) |

### WhatsApp & Storage

| Variable | Default | Keterangan |
|---|---|---|
| `WHATSAPP_DB_PATH` | `./data/whatsapp.db` | SQLite device store (session E2E) |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | `finwa` / `finwa_secret` / `finwa` | Kredensial PostgreSQL untuk docker-compose |
| `STORAGE_DRIVER` | `local` | Dev lokal; **produksi otomatis `minio` via docker-compose** |
| `STORAGE_LOCAL_DIR` | `./data/media` | Folder arsip media (driver local saja) |
| `STORAGE_ENDPOINT` | `localhost:9000` | Endpoint MinIO/S3 (di Docker: `minio:9000`) |
| `STORAGE_ACCESS_KEY` / `STORAGE_SECRET_KEY` | `minioadmin` | Kredensial MinIO |
| `STORAGE_BUCKET` | `finwa` | Nama bucket (auto-create saat startup) |
| `STORAGE_USE_SSL` | `false` | HTTPS ke storage |

> Console web MinIO tersedia di `http://localhost:9001` untuk inspeksi arsip media.

### AI & Token Saver

| Variable | Default | Keterangan |
|---|---|---|
| `GEMINI_MODEL` | `gemini-2.0-flash` | Model utama (STT / Vision / dokumen) |
| `GEMINI_MODEL_TEXT` | *(kosong)* | Model murah layer-2 teks; kosong = pakai model utama |
| `LLM_DAILY_BUDGET` | `300` | Maks panggilan LLM/hari WIB (circuit breaker); `0` = unlimited |
| `LLM_MAX_OUTPUT_TOKENS` | `150` | Batas token output ekstraksi teks |

### Perilaku Aplikasi

| Variable | Default | Keterangan |
|---|---|---|
| `APP_ENV` | `development` | `production` saat deploy |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `MAX_AUDIO_SIZE_MB` / `MAX_IMAGE_SIZE_MB` / `MAX_PDF_SIZE_MB` | `10` / `10` / `20` | Batas unduhan media |
| `DRAFT_EXPIRY_MINUTES` | `15` | Umur draft sebelum expired otomatis |
| `CONFIDENCE_AUTO_SAVE` | `0.80` | Ambang auto-save (layer deterministik saja) |
| `CONFIDENCE_ASK_CONFIRM` | `0.50` | Di bawah ambang ini → pesan diabaikan |

> Di docker-compose, `DATABASE_URL`, `WHATSAPP_DB_PATH`, dan `STORAGE_*`
> dioverride otomatis: DB menunjuk ke service `db`, storage **wajib MinIO**
> (`STORAGE_DRIVER=minio`, endpoint `minio:9000`) — sesuai target produksi.

---

## 💬 Panduan Penggunaan Bot

### Mencatat transaksi

```
beli kopi 15k                          → expense 15.000 (auto-save)
aku beli ketoprak 25k adn esteh 3k     → expense 28.000 (auto-save)
gajian 5jt                             → income 5.000.000
transfer 50k ke Budi                   → transfer ke Budi
kayaknya tadi bayar parkir 2rb deh     → confidence rendah → draft → "iya"/"tidak"
```

### Bertanya (bebas gaya bahasa)

Bot memahami maksud, bukan perintah kaku — sapaan/filler (`dong`, `coy`,
`deh`, `pls`…) diabaikan otomatis:

```
berapa pengeluaran hari ini?
rekap minggu ini dong          ← mingguan
rekap hari ini coy             ← harian
laporan bulan ini ya           ← bulanan
saldo gue sisa berapa?
uangku sisa apa
lihat transaksi terakhir
hapus transaksi terakhir
```

Rentang waktu dikenali dari frasa: *hari ini / kemarin / tadi* → harian,
*minggu(ini/an)* → mingguan, sisanya → bulanan.

### Media

| Kirim | Reaksi | Alur |
|---|---|---|
| Voice note | 👀 → ✅/❌ | STT → parser teks biasa |
| Foto struk | 👀 → ✅/❌ | Gemini Vision → draft konfirmasi |
| PDF | 👀 → ✅/❌ | Native → ekstraksi teks murah; Scan → vision |
| Gambar/PDF bukan struk | ❌ | Diamkan (tanpa balasan teks) |

Balas draft dengan jawaban natural — `iya`, `ok simpan`, `gaskeun` untuk
menyimpan; `tidak`, `ga usah`, `jangan dulu`, `batalin` untuk membatalkan.
Pesan berisi nominal (mis. transaksi baru) tidak akan salah terbaca sebagai
jawaban. Draft expired otomatis 15 menit.

---

## 🏗️ Arsitektur

Clean Architecture ringan — dependency mengarah ke dalam:

```
cmd/bot/              Entry point: wiring, migration, graceful shutdown
internal/
├── whatsapp/         Interface layer : koneksi, handler event, router intent,
│                     sender (teks & reaksi), logger bridge slog
├── application/      Use cases       : RecordTransaction (2 lapis), ProcessMedia,
│                                     ConfirmDraft, GetReport, GetBalance, ManageTransaction
├── domain/           Entitas inti    : Transaction, TransactionDraft, User, Category
├── parser/           Layer-1 engine  : amount (multi-item), numberwords, date,
│                                     category, type, prefilter
├── ai/               Layer-2         : interface AIClient, Gemini (genai),
│                                     validator JSON, token saver, prompts/*.txt
├── media/            Processor media : audio (STT), image (vision),
│                                     document (PDF native/scan), pdftext
├── repository/       Data access pgx : user, transaction, draft, category
├── database/         Pool koneksi pgx
├── config/           Loader env vars
├── format/           Format rupiah/tanggal
└── storage/          Arsip media     : local / MinIO
migrations/           Goose SQL migrations
tests/                Unit tests (parser, ai)
```

**Aturan dependency:** `whatsapp → application → {parser, ai, repository}` —
domain tidak mengimpor layer mana pun; AI/storage tidak menyentuh DB.

```mermaid
flowchart TD
    WA["internal/whatsapp<br/>koneksi · router · sender"] --> APP["internal/application<br/>use cases"]
    WA --> P1["internal/parser<br/>layer-1 deterministik"]
    APP --> P1
    APP --> AI["internal/ai<br/>layer-2 Gemini + token saver"]
    APP --> REP["internal/repository<br/>pgx → PostgreSQL"]
    APP --> MED["internal/media<br/>audio · image · PDF"]
    MED --> AI
    APP --> STO["internal/storage<br/>local · MinIO"]
    DOM["internal/domain<br/>entitas inti"] -.->|diimpor semua| WA

    style DOM fill:#f3e8ff,stroke:#9333ea,color:#000
```

### Teknologi

[Go 1.26](https://go.dev) · [whatsmeow](https://github.com/tulir/whatsmeow) (WhatsApp multi-device) ·
[pgx/v5](https://github.com/jackc/pgx) + [goose](https://github.com/pressly/goose) ·
[Google Gen AI SDK](https://pkg.go.dev/google.golang.org/genai) (Gemini) ·
[ledongthuc/pdf](https://github.com/ledongthuc/pdf) (ekstraksi teks PDF pure-Go) ·
[minio-go](https://github.com/minio/minio-go) · [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) (CGO-free)

---

## 🔒 Catatan Produksi

- **State persisten**: volume `wa_data` (session WhatsApp), `pgdata` (database),
  dan `minio_data` (arsip media). Backup ketiganya secara berkala.
- **Restart policy** `unless-stopped`: bot pulih sendiri setelah crash/host reboot.
- **Koneksi putus sesekali** (`websocket EOF`, `received message with old counter`)
  adalah normal — whatsmeow auto-reconnect dengan backoff, pesan duplikat dibuang otomatis.
- **Re-pair QR** (HP baru/logout paksa): hentikan bot, hapus volume session,
  nyalakan lagi:
  ```bash
  docker compose down bot && docker volume rm <project>_wa_data && docker compose up -d
  ```
- **Log rotasi** sudah diset (10 MB × 5 file) via logging driver json-file.
- Jangan commit `.env`.

---

## 📄 Lisensi

Belum ditentukan — gunakan internal/private sampai lisensi resmi diterbitkan.