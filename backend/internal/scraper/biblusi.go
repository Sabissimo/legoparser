package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"lego-parser/internal/config"
	"lego-parser/internal/models"
)

// BiblusiScraper scrapes LEGO products from Biblusi.ge REST API
type BiblusiScraper struct {
	name       string
	source     models.SiteSource
	categoryID int
	client     *http.Client
	limiter    *RateLimiter
	logger     *slog.Logger
}

func NewBiblusiXSScraper(cfg *config.Config, logger *slog.Logger) *BiblusiScraper {
	return newBiblusiScraper("biblusi_xs", models.SourceBiblusiXS, 456, logger)
}

func NewBiblusiPepelaScraper(cfg *config.Config, logger *slog.Logger) *BiblusiScraper {
	return newBiblusiScraper("biblusi_pepela", models.SourceBiblusiPepela, 474, logger)
}

func newBiblusiScraper(name string, source models.SiteSource, catID int, logger *slog.Logger) *BiblusiScraper {
	return &BiblusiScraper{
		name:       name,
		source:     source,
		categoryID: catID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: NewRateLimiter(200 * time.Millisecond),
		logger:  logger,
	}
}

func (s *BiblusiScraper) Name() string             { return s.name }
func (s *BiblusiScraper) Source() models.SiteSource { return s.source }

// Biblusi API response structures
type biblusiListResponse struct {
	CurrentPage int            `json:"current_page"`
	LastPage    int            `json:"last_page"`
	Total       int            `json:"total"`
	PerPage     int            `json:"per_page"`
	Data        []biblusiBook  `json:"data"`
}

type biblusiBook struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	CategoryID  int                `json:"category_id"`
	Price       json.Number        `json:"p"`
	StockOrNot  json.Number        `json:"stockOrNot"`
	IsNew       int                `json:"is_new"`
	Pictures    json.RawMessage    `json:"pictures"`
	Variations  []biblusiVariation `json:"variations"`
}

func (b biblusiBook) PriceF() float64 {
	f, _ := b.Price.Float64()
	return f
}
func (b biblusiBook) InStock() bool {
	i, _ := b.StockOrNot.Int64()
	return i == 1
}
func (b biblusiBook) PictureList() []string {
	var pics []string
	json.Unmarshal(b.Pictures, &pics)
	return pics
}

type biblusiDetailResponse struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	ISBN        string             `json:"isbn"`
	CategoryID  int                `json:"category_id"`
	BestPrice   json.Number        `json:"best_price"`
	Pictures    json.RawMessage    `json:"pictures"`
	MinPicture  string             `json:"min_picture"`
	StockOrNot  json.Number        `json:"stockOrNot"`
	Variations  []biblusiVariation `json:"variations"`
}

func (d biblusiDetailResponse) BestPriceF() float64 {
	f, _ := d.BestPrice.Float64()
	return f
}
func (d biblusiDetailResponse) PictureList() []string {
	var pics []string
	json.Unmarshal(d.Pictures, &pics)
	return pics
}

type biblusiVariation struct {
	Price         json.Number `json:"price"`
	Discount      json.Number `json:"discount"`
	DiscountValue json.Number `json:"discount_value"`
	StockCount    json.Number `json:"stock_count"`
}

func (v biblusiVariation) PriceF() float64 {
	f, _ := v.Price.Float64()
	return f
}
func (v biblusiVariation) DiscountF() float64 {
	f, _ := v.Discount.Float64()
	return f
}
func (v biblusiVariation) DiscountValueF() float64 {
	f, _ := v.DiscountValue.Float64()
	return f
}
func (v biblusiVariation) StockCountI() int {
	i, _ := v.StockCount.Int64()
	return int(i)
}

func (s *BiblusiScraper) Scrape(ctx context.Context, run *models.ScrapeRun) ([]ScrapedProduct, error) {
	s.logger.Info("scraping biblusi category", "category_id", s.categoryID, "source", s.name)
	products, err := s.scrapeCategory(ctx)
	if err != nil {
		return nil, err
	}
	s.logger.Info("scraped biblusi category", "category_id", s.categoryID, "count", len(products))
	return products, nil
}

func (s *BiblusiScraper) scrapeCategory(ctx context.Context) ([]ScrapedProduct, error) {
	var allProducts []ScrapedProduct
	page := 1
	perPage := 20

	for {
		if err := s.limiter.Wait(ctx); err != nil {
			return allProducts, err
		}

		url := fmt.Sprintf(
			"https://apiv1.biblusi.ge/api/book?category_id=%d&per_page=%d&page=%d&category=1&author=1",
			s.categoryID, perPage, page,
		)

		var resp biblusiListResponse
		err := WithRetry(ctx, 3, time.Second, func() error {
			return s.fetchJSON(ctx, url, &resp)
		})
		if err != nil {
			return allProducts, fmt.Errorf("fetch page %d: %w", page, err)
		}

		for _, book := range resp.Data {
			if !isLEGOProduct(book.Name) {
				continue
			}
			// Fetch detail to get ISBN (barcode)
			product, err := s.fetchProductDetail(ctx, book)
			if err != nil {
				s.logger.Warn("detail fetch failed, using list data", "id", book.ID, "error", err)
				product = s.bookToProduct(book, "")
			}
			allProducts = append(allProducts, product)
		}

		s.logger.Debug("scraped page", "page", page, "last_page", resp.LastPage, "products_so_far", len(allProducts))

		if page >= resp.LastPage {
			break
		}
		page++
	}

	return allProducts, nil
}

func (s *BiblusiScraper) fetchProductDetail(ctx context.Context, book biblusiBook) (ScrapedProduct, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return ScrapedProduct{}, err
	}

	url := fmt.Sprintf("https://apiv1.biblusi.ge/api/book/%d", book.ID)
	var detail biblusiDetailResponse

	err := WithRetry(ctx, 2, 500*time.Millisecond, func() error {
		return s.fetchJSON(ctx, url, &detail)
	})
	if err != nil {
		return ScrapedProduct{}, err
	}

	return s.detailToProduct(detail), nil
}

func (s *BiblusiScraper) bookToProduct(book biblusiBook, barcode string) ScrapedProduct {
	originalPrice, discountedPrice, discountPct := s.extractPrices(book.PriceF(), book.Variations)

	var imageURL string
	pics := book.PictureList()
	if len(pics) > 0 {
		imageURL = pics[0] // Pictures are already full URLs
	}

	return ScrapedProduct{
		MCode:           fmt.Sprintf("%d", book.ID),
		Barcode:         barcode,
		InvoiceCode:     ExtractLEGOID(book.Name),
		NameKa:          book.Name,
		InStock:         book.InStock(),
		OriginalPrice:   originalPrice,
		DiscountPercent: discountPct,
		DiscountedPrice: discountedPrice,
		ImageURL:        imageURL,
		SourceURL:       fmt.Sprintf("https://biblusi.ge/products/%d", book.ID),
		SourceProductID: fmt.Sprintf("%d", book.ID),
	}
}

func (s *BiblusiScraper) detailToProduct(detail biblusiDetailResponse) ScrapedProduct {
	originalPrice, discountedPrice, discountPct := s.extractPrices(detail.BestPriceF(), detail.Variations)

	var imageURL string
	if detail.MinPicture != "" {
		imageURL = detail.MinPicture // Already full URL
	} else {
		pics := detail.PictureList()
		if len(pics) > 0 {
			imageURL = pics[0]
		}
	}

	inStock := false
	for _, v := range detail.Variations {
		if v.StockCountI() > 0 {
			inStock = true
			break
		}
	}

	return ScrapedProduct{
		MCode:           fmt.Sprintf("%d", detail.ID),
		Barcode:         detail.ISBN,
		InvoiceCode:     ExtractLEGOID(detail.Name),
		NameKa:          detail.Name,
		InStock:         inStock,
		OriginalPrice:   originalPrice,
		DiscountPercent: discountPct,
		DiscountedPrice: discountedPrice,
		ImageURL:        imageURL,
		SourceURL:       fmt.Sprintf("https://biblusi.ge/products/%d", detail.ID),
		SourceProductID: fmt.Sprintf("%d", detail.ID),
	}
}

func (s *BiblusiScraper) extractPrices(listPrice float64, variations []biblusiVariation) (*float64, *float64, *float64) {
	if len(variations) == 0 {
		if listPrice > 0 {
			return &listPrice, nil, nil
		}
		return nil, nil, nil
	}

	// Find the best variation (lowest effective price with stock)
	var bestPrice, bestDiscount float64
	found := false

	for _, v := range variations {
		effectivePrice := v.PriceF()
		if v.DiscountF() > 0 {
			effectivePrice = v.PriceF() - v.DiscountValueF()
		}

		if !found || effectivePrice < bestPrice {
			bestPrice = effectivePrice
			bestDiscount = v.DiscountF()
			found = true
		}
	}

	if !found {
		return nil, nil, nil
	}

	if bestDiscount > 0 {
		// Has discount
		originalPrice := bestPrice / (1 - bestDiscount/100)
		if originalPrice < bestPrice {
			originalPrice = listPrice // fallback
		}
		return &originalPrice, &bestPrice, &bestDiscount
	}

	return &bestPrice, nil, nil
}

func (s *BiblusiScraper) fetchJSON(ctx context.Context, url string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body[:min(200, len(body))]))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	return nil
}
