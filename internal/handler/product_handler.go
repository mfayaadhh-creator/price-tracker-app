package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"price_tracker/internal/domain"
	"price_tracker/internal/repository"
	"price_tracker/internal/scraper"
	"price_tracker/internal/service"
)

// DTO untuk input request tracking produk baru
type TrackRequest struct {
	URL         string  `json:"url"`
	UserPhone   string  `json:"user_phone"`
	TargetPrice float64 `json:"target_price"`
}

// ProductHandler memegang dependensi ke Repository, TrackerManager, dan TrackerService
type ProductHandler struct {
	repo           *repository.ProductRepository
	trackManager   *scraper.TrackerManager
	trackerService *service.TrackerService
}

// Constructor untuk membuat ProductHandler baru
func NewProductHandler(repo *repository.ProductRepository, manager *scraper.TrackerManager, svc *service.TrackerService) *ProductHandler {
	return &ProductHandler{
		repo:           repo,
		trackManager:   manager,
		trackerService: svc,
	}
}

// AddTrack menangani POST /api/v1/track
func (h *ProductHandler) AddTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req TrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON tidak valid"})
		return
	}

	if req.URL == "" || req.UserPhone == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "'url' dan 'user_phone' wajib diisi"})
		return
	}

	// 1. Ambil data produk secara live dari Scraper
	productInfo, err := h.trackManager.FetchProduct(req.URL)
	if err != nil {
		slog.Error("Gagal scraping saat pendaftaran", "url", req.URL, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// 2. Susun entity TrackedProduct
	tp := domain.TrackedProduct{
		UserPhone:   req.UserPhone,
		URL:         req.URL,
		Platform:    productInfo.Platform,
		ProductID:   productInfo.ProductID,
		Name:        productInfo.Name,
		BasePrice:   productInfo.BasePrice,
		LastPrice:   productInfo.CurrentPrice,
		TargetPrice: req.TargetPrice,
		IsDiscount:  productInfo.IsDiscount,
	}

	// 3. Simpan ke database Supabase
	id, err := h.repo.AddTrackedProduct(r.Context(), tp)
	if err != nil {
		slog.Error("Gagal menyimpan ke database", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Gagal menyimpan data ke database"})
		return
	}

	tp.ID = id
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Produk berhasil didaftarkan untuk dipantau",
		"data":    tp,
	})
}

// ListTracks menangani GET /api/v1/tracks
func (h *ProductHandler) ListTracks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	products, err := h.repo.GetAllTrackedProducts(r.Context())
	if err != nil {
		slog.Error("Gagal mengambil data dari database", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Gagal mengambil data produk"})
		return
	}

	if products == nil {
		products = []domain.TrackedProduct{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"total": len(products),
		"data":  products,
	})
}

// DeleteTrack menangani DELETE /api/v1/tracks/{id}
func (h *ProductHandler) DeleteTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	if err := h.repo.RemoveTrackedProduct(r.Context(), id); err != nil {
		slog.Error("Gagal menghapus produk dari database", "id", id, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Gagal menghapus produk"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Produk berhasil dihapus dari tracking list",
		"id":      id,
	})
}

// TestScrape menangani GET /test-scrape (untuk debug)
func (h *ProductHandler) TestScrape(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, `{"error": "parameter 'url' wajib diisi"}`, http.StatusBadRequest)
		return
	}

	productInfo, err := h.trackManager.FetchProduct(targetURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(productInfo)
}

// CronEvaluate menangani GET /api/cron untuk memicu evaluasi harga berkala
func (h *ProductHandler) CronEvaluate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validasi Secret Token opsional jika diset di env
	cronSecret := os.Getenv("CRON_SECRET")
	if cronSecret != "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+cronSecret && r.URL.Query().Get("key") != cronSecret {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized cron trigger"})
			return
		}
	}

	result, err := h.trackerService.EvaluateAllProducts(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Evaluasi harga selesai dijalankan",
		"result":  result,
	})
}