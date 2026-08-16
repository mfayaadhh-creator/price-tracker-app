# Project Context: Multi-Source E-Commerce Price Drop Tracker & Notification Engine

## 1. Executive Summary & Objective
Proyek ini bertujuan untuk membangun sistem pelacak penurunan harga (*price drop alert*) otomatis pada platform retail e-commerce (dimulai dari Uniqlo dan dirancang modular untuk multi-platform seperti Zara, H&M, dll). 
Sistem akan memantau URL produk yang diinput oleh pengguna, mendeteksi diskon (*limited offer*, *value buy*, atau perubahan harga), dan mengirimkan notifikasi instan (misalnya via Telegram Bot).

---

## 2. Technical Stack & Deployment Architecture
- **Language / Runtime:** Go (Golang)
- **Design Pattern:** Strategy Pattern & Interface-driven Architecture (untuk mendukung multi-source/multi-store)
- **Deployment & Hosting:**
  - **Vercel Serverless Functions:** Native Go support via AWS Lambda environment (`api/` handlers + `vercel.json` Cron Triggers).
  - **Alternative PaaS:** Fly.io / Render / Koyeb (Dockerized Go binary daemon) jika membutuhkan long-running background worker.
  - **Edge Note:** Cloudflare Workers memerlukan WebAssembly (`.wasm`), sehingga Vercel / PaaS standar lebih direkomendasikan untuk Go murni.
- **Database (Lightweight / Serverless):** PostgreSQL (Supabase / Neon) atau SQLite / Turso.
- **Notification Service:** Telegram Bot API (webhook / HTTP request).

---

## 3. Data Extraction Methodologies
1. **Reverse Engineering Internal API (Primary Approach):**
   - Menembak endpoint JSON internal/mobile API dari e-commerce (misalnya `https://www.uniqlo.com/id/api/commerce/v5/id/products/{PRODUCT_ID}`).
   - **Kelebihan:** Sangat cepat, hemat resource/bandwidth, struktur data JSON teratur, meminimalkan resiko *layout breaking*.
2. **Automated URL Parsing:**
   - Pengguna cukup menyalin URL halaman web produk biasa (`https://www.uniqlo.com/id/id/products/E464851-000/00`).
   - Sistem melakukan RegEx extraction untuk mengambil Product ID (`E464851-000`) dan menyusun URL endpoint API secara otomatis.

---

## 4. Multi-Source Architecture (Strategy Pattern in Go)

### A. Universal Data Contract
```go
type ProductInfo struct {
    Platform     string    `json:"platform"`
    ProductID    string    `json:"product_id"`
    Name         string    `json:"name"`
    BasePrice    float64   `json:"base_price"`
    CurrentPrice float64   `json:"current_price"`
    IsDiscount   bool      `json:"is_discount"`
    CheckedAt    time.Time `json:"checked_at"`
}

type Scraper interface {
    CanHandle(rawURL string) bool
    FetchPrice(rawURL string) (*ProductInfo, error)
}
```

### B. Core Components
- `UniqloScraper`: Mengimplementasikan `Scraper` khusus Uniqlo.
- `ZaraScraper`, `HmScraper`: Parser untuk platform tambahan di masa depan.
- `TrackerManager`: Dispatcher / Router yang menerima URL input, memilih scraper yang cocok berdasarkan domain, dan mengembalikan data produk seragam.

---

## 5. System Workflow
1. **User Input:** Pengguna mengirimkan link produk via Telegram / Web UI.
2. **Validation & Register:** `TrackerManager` memvalidasi URL, mengekstrak data awal, lalu menyimpan ke database (`user_id`, `url`, `platform`, `target_price`, `last_price`).
3. **Cron Job Evaluation:** Cron scheduler (Vercel Cron / Go Scheduler) berjalan secara periodik:
   - Mengambil daftar produk aktif dari database.
   - Melakukan `FetchPrice` ke API masing-masing platform dengan jeda (*polite rate limiting*).
   - Membandingkan `current_price` dengan `last_price` / `target_price`.
4. **Alert Trigger:** Jika terdeteksi diskon atau harga turun, kirim pesan notifikasi ke Telegram user.

---

## 6. Implementation Roadmap & Next Steps
- [ ] Setup repository Go & konfigurasi Vercel Serverless (`vercel.json`).
- [ ] Implementasi `UniqloScraper` dengan parser Product ID + fallback headers/User-Agent.
- [ ] Setup database schema untuk tracking list.
- [ ] Integrasi Telegram Bot webhook & command parser (`/add <url>`, `/list`, `/remove`).
- [ ] Konfigurasi cron job dan logic pendeteksi perubahan harga.
