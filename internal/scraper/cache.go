package scraper

import (
	"sync"
	"time"

	"price_tracker/internal/domain"
)

type cacheItem struct {
	info      *domain.ProductInfo
	expiresAt time.Time
}

// ScrapeCache menyediakan in-memory caching thread-safe untuk hasil scraping
type ScrapeCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
	ttl   time.Duration
}

// NewScrapeCache membuat instance cache baru dengan durasi TTL tertentu (default 5-10 menit)
func NewScrapeCache(ttl time.Duration) *ScrapeCache {
	c := &ScrapeCache{
		items: make(map[string]cacheItem),
		ttl:   ttl,
	}

	// Background goroutine untuk membersihkan item yang sudah kadaluarsa berkala
	go func() {
		ticker := time.NewTicker(ttl * 2)
		for range ticker.C {
			c.cleanupExpired()
		}
	}()

	return c
}

// Get mengambil info produk dari cache jika masih valid
func (c *ScrapeCache) Get(url string) (*domain.ProductInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[url]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		return nil, false
	}

	// Kembalikan salinan shallow copy agar thread-safe
	copied := *item.info
	return &copied, true
}

// Set menyimpan info produk ke dalam cache dengan batas waktu TTL
func (c *ScrapeCache) Set(url string, info *domain.ProductInfo) {
	if info == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[url] = cacheItem{
		info:      info,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// cleanupExpired menghapus item yang sudah lewat masa berlakunya
func (c *ScrapeCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, v := range c.items {
		if now.After(v.expiresAt) {
			delete(c.items, k)
		}
	}
}
