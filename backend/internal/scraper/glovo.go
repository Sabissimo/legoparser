package scraper

import (
	"context"
	"log/slog"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"lego-parser/internal/config"
	"lego-parser/internal/models"
)

type GlovoScraper struct {
	name   string
	source models.SiteSource
	url    string
	cfg    *config.Config
	logger *slog.Logger
}

func NewGlovoXSScraper(cfg *config.Config, logger *slog.Logger) *GlovoScraper {
	return &GlovoScraper{
		name:   "glovo_xs",
		source: models.SourceGlovoXS,
		url:    "https://glovoapp.com/en/ge/tbilisi/stores/xs-toys-tbi?content=lego-konstruqtorebi-c.21175679",
		cfg:    cfg, logger: logger,
	}
}

func NewGlovoPepelaScraper(cfg *config.Config, logger *slog.Logger) *GlovoScraper {
	return &GlovoScraper{
		name:   "glovo_pepela",
		source: models.SourceGlovoPepela,
		url:    "https://glovoapp.com/en/ge/tbilisi/stores/pepela-tbi?content=lego-sc.21169575%2Flego-konstruqtorebi-c.21169580",
		cfg:    cfg, logger: logger,
	}
}

func (s *GlovoScraper) Name() string             { return s.name }
func (s *GlovoScraper) Source() models.SiteSource { return s.source }

func (s *GlovoScraper) Scrape(ctx context.Context, run *models.ScrapeRun) ([]ScrapedProduct, error) {
	doc, err := FetchPageHTML(ctx, s.cfg.ChromeWSURL, s.url, s.logger)
	if err != nil {
		s.logger.Error("glovo page failed", "source", s.name, "error", err)
		return nil, nil
	}

	var products []ScrapedProduct

	doc.Find("[class*='ItemTile_itemTile']").Each(func(i int, sel *goquery.Selection) {
		name := strings.TrimSpace(sel.Find("[class*='ItemTile_title'], h3").First().Text())
		if name == "" {
			return
		}

		priceText := strings.TrimSpace(sel.Find("[class*='ItemTile_discountedPrice'], [class*='price']").First().Text())
		price, _ := ParsePrice(priceText)

		imgURL, _ := sel.Find("[class*='ItemTile_image'] img, img").First().Attr("src")
		legoID := ExtractLEGOID(name)

		products = append(products, ScrapedProduct{
			InvoiceCode:     legoID,
			NameKa:          name,
			OriginalPrice:   price,
			InStock:         true,
			ImageURL:        imgURL,
			SourceURL:       s.url,
			SourceProductID: legoID,
		})
	})

	s.logger.Info("glovo scraping complete", "source", s.name, "found", len(products))
	return products, nil
}
