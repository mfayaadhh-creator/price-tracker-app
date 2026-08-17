package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"price_tracker/internal/handler"
	"price_tracker/internal/repository"
	"price_tracker/internal/scraper"
	"price_tracker/internal/service"
	"price_tracker/internal/telegram"
)

func main() {
	// Logger & Env
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL wajib diisi di file .env")
		os.Exit(1)
	}

	// 1. Database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo, err := repository.NewProductRepository(ctx, dbURL)
	if err != nil {
		slog.Error("Gagal terhubung ke Database Supabase", "error", err)
		os.Exit(1)
	}
	defer repo.Close()
	slog.Info("Berhasil terhubung ke Database Supabase")

	// 2. Scraper, Notifier, Service & Handler
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramNotifier := telegram.NewTelegramNotifier(botToken)

	uniqloScraper := scraper.NewUniqloScraper()
	trackManager := scraper.NewTrackerManager(uniqloScraper)
	trackerService := service.NewTrackerService(repo, trackManager, telegramNotifier)
	productHandler := handler.NewProductHandler(repo, trackManager, trackerService)

	// 3. Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "OK",
			"timestamp": time.Now(),
			"service":   "price-tracker-api",
		})
	})

	r.Get("/test-scrape", productHandler.TestScrape)
	r.Get("/api/cron", productHandler.CronEvaluate)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/track", productHandler.AddTrack)
		r.Get("/tracks", productHandler.ListTracks)
		r.Delete("/tracks/{id}", productHandler.DeleteTrack)
	})

	// 4. Server Start
	addr := ":" + port
	slog.Info("Server Price Tracker siap berjalan", "port", port, "healthcheck", "http://localhost:"+port+"/healthz")
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("Gagal menjalankan server", "error", err)
		os.Exit(1)
	}
}
