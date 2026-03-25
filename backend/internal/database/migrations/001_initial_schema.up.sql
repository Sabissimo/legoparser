CREATE TYPE site_source AS ENUM (
    'biblusi_xs',
    'biblusi_pepela',
    'wolt_xs',
    'wolt_pepela',
    'glovo_xs',
    'glovo_pepela',
    'wishlist',
    'piccolatoys',
    'kubiki'
);

CREATE TABLE products (
    id              BIGSERIAL PRIMARY KEY,
    mcode           VARCHAR(50),
    barcode         VARCHAR(50),
    invoice_code    VARCHAR(100),
    name_ka         TEXT NOT NULL,
    name_en         TEXT,
    image_url       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_products_barcode ON products(barcode) WHERE barcode IS NOT NULL;
CREATE UNIQUE INDEX idx_products_invoice_code ON products(invoice_code) WHERE invoice_code IS NOT NULL;

CREATE TABLE scrape_runs (
    id              BIGSERIAL PRIMARY KEY,
    source          site_source NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    trigger_type    VARCHAR(20) NOT NULL DEFAULT 'manual',
    products_found  INT DEFAULT 0,
    products_saved  INT DEFAULT 0,
    errors_count    INT DEFAULT 0,
    error_log       TEXT,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scrape_runs_source ON scrape_runs(source);

CREATE TABLE prices (
    id                  BIGSERIAL PRIMARY KEY,
    product_id          BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    source              site_source NOT NULL,
    original_price      NUMERIC(10, 2),
    discount_percent    NUMERIC(5, 2),
    discounted_price    NUMERIC(10, 2),
    in_stock            BOOLEAN DEFAULT TRUE,
    source_url          TEXT,
    source_product_id   VARCHAR(100),
    scrape_run_id       BIGINT NOT NULL REFERENCES scrape_runs(id) ON DELETE CASCADE,
    scraped_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prices_product_source ON prices(product_id, source);
CREATE INDEX idx_prices_scrape_run ON prices(scrape_run_id);
CREATE INDEX idx_prices_scraped_at ON prices(scraped_at DESC);

CREATE VIEW latest_prices AS
SELECT DISTINCT ON (product_id, source)
    id, product_id, source, original_price, discount_percent,
    discounted_price, in_stock, source_url, scraped_at
FROM prices
ORDER BY product_id, source, scraped_at DESC;

CREATE VIEW price_comparison AS
SELECT
    p.id AS product_id,
    p.mcode,
    p.barcode,
    p.invoice_code,
    p.name_ka,
    p.name_en,
    MAX(CASE WHEN lp.source = 'biblusi_xs' THEN COALESCE(lp.discounted_price, lp.original_price) END) AS biblusi_xs_price,
    MAX(CASE WHEN lp.source = 'biblusi_pepela' THEN COALESCE(lp.discounted_price, lp.original_price) END) AS biblusi_pepela_price,
    MAX(CASE WHEN lp.source = 'wolt_xs' THEN COALESCE(lp.discounted_price, lp.original_price) END) AS wolt_xs_price,
    MAX(CASE WHEN lp.source = 'wolt_pepela' THEN COALESCE(lp.discounted_price, lp.original_price) END) AS wolt_pepela_price,
    MAX(CASE WHEN lp.source = 'glovo_xs' THEN COALESCE(lp.discounted_price, lp.original_price) END) AS glovo_xs_price,
    MAX(CASE WHEN lp.source = 'glovo_pepela' THEN COALESCE(lp.discounted_price, lp.original_price) END) AS glovo_pepela_price,
    MAX(CASE WHEN lp.source = 'wishlist' THEN COALESCE(lp.discounted_price, lp.original_price) END) AS wishlist_price,
    MAX(CASE WHEN lp.source = 'piccolatoys' THEN COALESCE(lp.discounted_price, lp.original_price) END) AS piccolatoys_price,
    MAX(CASE WHEN lp.source = 'kubiki' THEN COALESCE(lp.discounted_price, lp.original_price) END) AS kubiki_price
FROM products p
LEFT JOIN latest_prices lp ON lp.product_id = p.id
GROUP BY p.id, p.mcode, p.barcode, p.invoice_code, p.name_ka, p.name_en;
