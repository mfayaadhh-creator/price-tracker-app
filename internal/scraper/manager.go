package scraper

import (
	"fmt"
	"time"

	"price_tracker/internal/domain"
)

type TrackerManager struct {
	scrapers []domain.Scraper
	cache    *ScrapeCache
}

func NewTrackerManager(scrapers ...domain.Scraper) *TrackerManager {
	return &TrackerManager{
		scrapers: scrapers,
		cache:    NewScrapeCache(5 * time.Minute),
	}
}

func (m *TrackerManager) FetchProduct(rawURL string) (*domain.ProductInfo, error) {
	// 1. Cek Cache terlebih dahulu
	if m.cache != nil {
		if cached, ok := m.cache.Get(rawURL); ok {
			return cached, nil
		}
	}

	// 2. Jika tidak ada di cache, scrape langsung
	for _, s := range m.scrapers {
		if s.CanHandle(rawURL) {
			info, err := s.FetchPrice(rawURL)
			if err != nil {
				return nil, err
			}
			// Simpan ke cache jika sukses
			if m.cache != nil && info != nil && info.CurrentPrice > 0 {
				m.cache.Set(rawURL, info)
			}
			return info, nil
		}
	}
	return nil, fmt.Errorf("platform e-commerce untuk URL ini belum didukung: %s", rawURL)
}