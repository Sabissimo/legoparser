# LEGO Price Comparison Parser - Architecture

## Overview

A web application that scrapes LEGO product data from multiple Georgian e-commerce websites and provides price comparison.

## Tech Stack

- **Frontend**: React 18 + TypeScript + Vite + TailwindCSS
- **Backend**: Go (net/http, chi patterns)
- **Database**: PostgreSQL 16
- **Browser Automation**: chromedp (headless Chrome)
- **HTML Scraping**: Colly
- **Scheduling**: robfig/cron

## Architecture Diagram

```
React (localhost:3000)  -->  Go API (:8081)  -->  PostgreSQL (:5433)
                                  |
                          Chrome headless (:9222)
```

## Sites Scraped

| Site | Method | Search By | Technology |
|------|--------|-----------|------------|
| Biblusi.ge (cat 474=Pepela, 456=XS) | REST API | barcode | HTTP GET |
| PiccolaToys.ge | HTML scraping | invoice code | Colly |
| Kubiki.ge | HTML scraping | invoice code | Colly |
| Wishlist.ge | Browser automation | invoice code | chromedp |
| Wolt | REST API | text search | HTTP POST |
| Glovo | Browser automation | invoice code | chromedp |

## Database Schema

### Tables
- **products** - canonical product identity (id, subtype, mcode, barcode, invoice_code, name_ka, name_en, image_url)
- **prices** - append-only price observations (product_id, source, original_price, discount_percent, discounted_price, scrape_run_id)
- **scrape_runs** - scraping job history (source, status, trigger_type, counts, timestamps)

### Views
- **latest_prices** - most recent price per product per source
- **price_comparison** - pivot table with all site prices in one row per product

## API Endpoints

```
GET  /api/health                 # Health check
GET  /api/products               # Paginated product list
GET  /api/products/:id           # Single product with prices
GET  /api/products/comparison    # Price comparison view
GET  /api/export/products        # Excel export
GET  /api/export/comparison      # Comparison Excel export
POST /api/scrape/run             # Trigger scraping
GET  /api/scrape/runs            # Scrape history
GET  /api/scrape/status          # Current scrape status
```

## Frontend Pages

- **Dashboard** - overview with stats and recent scrapes
- **Products** - searchable/filterable/sortable product table
- **Comparison** - price comparison across all sites (color-coded)
- **Scraper** - manual trigger controls and run history

## Data Flow

1. **Biblusi scraped first** (seed catalog via API, ~3000 products)
2. Other sites scrape independently and products match by barcode/invoice_code
3. Prices stored append-only for history tracking
4. Scheduler runs daily at 6:00 AM (configurable via CRON_SCHEDULE)
