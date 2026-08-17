package domain

// Scraper adalah kontrak yang harus diimplementasikan oleh Uniqlo, Zara, H&M, dll
type Scraper interface {
	CanHandle(rawURL string) bool
	FetchPrice(rawURL string) (*ProductInfo, error)
}

// Notifier adalah kontrak untuk mengirim pesan notifikasi (Telegram, WhatsApp, dll)
type Notifier interface {
	SendAlert(target string, message string) error
}