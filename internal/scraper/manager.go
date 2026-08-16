package scraper

import (
	"fmt"
	"price_tracker/internal/domain"
)

type TrackerManager struct {
	scrapers []domain.Scraper
}

func NewTrackerManager(scrapers ...domain.Scraper) *TrackerManager {
	return &TrackerManager{
		scrapers: scrapers,
	}
}

func (m *TrackerManager) FetchProduct(rawURL string) (*domain.ProductInfo, error) {
	for _, s := range m.scrapers {
		if s.CanHandle(rawURL) {
			return s.FetchPrice(rawURL)
		}
	}
	return nil, fmt.Errorf("platform e-commerce untuk URL ini belum didukung: %s", rawURL)
}