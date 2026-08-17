package scraper

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	http "github.com/bogdanfinn/fhttp"

	"price_tracker/internal/domain"
)

type UniversalScraper struct {
	client tls_client.HttpClient
}

func NewUniversalScraper() *UniversalScraper {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(25),
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		// Fallback to default options if error
		client, _ = tls_client.NewHttpClient(tls_client.NewNoopLogger())
	}

	return &UniversalScraper{
		client: client,
	}
}

func (s *UniversalScraper) CanHandle(rawURL string) bool {
	return strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://")
}

func (s *UniversalScraper) FetchPrice(rawURL string) (*domain.ProductInfo, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("URL tidak valid: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil halaman produk: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 || resp.StatusCode == 410 {
		return nil, fmt.Errorf("PRODUK_UNAVAILABLE: halaman produk tidak ditemukan (status %d)", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca data HTML: %w", err)
	}

	htmlContent := string(bodyBytes)

	// 0. Auto-Solver: Deteksi & Selesaikan Interstitial Challenge (Akamai WAF seperti pada Zara)
	if strings.Contains(htmlContent, "triggerInterstitialChallenge") {
		resolvedHTML, err := s.solveAkamaiChallenge(rawURL, htmlContent)
		if err == nil && len(resolvedHTML) > 0 {
			htmlContent = resolvedHTML
		}
	}

	var name, imageURL string
	var currentPrice, basePrice float64
	var sku string

	// 1. Ekstraksi Tier 1: Next.js __NEXT_DATA__ (Zalora, Tokopedia, dll)
	name, imageURL, currentPrice, basePrice = parseNextData(htmlContent)

	// 2. Ekstraksi Tier 2: Schema.org JSON-LD (Shopify, WooCommerce, H&M, dll)
	if name == "" || currentPrice == 0 {
		ldName, ldImg, ldPrice, ldBase, ldSku := parseJSONLD(htmlContent)
		if name == "" {
			name = ldName
		}
		if imageURL == "" {
			imageURL = ldImg
		}
		if currentPrice == 0 {
			currentPrice = ldPrice
		}
		if basePrice == 0 {
			basePrice = ldBase
		}
		if sku == "" {
			sku = ldSku
		}
	}

	// 3. Ekstraksi Tier 3: OpenGraph & Product Meta Tags
	if name == "" {
		name = extractMetaContent(htmlContent, "og:title", "twitter:title", "title", "name")
		if name == "" {
			name = extractTitleTag(htmlContent)
		}
	}

	if imageURL == "" || strings.HasSuffix(imageURL, "/") {
		ogImg := extractMetaContent(htmlContent, "og:image", "og:image:secure_url", "twitter:image")
		if ogImg != "" && !strings.HasSuffix(ogImg, "/") {
			imageURL = ogImg
		} else {
			imageURL = extractFallbackImage(htmlContent)
		}
	}

	if currentPrice == 0 {
		priceStr := extractMetaContent(htmlContent, "product:price:amount", "og:price:amount", "itemprop=\"price\"", "price", "twitter:data1")
		currentPrice = parsePrice(priceStr)
	}

	// 4. Ekstraksi Tier 4: Fallback Pencarian Regex Harga (Class-based & Embedded Analytics)
	if currentPrice == 0 {
		currentPrice, basePrice = extractPriceFromHTMLClasses(htmlContent)
	}

	if name == "" && currentPrice == 0 {
		return nil, fmt.Errorf("tidak dapat mengekstrak data produk (nama & harga tidak ditemukan di metadata web)")
	}

	if basePrice == 0 {
		basePrice = currentPrice
	}

	isDiscount := currentPrice < basePrice && currentPrice > 0

	// 5. Platform & Product ID
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
		CurrentPrice: currentPrice,
		IsDiscount:   isDiscount,
		IsAvailable:  true,
		URL:          rawURL,
		CheckedAt:    time.Now(),
	}, nil
}

func (s *UniversalScraper) solveAkamaiChallenge(rawURL string, challengeHTML string) (string, error) {
	reI := regexp.MustCompile(`var\s+i\s*=\s*([0-9]+);`)
	reJ := regexp.MustCompile(`var\s+j\s*=\s*i\s*\+\s*Number\("([0-9]+)"\s*\+\s*"([0-9]+)"\);`)
	reVerify := regexp.MustCompile(`"bm-verify":\s*"([^"]+)"`)

	matchI := reI.FindStringSubmatch(challengeHTML)
	matchJ := reJ.FindStringSubmatch(challengeHTML)
	matchVerify := reVerify.FindStringSubmatch(challengeHTML)

	if len(matchI) <= 1 || len(matchJ) <= 2 || len(matchVerify) <= 1 {
		return "", fmt.Errorf("pola challenge tidak cocok")
	}

	iVal, _ := strconv.ParseInt(matchI[1], 10, 64)
	numStr := matchJ[1] + matchJ[2]
	numVal, _ := strconv.ParseInt(numStr, 10, 64)
	powVal := iVal + numVal
	bmVerify := matchVerify[1]

	u, _ := url.Parse(rawURL)
	verifyURL := fmt.Sprintf("%s://%s/_sec/verify?provider=interstitial", u.Scheme, u.Host)
	payload := map[string]interface{}{
		"bm-verify": bmVerify,
		"pow":       powVal,
	}
	jsonPayload, _ := json.Marshal(payload)

	postReq, _ := http.NewRequest("POST", verifyURL, bytes.NewBuffer(jsonPayload))
	postReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Referer", rawURL)
	postReq.Header.Set("Origin", fmt.Sprintf("%s://%s", u.Scheme, u.Host))

	postResp, err := s.client.Do(postReq)
	if err != nil {
		return "", err
	}
	postResp.Body.Close()

	// Ambil kembali halaman produk yang sebenarnya setelah verifikasi
	finalReq, _ := http.NewRequest("GET", rawURL, nil)
	finalReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	finalReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	finalResp, err := s.client.Do(finalReq)
	if err != nil {
		return "", err
	}
	defer finalResp.Body.Close()

	finalBody, err := io.ReadAll(finalResp.Body)
	if err != nil {
		return "", err
	}

	return string(finalBody), nil
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

func extractMetaContent(htmlContent string, propertyNames ...string) string {
	for _, prop := range propertyNames {
		re1 := regexp.MustCompile(`(?i)<meta[^>]+(?:property|name|itemprop)=["']` + regexp.QuoteMeta(prop) + `["'][^>]+content=["']([^"']*)["']`)
		if match := re1.FindStringSubmatch(htmlContent); len(match) > 1 && strings.TrimSpace(match[1]) != "" {
			return html.UnescapeString(strings.TrimSpace(match[1]))
		}
		re2 := regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']*)["'][^>]+(?:property|name|itemprop)=["']` + regexp.QuoteMeta(prop) + `["']`)
		if match := re2.FindStringSubmatch(htmlContent); len(match) > 1 && strings.TrimSpace(match[1]) != "" {
			return html.UnescapeString(strings.TrimSpace(match[1]))
		}
	}
	return ""
}

func extractPriceFromHTMLClasses(htmlContent string) (currentPrice float64, basePrice float64) {
	// 1. Coba cari analytics data (seperti pada Zara: "mainPrice": 1699000)
	reAnalytics := regexp.MustCompile(`(?i)"mainPrice":\s*([0-9.]+)`)
	if m := reAnalytics.FindStringSubmatch(htmlContent); len(m) > 1 {
		p := parsePrice(m[1])
		if p > 0 {
			return p, p
		}
	}

	// 2. Coba cari span class="price..." atau "money..."
	rePrice := regexp.MustCompile(`(?i)(?:class="[^"]*(?:price|money|amount)[^"]*"[^>]*>|<span[^>]*class="[^"]*amount[^"]*"[^>]*>)([^<]+)<`)
	matches := rePrice.FindAllStringSubmatch(htmlContent, -1)
	for _, m := range matches {
		txt := strings.TrimSpace(m[1])
		if strings.Contains(txt, "IDR") || strings.Contains(txt, "Rp") || strings.Contains(txt, ".") || strings.Contains(txt, ",") {
			p := parsePrice(txt)
			if p > 1000 {
				return p, p
			}
		}
	}
	return 0, 0
}

func parseNextData(htmlContent string) (name string, img string, currentPrice float64, basePrice float64) {
	reNext := regexp.MustCompile(`(?is)<script[^>]*id=["']__NEXT_DATA__["'][^>]*>(.*?)</script>`)
	match := reNext.FindStringSubmatch(htmlContent)
	if len(match) < 2 {
		return "", "", 0, 0
	}

	var root map[string]interface{}
	if err := json.Unmarshal([]byte(match[1]), &root); err != nil {
		return "", "", 0, 0
	}

	prodMap := findProductMap(root)
	if prodMap != nil {
		if title, ok := prodMap["title"].(string); ok && title != "" {
			name = html.UnescapeString(title)
		} else if n, ok := prodMap["name"].(string); ok && n != "" {
			name = html.UnescapeString(n)
		}

		if imgStr, ok := prodMap["image"].(string); ok && imgStr != "" {
			img = imgStr
		}

		if sp, exists := prodMap["SpecialPrice"]; exists {
			currentPrice = parsePrice(fmt.Sprintf("%v", sp))
		}
		if p, exists := prodMap["Price"]; exists {
			basePrice = parsePrice(fmt.Sprintf("%v", p))
			if currentPrice == 0 {
				currentPrice = basePrice
			}
		} else if p, exists := prodMap["price"]; exists {
			basePrice = parsePrice(fmt.Sprintf("%v", p))
			if currentPrice == 0 {
				currentPrice = basePrice
			}
		}
	}

	return name, img, currentPrice, basePrice
}

func findProductMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		if _, hasProductKey := m["SpecialPrice"]; hasProductKey {
			return m
		}
		if _, hasProductKey := m["price"]; hasProductKey {
			if _, hasTitle := m["title"]; hasTitle {
				return m
			}
		}
		for _, child := range m {
			if res := findProductMap(child); res != nil {
				return res
			}
		}
	} else if arr, ok := v.([]interface{}); ok {
		for _, child := range arr {
			if res := findProductMap(child); res != nil {
				return res
			}
		}
	}
	return nil
}

func extractFallbackImage(htmlContent string) string {
	reImg := regexp.MustCompile(`https?://[^"'\s>]+\.(?:jpg|jpeg|png|webp)`)
	matches := reImg.FindAllString(htmlContent, -1)
	for _, m := range matches {
		if strings.Contains(m, "catalog/product/large") || strings.Contains(m, "static.zara.net") || strings.Contains(m, "imagesgoods") || strings.Contains(m, "static-id.zacdn.com/p/") {
			return m
		}
	}
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func extractTitleTag(htmlContent string) string {
	re := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	if match := re.FindStringSubmatch(htmlContent); len(match) > 1 {
		t := strings.TrimSpace(html.UnescapeString(match[1]))
		if idx := strings.Index(t, " | "); idx != -1 {
			t = t[:idx]
		} else if idx := strings.Index(t, " - "); idx != -1 {
			t = t[:idx]
		}
		return strings.TrimSpace(t)
	}
	return ""
}

func extractPlatform(rawURL string, htmlContent string) string {
	siteName := extractMetaContent(htmlContent, "og:site_name", "twitter:site")
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
		p := parts[0]
		if p == "id" && len(parts) > 1 {
			p = parts[1]
		}
		return strings.ToUpper(p)
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

func parseJSONLD(htmlContent string) (name string, img string, price float64, basePrice float64, sku string) {
	reScript := regexp.MustCompile(`(?is)<script[^>]*type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	matches := reScript.FindAllStringSubmatch(htmlContent, -1)

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

	if n, ok := m["name"].(string); ok {
		name = html.UnescapeString(n)
	}
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
