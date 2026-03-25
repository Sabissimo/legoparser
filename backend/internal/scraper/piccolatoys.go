package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"lego-parser/internal/config"
	"lego-parser/internal/models"
)

type PiccolaToysScraper struct {
	cfg    *config.Config
	logger *slog.Logger
}

func NewPiccolaToysScraper(cfg *config.Config, logger *slog.Logger) *PiccolaToysScraper {
	return &PiccolaToysScraper{cfg: cfg, logger: logger}
}

func (s *PiccolaToysScraper) Name() string              { return "piccolatoys" }
func (s *PiccolaToysScraper) Source() models.SiteSource  { return models.SourcePiccolaToys }

func (s *PiccolaToysScraper) Scrape(ctx context.Context, run *models.ScrapeRun) ([]ScrapedProduct, error) {
	var products []ScrapedProduct
	var detailURLs []string

	// Collector for listing pages
	listCollector := colly.NewCollector(
		colly.AllowedDomains("piccolatoys.ge", "www.piccolatoys.ge"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)
	listCollector.SetRequestTimeout(30 * time.Second)
	listCollector.Limit(&colly.LimitRule{
		DomainGlob: "*piccolatoys.ge*",
		Delay:      500 * time.Millisecond,
	})

	// Collector for product detail pages
	detailCollector := listCollector.Clone()

	// Debug: log response
	listCollector.OnResponse(func(r *colly.Response) {
		s.logger.Info("piccolatoys response", "url", r.Request.URL.String(), "status", r.StatusCode, "body_len", len(r.Body))
	})

	// Extract product URLs from listing pages - try multiple selectors
	listCollector.OnHTML(".wd-product", func(e *colly.HTMLElement) {
		productURL := e.ChildAttr(".product-image-link", "href")
		if productURL == "" {
			productURL = e.ChildAttr("h3.product-title a", "href")
		}
		if productURL == "" {
			productURL = e.ChildAttr(".product-title a", "href")
		}
		if productURL == "" {
			productURL = e.ChildAttr("a", "href")
		}
		if productURL != "" {
			detailURLs = append(detailURLs, productURL)
		}
	})

	// Fallback: also try standard WooCommerce selectors
	listCollector.OnHTML("li.product a.woocommerce-LoopProduct-link", func(e *colly.HTMLElement) {
		productURL := e.Attr("href")
		if productURL != "" && strings.Contains(productURL, "/product/") {
			detailURLs = append(detailURLs, productURL)
		}
	})

	// Follow pagination
	listCollector.OnHTML("a.next.page-numbers", func(e *colly.HTMLElement) {
		nextURL := e.Attr("href")
		if nextURL != "" {
			s.logger.Debug("following next page", "url", nextURL)
			e.Request.Visit(nextURL)
		}
	})

	listCollector.OnError(func(r *colly.Response, err error) {
		s.logger.Error("piccolatoys list error", "url", r.Request.URL.String(), "error", err)
	})

	// Extract product details
	detailCollector.OnHTML("body", func(e *colly.HTMLElement) {
		product := s.parseProductDetail(e)
		if product.NameKa != "" {
			products = append(products, product)
		}
	})

	detailCollector.OnError(func(r *colly.Response, err error) {
		s.logger.Error("piccolatoys detail error", "url", r.Request.URL.String(), "error", err)
	})

	// Start scraping LEGO products
	s.logger.Info("scraping piccolatoys.ge LEGO products")

	// Search for LEGO products
	err := listCollector.Visit("https://piccolatoys.ge/?s=lego&post_type=product")
	if err != nil {
		return nil, fmt.Errorf("visit piccolatoys search: %w", err)
	}
	listCollector.Wait()

	s.logger.Info("found product URLs", "count", len(detailURLs))

	// Visit each product detail page
	for _, url := range detailURLs {
		select {
		case <-ctx.Done():
			return products, ctx.Err()
		default:
		}
		detailCollector.Visit(url)
	}
	detailCollector.Wait()

	s.logger.Info("scraped piccolatoys products", "count", len(products))
	return products, nil
}

func (s *PiccolaToysScraper) parseProductDetail(e *colly.HTMLElement) ScrapedProduct {
	name := strings.TrimSpace(e.ChildText("h1.product_title"))
	if name == "" {
		name = strings.TrimSpace(e.ChildText(".product-title"))
	}

	sku := strings.TrimSpace(e.ChildText("span.sku"))
	if sku == "" {
		sku = strings.TrimSpace(e.ChildText(".sku-value"))
	}

	// Price extraction
	var originalPrice, discountedPrice *float64
	var discountPct *float64

	// Check for sale price first
	delPrice := strings.TrimSpace(e.ChildText("p.price del .amount, .price del .woocommerce-Price-amount"))
	insPrice := strings.TrimSpace(e.ChildText("p.price ins .amount, .price ins .woocommerce-Price-amount"))

	if delPrice != "" && insPrice != "" {
		// Has both original and sale price
		originalPrice, _ = ParsePrice(delPrice)
		discountedPrice, _ = ParsePrice(insPrice)
		if originalPrice != nil && discountedPrice != nil {
			discountPct = CalcDiscountPercent(*originalPrice, *discountedPrice)
		}
	} else {
		// Single price
		priceStr := strings.TrimSpace(e.ChildText("p.price .amount, .price .woocommerce-Price-amount"))
		if priceStr == "" {
			priceStr = strings.TrimSpace(e.ChildText(".price-section .price"))
		}
		originalPrice, _ = ParsePrice(priceStr)
	}

	imageURL := e.ChildAttr(".woocommerce-product-gallery__image img", "src")
	if imageURL == "" {
		imageURL = e.ChildAttr(".product-image img", "src")
	}

	// Stock status
	inStock := true
	stockText := strings.TrimSpace(e.ChildText(".stock"))
	if strings.Contains(stockText, "არ არის") || strings.Contains(stockText, "out") {
		inStock = false
	}

	sourceURL := e.Request.URL.String()

	return ScrapedProduct{
		InvoiceCode:     sku,
		NameKa:          name,
		OriginalPrice:   originalPrice,
		DiscountPercent: discountPct,
		DiscountedPrice: discountedPrice,
		InStock:         inStock,
		ImageURL:        imageURL,
		SourceURL:       sourceURL,
		SourceProductID: sku,
	}
}
