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

	"price_tracker/internal/auth"
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

	// 2. Scraper, Notifier, Service, Auth & Handler
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramNotifier := telegram.NewTelegramNotifier(botToken)
	authManager := auth.NewTelegramAuthManager(botToken, "mf_pricetracker_bot")

	uniqloScraper := scraper.NewUniqloScraper()
	universalScraper := scraper.NewUniversalScraper()
	trackManager := scraper.NewTrackerManager(uniqloScraper, universalScraper)
	trackerService := service.NewTrackerService(repo, trackManager, telegramNotifier)
	productHandler := handler.NewProductHandler(repo, trackManager, trackerService, authManager)

	// 3. Router
	r := chi.NewRouter()
	
	// Logger yang otomatis mengabaikan log polling /poll agar terminal tetap bersih
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/auth/telegram/poll" {
				next.ServeHTTP(w, r)
				return
			}
			middleware.Logger(next).ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Recoverer)

	// CORS Middleware untuk Frontend Svelte
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

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

		// Auth Routes
		r.Post("/auth/telegram/init", productHandler.InitTelegramAuth)
		r.Get("/auth/telegram/poll", productHandler.PollTelegramAuth)
		r.Post("/auth/instant", productHandler.InstantLogin)
	})

	// 4. Server Start
	addr := ":" + port
	slog.Info("Server Price Tracker siap berjalan", "port", port, "healthcheck", "http://localhost:"+port+"/healthz")
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("Gagal menjalankan server", "error", err)
		os.Exit(1)
	}
}
