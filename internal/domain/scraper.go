package domain

type Scraper interface {
	CanHandle(rawURL string) bool
	FetchPrice(rawURL string) (*ProductInfo, error)
}