import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getComparison, exportComparisonUrl, exportComparisonCsvUrl } from '../api/client'
import { Download, Search, ChevronLeft, ChevronRight } from 'lucide-react'

const priceCols = [
  { key: 'biblusi_xs_price', label: 'Biblusi XS' },
  { key: 'biblusi_pepela_price', label: 'Biblusi Pepela' },
  { key: 'wolt_xs_price', label: 'Wolt XS' },
  { key: 'wolt_pepela_price', label: 'Wolt Pepela' },
  { key: 'glovo_xs_price', label: 'Glovo XS' },
  { key: 'glovo_pepela_price', label: 'Glovo Pepela' },
  { key: 'wishlist_price', label: 'Wishlist' },
  { key: 'piccolatoys_price', label: 'Piccola' },
  { key: 'kubiki_price', label: 'Kubiki' },
]

function priceColor(price: number | null, allPrices: (number | null)[]): string {
  if (price == null) return ''
  const valid = allPrices.filter((p): p is number => p != null)
  if (valid.length < 2) return ''
  const min = Math.min(...valid)
  const max = Math.max(...valid)
  if (price === min) return 'bg-green-100 text-green-800'
  if (price === max) return 'bg-red-100 text-red-800'
  return ''
}

export default function ComparisonPage() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [sortBy, setSortBy] = useState('name_ka')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc')
  const perPage = 20

  const { data, isLoading } = useQuery({
    queryKey: ['comparison', { page, per_page: perPage, search, sort_by: sortBy, sort_order: sortOrder }],
    queryFn: () => getComparison({ page, per_page: perPage, search, sort_by: sortBy, sort_order: sortOrder }),
  })

  const items = data?.data ?? []
  const totalPages = data?.total_pages ?? 0

  const handleSort = (col: string) => {
    if (sortBy === col) {
      setSortOrder(prev => prev === 'asc' ? 'desc' : 'asc')
    } else {
      setSortBy(col)
      setSortOrder('asc')
    }
    setPage(1)
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Price Comparison</h2>
        <div className="flex gap-2">
          <a href={exportComparisonUrl({ search })}
            className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg text-sm hover:bg-green-700 transition-colors">
            <Download size={16} /> Excel
          </a>
          <a href={exportComparisonCsvUrl({ search })}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700 transition-colors">
            <Download size={16} /> CSV (Qlik)
          </a>
        </div>
      </div>

      <div className="flex gap-4 mb-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
          <input
            type="text"
            placeholder="Search by name, LEGO ID..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      </div>

      <div className="bg-white rounded-xl border border-gray-200 overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-gray-50 text-left text-gray-600 border-b">
              {[
                { key: 'name_ka', label: 'Name' },
                { key: 'invoice_code', label: 'LEGO ID' },
              ].map(({ key, label }) => (
                <th
                  key={key}
                  className="px-3 py-3 font-medium cursor-pointer hover:text-gray-900"
                  onClick={() => handleSort(key)}
                >
                  {label}
                  {sortBy === key && (sortOrder === 'asc' ? ' ↑' : ' ↓')}
                </th>
              ))}
              {priceCols.map(s => (
                <th
                  key={s.key}
                  className="px-2 py-3 font-medium text-center text-xs cursor-pointer hover:text-gray-900"
                  onClick={() => handleSort(s.key)}
                >
                  {s.label} ₾
                  {sortBy === s.key && (sortOrder === 'asc' ? ' ↑' : ' ↓')}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={10} className="px-4 py-8 text-center text-gray-400">Loading...</td></tr>
            ) : items.length === 0 ? (
              <tr><td colSpan={10} className="px-4 py-8 text-center text-gray-400">No data</td></tr>
            ) : (
              items.map(item => {
                const prices = priceCols.map(c => (item as Record<string, unknown>)[c.key] as number | null)
                return (
                  <tr key={item.product_id} className="border-b last:border-0 hover:bg-gray-50">
                    <td className="px-3 py-2 font-medium text-gray-900 max-w-xs truncate">{item.name_ka}</td>
                    <td className="px-3 py-2 text-blue-600 font-mono text-xs">{item.invoice_code ?? '-'}</td>
                    {prices.map((price, i) => (
                      <td key={priceCols[i].key} className={`px-2 py-2 text-center font-mono text-xs ${priceColor(price, prices)}`}>
                        {price != null ? price.toFixed(2) : '-'}
                      </td>
                    ))}
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between mt-4">
          <p className="text-sm text-gray-500">Page {page} of {totalPages}</p>
          <div className="flex gap-2">
            <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page <= 1}
              className="p-2 rounded-lg border border-gray-300 disabled:opacity-40 hover:bg-gray-100">
              <ChevronLeft size={16} />
            </button>
            <button onClick={() => setPage(p => Math.min(totalPages, p + 1))} disabled={page >= totalPages}
              className="p-2 rounded-lg border border-gray-300 disabled:opacity-40 hover:bg-gray-100">
              <ChevronRight size={16} />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
