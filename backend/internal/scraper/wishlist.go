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

type WishlistScraper struct {
	cfg    *config.Config
	logger *slog.Logger
}

func NewWishlistScraper(cfg *config.Config, logger *slog.Logger) *WishlistScraper {
	return &WishlistScraper{cfg: cfg, logger: logger}
}

func (s *WishlistScraper) Name() string             { return "wishlist" }
func (s *WishlistScraper) Source() models.SiteSource { return models.SourceWishlist }

const wishlistCategoryBase = "https://wishlist.ge/%E1%83%91%E1%83%90%E1%83%95%E1%83%A8%E1%83%95%E1%83%97%E1%83%90-%E1%83%A1%E1%83%90%E1%83%9B%E1%83%A7%E1%83%90%E1%83%A0%E1%83%9D/%E1%83%A1%E1%83%90%E1%83%97%E1%83%90%E1%83%9B%E1%83%90%E1%83%A8%E1%83%9D%E1%83%94%E1%83%91%E1%83%98/lego-%E1%83%9A%E1%83%94%E1%83%92%E1%83%9D-ka"

func (s *WishlistScraper) Scrape(ctx context.Context, run *models.ScrapeRun) ([]ScrapedProduct, error) {
	var allProducts []ScrapedProduct

	// Find the last valid page by going from page-20 downward
	maxPage := s.findLastPage(ctx)
	s.logger.Info("wishlist last page found", "maxPage", maxPage)

	// Scrape all pages with 192 items per page
	for page := 1; page <= maxPage; page++ {
		url := fmt.Sprintf("%s/page-%d/?items_per_page=192", wishlistCategoryBase, page)

		doc, err := FetchPageHTML(ctx, s.cfg.ChromeWSURL, url, s.logger)
		if err != nil {
			s.logger.Error("wishlist page failed", "page", page, "error", err)
			break
		}

		pageProducts := s.parsePage(doc)
		s.logger.Info("wishlist page scraped", "page", page, "found", len(pageProducts))

		if len(pageProducts) == 0 {
			break
		}
		allProducts = append(allProducts, pageProducts...)
	}

	s.logger.Info("wishlist scraping complete", "total", len(allProducts))
	return allProducts, nil
}

// findLastPage checks from page-20 down to find the highest valid page
func (s *WishlistScraper) findLastPage(ctx context.Context) int {
	for page := 20; page >= 1; page-- {
		url := fmt.Sprintf("%s/page-%d/?items_per_page=192", wishlistCategoryBase, page)

		doc, err := FetchPageHTML(ctx, s.cfg.ChromeWSURL, url, s.logger)
		if err != nil {
			continue
		}

		// Check for 404 text
		bodyText := doc.Find("body").Text()
		if strings.Contains(bodyText, "ვერ იქნა მოძიებული") || strings.Contains(bodyText, "404") {
			continue
		}

		// Check if there are products on this page
		count := doc.Find(".ty-grid-list__item").Length()
		if count > 0 {
			s.logger.Info("found valid page", "page", page, "products", count)
			return page
		}
	}
	return 1
}

func (s *WishlistScraper) parsePage(doc *goquery.Document) []ScrapedProduct {
	var products []ScrapedProduct

	doc.Find(".ty-grid-list__item").Each(func(i int, sel *goquery.Selection) {
		name := strings.TrimSpace(sel.Find("a.product-title").Text())
		if name == "" {
			name = strings.TrimSpace(sel.Find(".ty-grid-list__item-name a").Text())
		}
		if name == "" {
			return
		}

		// Product ID from CS-Cart hidden input
		productID, _ := sel.Find("input[name*='product_data'][name*='product_id']").Attr("value")

		// LEGO set number from name (5-6 digits, anywhere in wishlist names)
		legoID := ExtractLEGOIDAnywhere(name)

		var invoiceCode, mcode string
		if legoID != "" {
			// Found LEGO ID in name → lego_id goes to invoice_code, CS-Cart ID to mcode
			invoiceCode = legoID
			mcode = productID
		} else {
			// No LEGO ID → CS-Cart ID goes to invoice_code, mcode empty
			invoiceCode = productID
		}

		// Prices
		var originalPrice, discountedPrice *float64
		var discountPct *float64

		oldPriceText := strings.TrimSpace(sel.Find(".ty-strike .ty-list-price").Text())
		newPriceText := strings.TrimSpace(sel.Find(".ty-price .ty-price-num").Text())

		if oldPriceText != "" && newPriceText != "" {
			originalPrice, _ = ParsePrice(oldPriceText)
			discountedPrice, _ = ParsePrice(newPriceText)
			if originalPrice != nil && discountedPrice != nil {
				if *discountedPrice < *originalPrice {
					discountPct = CalcDiscountPercent(*originalPrice, *discountedPrice)
				} else {
					discountedPrice = nil
				}
			}
		} else if newPriceText != "" {
			originalPrice, _ = ParsePrice(newPriceText)
		}

		productURL, _ := sel.Find("a.product-title").Attr("href")
		if productURL == "" {
			productURL, _ = sel.Find(".abt-single-image").Attr("href")
		}

		imgURL, _ := sel.Find("img.ty-pict").Attr("src")
		if imgURL == "" {
			imgURL, _ = sel.Find("img.ty-pict").Attr("data-src")
		}

		products = append(products, ScrapedProduct{
			MCode:           mcode,
			InvoiceCode:     invoiceCode,
			NameKa:          name,
			OriginalPrice:   originalPrice,
			DiscountPercent: discountPct,
			DiscountedPrice: discountedPrice,
			InStock:         true,
			ImageURL:        imgURL,
			SourceURL:       productURL,
			SourceProductID: productID,
		})
	})

	return products
}
