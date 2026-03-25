package models

import "time"

type SiteSource string

const (
	SourceBiblusiXS      SiteSource = "biblusi_xs"
	SourceBiblusiPepela  SiteSource = "biblusi_pepela"
	SourceWoltXS         SiteSource = "wolt_xs"
	SourceWoltPepela     SiteSource = "wolt_pepela"
	SourceGlovoXS        SiteSource = "glovo_xs"
	SourceGlovoPepela    SiteSource = "glovo_pepela"
	SourceWishlist       SiteSource = "wishlist"
	SourcePiccolaToys    SiteSource = "piccolatoys"
	SourceKubiki         SiteSource = "kubiki"
)

type Product struct {
	ID          int64    `json:"id"`
	MCode       *string  `json:"mcode"`
	Barcode     *string  `json:"barcode"`
	InvoiceCode *string  `json:"invoice_code"`
	NameKa      string   `json:"name_ka"`
	NameEn      *string  `json:"name_en"`
	ImageURL    *string  `json:"image_url"`
	Sources     []string `json:"sources"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	LatestPrices []Price `json:"latest_prices,omitempty"`
}

type Price struct {
	ID              int64      `json:"id"`
	ProductID       int64      `json:"product_id"`
	Source          SiteSource `json:"source"`
	OriginalPrice   *float64   `json:"original_price"`
	DiscountPercent *float64   `json:"discount_percent"`
	DiscountedPrice *float64   `json:"discounted_price"`
	InStock         bool       `json:"in_stock"`
	SourceURL       *string    `json:"source_url"`
	SourceProductID *string    `json:"source_product_id"`
	ScrapeRunID     int64      `json:"scrape_run_id"`
	ScrapedAt       time.Time  `json:"scraped_at"`
}

type ScrapeRun struct {
	ID            int64      `json:"id"`
	Source        SiteSource `json:"source"`
	Status        string     `json:"status"`
	TriggerType   string     `json:"trigger_type"`
	ProductsFound int        `json:"products_found"`
	ProductsSaved int        `json:"products_saved"`
	ErrorsCount   int        `json:"errors_count"`
	ErrorLog      *string    `json:"error_log"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type PriceComparison struct {
	ProductID         int64    `json:"product_id"`
	MCode             *string  `json:"mcode"`
	Barcode           *string  `json:"barcode"`
	InvoiceCode       *string  `json:"invoice_code"`
	NameKa            string   `json:"name_ka"`
	NameEn            *string  `json:"name_en"`
	BiblusiXSPrice    *float64 `json:"biblusi_xs_price"`
	BiblusiPepelaPrice *float64 `json:"biblusi_pepela_price"`
	WoltXSPrice       *float64 `json:"wolt_xs_price"`
	WoltPepelaPrice   *float64 `json:"wolt_pepela_price"`
	GlovoXSPrice      *float64 `json:"glovo_xs_price"`
	GlovoPepelaPrice  *float64 `json:"glovo_pepela_price"`
	WishlistPrice     *float64 `json:"wishlist_price"`
	PiccolaToysPrice  *float64 `json:"piccolatoys_price"`
	KubikiPrice       *float64 `json:"kubiki_price"`
}
