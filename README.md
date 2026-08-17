# 🏷️ PRICE_TRACKER // Multi-Store E-Commerce Price Monitor & Alert Engine

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?style=flat-square&logo=svelte)](https://svelte.dev)
[![Router](https://img.shields.io/badge/Router-go--chi%2Fchi-blue?style=flat-square)](https://github.com/go-chi/chi)
[![Database](https://img.shields.io/badge/Database-Supabase%20PostgreSQL-3ECF8E?style=flat-square&logo=supabase)](https://supabase.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](https://opensource.org/licenses/MIT)

Aplikasi pemantau harga dan pemburu diskon otomatis untuk berbagai platform toko online (*e-commerce*). Dibangun menggunakan **Go (Golang)** dengan router **`go-chi/chi`**, **Svelte**, **Supabase PostgreSQL**, dan integrasi bot **Telegram** untuk pengiriman notifikasi instan secara *real-time*.

---

## 📖 Latar Belakang & Tujuan Proyek

Proyek ini dikembangkan sebagai **sarana pembelajaran dan eksplorasi rekayasa perangkat lunak (software engineering)**, dengan fokus pada:
1. **Eksplorasi Framework `go-chi/chi`**: Membangun RESTful API yang modular, terstruktur, dan idiomatik di Go menggunakan middleware standar (*logging, recovery, CORS, rate limiting, and routing group*).
2. **Reverse Engineering & Scraping Multi-Platform**: Mempelajari cara kerja ekstraksi metadata web modern (*Next.js `__NEXT_DATA__` state, JSON-LD Schema.org, OpenGraph/Twitter Meta, serta penanganan tantangan bot Akamai Interstitial Solver*).
3. **Concurrency & Real-Time Alerting**: Mengoptimalkan evaluasi harga ribuan produk secara paralel menggunakan *Goroutines & Worker Pools (`sync.WaitGroup`)* serta sistem antrean notifikasi Telegram.
4. **Passwordless Authentication**: Merancang alur login modern yang aman dan praktis tanpa password menggunakan *Telegram Deep-Link Bot Auth*.

---

## ⚡ Fitur Utama

- 🛍️ **Universal Multi-Store Engine**: Mendukung berbagai kategori e-commerce:
  - **Laptop & Elektronik**: Agres.id, toko gadget/komputer.
  - **Fashion & Pakaian**: Uniqlo, Zara (dengan Akamai PoW Solver), Zalora, H&M.
  - **Toko Mandiri**: Toko berbasis platform *Shopify* dan *WooCommerce*.
- 🛡️ **Anti-Bot & Akamai Interstitial Solver**: Memanfaatkan TLS Client Fingerprint impersonation (*Chrome 120 profile*) dan pemecah tantangan Proof-of-Work JavaScript internal.
- ⚡ **In-Memory Scraping Cache (TTL 5 Menit)**: Mencegah pemanggilan berulang ke server toko untuk URL yang sama, menghemat bandwidth, dan mencegah pembatasan IP.
- 🤖 **Telegram Passwordless Deep-Link Login**: Pengguna cukup menekan tombol *Start* di Telegram untuk login otomatis tanpa perlu mengetik atau menghafal Chat ID.
- 📡 **Dual-Mode Telegram Architecture**:
  - **Webhook Mode**: Diaktifkan di lingkungan produksi (Vercel Serverless / VPS dengan domain HTTPS).
  - **Long-Polling Mode**: Berjalan otomatis di lingkungan lokal pengembangan (*localhost*).
- 🎯 **Target Budget Alerts**: Pengguna dapat mengatur batas budget tertentu. Bot akan memprioritaskan peringatan ketika harga menyentuh target.
- 🧭 **Neo-Brutalist Technical UI**: Antarmuka responsif berbasis Svelte yang bebas dependensi CSS berat, dilengkapi sistem ikon vektor SVG kustom dan drag-and-drop reordering.
- 🔒 **Cron Security Protection**: Endpoint evaluasi harga diamankan menggunakan token `CRON_SECRET`.

---

## 🏗️ Arsitektur Proyek

Struktur folder mengadopsi prinsip *Clean Architecture* dan *Standard Go Project Layout*:

```text
price_tracker/
├── api/                       # Entrypoint Vercel Serverless Function (index.go)
├── cmd/
│   └── server/                # Entrypoint server Go standalone (main.go)
├── frontend/                  # Frontend SPA (Svelte + Vite)
│   ├── public/                # Favicon SVG, Apple Touch Icon, static assets
│   └── src/
│       ├── lib/               # Svelte Components (Header, BlockCreator, ProductBlock, Modal, Toast, Icon)
│       ├── App.svelte         # Main Orchestrator UI
│       └── main.js
├── pkg/
│   ├── auth/                  # Telegram Deep-Link Auth & Webhook Handler
│   ├── domain/                # Entity models & interface contracts
│   ├── handler/               # HTTP Handlers (Chi Router endpoints)
│   ├── repository/            # PostgreSQL Database Layer (pgx/v5)
│   ├── scraper/               # Universal Engine, Akamai Solver & Cache
│   ├── service/               # Core Business Logic & Concurrency Evaluator
│   └── telegram/              # Telegram Bot API Dispatcher
├── .env.example               # Template environment variables
├── vercel.json                # Vercel Deployment & Cron Configuration
└── README.md
```

---

## 🚀 Panduan Menjalankan Secara Lokal (Local Development)

### 1. Prasyarat
- **Go**: Versi `1.22` atau lebih baru
- **Node.js**: Versi `18` atau lebih baru
- **Database**: PostgreSQL (direkomendasikan menggunakan instance gratis dari [Supabase](https://supabase.com))
- **Bot Telegram**: Token bot dari [@BotFather](https://t.me/BotFather)

### 2. Kloning Repository
```bash
git clone https://github.com/mfayaadhh-creator/price-tracker-app.git
cd price-tracker-app
```

### 3. Konfigurasi Environment Variables
Salin file `.env.example` menjadi `.env`:
```bash
cp .env.example .env
```
Isi variabel yang dibutuhkan di dalam file `.env`:
```env
PORT=8080
DATABASE_URL=postgres://postgres:password@your-db-host:5432/postgres?sslmode=require
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
WEBHOOK_URL=
CRON_SECRET=
```

### 4. Menjalankan Backend Go
```bash
go run cmd/server/main.go
```
*Server API akan berjalan di `http://localhost:8080`*.

### 5. Menjalankan Frontend Svelte
Di terminal terpisah:
```bash
cd frontend
npm install
npm run dev
```
*Frontend akan berjalan di `http://localhost:5173`*.

---

## 🧪 Menjalankan Automated Tests

Aplikasi dilengkapi unit test live untuk memverifikasi fungsionalitas scraper multi-platform:

```bash
go test -v ./internal/scraper/...
```

---

## 🌐 Panduan Deployment ke Vercel

Proyek ini telah dikonfigurasi agar dapat di-deploy secara *Fullstack Serverless* di Vercel:

1. Hubungkan repository GitHub ini ke akun **Vercel** Anda.
2. Tambahkan **Environment Variables** di Vercel Dashboard Settings:
   - `DATABASE_URL`
   - `TELEGRAM_BOT_TOKEN`
   - `WEBHOOK_URL` (Contoh: `https://your-project.vercel.app`)
   - `CRON_SECRET`
3. Deploy! Vercel akan otomatis membangun frontend Svelte, menyajikan endpoint API Go di `/api`, dan menjalankan jadwal evaluasi harga berkala sesuai [`vercel.json`](vercel.json).

---

## ⚠️ Catatan & Keterbatasan (Disclaimer)

Aplikasi ini dikembangkan sebagai **eksperimen teknis dan proyek studi kasus rekayasa web**:
- Website e-commerce eksternal dapat sewaktu-waktu memperbarui struktur DOM/HTML, kelas CSS, atau memperketat proteksi *anti-bot (Cloudflare, Akamai, Datadome)* yang dapat memengaruhi keberhasilan scraping.
- Super-app marketplace tertentu (seperti Shopee atau Tokopedia versi web aplikasi) menggunakan arsitektur GraphQL terenkripsi dinamis yang memerlukan perlakuan khusus.
- Penggunaan alat ini diharapkan tetap mematuhi ketentuan layanan (*Terms of Service*) dari masing-masing platform toko terkait.

---

## 📄 Lisensi

Proyek ini dilisensikan di bawah [MIT License](LICENSE).
