package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"price_tracker/internal/handler"
	"price_tracker/internal/repository"
	"price_tracker/internal/scraper"
	"price_tracker/internal/service"
	"price_tracker/internal/telegram"
)

var app http.Handler

func init() {
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

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		repo, err := repository.NewProductRepository(ctx, dbURL)
		if err == nil {
			botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
			telegramNotifier := telegram.NewTelegramNotifier(botToken)

			uniqloScraper := scraper.NewUniqloScraper()
			trackManager := scraper.NewTrackerManager(uniqloScraper)
			trackerService := service.NewTrackerService(repo, trackManager, telegramNotifier)
			productHandler := handler.NewProductHandler(repo, trackManager, trackerService)

			r.Get("/test-scrape", productHandler.TestScrape)
			r.Get("/api/cron", productHandler.CronEvaluate)

			r.Route("/api/v1", func(r chi.Router) {
				r.Post("/track", productHandler.AddTrack)
				r.Get("/tracks", productHandler.ListTracks)
				r.Delete("/tracks/{id}", productHandler.DeleteTrack)
			})
		} else {
			slog.Error("Gagal inisialisasi DB di Vercel", "error", err)
		}
	}

	app = r
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
 