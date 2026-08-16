package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
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

	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "OK",
			"timestamp": time.Now(),
			"service":   "price-tracker-api",
		})
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
