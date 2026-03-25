package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"lego-parser/internal/models"
)

type PriceRepo struct {
	pool *pgxpool.Pool
}

func NewPriceRepo(pool *pgxpool.Pool) *PriceRepo {
	return &PriceRepo{pool: pool}
}

func (r *PriceRepo) Insert(ctx context.Context, p *models.Price) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO prices (product_id, source, original_price, discount_percent,
			discounted_price, in_stock, source_url, source_product_id, scrape_run_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		p.ProductID, p.Source, p.OriginalPrice, p.DiscountPercent,
		p.DiscountedPrice, p.InStock, p.SourceURL, p.SourceProductID, p.ScrapeRunID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert price: %w", err)
	}
	return id, nil
}

func (r *PriceRepo) GetLatestByProductID(ctx context.Context, productID int64) ([]models.Price, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, source, original_price, discount_percent,
				discounted_price, in_stock, source_url, scraped_at
		 FROM latest_prices
		 WHERE product_id = $1`, productID,
	)
	if err != nil {
		return nil, fmt.Errorf("get latest prices: %w", err)
	}
	defer rows.Close()

	var prices []models.Price
	for rows.Next() {
		var p models.Price
		if err := rows.Scan(
			&p.ID, &p.ProductID, &p.Source, &p.OriginalPrice, &p.DiscountPercent,
			&p.DiscountedPrice, &p.InStock, &p.SourceURL, &p.ScrapedAt,
		); err != nil {
			return nil, fmt.Errorf("scan price: %w", err)
		}
		prices = append(prices, p)
	}

	return prices, nil
}
