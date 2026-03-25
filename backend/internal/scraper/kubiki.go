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

type KubikiScraper struct {
	cfg    *config.Config
	logger *slog.Logger
}

func NewKubikiScraper(cfg *config.Config, logger *slog.Logger) *KubikiScraper {
	return &KubikiScraper{cfg: cfg, logger: logger}
}

func (s *KubikiScraper) Name() string             { return "kubiki" }
func (s *KubikiScraper) Source() models.SiteSource { return models.SourceKubiki }

func (s *KubikiScraper) Scrape(ctx context.Context, run *models.ScrapeRun) ([]ScrapedProduct, error) {
	var allProducts []ScrapedProduct

	for page := 1; page <= 40; page++ {
		url := fmt.Sprintf("https://kubiki.ge/shop/page/%d/?count=36", page)

		doc, err := FetchPageHTML(ctx, s.cfg.ChromeWSURL, url, s.logger)
		if err != nil {
			s.logger.Error("kubiki page failed", "page", page, "error", err)
			break
		}

		var pageProducts []ScrapedProduct

		doc.Find("li.product, .product-item, .product").Each(func(i int, sel *goquery.Selection) {
			name := strings.TrimSpace(sel.Find(".product-title a, h3 a, h2 a, .woocommerce-loop-product__title").First().Text())
			if name == "" {
				return
			}
			// Kubiki is LEGO-only store, no need to filter by name

			var originalPrice, discountedPrice *float64
			var discountPct *float64

			delText := strings.TrimSpace(sel.Find("del .woocommerce-Price-amount, del .amount").First().Text())
			insText := strings.TrimSpace(sel.Find("ins .woocommerce-Price-amount, ins .amount").First().Text())

			if delText != "" && insText != "" {
				originalPrice, _ = ParsePrice(delText)
				discountedPrice, _ = ParsePrice(insText)
				if originalPrice != nil && discountedPrice != nil {
					discountPct = CalcDiscountPercent(*originalPrice, *discountedPrice)
				}
			} else {
				priceText := strings.TrimSpace(sel.Find(".price .woocommerce-Price-amount, .price .amount, .price").First().Text())
				originalPrice, _ = ParsePrice(priceText)
			}

			productURL, _ := sel.Find("a").First().Attr("href")
			imgURL, _ := sel.Find("img").First().Attr("src")
			if imgURL == "" {
				imgURL, _ = sel.Find("img").First().Attr("data-src")
			}

			legoID := ExtractLEGOID(name)

			pageProducts = append(pageProducts, ScrapedProduct{
				InvoiceCode:     legoID,
				NameKa:          name,
				OriginalPrice:   originalPrice,
				DiscountPercent: discountPct,
				DiscountedPrice: discountedPrice,
				InStock:         true,
				ImageURL:        imgURL,
				SourceURL:       productURL,
				SourceProductID: legoID,
			})
		})

		s.logger.Info("kubiki page scraped", "page", page, "found", len(pageProducts))

		if len(pageProducts) == 0 {
			break
		}
		allProducts = append(allProducts, pageProducts...)
	}

	s.logger.Info("kubiki scraping complete", "total", len(allProducts))
	return allProducts, nil
}
