import { useQuery } from '@tanstack/react-query'
import { getProducts, getScrapeRuns } from '../api/client'
import { Activity, Package, Clock } from 'lucide-react'

export default function DashboardPage() {
  const { data: products } = useQuery({
    queryKey: ['products', { page: 1, per_page: 1 }],
    queryFn: () => getProducts({ page: 1, per_page: 1 }),
  })

  const { data: runs } = useQuery({
    queryKey: ['scrapeRuns'],
    queryFn: getScrapeRuns,
  })

  const totalProducts = products?.total ?? 0
  const lastRun = runs?.[0]
  const completedRuns = runs?.filter(r => r.status === 'completed').length ?? 0

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Dashboard</h2>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <div className="flex items-center gap-3 mb-2">
            <Package className="text-blue-600" size={24} />
            <span className="text-sm text-gray-500">Total Products</span>
          </div>
          <p className="text-3xl font-bold text-gray-900">{totalProducts}</p>
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <div className="flex items-center gap-3 mb-2">
            <Activity className="text-green-600" size={24} />
            <span className="text-sm text-gray-500">Completed Scrapes</span>
          </div>
          <p className="text-3xl font-bold text-gray-900">{completedRuns}</p>
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <div className="flex items-center gap-3 mb-2">
            <Clock className="text-orange-600" size={24} />
            <span className="text-sm text-gray-500">Last Scrape</span>
          </div>
          <p className="text-lg font-medium text-gray-900">
            {lastRun
              ? new Date(lastRun.created_at).toLocaleString('ka-GE')
              : 'N/A'}
          </p>
          {lastRun && (
            <span className={`text-xs px-2 py-0.5 rounded-full ${
              lastRun.status === 'completed' ? 'bg-green-100 text-green-700' :
              lastRun.status === 'running' ? 'bg-blue-100 text-blue-700' :
              lastRun.status === 'failed' ? 'bg-red-100 text-red-700' :
              'bg-gray-100 text-gray-700'
            }`}>
              {lastRun.status}
            </span>
          )}
        </div>
      </div>

      {runs && runs.length > 0 && (
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Recent Scrape Runs</h3>
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-500 border-b">
                <th className="pb-2">Source</th>
                <th className="pb-2">Status</th>
                <th className="pb-2">Trigger</th>
                <th className="pb-2">Found</th>
                <th className="pb-2">Saved</th>
                <th className="pb-2">Errors</th>
                <th className="pb-2">Date</th>
              </tr>
            </thead>
            <tbody>
              {runs.slice(0, 10).map(run => (
                <tr key={run.id} className="border-b last:border-0">
                  <td className="py-2 font-medium">{run.source}</td>
                  <td className="py-2">
                    <span className={`px-2 py-0.5 rounded-full text-xs ${
                      run.status === 'completed' ? 'bg-green-100 text-green-700' :
                      run.status === 'running' ? 'bg-blue-100 text-blue-700' :
                      run.status === 'failed' ? 'bg-red-100 text-red-700' :
                      'bg-gray-100 text-gray-700'
                    }`}>
                      {run.status}
                    </span>
                  </td>
                  <td className="py-2">{run.trigger_type}</td>
                  <td className="py-2">{run.products_found}</td>
                  <td className="py-2">{run.products_saved}</td>
                  <td className="py-2">{run.errors_count}</td>
                  <td className="py-2 text-gray-500">
                    {new Date(run.created_at).toLocaleString('ka-GE')}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
