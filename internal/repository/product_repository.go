package repository

import (
	"context"
	"price_tracker/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(ctx context.Context, connString string) (*ProductRepository, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	// Auto-migration: pastikan kolom image_url ada
	_, _ = pool.Exec(ctx, "ALTER TABLE tracked_products ADD COLUMN IF NOT EXISTS image_url TEXT DEFAULT '';")

	return &ProductRepository{pool: pool}, nil
}

func (r *ProductRepository) Close() {
	r.pool.Close()
}

func (r *ProductRepository) AddTrackedProduct(ctx context.Context, tp domain.TrackedProduct) (string, error) {
	query := `INSERT INTO tracked_products (user_phone, url, platform, product_id, name, image_url, base_price, last_price, target_price, is_discount)
                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
                  RETURNING id`
	var id string
	err := r.pool.QueryRow(ctx, query, tp.UserPhone, tp.URL, tp.Platform, tp.ProductID,
		tp.Name, tp.ImageURL, tp.BasePrice, tp.LastPrice, tp.TargetPrice, tp.IsDiscount,
	).Scan(&id)
	return id, err
}

func (r *ProductRepository) GetAllTrackedProducts(ctx context.Context) ([]domain.TrackedProduct, error) {
	query := `SELECT id, user_phone, url, platform, product_id, name,
  COALESCE(image_url, ''), base_price, last_price, target_price, is_discount, created_at,
  updated_at
                  FROM tracked_products ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.TrackedProduct
	for rows.Next() {
		var tp domain.TrackedProduct
		err := rows.Scan(&tp.ID, &tp.UserPhone, &tp.URL, &tp.Platform,
			&tp.ProductID, &tp.Name, &tp.ImageURL, &tp.BasePrice, &tp.LastPrice,
			&tp.TargetPrice, &tp.IsDiscount, &tp.CreatedAt, &tp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, tp)
	}
	return products, nil
}

func (r *ProductRepository) GetTrackedProductsByUser(ctx context.Context, userPhone string) ([]domain.TrackedProduct, error) {
	query := `SELECT id, user_phone, url, platform, product_id, name,
  COALESCE(image_url, ''), base_price, last_price, target_price, is_discount, created_at,
  updated_at
                  FROM tracked_products WHERE user_phone = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, userPhone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.TrackedProduct
	for rows.Next() {
		var tp domain.TrackedProduct
		err := rows.Scan(&tp.ID, &tp.UserPhone, &tp.URL, &tp.Platform,
			&tp.ProductID, &tp.Name, &tp.ImageURL, &tp.BasePrice, &tp.LastPrice,
			&tp.TargetPrice, &tp.IsDiscount, &tp.CreatedAt, &tp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, tp)
	}
	return products, nil
}
func (r *ProductRepository) UpdateLastPrice(ctx context.Context, id string, lastPrice float64, isDiscount bool) error {
	query := `UPDATE tracked_products SET last_price = $1, is_discount = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, lastPrice, isDiscount, id)
	return err
}

func (r *ProductRepository) RemoveTrackedProduct(ctx context.Context, id string) error {
	query := `DELETE FROM tracked_products WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
