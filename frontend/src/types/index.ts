export type SiteSource = 'biblusi_xs' | 'biblusi_pepela' | 'wolt_xs' | 'wolt_pepela' | 'glovo_xs' | 'glovo_pepela' | 'wishlist' | 'piccolatoys' | 'kubiki'

export interface Product {
  id: number
  mcode: string | null
  barcode: string | null
  invoice_code: string | null
  name_ka: string
  name_en: string | null
  image_url: string | null
  sources: string[]
  created_at: string
  updated_at: string
  latest_prices?: PriceEntry[]
}

export interface PriceEntry {
  id: number
  product_id: number
  source: SiteSource
  original_price: number | null
  discount_percent: number | null
  discounted_price: number | null
  in_stock: boolean
  source_url: string | null
  scraped_at: string
}

export interface PriceComparison {
  product_id: number
  mcode: string | null
  barcode: string | null
  invoice_code: string | null
  name_ka: string
  name_en: string | null
  biblusi_xs_price: number | null
  biblusi_pepela_price: number | null
  wolt_xs_price: number | null
  wolt_pepela_price: number | null
  glovo_xs_price: number | null
  glovo_pepela_price: number | null
  wishlist_price: number | null
  piccolatoys_price: number | null
  kubiki_price: number | null
}

export interface ScrapeRun {
  id: number
  source: SiteSource
  status: 'pending' | 'running' | 'completed' | 'failed'
  trigger_type: 'manual' | 'scheduled'
  products_found: number
  products_saved: number
  errors_count: number
  error_log: string | null
  started_at: string | null
  completed_at: string | null
  created_at: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface ScrapeStatus {
  is_running: boolean
  runs: ScrapeRun[]
}
