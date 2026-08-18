package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"price_tracker/pkg/domain"
	"price_tracker/pkg/repository"
	"price_tracker/pkg/scraper"
)

// EvaluationResult menyimpan ringkasan hasil evaluasi cron
type EvaluationResult struct {
	TotalChecked int      `json:"total_checked"`
	PriceDrops   int      `json:"price_drops"`
	Errors       []string `json:"errors,omitempty"`
}

type TrackerService struct {
	repo         *repository.ProductRepository
	trackManager *scraper.TrackerManager
	notifier     domain.Notifier
}

func NewTrackerService(repo *repository.ProductRepository, manager *scraper.TrackerManager, notifier domain.Notifier) *TrackerService {
	return &TrackerService{
		repo:         repo,
		trackManager: manager,
		notifier:     notifier,
	}
}

// EvaluateAllProducts memeriksa seluruh produk secara paralel menggunakan Goroutines
func (s *TrackerService) EvaluateAllProducts(ctx context.Context) (*EvaluationResult, error) {
	products, err := s.repo.GetAllTrackedProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil produk dari database: %w", err)
	}

	result := &EvaluationResult{
		TotalChecked: len(products),
		PriceDrops:   0,
		Errors:       []string{},
	}

	if len(products) == 0 {
		return result, nil
	}

	slog.Info("Memulai evaluasi harga berkala (Paralel)", "total_products", len(products))

	// Menggunakan sync.WaitGroup dan Mutex untuk keamanan konkurensi Go
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Semaphore untuk membatasi maksimal 5 request konkuren sekaligus (agar aman & sopan)
	maxWorkers := 5
	sem := make(chan struct{}, maxWorkers)

	for _, p := range products {
		wg.Add(1)
		sem <- struct{}{} // Ambil slot worker

		go func(prod domain.TrackedProduct) {
			defer wg.Done()
			defer func() { <-sem }() // Lepas slot worker

			// 1. Tarik harga terbaru secara live
			latestInfo, err := s.trackManager.FetchProduct(prod.URL)
			if err != nil {
				errMsg := fmt.Sprintf("Gagal cek produk ID %s (%s): %v", prod.ID, prod.Name, err)
				slog.Warn(errMsg)

				mu.Lock()
				result.Errors = append(result.Errors, errMsg)
				mu.Unlock()

				// Jika produk sebelumnya aktif dan sekarang tidak tersedia (404/habis), update status ke unavailable
				if prod.Status != "unavailable" {
					_ = s.repo.UpdateProductStatus(ctx, prod.ID, "unavailable")
					slog.Info("Status produk diubah menjadi unavailable di DB", "id", prod.ID, "name", prod.Name)

					if s.notifier != nil {
						msg := fmt.Sprintf(
							"⚠️ <b>INFO: PRODUK TIDAK TERSEDIA / HABIS</b>\n\n"+
								"Produk incaran Anda: <b>%s</b> tampaknya sudah tidak tersedia atau dihapus dari katalog Uniqlo.\n\n"+
								"🛒 <a href=\"%s\">Cek Link Produk</a>",
							prod.Name, prod.URL,
						)
						_ = s.notifier.SendAlert(prod.UserPhone, msg)
					}
				}
				return
			}

			// Jika produk sebelumnya unavailable tapi sekarang kembali aktif, pulihkan status ke active
			if prod.Status == "unavailable" {
				_ = s.repo.UpdateProductStatus(ctx, prod.ID, "active")
				slog.Info("Produk kembali tersedia, status dipulihkan ke active", "id", prod.ID, "name", prod.Name)
			}

			oldPrice := prod.LastPrice
			newPrice := latestInfo.CurrentPrice

			// 2. Cek apakah harga turun atau target tercapai
			isPriceDropped := newPrice < oldPrice || (latestInfo.IsDiscount && !prod.IsDiscount)
			isTargetHit := prod.TargetPrice > 0 && newPrice <= prod.TargetPrice

			if isPriceDropped || isTargetHit {
				mu.Lock()
				result.PriceDrops++
				mu.Unlock()

				slog.Info("🎉 PERUBAHAN HARGA MENGUNTUNGKAN TERDETEKSI!",
					"product", prod.Name,
					"old_price", oldPrice,
					"new_price", newPrice,
					"target_hit", isTargetHit,
					"chat_id", prod.UserPhone,
				)

				// Kirim notifikasi alert via Telegram Bot
				if s.notifier != nil {
					hemat := oldPrice - newPrice
					if hemat <= 0 {
						hemat = latestInfo.BasePrice - newPrice
					}

					platformName := prod.Platform
					if platformName == "" {
						platformName = "Website Toko"
					}

					var msg string
					if isTargetHit {
						// 🎯 Format Khusus: Target Budget Tercapai!
						msg = fmt.Sprintf(
							"🎯 <b>TARGET HARGA TERCAPAI!</b>\n\n"+
								"Kabar gembira! Produk incaran Anda sudah menyentuh target budget!\n\n"+
								"📦 <b>Produk:</b> %s\n"+
								"🏷️ <b>Platform:</b> %s\n"+
								"🎯 <b>Target Budget:</b> Rp %.0f\n"+
								"🔥 <b>Harga Sekarang:</b> <b>Rp %.0f</b>\n"+
								"💰 <b>Total Hemat:</b> Rp %.0f\n\n"+
								"🛍️ <a href=\"%s\">Beli Sekarang di %s</a>",
							prod.Name, platformName, prod.TargetPrice, newPrice, hemat, prod.URL, platformName,
						)
					} else {
						// 📉 Format: Penurunan Harga Biasa
						targetInfo := ""
						if prod.TargetPrice > 0 {
							targetInfo = fmt.Sprintf("\n🎯 <i>Target Budget Anda: Rp %.0f (Belum tercapai)</i>", prod.TargetPrice)
						}

						msg = fmt.Sprintf(
							"🔔 <b>PENURUNAN HARGA TERDETEKSI!</b>\n\n"+
								"📦 <b>Produk:</b> %s\n"+
								"🏷️ <b>Platform:</b> %s\n"+
								"🏷️ <b>Harga Sebelumnya:</b> <s>Rp %.0f</s>\n"+
								"🔥 <b>Harga Sekarang:</b> <b>Rp %.0f</b>\n"+
								"💰 <b>Hemat:</b> Rp %.0f%s\n\n"+
								"🛍️ <a href=\"%s\">Klik di Sini untuk Beli di %s</a>",
							prod.Name, platformName, oldPrice, newPrice, hemat, targetInfo, prod.URL, platformName,
						)
					}

					if err := s.notifier.SendAlert(prod.UserPhone, msg); err != nil {
						slog.Error("Gagal mengirim notifikasi Telegram", "chat_id", prod.UserPhone, "error", err)
					}
				}
			}

			// 3. Update database jika ada perubahan harga
			if newPrice != oldPrice || latestInfo.IsDiscount != prod.IsDiscount {
				if err := s.repo.UpdateLastPrice(ctx, prod.ID, newPrice, latestInfo.IsDiscount); err != nil {
					slog.Error("Gagal update harga di DB", "id", prod.ID, "error", err)
				} else {
					slog.Info("Database harga berhasil diperbarui", "id", prod.ID, "new_price", newPrice)
				}
			}
		}(p)
	}

	wg.Wait() // Tunggu seluruh goroutine selesai secara bersamaan

	slog.Info("Evaluasi harga selesai",
		"checked", result.TotalChecked,
		"drops", result.PriceDrops,
		"errors", len(result.Errors),
	)

	return result, nil
}

// SendRegistrationAlert mengirimkan notifikasi konfirmasi pendaftaran produk ke Telegram user
func (s *TrackerService) SendRegistrationAlert(p domain.TrackedProduct) {
	if s.notifier == nil || p.UserPhone == "" {
		return
	}

	platformName := p.Platform
	if platformName == "" {
		platformName = "Toko Online"
	}

	targetStr := ""
	if p.TargetPrice > 0 {
		targetStr = fmt.Sprintf("\n🎯 <b>Target Budget:</b> Rp %.0f", p.TargetPrice)
	}

	msg := fmt.Sprintf(
		"🎉 <b>PRODUK BERHASIL DIPANTAU!</b>\n\n"+
			"📦 <b>Produk:</b> %s\n"+
			"🏪 <b>Platform:</b> %s\n"+
			"💰 <b>Harga Terpantau:</b> <b>Rp %.0f</b>%s\n\n"+
			"✅ <i>Sistem akan memantau produk ini 24/7. Notifikasi akan segera dikirim jika harga turun!</i>\n\n"+
			"🛍️ <a href=\"%s\">Buka Produk di %s</a>",
		p.Name, platformName, p.LastPrice, targetStr, p.URL, platformName,
	)

	go func() {
		if err := s.notifier.SendAlert(p.UserPhone, msg); err != nil {
			slog.Warn("Gagal mengirim notifikasi registrasi ke Telegram", "chat_id", p.UserPhone, "error", err)
		}
	}()
}

