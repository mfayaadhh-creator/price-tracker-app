package domain

import "time"

type TrackedProduct struct {
	ID          string    `json:"id"`
	UserPhone   string    `json:"user_phone"`
	URL         string    `json:"url"`
	Platform    string    `json:"platform"`
	ProductID   string    `json:"product_id"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"image_url"`
	BasePrice   float64   `json:"base_price"`
	LastPrice   float64   `json:"last_price"`
	TargetPrice float64   `json:"target_price"`
	IsDiscount  bool      `json:"is_discount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}