import type { PaginatedResponse, Product, PriceComparison, ScrapeRun, ScrapeStatus } from '../types'

const BASE_URL = '/api'

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const resp = await fetch(`${BASE_URL}${url}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!resp.ok) {
    throw new Error(`HTTP ${resp.status}: ${resp.statusText}`)
  }
  return resp.json()
}

function toQueryString(params: Record<string, unknown>): string {
  const qs = new URLSearchParams()
  for (const [key, val] of Object.entries(params)) {
    if (val !== undefined && val !== null && val !== '') {
      qs.set(key, String(val))
    }
  }
  return qs.toString()
}

export interface ProductParams {
  page?: number
  per_page?: number
  sort_by?: string
  sort_order?: 'asc' | 'desc'
  search?: string
}

export async function getProducts(params: ProductParams = {}): Promise<PaginatedResponse<Product>> {
  const qs = toQueryString(params)
  return fetchJSON(`/products?${qs}`)
}

export async function getProductById(id: number): Promise<Product> {
  return fetchJSON(`/products/${id}`)
}

export async function getComparison(params: ProductParams = {}): Promise<PaginatedResponse<PriceComparison>> {
  const qs = toQueryString(params)
  return fetchJSON(`/products/comparison?${qs}`)
}

export async function startScrape(source: string = 'all'): Promise<void> {
  await fetchJSON('/scrape/run', {
    method: 'POST',
    body: JSON.stringify({ source }),
  })
}

export async function getScrapeRuns(): Promise<ScrapeRun[]> {
  return fetchJSON('/scrape/runs')
}

export async function getScrapeStatus(): Promise<ScrapeStatus> {
  return fetchJSON('/scrape/status')
}

export function exportProductsUrl(params: ProductParams = {}): string {
  const qs = toQueryString(params)
  return `/api/export/products?${qs}`
}

export function exportComparisonUrl(params: ProductParams = {}): string {
  const qs = toQueryString(params)
  return `/api/export/comparison?${qs}`
}

export function exportProductsCsvUrl(params: ProductParams = {}): string {
  const qs = toQueryString(params)
  return `/api/export/products/csv?${qs}`
}

export function exportComparisonCsvUrl(params: ProductParams = {}): string {
  const qs = toQueryString(params)
  return `/api/export/comparison/csv?${qs}`
}
