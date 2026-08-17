package domain

import "time"

type ProductInfo struct {
	Platform     string    `json:"platform"`
	ProductID    string    `json:"product_id"`
	Name         string    `json:"name"`
	ImageURL     string    `json:"image_url"`
	BasePrice    float64   `json:"base_price"`
	CurrentPrice float64   `json:"current_price"`
	IsDiscount   bool      `json:"is_discount"`
	URL          string    `json:"url"`
	CheckedAt    time.Time `json:"checked_at"`
}
