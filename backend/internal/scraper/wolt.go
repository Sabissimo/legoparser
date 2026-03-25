package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"lego-parser/internal/config"
	"lego-parser/internal/models"
)

type WoltScraper struct {
	name   string
	source models.SiteSource
	stores []string
	cfg    *config.Config
	logger *slog.Logger
}

func NewWoltXSScraper(cfg *config.Config, logger *slog.Logger) *WoltScraper {
	return &WoltScraper{
		name:   "wolt_xs",
		source: models.SourceWoltXS,
		stores: []string{
			"xs-toys-galleria-tbilisi1",
			"xs-toys-city-mall",
			"lego-galleria-tbilisi",
			"xs-toys-tbilisi-mall",
		},
		cfg: cfg, logger: logger,
	}
}

func NewWoltPepelaScraper(cfg *config.Config, logger *slog.Logger) *WoltScraper {
	return &WoltScraper{
		name:   "wolt_pepela",
		source: models.SourceWoltPepela,
		stores: []string{
			"pepela-vake-park",
			"pepela-city-mall",
			"pepela-vake",
			"pepela-marjanishvili",
			"pepela-saburtalo",
			"pepela-aghmashenebeli",
		},
		cfg: cfg, logger: logger,
	}
}

func (s *WoltScraper) Name() string             { return s.name }
func (s *WoltScraper) Source() models.SiteSource { return s.source }

func (s *WoltScraper) Scrape(ctx context.Context, run *models.ScrapeRun) ([]ScrapedProduct, error) {
	var allProducts []ScrapedProduct

	for _, slug := range s.stores {
		legoURL := fmt.Sprintf("https://wolt.com/en/geo/tbilisi/venue/%s/items/lego-3", slug)

		doc, err := FetchPageHTML(ctx, s.cfg.ChromeWSURL, legoURL, s.logger)
		if err != nil {
			s.logger.Warn("wolt store failed", "slug", slug, "error", err)
			continue
		}

		products := s.parseProducts(doc, slug)
		if len(products) == 0 {
			mainURL := fmt.Sprintf("https://wolt.com/en/geo/tbilisi/venue/%s", slug)
			doc, err = FetchPageHTML(ctx, s.cfg.ChromeWSURL, mainURL, s.logger)
			if err == nil {
				products = s.parseProducts(doc, slug)
			}
		}

		allProducts = append(allProducts, products...)
		s.logger.Info("wolt store scraped", "slug", slug, "found", len(products))
	}

	s.logger.Info("wolt scraping complete", "source", s.name, "total", len(allProducts))
	return allProducts, nil
}

func (s *WoltScraper) parseProducts(doc *goquery.Document, slug string) []ScrapedProduct {
	var products []ScrapedProduct

	doc.Find("[data-test-id='ItemCard']").Each(func(i int, sel *goquery.Selection) {
		name := strings.TrimSpace(sel.Find("[data-test-id='ImageCentricProductCard.Title'], h3").First().Text())
		if name == "" || !isLEGOProduct(name) {
			return
		}

		priceText := strings.TrimSpace(sel.Find("[class*='p1yjaao3'], [aria-label*='Price']").First().Text())
		priceText = strings.ReplaceAll(priceText, "GEL", "")
		price, _ := ParsePrice(priceText)

		imgURL, _ := sel.Find("[data-test-id='ImageCentricProductCard.ProductImage'], img").First().Attr("src")
		productURL, _ := sel.Find("[data-test-id='CardLinkButton'], a").First().Attr("href")
		if productURL != "" && !strings.HasPrefix(productURL, "http") {
			productURL = "https://wolt.com" + productURL
		}

		legoID := ExtractLEGOID(name)

		products = append(products, ScrapedProduct{
			InvoiceCode:     legoID,
			NameKa:          name,
			OriginalPrice:   price,
			InStock:         true,
			ImageURL:        imgURL,
			SourceURL:       productURL,
			SourceProductID: legoID,
		})
	})

	return products
}
