package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"lego-parser/internal/models"
	"lego-parser/internal/repository"
)

// ScrapedProduct is the intermediate struct returned by each scraper
type ScrapedProduct struct {
	MCode           string
	Barcode         string
	InvoiceCode     string
	NameKa          string
	NameEn          string
	ImageURL        string
	OriginalPrice   *float64
	DiscountPercent *float64
	DiscountedPrice *float64
	InStock         bool
	SourceURL       string
	SourceProductID string
}

// Scraper is the interface each site scraper must implement
type Scraper interface {
	Name() string
	Source() models.SiteSource
	Scrape(ctx context.Context, run *models.ScrapeRun) ([]ScrapedProduct, error)
}

// Orchestrator manages all scrapers and persists results
type Orchestrator struct {
	scrapers    map[string]Scraper
	productRepo *repository.ProductRepo
	priceRepo   *repository.PriceRepo
	scrapeRepo  *repository.ScrapeRepo
	logger      *slog.Logger

	mu      sync.Mutex
	running bool
}

func NewOrchestrator(
	productRepo *repository.ProductRepo,
	priceRepo *repository.PriceRepo,
	scrapeRepo *repository.ScrapeRepo,
	logger *slog.Logger,
) *Orchestrator {
	return &Orchestrator{
		scrapers:    make(map[string]Scraper),
		productRepo: productRepo,
		priceRepo:   priceRepo,
		scrapeRepo:  scrapeRepo,
		logger:      logger,
	}
}

func (o *Orchestrator) Register(s Scraper) {
	o.scrapers[s.Name()] = s
}

func (o *Orchestrator) IsRunning() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.running
}

func (o *Orchestrator) RunAll(ctx context.Context, triggerType string) error {
	o.mu.Lock()
	if o.running {
		o.mu.Unlock()
		return fmt.Errorf("scraper already running")
	}
	o.running = true
	o.mu.Unlock()
	defer func() {
		o.mu.Lock()
		o.running = false
		o.mu.Unlock()
	}()

	// Run scrapers in order: biblusi first (seed), then others
	order := []string{
		"biblusi_xs", "biblusi_pepela",
		"wishlist", "piccolatoys", "kubiki",
		"wolt_xs", "wolt_pepela",
		"glovo_xs", "glovo_pepela",
	}
	for _, name := range order {
		s, ok := o.scrapers[name]
		if !ok {
			continue
		}
		if err := o.runOne(ctx, s, triggerType); err != nil {
			o.logger.Error("scraper failed", "source", name, "error", err)
		}
	}

	return nil
}

func (o *Orchestrator) RunSource(ctx context.Context, source string, triggerType string) error {
	o.mu.Lock()
	if o.running {
		o.mu.Unlock()
		return fmt.Errorf("scraper already running")
	}
	o.running = true
	o.mu.Unlock()
	defer func() {
		o.mu.Lock()
		o.running = false
		o.mu.Unlock()
	}()

	s, ok := o.scrapers[source]
	if !ok {
		return fmt.Errorf("unknown scraper source: %s", source)
	}

	return o.runOne(ctx, s, triggerType)
}

func (o *Orchestrator) runOne(ctx context.Context, s Scraper, triggerType string) error {
	o.logger.Info("starting scraper", "source", s.Name())

	run, err := o.scrapeRepo.Create(ctx, s.Source(), triggerType)
	if err != nil {
		return fmt.Errorf("create scrape run: %w", err)
	}

	products, err := s.Scrape(ctx, run)
	if err != nil {
		errMsg := err.Error()
		_ = o.scrapeRepo.Complete(ctx, run.ID, 0, 0, 1, errMsg)
		return fmt.Errorf("scrape %s: %w", s.Name(), err)
	}

	saved := 0
	var errorMsgs []string

	for _, sp := range products {
		product := &models.Product{
			NameKa:   sp.NameKa,
			ImageURL: strPtr(sp.ImageURL),
		}
		if sp.MCode != "" {
			product.MCode = strPtr(sp.MCode)
		}
		if sp.Barcode != "" {
			product.Barcode = strPtr(sp.Barcode)
		}
		if sp.InvoiceCode != "" {
			product.InvoiceCode = strPtr(sp.InvoiceCode)
		}
		if sp.NameEn != "" {
			product.NameEn = strPtr(sp.NameEn)
		}

		productID, err := o.productRepo.Upsert(ctx, product)
		if err != nil {
			errorMsgs = append(errorMsgs, fmt.Sprintf("upsert product %q: %v", sp.NameKa, err))
			continue
		}

		price := &models.Price{
			ProductID:       productID,
			Source:          s.Source(),
			OriginalPrice:   sp.OriginalPrice,
			DiscountPercent: sp.DiscountPercent,
			DiscountedPrice: sp.DiscountedPrice,
			InStock:         sp.InStock,
			SourceURL:       strPtr(sp.SourceURL),
			SourceProductID: strPtr(sp.SourceProductID),
			ScrapeRunID:     run.ID,
		}

		if _, err := o.priceRepo.Insert(ctx, price); err != nil {
			errorMsgs = append(errorMsgs, fmt.Sprintf("insert price for %q: %v", sp.NameKa, err))
			continue
		}

		saved++
	}

	errLog := strings.Join(errorMsgs, "\n")
	if err := o.scrapeRepo.Complete(ctx, run.ID, len(products), saved, len(errorMsgs), errLog); err != nil {
		o.logger.Error("failed to complete scrape run", "error", err)
	}

	o.logger.Info("scraper completed", "source", s.Name(), "found", len(products), "saved", saved, "errors", len(errorMsgs))
	return nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
