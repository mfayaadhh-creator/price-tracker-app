package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"price_tracker/internal/repository"
	"price_tracker/internal/scraper"
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
}

func NewTrackerService(repo *repository.ProductRepository, manager *scraper.TrackerManager) *TrackerService {
	return &TrackerService{
		repo:         repo,
		trackManager: manager,
	}
}

// EvaluateAllProducts memeriksa seluruh produk di DB dan mendeteksi penurunan harga
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

	slog.Info("Memulai evaluasi harga berkala", "total_products", len(products))

	for _, p := range products {
		// 1. Tarik harga terbaru secara live
		latestInfo, err := s.trackManager.FetchProduct(p.URL)
		if err != nil {
			errMsg := fmt.Sprintf("Gagal cek produk ID %s (%s): %v", p.ID, p.Name, err)
			slog.Error(errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}

		oldPrice := p.LastPrice
		newPrice := latestInfo.CurrentPrice

		// 2. Evaluasi: Apakah harga turun?
		if newPrice < oldPrice {
			result.PriceDrops++
			slog.Info("🎉 HARGA TURUN TERDETEKSI!",
				"product", p.Name,
				"old_price", oldPrice,
				"new_price", newPrice,
				"user_phone", p.UserPhone,
			)

			// TODO: Di tahap berikutnya, di sini kita panggil Notifier WhatsApp
		}

		// 3. Update database jika harga berubah
		if newPrice != oldPrice || latestInfo.IsDiscount != p.IsDiscount {
			if err := s.repo.UpdateLastPrice(ctx, p.ID, newPrice, latestInfo.IsDiscount); err != nil {
				slog.Error("Gagal update harga di DB", "id", p.ID, "error", err)
			} else {
				slog.Info("Database harga berhasil diperbarui", "id", p.ID, "new_price", newPrice)
			}
		}

		// 4. Polite delay (1 detik) sebelum memeriksa produk berikutnya
		time.Sleep(1 * time.Second)
	}

	slog.Info("Evaluasi harga selesai",
		"checked", result.TotalChecked,
		"drops", result.PriceDrops,
		"errors", len(result.Errors),
	)

	return result, nil
}
