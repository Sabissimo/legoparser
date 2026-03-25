# LEGO Parser - Current Status

## All 9 Scrapers

| Source | Products | Description |
|--------|----------|-------------|
| biblusi_xs | 218 | Biblusi cat 456, REST API, MCode=Biblusi ID |
| biblusi_pepela | 37 | Biblusi cat 474, REST API, MCode=Biblusi ID |
| wishlist | 1,453 | Wishlist.ge, chromedp, 192/page, LEGO ID anywhere in name |
| piccolatoys | 51 | PiccolaToys.ge, Colly |
| kubiki | 741 | Kubiki.ge, chromedp, 36/page paginated, no LEGO filter (LEGO-only store) |
| wolt_xs | 200 | 4 XS stores on Wolt, chromedp |
| wolt_pepela | 294 | 6 Pepela stores on Wolt, chromedp |
| glovo_xs | 218 | Glovo XS Toys, chromedp |
| glovo_pepela | 93 | Glovo Pepela, chromedp |
| **Total** | **~2,356** | unique products |

## Data Model
- **No subtypes** - unified product table
- `invoice_code` = LEGO set number (from name)
- `mcode` = Biblusi internal ID or Wishlist CS-Cart ID
- `sources[]` = JSON array of which price sources found this product (tags in UI)

## LEGO ID Extraction (ExtractLEGOID)
- Start of name: `"76452 LEGO..."` → 76452
- After "LEGO"/"ლეგო": `"LEGO 10713 Creative..."` → 10713
- End of name: `"...Quidditch 76452"` → 76452
- Dashes normalized to spaces
- **Wishlist only**: `ExtractLEGOIDAnywhere` - fallback to any 5-6 digit number in name
- Wishlist logic: LEGO ID found → invoice_code=LEGO ID + mcode=CS-Cart ID; not found → invoice_code=CS-Cart ID, mcode empty

## Comparison (9 price columns, all sortable with NULLS LAST)
Biblusi XS | Biblusi Pepela | Wolt XS | Wolt Pepela | Glovo XS | Glovo Pepela | Wishlist | Piccola | Kubiki

## Export Formats
- **Excel** (.xlsx) - Products + Price Comparison
- **CSV/QVD** (.csv, tab-delimited, UTF-8 BOM) - Qlik compatible, Products + Price Comparison

## API Endpoints
```
GET /api/products                    # Products list (paginated, sortable, sources tags)
GET /api/products/:id                # Single product
GET /api/products/comparison         # Price comparison (9 cols, sortable)
GET /api/export/products             # Excel export
GET /api/export/comparison           # Comparison Excel
GET /api/export/products/csv         # CSV for Qlik
GET /api/export/comparison/csv       # Comparison CSV for Qlik
POST /api/scrape/run                 # Trigger scraper
GET /api/scrape/runs                 # History
GET /api/scrape/status               # Running status
```

## Architecture
```
React (:3001) → Go API (:8081) → PostgreSQL (:5433)
                     ↕
              Chrome headless (:9222)
```

## Key Fixes History
- Biblusi: LEGO-only filter, detail fetch for ISBN/barcode, MCode=internal ID
- Wishlist: category page 192/page, ExtractLEGOIDAnywhere for IDs
- Kubiki: paginated 36/page, no LEGO filter (LEGO-only store), 40 page max
- Upsert: invoice_code → barcode → name fallback matching
- Comparison sorting: NULLS LAST for correct price sorting
- Sources tags: JSON aggregate query for pgx compatibility
