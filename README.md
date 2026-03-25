# LEGO Price Comparison Parser

A web application that scrapes LEGO product data from multiple Georgian e-commerce websites and provides price comparison across 9 sources.

## Sources

| Source | Method | Description |
|--------|--------|-------------|
| **Biblusi XS** | REST API | Biblusi.ge category 456 (XS Toys) |
| **Biblusi Pepela** | REST API | Biblusi.ge category 474 (Pepela) |
| **Wishlist** | Headless Chrome | Wishlist.ge LEGO category |
| **PiccolaToys** | HTML Scraping | PiccolaToys.ge |
| **Kubiki** | Headless Chrome | Kubiki.ge (LEGO specialist store) |
| **Wolt XS** | Headless Chrome | 4 XS Toys stores on Wolt |
| **Wolt Pepela** | Headless Chrome | 6 Pepela stores on Wolt |
| **Glovo XS** | Headless Chrome | Glovo XS Toys LEGO section |
| **Glovo Pepela** | Headless Chrome | Glovo Pepela LEGO section |

## Tech Stack

- **Frontend**: React 18, TypeScript, Vite, TailwindCSS, TanStack Query/Table
- **Backend**: Go (net/http, chromedp, Colly, pgx, excelize)
- **Database**: PostgreSQL 16
- **Browser Automation**: chromedp + headless Chrome container
- **Containerization**: Docker Compose

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.25+
- Node.js 18+

### 1. Start infrastructure

```bash
docker-compose up -d db chrome
```

### 2. Start backend

```bash
cd backend
cp ../.env.example ../.env  # edit if needed
export DATABASE_URL="postgres://parser:parserpass@localhost:5433/lego_parser?sslmode=disable"
export MIGRATIONS_PATH="internal/database/migrations"
export PORT=8081
go run ./cmd/server/
```

### 3. Start frontend

```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:3001

### 4. Run scrapers

Either click "Start" on the Scraper page, or:

```bash
curl -X POST http://localhost:8081/api/scrape/run \
  -H "Content-Type: application/json" \
  -d '{"source":"all"}'
```

Individual sources: `biblusi_xs`, `biblusi_pepela`, `wishlist`, `piccolatoys`, `kubiki`, `wolt_xs`, `wolt_pepela`, `glovo_xs`, `glovo_pepela`

## Features

### Products Page
- Unified LEGO product table (no categories - products are unique LEGO sets)
- Search by name, LEGO ID, MCode, barcode
- Sortable columns
- Source tags showing which sites carry each product
- Excel + CSV (Qlik) export

### Price Comparison
- 9 price columns side-by-side
- Color-coded: green = cheapest, red = most expensive
- All columns sortable (NULLS LAST)
- Excel + CSV (Qlik) export

### Scraper Control
- Manual trigger per source or all at once
- Live status monitoring
- Scrape history with duration and error tracking
- Automatic daily scheduling (configurable via CRON_SCHEDULE)

## Data Model

- `invoice_code` = LEGO set number (extracted from product name)
- `mcode` = Source-specific internal ID (Biblusi product ID, Wishlist CS-Cart ID)
- `barcode` = EAN/UPC barcode (from Biblusi detail API)
- Products matched across sources by: invoice_code → barcode → name

## API Endpoints

```
GET  /api/products                 # Products list (paginated, sortable)
GET  /api/products/:id             # Single product with price history
GET  /api/products/comparison      # Price comparison (9 columns)
POST /api/scrape/run               # Trigger scraper {"source":"all"|"biblusi_xs"|...}
GET  /api/scrape/runs              # Scrape history
GET  /api/scrape/status            # Current running status
GET  /api/export/products          # Excel export
GET  /api/export/comparison        # Comparison Excel
GET  /api/export/products/csv      # CSV for Qlik
GET  /api/export/comparison/csv    # Comparison CSV for Qlik
GET  /api/health                   # Health check
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://parser:parserpass@localhost:5433/lego_parser?sslmode=disable` | PostgreSQL connection |
| `CHROME_WS_URL` | `ws://localhost:9222` | Headless Chrome WebSocket |
| `PORT` | `8081` | Backend HTTP port |
| `CRON_SCHEDULE` | `0 0 6 * * *` | Daily scrape schedule (6 AM) |
| `MIGRATIONS_PATH` | `migrations` | Path to SQL migrations |

## Architecture

```
React (:3001)  →  Go API (:8081)  →  PostgreSQL (:5433)
                       ↕
                Chrome headless (:9222)
```

## Project Structure

```
├── backend/
│   ├── cmd/server/main.go          # Entry point
│   └── internal/
│       ├── config/                  # Environment config
│       ├── database/                # DB connection + migrations
│       ├── handler/                 # HTTP handlers + export
│       ├── models/                  # Data models
│       ├── repository/              # DB queries
│       ├── router/                  # HTTP routing + CORS
│       ├── scheduler/               # Cron scheduler
│       └── scraper/                 # All 9 scrapers + helpers
├── frontend/
│   └── src/
│       ├── api/client.ts            # API client
│       ├── components/Layout/       # Navigation layout
│       ├── pages/                   # Dashboard, Products, Comparison, Scraper
│       └── types/                   # TypeScript interfaces
├── docker-compose.yml
├── docs/architecture.md
└── plan/lego-parser-plan.md
```
