package scraper

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"price_tracker/internal/domain"
)

type UniversalScraper struct {
	httpClient *http.Client
}

func NewUniversalScraper() *UniversalScraper {
	return &UniversalScraper{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *UniversalScraper) CanHandle(rawURL string) bool {
	// Universal Scraper bisa menangani URL http/https apa pun sebagai fallback
	return strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://")
}

func (s *UniversalScraper) FetchPrice(rawURL string) (*domain.ProductInfo, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("URL tidak valid: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil halaman produk: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return nil, fmt.Errorf("PRODUK_UNAVAILABLE: halaman produk tidak ditemukan (status %d)", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("halaman web mengembalikan status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca HTML: %w", err)
	}

	htmlContent := string(bodyBytes)

	// 1. Ekstraksi Tier 1: Schema.org JSON-LD
	name, imageURL, price, basePrice, sku := parseJSONLD(htmlContent)

	// 2. Ekstraksi Tier 2: OpenGraph & Meta Tags Fallback
	if name == "" {
		name = extractMetaContent(htmlContent, "og:title", "twitter:title")
		if name == "" {
			name = extractTitleTag(htmlContent)
		}
	}

	if imageURL == "" {
		imageURL = extractMetaContent(htmlContent, "og:image", "og:image:secure_url", "twitter:image")
	}

	if price == 0 {
		priceStr := extractMetaContent(htmlContent, "product:price:amount", "og:price:amount", "price", "twitter:data1")
		price = parsePrice(priceStr)
	}

	if name == "" && price == 0 {
		return nil, fmt.Errorf("tidak dapat mengekstrak data produk (nama & harga tidak ditemukan di metadata web)")
	}

	if basePrice == 0 {
		basePrice = price
	}

	isDiscount := price < basePrice && price > 0

	// 3. Platform & Product ID Fallback
	platform := extractPlatform(rawURL, htmlContent)
	productID := sku
	if productID == "" {
		productID = generateProductIDFromURL(rawURL)
	}

	return &domain.ProductInfo{
		Platform:     platform,
		ProductID:    productID,
		Name:         name,
		ImageURL:     imageURL,
		BasePrice:    basePrice,
		CurrentPrice: price,
		IsDiscount:   isDiscount,
		IsAvailable:  true,
		URL:          rawURL,
		CheckedAt:    time.Now(),
	}, nil
}

func parsePrice(val string) float64 {
	val = strings.TrimSpace(val)
	reClean := regexp.MustCompile(`(?i)(rp\.?|idr|usd|\$|eur|€|sgd|myr|\s+)`)
	cleaned := reClean.ReplaceAllString(val, "")

	if strings.Contains(cleaned, ".") && strings.Contains(cleaned, ",") {
		cleaned = strings.ReplaceAll(cleaned, ".", "")
		cleaned = strings.ReplaceAll(cleaned, ",", ".")
	} else if strings.Contains(cleaned, ".") && !strings.Contains(cleaned, ",") {
		parts := strings.Split(cleaned, ".")
		if len(parts) == 2 && len(parts[1]) == 3 {
			cleaned = parts[0] + parts[1]
		}
	} else if strings.Contains(cleaned, ",") {
		cleaned = strings.ReplaceAll(cleaned, ",", "")
	}

	f, _ := strconv.ParseFloat(cleaned, 64)
	return f
}

func extractMetaContent(html string, propertyNames ...string) string {
	for _, prop := range propertyNames {
		re1 := regexp.MustCompile(`(?i)<meta[^>]+(?:property|name)=["']` + regexp.QuoteMeta(prop) + `["'][^>]+content=["']([^"']*)["']`)
		if match := re1.FindStringSubmatch(html); len(match) > 1 && strings.TrimSpace(match[1]) != "" {
			return strings.TrimSpace(match[1])
		}
		re2 := regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']*)["'][^>]+(?:property|name)=["']` + regexp.QuoteMeta(prop) + `["']`)
		if match := re2.FindStringSubmatch(html); len(match) > 1 && strings.TrimSpace(match[1]) != "" {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func extractTitleTag(html string) string {
	re := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	if match := re.FindStringSubmatch(html); len(match) > 1 {
		t := strings.TrimSpace(match[1])
		if idx := strings.Index(t, " | "); idx != -1 {
			t = t[:idx]
		} else if idx := strings.Index(t, " - "); idx != -1 {
			t = t[:idx]
		}
		return strings.TrimSpace(t)
	}
	return ""
}

func extractPlatform(rawURL string, html string) string {
	siteName := extractMetaContent(html, "og:site_name", "twitter:site")
	if siteName != "" {
		return strings.ToUpper(siteName)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "E-COMMERCE"
	}
	host := strings.TrimPrefix(u.Hostname(), "www.")
	parts := strings.Split(host, ".")
	if len(parts) > 0 && parts[0] != "" {
		return strings.ToUpper(parts[0])
	}
	return "E-COMMERCE"
}

func generateProductIDFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err == nil {
		segments := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(segments) > 0 && len(segments[len(segments)-1]) >= 4 {
			return segments[len(segments)-1]
		}
	}
	hash := md5.Sum([]byte(rawURL))
	return hex.EncodeToString(hash[:])[:8]
}

func parseJSONLD(html string) (name string, img string, price float64, basePrice float64, sku string) {
	reScript := regexp.MustCompile(`(?is)<script[^>]*type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	matches := reScript.FindAllStringSubmatch(html, -1)

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		rawJSON := strings.TrimSpace(m[1])

		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(rawJSON), &obj); err == nil {
			if n, i, p, bp, s := extractFromMap(obj); n != "" && p > 0 {
				return n, i, p, bp, s
			}
		}

		var arr []map[string]interface{}
		if err := json.Unmarshal([]byte(rawJSON), &arr); err == nil {
			for _, item := range arr {
				if n, i, p, bp, s := extractFromMap(item); n != "" && p > 0 {
					return n, i, p, bp, s
				}
			}
		}
	}
	return "", "", 0, 0, ""
}

func extractFromMap(m map[string]interface{}) (name string, img string, price float64, basePrice float64, sku string) {
	if graph, ok := m["@graph"].([]interface{}); ok {
		for _, g := range graph {
			if gm, ok := g.(map[string]interface{}); ok {
				if n, i, p, bp, s := extractFromMap(gm); n != "" && p > 0 {
					return n, i, p, bp, s
				}
			}
		}
	}

	typeVal, _ := m["@type"].(string)
	if !strings.EqualFold(typeVal, "Product") && !strings.Contains(strings.ToLower(typeVal), "product") {
		return "", "", 0, 0, ""
	}

	name, _ = m["name"].(string)
	if s, ok := m["sku"].(string); ok {
		sku = s
	}

	if imgStr, ok := m["image"].(string); ok {
		img = imgStr
	} else if imgArr, ok := m["image"].([]interface{}); ok && len(imgArr) > 0 {
		if s, ok := imgArr[0].(string); ok {
			img = s
		}
	} else if imgObj, ok := m["image"].(map[string]interface{}); ok {
		if u, ok := imgObj["url"].(string); ok {
			img = u
		}
	}

	if offers, ok := m["offers"].(map[string]interface{}); ok {
		if pVal, exists := offers["price"]; exists {
			price = parsePrice(fmt.Sprintf("%v", pVal))
		}
		if pVal, exists := offers["lowPrice"]; exists && price == 0 {
			price = parsePrice(fmt.Sprintf("%v", pVal))
		}
		if pVal, exists := offers["highPrice"]; exists {
			basePrice = parsePrice(fmt.Sprintf("%v", pVal))
		}
	} else if offersArr, ok := m["offers"].([]interface{}); ok && len(offersArr) > 0 {
		if off0, ok := offersArr[0].(map[string]interface{}); ok {
			if pVal, exists := off0["price"]; exists {
				price = parsePrice(fmt.Sprintf("%v", pVal))
			}
		}
	}

	return name, img, price, basePrice, sku
}
