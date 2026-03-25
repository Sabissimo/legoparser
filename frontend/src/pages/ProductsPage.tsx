import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getProducts, exportProductsUrl, exportProductsCsvUrl } from '../api/client'
import { Search, Download, ChevronLeft, ChevronRight } from 'lucide-react'

export default function ProductsPage() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [sortBy, setSortBy] = useState('name_ka')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc')
  const perPage = 20

  const { data, isLoading } = useQuery({
    queryKey: ['products', { page, per_page: perPage, search, sort_by: sortBy, sort_order: sortOrder }],
    queryFn: () => getProducts({ page, per_page: perPage, search, sort_by: sortBy, sort_order: sortOrder }),
  })

  const products = data?.data ?? []
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
        <h2 className="text-2xl font-bold text-gray-900">Products</h2>
        <div className="flex gap-2">
          <a href={exportProductsUrl({ search })}
            className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg text-sm hover:bg-green-700 transition-colors">
            <Download size={16} /> Excel
          </a>
          <a href={exportProductsCsvUrl({ search })}
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
            placeholder="Search by name, LEGO ID, barcode, MCode..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      </div>

      <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-gray-50 text-left text-gray-600 border-b">
              {[
                { key: 'name_ka', label: 'Name' },
                { key: 'invoice_code', label: 'LEGO ID' },
                { key: 'mcode', label: 'MCode' },
                { key: 'barcode', label: 'Barcode' },
              ].map(({ key, label }) => (
                <th
                  key={key}
                  className="px-4 py-3 font-medium cursor-pointer hover:text-gray-900"
                  onClick={() => handleSort(key)}
                >
                  {label}
                  {sortBy === key && (sortOrder === 'asc' ? ' ↑' : ' ↓')}
                </th>
              ))}
              <th className="px-4 py-3 font-medium">Sources</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-400">Loading...</td></tr>
            ) : products.length === 0 ? (
              <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-400">No products found</td></tr>
            ) : (
              products.map(product => (
                <tr key={product.id} className="border-b last:border-0 hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-900">{product.name_ka}</td>
                  <td className="px-4 py-3 text-blue-600 font-mono text-xs">{product.invoice_code ?? '-'}</td>
                  <td className="px-4 py-3 text-gray-500 font-mono text-xs">{product.mcode ?? '-'}</td>
                  <td className="px-4 py-3 text-gray-500 font-mono text-xs">{product.barcode ?? '-'}</td>
                  <td className="px-3 py-3">
                    <div className="flex flex-wrap gap-1">
                      {(product.sources ?? []).map(src => (
                        <span key={src} className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-gray-100 text-gray-600">
                          {src.replace('_', ' ')}
                        </span>
                      ))}
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between mt-4">
          <p className="text-sm text-gray-500">
            Page {page} of {totalPages} ({data?.total ?? 0} total)
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="p-2 rounded-lg border border-gray-300 disabled:opacity-40 hover:bg-gray-100"
            >
              <ChevronLeft size={16} />
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="p-2 rounded-lg border border-gray-300 disabled:opacity-40 hover:bg-gray-100"
            >
              <ChevronRight size={16} />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
