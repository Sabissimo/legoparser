package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"lego-parser/internal/models"
)

type ProductRepo struct {
	pool *pgxpool.Pool
}

func NewProductRepo(pool *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{pool: pool}
}

func (r *ProductRepo) Upsert(ctx context.Context, p *models.Product) (int64, error) {
	var existingID int64

	if p.Barcode != nil && *p.Barcode != "" {
		err := r.pool.QueryRow(ctx,
			"SELECT id FROM products WHERE barcode = $1", *p.Barcode,
		).Scan(&existingID)
		if err == nil {
			return r.update(ctx, existingID, p)
		}
	}

	if p.InvoiceCode != nil && *p.InvoiceCode != "" {
		err := r.pool.QueryRow(ctx,
			"SELECT id FROM products WHERE invoice_code = $1", *p.InvoiceCode,
		).Scan(&existingID)
		if err == nil {
			return r.update(ctx, existingID, p)
		}
	}

	// Fallback: match by name to prevent duplicates
	if p.NameKa != "" {
		err := r.pool.QueryRow(ctx,
			"SELECT id FROM products WHERE name_ka = $1", p.NameKa,
		).Scan(&existingID)
		if err == nil {
			return r.update(ctx, existingID, p)
		}
	}

	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO products (mcode, barcode, invoice_code, name_ka, name_en, image_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		p.MCode, p.Barcode, p.InvoiceCode, p.NameKa, p.NameEn, p.ImageURL,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert product: %w", err)
	}
	return id, nil
}

func (r *ProductRepo) update(ctx context.Context, id int64, p *models.Product) (int64, error) {
	_, err := r.pool.Exec(ctx,
		`UPDATE products SET
			mcode = COALESCE($2, mcode),
			barcode = COALESCE($3, barcode),
			invoice_code = COALESCE($4, invoice_code),
			name_ka = COALESCE(NULLIF($5, ''), name_ka),
			name_en = COALESCE($6, name_en),
			image_url = COALESCE($7, image_url),
			updated_at = NOW()
		 WHERE id = $1`,
		id, p.MCode, p.Barcode, p.InvoiceCode, p.NameKa, p.NameEn, p.ImageURL,
	)
	if err != nil {
		return 0, fmt.Errorf("update product %d: %w", id, err)
	}
	return id, nil
}

var allowedSortColumns = map[string]string{
	"name_ka":      "p.name_ka",
	"name_en":      "p.name_en",
	"barcode":      "p.barcode",
	"invoice_code": "p.invoice_code",
	"mcode":        "p.mcode",
	"created_at":   "p.created_at",
	"updated_at":   "p.updated_at",
}

func (r *ProductRepo) List(ctx context.Context, params models.QueryParams) ([]models.Product, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(p.name_ka ILIKE $%d OR p.name_en ILIKE $%d OR p.invoice_code ILIKE $%d OR p.barcode ILIKE $%d OR p.mcode ILIKE $%d)",
			argIdx, argIdx, argIdx, argIdx, argIdx,
		))
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p %s", where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	sortCol, ok := allowedSortColumns[params.SortBy]
	if !ok {
		sortCol = "p.name_ka"
	}
	sortDir := "ASC NULLS LAST"
	if params.SortOrder == "desc" {
		sortDir = "DESC NULLS LAST"
	}

	offset := (params.Page - 1) * params.PerPage

	query := fmt.Sprintf(`
		SELECT p.id, p.mcode, p.barcode, p.invoice_code,
			   p.name_ka, p.name_en, p.image_url,
			   COALESCE((SELECT json_agg(DISTINCT source::text ORDER BY source::text) FROM latest_prices lp WHERE lp.product_id = p.id), '[]'::json) AS sources,
			   p.created_at, p.updated_at
		FROM products p
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, argIdx, argIdx+1)

	args = append(args, params.PerPage, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		var sourcesJSON []byte
		if err := rows.Scan(
			&p.ID, &p.MCode, &p.Barcode, &p.InvoiceCode,
			&p.NameKa, &p.NameEn, &p.ImageURL, &sourcesJSON,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan product: %w", err)
		}
		json.Unmarshal(sourcesJSON, &p.Sources)
		if p.Sources == nil {
			p.Sources = []string{}
		}
		products = append(products, p)
	}

	return products, total, nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	var p models.Product
	var sourcesJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, mcode, barcode, invoice_code,
				name_ka, name_en, image_url,
				COALESCE((SELECT json_agg(DISTINCT source::text ORDER BY source::text) FROM latest_prices lp WHERE lp.product_id = products.id), '[]'::json),
				created_at, updated_at
		 FROM products WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.MCode, &p.Barcode, &p.InvoiceCode,
		&p.NameKa, &p.NameEn, &p.ImageURL, &sourcesJSON,
		&p.CreatedAt, &p.UpdatedAt,
	)
	json.Unmarshal(sourcesJSON, &p.Sources)
	if p.Sources == nil {
		p.Sources = []string{}
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get product %d: %w", id, err)
	}
	return &p, nil
}

func (r *ProductRepo) GetComparison(ctx context.Context, params models.QueryParams) ([]models.PriceComparison, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(name_ka ILIKE $%d OR name_en ILIKE $%d OR invoice_code ILIKE $%d OR barcode ILIKE $%d)",
			argIdx, argIdx, argIdx, argIdx,
		))
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM price_comparison %s", where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count comparison: %w", err)
	}

	compSortCols := map[string]string{
		"name_ka":               "name_ka",
		"name_en":               "name_en",
		"barcode":               "barcode",
		"invoice_code":          "invoice_code",
		"mcode":                 "mcode",
		"biblusi_xs_price":      "biblusi_xs_price",
		"biblusi_pepela_price":  "biblusi_pepela_price",
		"wolt_xs_price":         "wolt_xs_price",
		"wolt_pepela_price":     "wolt_pepela_price",
		"glovo_xs_price":        "glovo_xs_price",
		"glovo_pepela_price":    "glovo_pepela_price",
		"wishlist_price":        "wishlist_price",
		"piccolatoys_price":     "piccolatoys_price",
		"kubiki_price":          "kubiki_price",
	}
	sortCol := "name_ka"
	if col, ok := compSortCols[params.SortBy]; ok {
		sortCol = col
	}
	sortDir := "ASC NULLS LAST"
	if params.SortOrder == "desc" {
		sortDir = "DESC NULLS LAST"
	}

	offset := (params.Page - 1) * params.PerPage

	query := fmt.Sprintf(`
		SELECT product_id, mcode, barcode, invoice_code, name_ka, name_en,
			   biblusi_xs_price, biblusi_pepela_price,
			   wolt_xs_price, wolt_pepela_price,
			   glovo_xs_price, glovo_pepela_price,
			   wishlist_price, piccolatoys_price, kubiki_price
		FROM price_comparison
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, argIdx, argIdx+1)

	args = append(args, params.PerPage, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list comparison: %w", err)
	}
	defer rows.Close()

	var results []models.PriceComparison
	for rows.Next() {
		var c models.PriceComparison
		if err := rows.Scan(
			&c.ProductID, &c.MCode, &c.Barcode, &c.InvoiceCode,
			&c.NameKa, &c.NameEn,
			&c.BiblusiXSPrice, &c.BiblusiPepelaPrice,
			&c.WoltXSPrice, &c.WoltPepelaPrice,
			&c.GlovoXSPrice, &c.GlovoPepelaPrice,
			&c.WishlistPrice, &c.PiccolaToysPrice, &c.KubikiPrice,
		); err != nil {
			return nil, 0, fmt.Errorf("scan comparison: %w", err)
		}
		results = append(results, c)
	}

	return results, total, nil
}
