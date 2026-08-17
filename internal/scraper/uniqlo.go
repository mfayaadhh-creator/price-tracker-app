package scraper

import (
	"encoding/json"
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
			Timeout: 15 * time.Second,
		},
	}
}

func (s *UniqloScraper) CanHandle(rawURL string) bool {
	return strings.Contains(rawURL, "uniqlo.com")
}

func (s *UniqloScraper) ExtractProductID(rawURL string) (string, error) {
	re := regexp.MustCompile(`products/([A-Za-z]?[0-9]{6}-[0-9]{3})`)
	matches := re.FindStringSubmatch(rawURL)
	if len(matches) >= 2 {
		return matches[1], nil
	}
	return "", fmt.Errorf("gagal menemukan product ID dari URL Uniqlo")
}

// Struct untuk membaca data harga dari __PRELOADED_STATE__
type uqPriceDetail struct {
	Currency struct {
		Code   string `json:"code"`
		Symbol string `json:"symbol"`
	} `json:"currency"`
	Value float64 `json:"value"`
}

type uqPrices struct {
	Base  uqPriceDetail `json:"base"`
	Promo uqPriceDetail `json:"promo"`
}

type uqProduct struct {
	Name      string   `json:"name"`
	ProductID string   `json:"productId"`
	Prices    uqPrices `json:"prices"`
	Images    struct {
		Main map[string]struct {
			Image string `json:"image"`
			URL   string `json:"url"`
		} `json:"main"`
		Sub []struct {
			Image string `json:"image"`
			URL   string `json:"url"`
		} `json:"sub"`
	} `json:"images"`
	Representative struct {
		Color struct {
			DisplayCode string `json:"displayCode"`
		} `json:"color"`
	} `json:"representative"`
}

type uqProductWrapper struct {
	Product uqProduct `json:"product"`
}

// extractPreloadedState mengekstrak blok JSON __PRELOADED_STATE__ dari HTML Uniqlo
func extractPreloadedState(htmlContent string) ([]byte, error) {
	marker := "window.__PRELOADED_STATE__"
	startIdx := strings.Index(htmlContent, marker)
	if startIdx == -1 {
		return nil, fmt.Errorf("__PRELOADED_STATE__ tidak ditemukan di HTML")
	}

	afterMarker := htmlContent[startIdx+len(marker):]
	eqIdx := strings.Index(afterMarker, "=")
	if eqIdx == -1 {
		return nil, fmt.Errorf("karakter '=' tidak ditemukan setelah __PRELOADED_STATE__")
	}

	jsonStart := startIdx + len(marker) + eqIdx + 1
	remaining := htmlContent[jsonStart:]

	scriptEnd := strings.Index(remaining, "</script>")
	if scriptEnd == -1 {
		return nil, fmt.Errorf("tag </script> tidak ditemukan setelah __PRELOADED_STATE__")
	}

	jsonStr := strings.TrimSpace(remaining[:scriptEnd])
	jsonStr = strings.TrimRight(jsonStr, "; \n\r\t")

	return []byte(jsonStr), nil
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

	// 1. Ekstrak blok JSON __PRELOADED_STATE__ secara utuh
	stateJSON, err := extractPreloadedState(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("gagal mengekstrak state JSON: %w", err)
	}

	// 2. Parse top-level state
	var state map[string]json.RawMessage
	if err := json.Unmarshal(stateJSON, &state); err != nil {
		return nil, fmt.Errorf("gagal parse __PRELOADED_STATE__: %w", err)
	}

	// 3. Navigasi ke entity → pdpEntity
	var entity map[string]json.RawMessage
	if err := json.Unmarshal(state["entity"], &entity); err != nil {
		return nil, fmt.Errorf("gagal parse entity: %w", err)
	}

	var pdpEntity map[string]json.RawMessage
	if err := json.Unmarshal(entity["pdpEntity"], &pdpEntity); err != nil {
		return nil, fmt.Errorf("gagal parse pdpEntity: %w", err)
	}

	// 4. Ambil entry produk pertama dari pdpEntity
	var productData uqProductWrapper
	for _, rawProduct := range pdpEntity {
		if err := json.Unmarshal(rawProduct, &productData); err == nil && productData.Product.Name != "" {
			break
		}
	}

	if productData.Product.Name == "" {
		return nil, fmt.Errorf("data produk tidak ditemukan di pdpEntity")
	}

	prod := productData.Product
	basePrice := prod.Prices.Base.Value
	currentPrice := prod.Prices.Promo.Value

	if currentPrice == 0 {
		currentPrice = basePrice
	}
	if basePrice == 0 {
		basePrice = currentPrice
	}

	isDiscount := currentPrice < basePrice && currentPrice > 0

	// Ekstrak Foto Produk Resmi Uniqlo
	imageURL := ""
	repColor := prod.Representative.Color.DisplayCode
	if repColor != "" && prod.Images.Main != nil {
		if imgObj, ok := prod.Images.Main[repColor]; ok && imgObj.Image != "" {
			imageURL = imgObj.Image
		}
	}

	if imageURL == "" && prod.Images.Main != nil {
		for _, imgObj := range prod.Images.Main {
			if imgObj.Image != "" {
				imageURL = imgObj.Image
				break
			}
		}
	}

	if imageURL == "" && len(prod.Images.Sub) > 0 {
		imageURL = prod.Images.Sub[0].Image
	}

	return &domain.ProductInfo{
		Platform:     "Uniqlo",
		ProductID:    productID,
		Name:         prod.Name,
		ImageURL:     imageURL,
		BasePrice:    basePrice,
		CurrentPrice: currentPrice,
		IsDiscount:   isDiscount,
		URL:          rawURL,
		CheckedAt:    time.Now(),
	}, nil
}

