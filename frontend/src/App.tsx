import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout/Layout'
import DashboardPage from './pages/DashboardPage'
import ProductsPage from './pages/ProductsPage'
import ComparisonPage from './pages/ComparisonPage'
import ScraperPage from './pages/ScraperPage'

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/products" element={<ProductsPage />} />
        <Route path="/comparison" element={<ComparisonPage />} />
        <Route path="/scraper" element={<ScraperPage />} />
      </Route>
    </Routes>
  )
}
