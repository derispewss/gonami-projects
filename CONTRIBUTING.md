# Panduan Berkontribusi

Terima kasih ingin berkontribusi pada **gonami**! Dokumen ini memandu kamu dari
nol sampai pull request pertama. Seluruh siklus pengembangan berjalan di
Docker — tidak perlu install Go, PostgreSQL, MinIO, atau goose di mesin lokal.

---

## 1. Prasyarat

| Tool | Keterangan |
|---|---|
| Docker + Docker Compose v2 | Satu-satunya keharusan |
| Akun GitHub | Untuk fork & pull request |
| `git` | Version control |

## 2. Setup Awal (sekali saja)

```bash
# 1. Fork repo di GitHub, lalu clone fork-mu
git clone https://github.com/<username>/gonami-projects.git
cd gonami-projects

# 2. Tambahkan upstream agar selalu bisa sinkron
git remote add upstream https://github.com/derispewss/gonami-projects.git

# 3. Siapkan konfigurasi
make setup            # membuat .env dari .env.example
# edit .env → isi GEMINI_API_KEY (wajib untuk fitur AI/media)

# 4. Build & jalankan stack development lengkap
make dev-up           # PostgreSQL + MinIO + bot (image build lokal)

# 5. Pairing WhatsApp — scan QR dari log
make dev-logs         # WhatsApp > Perangkat Tertaut > Scan
```

Sesi WhatsApp tersimpan di volume `wa_data` — cukup scan sekali; aman melewati
`down`/`up`. Migrasi database dijalankan otomatis saat bot start.

## 3. Alur Kerja Harian

```bash
git switch main
git pull upstream main

git switch -c feat/nama-fitur     # atau fix/nama-bug

# ... tulis kode ...

make test                         # unit test di container Go
make dev-up                       # rebuild cepat & uji manual via WA

git add -p
git commit -m "feat: deskripsi singkat"
git push origin feat/nama-fitur
```

Lalu buka **Pull Request** dari fork-mu ke `main` di repo upstream.

### Konvensi commit

Pakai awalan tipe + deskripsi singkat bahasa Indonesia:

```
feat: tambah intent laporan tahunan
fix(compose): mount volume wa_data
chore: bump whatsmeow
docs: perbarui panduan kontribusi
```

## 4. Testing

```bash
make test        # semua unit test (jalan di golang:1.26-alpine)
```

- Test parser ada di `tests/parser/`, test AI di `tests/ai/`.
- Tambahkan test untuk setiap perilaku baru — khususnya parser (kasus teks
  natural user sangat beragam).
- CI (`test → docker`) wajib hijau sebelum PR bisa di-merge.

## 5. Struktur Proyek

| Path | Isi |
|---|---|
| `cmd/bot` | Entry point, wiring, migrasi |
| `internal/domain` | Entitas inti (tidak mengimpor layer lain) |
| `internal/whatsapp` | Koneksi, router pesan, sender, reaksi |
| `internal/application` | Use case: catat, konfirmasi, laporan, media |
| `internal/parser` | Layer-1 deterministik: regex, terbilang, multi-item, intent natural |
| `internal/ai` | Layer-2 Gemini (STT/vision/PDF) + token saver |
| `internal/repository` | Akses PostgreSQL (pgx) |
| `internal/storage` | Arsip media: local / MinIO |
| `migrations/` | SQL goose (otomatis dieksekusi saat start) |

Aturan dependency: `whatsapp → application → {parser, ai, repository}`.
Domain tidak mengimpor apa pun; AI/storage tidak menyentuh DB.

## 6. Aturan Desain (penting!)

1. **Komentar itu opsional.** Tidak ada aturan kaku soal komentar —
   gunakan seperlunya bila memang membantu pembaca kode.
2. **Silence rule.** Pesan yang tidak dikenali didiamkan total, tanpa balasan
   teks. Jangan menambah balasan "maaf, tidak paham".
3. **Hasil LLM selalu draft.** Output AI tidak pernah auto-save tanpa
   konfirmasi user.
4. **Layer-1 dulu baru LLM.** Segala hal yang bisa dipahami deterministik
   (0 token) harus ditangani `internal/parser` sebelum menyentuh Gemini.
5. **Bahasa Indonesia** untuk log, pesan bot, dan dokumen.

## 7. Menambah Migrasi Database

Buat file baru bernomor di `migrations/`, contoh `003_tabel_tabungan.sql`,
dengan direktif goose:

```sql
-- +goose Up
CREATE TABLE tabungan (...);

-- +goose Down
DROP TABLE tabungan;
```

Tidak perlu menjalankan apa pun — migrasi otomatis saat bot restart.

## 8. Checklist Pull Request

- [ ] Branch baru dari `main` terbaru
- [ ] Commit mengikuti konvensi
- [ ] `make test` hijau
- [ ] Perilaku sudah dicoba manual lewat `make dev-up`
- [ ] Tanpa secret/`.env` ter-commit
