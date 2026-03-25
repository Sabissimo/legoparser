package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"lego-parser/internal/models"
)

type ScrapeRepo struct {
	pool *pgxpool.Pool
}

func NewScrapeRepo(pool *pgxpool.Pool) *ScrapeRepo {
	return &ScrapeRepo{pool: pool}
}

func (r *ScrapeRepo) Create(ctx context.Context, source models.SiteSource, triggerType string) (*models.ScrapeRun, error) {
	now := time.Now()
	var run models.ScrapeRun
	err := r.pool.QueryRow(ctx,
		`INSERT INTO scrape_runs (source, status, trigger_type, started_at)
		 VALUES ($1, 'running', $2, $3)
		 RETURNING id, source, status, trigger_type, products_found, products_saved,
				   errors_count, error_log, started_at, completed_at, created_at`,
		source, triggerType, now,
	).Scan(
		&run.ID, &run.Source, &run.Status, &run.TriggerType,
		&run.ProductsFound, &run.ProductsSaved, &run.ErrorsCount, &run.ErrorLog,
		&run.StartedAt, &run.CompletedAt, &run.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create scrape run: %w", err)
	}
	return &run, nil
}

func (r *ScrapeRepo) Complete(ctx context.Context, id int64, found, saved, errors int, errorLog string) error {
	status := "completed"
	if errors > 0 && saved == 0 {
		status = "failed"
	}

	var errLogPtr *string
	if errorLog != "" {
		errLogPtr = &errorLog
	}

	_, err := r.pool.Exec(ctx,
		`UPDATE scrape_runs SET
			status = $2, products_found = $3, products_saved = $4,
			errors_count = $5, error_log = $6, completed_at = NOW()
		 WHERE id = $1`,
		id, status, found, saved, errors, errLogPtr,
	)
	if err != nil {
		return fmt.Errorf("complete scrape run %d: %w", id, err)
	}
	return nil
}

func (r *ScrapeRepo) List(ctx context.Context, limit int) ([]models.ScrapeRun, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, source, status, trigger_type, products_found, products_saved,
				errors_count, error_log, started_at, completed_at, created_at
		 FROM scrape_runs
		 ORDER BY created_at DESC
		 LIMIT $1`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list scrape runs: %w", err)
	}
	defer rows.Close()

	var runs []models.ScrapeRun
	for rows.Next() {
		var run models.ScrapeRun
		if err := rows.Scan(
			&run.ID, &run.Source, &run.Status, &run.TriggerType,
			&run.ProductsFound, &run.ProductsSaved, &run.ErrorsCount, &run.ErrorLog,
			&run.StartedAt, &run.CompletedAt, &run.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan scrape run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}

func (r *ScrapeRepo) GetRunning(ctx context.Context) ([]models.ScrapeRun, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, source, status, trigger_type, products_found, products_saved,
				errors_count, error_log, started_at, completed_at, created_at
		 FROM scrape_runs
		 WHERE status = 'running'
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("get running scrapes: %w", err)
	}
	defer rows.Close()

	var runs []models.ScrapeRun
	for rows.Next() {
		var run models.ScrapeRun
		if err := rows.Scan(
			&run.ID, &run.Source, &run.Status, &run.TriggerType,
			&run.ProductsFound, &run.ProductsSaved, &run.ErrorsCount, &run.ErrorLog,
			&run.StartedAt, &run.CompletedAt, &run.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan running scrape: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}
