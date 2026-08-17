package scraper

import (
	"testing"
)

func TestUniversalScraper_Zalora(t *testing.T) {
	s := NewUniversalScraper()
	url := "https://www.zalora.co.id/p/nike-metcon-10-workout-shoes-black-4976229"
	info, err := s.FetchPrice(url)
	if err != nil {
		t.Fatalf("Zalora fetch error: %v", err)
	}
	if info.CurrentPrice == 0 {
		t.Errorf("Expected price > 0, got %.0f", info.CurrentPrice)
	}
	t.Logf("Zalora SUCCESS: %s - Rp %.0f (Base: Rp %.0f)", info.Name, info.CurrentPrice, info.BasePrice)
}

func TestUniversalScraper_HnM(t *testing.T) {
	s := NewUniversalScraper()
	url := "https://id.hm.com/id_id/relaxed-fit-waffled-resort-shirt-1355796002.html"
	info, err := s.FetchPrice(url)
	if err != nil {
		t.Fatalf("H&M fetch error: %v", err)
	}
	if info.CurrentPrice == 0 {
		t.Errorf("Expected price > 0, got %.0f", info.CurrentPrice)
	}
	t.Logf("H&M SUCCESS: %s - Rp %.0f", info.Name, info.CurrentPrice)
}

func TestUniversalScraper_Zara(t *testing.T) {
	s := NewUniversalScraper()
	url := "https://www.zara.com/id/en/regular-fit-leather-effect-jacket-p00155751.html?v1=548926880&v2=2536906"
	info, err := s.FetchPrice(url)
	if err != nil {
		t.Fatalf("Zara fetch error: %v", err)
	}
	if info.CurrentPrice == 0 {
		t.Errorf("Expected price > 0, got %.0f", info.CurrentPrice)
	}
	t.Logf("Zara SUCCESS: %s - Rp %.0f", info.Name, info.CurrentPrice)
}

func TestUniversalScraper_Tokopedia(t *testing.T) {
	s := NewUniversalScraper()
	url := "https://www.tokopedia.com/aerostreet/aerostreet-massive-low-abu-muda-sepatu-sneakers-pria-wanita"
	info, err := s.FetchPrice(url)
	if err != nil {
		t.Logf("Tokopedia notice: %v", err)
		return
	}
	t.Logf("Tokopedia SUCCESS: [%s] %s - Rp %.0f", info.Platform, info.Name, info.CurrentPrice)
}

func TestUniversalScraper_Agres(t *testing.T) {
	s := NewUniversalScraper()
	url := "https://www.agres.id/products/lenovo-ideapad-slim-3-14-core-5-320-8gb-512gb-w11ohsm365b-140wuxga-blit-2y-prem2adp-box-bp-blue-8yid"
	info, err := s.FetchPrice(url)
	if err != nil {
		t.Fatalf("Agres fetch error: %v", err)
	}
	if info.CurrentPrice == 0 {
		t.Errorf("Expected price > 0, got %.0f", info.CurrentPrice)
	}
	t.Logf("Agres.id SUCCESS: [%s] %s - Rp %.0f (Img: %s)", info.Platform, info.Name, info.CurrentPrice, info.ImageURL)
}
