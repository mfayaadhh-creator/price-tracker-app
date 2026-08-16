package scraper

import (
	"fmt"
	"io"
	"net/http"
	"price_tracker/internal/domain"
	"regexp"
	"strings"
	"time"
)

type UniqloScraper struct {
	httpClient *http.Client
}

func NewUniqloScraper() *UniqloScraper {
	return &UniqloScraper{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *UniqloScraper) CanHandle(rawURL string) bool {
	return strings.Contains(rawURL, "uniqlo.com")
}

func (s *UniqloScraper) ExtractProductID(rawURL string) (string, error) {
	re := regexp.MustCompile(`products/[A-Za-z]*([0-9]{6})`)
	matches := re.FindStringSubmatch(rawURL)
	if len(matches) >= 2 {
		return matches[1], nil
	}
	return "", fmt.Errorf("gagal menemukan product ID dari URL Uniqlo")
}

func (s *UniqloScraper) FetchPrice(rawURL string) (*domain.ProductInfo, error) {
	productID, err := s.ExtractProductID(rawURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil web Uniqlo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("halaman Uniqlo mengembalikan status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca HTML Uniqlo: %w", err)
	}

	htmlContent := string(bodyBytes)

	// 1. Ekstrak nama produk yang valid dari JSON-state
	name := ""
	reName := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`)
	matchesName := reName.FindAllStringSubmatch(htmlContent, 30)
	ignoredNames := map[string]bool{
		"UNIQLO APP": true, "StyleHint APP": true, "tops": true, "t-shirts": true,
		"crew neck": true, "WHITE": true, "BLACK": true, "Uniseks": true,
	}
	for _, m := range matchesName {
		if len(m) >= 2 {
			candidate := strings.TrimSpace(m[1])
			if !ignoredNames[candidate] && !strings.Contains(candidate, "http") && len(candidate) > 3 {
				name = candidate
				break
			}
		}
	}

	// 2. Ekstrak Base Price & Promo Price
	var basePrice, currentPrice float64

	// Pattern Base Price
	reBasePrice := regexp.MustCompile(`"base"\s*:\s*\{"currency"[^}]+\},\s*"value"\s*:\s*([0-9]+)`)
	if m := reBasePrice.FindStringSubmatch(htmlContent); len(m) >= 2 {
		fmt.Sscanf(m[1], "%f", &basePrice)
	}

	// Pattern Promo Price (jika diskon)
	rePromoPrice := regexp.MustCompile(`"promo"\s*:\s*\{"currency"[^}]+\},\s*"value"\s*:\s*([0-9]+)`)
	if m := rePromoPrice.FindStringSubmatch(htmlContent); len(m) >= 2 {
		fmt.Sscanf(m[1], "%f", &currentPrice)
	}

	// Jika tidak sedang promo, currentPrice = basePrice
	if currentPrice == 0 {
		currentPrice = basePrice
	}
	if basePrice == 0 {
		basePrice = currentPrice
	}

	isDiscount := currentPrice < basePrice && currentPrice > 0

	return &domain.ProductInfo{
		Platform:     "Uniqlo",
		ProductID:    productID,
		Name:         name,
		BasePrice:    basePrice,
		CurrentPrice: currentPrice,
		IsDiscount:   isDiscount,
		URL:          rawURL,
		CheckedAt:    time.Now(),
	}, nil
}
