package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"price_tracker/internal/scraper"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback
	}

	// 1. Initiate Scraper Uniqlo & TrackerManager
	uniqloScraper := scraper.NewUniqloScraper()
	trackManager := scraper.NewTrackerManager(uniqloScraper)

	r := chi.NewRouter()

	// Endpoint Healthcheck
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "OK",
			"timestamp": time.Now(),
			"service":   "price-tracker-api",
		})
	})

	// 2. Endpoint Scraping
	r.Get("/test-scrape", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Ambil query param
		targetURL := r.URL.Query().Get("url")
		if targetURL == "" {
			http.Error(w, `{"error": "parameter 'url' wajib diisi"}`, http.StatusBadRequest)
			return
		}

		slog.Info("Mencoba melakukan scraping live", "url", targetURL)

		// Panggil TrackerManager
		productInfo, err := trackManager.FetchProduct(targetURL)
		if err != nil {
			slog.Error("Gagal mengambil harga produk", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Return data product dalam JSON
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(productInfo)
	})

	slog.Info("Server Price Tracker siap berjalan",
		"port", port,
		"healthcheck", "http://localhost:"+port+"/healthz")

	addr := ":" + port
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("Gagal menjalankan server", "error", err)
		os.Exit(1)
	}
}
