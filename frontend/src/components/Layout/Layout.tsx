import { NavLink, Outlet } from 'react-router-dom'
import { LayoutDashboard, Package, GitCompare, Play } from 'lucide-react'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/products', icon: Package, label: 'Products' },
  { to: '/comparison', icon: GitCompare, label: 'Comparison' },
  { to: '/scraper', icon: Play, label: 'Scraper' },
]

export default function Layout() {
  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white border-b border-gray-200 px-6 py-3">
        <div className="flex items-center gap-8">
          <h1 className="text-xl font-bold text-gray-900">LEGO Parser</h1>
          <div className="flex gap-1">
            {navItems.map(({ to, icon: Icon, label }) => (
              <NavLink
                key={to}
                to={to}
                end={to === '/'}
                className={({ isActive }) =>
                  `flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-blue-50 text-blue-700'
                      : 'text-gray-600 hover:bg-gray-100'
                  }`
                }
              >
                <Icon size={18} />
                {label}
              </NavLink>
            ))}
          </div>
        </div>
      </nav>
      <main className="p-6">
        <Outlet />
      </main>
    </div>
  )
}
