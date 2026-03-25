import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { startScrape, getScrapeStatus, getScrapeRuns } from '../api/client'
import { Play, Loader2 } from 'lucide-react'
import type { SiteSource } from '../types'

const sources: { value: string; label: string }[] = [
  { value: 'all', label: 'All Sources' },
  { value: 'biblusi_xs', label: 'Biblusi XS' },
  { value: 'biblusi_pepela', label: 'Biblusi Pepela' },
  { value: 'wishlist', label: 'Wishlist' },
  { value: 'piccolatoys', label: 'PiccolaToys' },
  { value: 'kubiki', label: 'Kubiki' },
  { value: 'wolt_xs', label: 'Wolt XS' },
  { value: 'wolt_pepela', label: 'Wolt Pepela' },
  { value: 'glovo_xs', label: 'Glovo XS' },
  { value: 'glovo_pepela', label: 'Glovo Pepela' },
]

export default function ScraperPage() {
  const [source, setSource] = useState('all')
  const queryClient = useQueryClient()

  const { data: status } = useQuery({
    queryKey: ['scrapeStatus'],
    queryFn: getScrapeStatus,
    refetchInterval: 2000,
  })

  const { data: runs } = useQuery({
    queryKey: ['scrapeRuns'],
    queryFn: getScrapeRuns,
  })

  const mutation = useMutation({
    mutationFn: (src: string) => startScrape(src),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['scrapeStatus'] })
      queryClient.invalidateQueries({ queryKey: ['scrapeRuns'] })
    },
  })

  const isRunning = status?.is_running ?? false

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Scraper Control</h2>

      <div className="bg-white rounded-xl border border-gray-200 p-6 mb-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Start Scrape</h3>
        <div className="flex gap-4 items-center">
          <select
            value={source}
            onChange={(e) => setSource(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg text-sm"
          >
            {sources.map(s => (
              <option key={s.value} value={s.value}>{s.label}</option>
            ))}
          </select>
          <button
            onClick={() => mutation.mutate(source)}
            disabled={isRunning || mutation.isPending}
            className="flex items-center gap-2 px-6 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            {isRunning ? <Loader2 size={16} className="animate-spin" /> : <Play size={16} />}
            {isRunning ? 'Running...' : 'Start'}
          </button>
        </div>

        {isRunning && status?.runs && status.runs.length > 0 && (
          <div className="mt-4 p-4 bg-blue-50 rounded-lg">
            <p className="text-sm font-medium text-blue-800">Currently running:</p>
            {status.runs.map(run => (
              <div key={run.id} className="mt-2 text-sm text-blue-700">
                <span className="font-medium">{run.source}</span>
                {' '} - Found: {run.products_found}, Saved: {run.products_saved}
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="bg-white rounded-xl border border-gray-200 p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Scrape History</h3>
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-gray-500 border-b">
              <th className="pb-2 px-2">Source</th>
              <th className="pb-2 px-2">Status</th>
              <th className="pb-2 px-2">Trigger</th>
              <th className="pb-2 px-2">Found</th>
              <th className="pb-2 px-2">Saved</th>
              <th className="pb-2 px-2">Errors</th>
              <th className="pb-2 px-2">Started</th>
              <th className="pb-2 px-2">Duration</th>
            </tr>
          </thead>
          <tbody>
            {runs?.map(run => {
              const duration = run.started_at && run.completed_at
                ? `${Math.round((new Date(run.completed_at).getTime() - new Date(run.started_at).getTime()) / 1000)}s`
                : run.status === 'running' ? '...' : '-'

              return (
                <tr key={run.id} className="border-b last:border-0">
                  <td className="py-2 px-2 font-medium">{run.source}</td>
                  <td className="py-2 px-2">
                    <span className={`px-2 py-0.5 rounded-full text-xs ${
                      run.status === 'completed' ? 'bg-green-100 text-green-700' :
                      run.status === 'running' ? 'bg-blue-100 text-blue-700' :
                      run.status === 'failed' ? 'bg-red-100 text-red-700' :
                      'bg-gray-100 text-gray-700'
                    }`}>
                      {run.status}
                    </span>
                  </td>
                  <td className="py-2 px-2">{run.trigger_type}</td>
                  <td className="py-2 px-2">{run.products_found}</td>
                  <td className="py-2 px-2">{run.products_saved}</td>
                  <td className="py-2 px-2">{run.errors_count}</td>
                  <td className="py-2 px-2 text-gray-500">
                    {run.started_at ? new Date(run.started_at).toLocaleString('ka-GE') : '-'}
                  </td>
                  <td className="py-2 px-2">{duration}</td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
