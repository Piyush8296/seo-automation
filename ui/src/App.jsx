import { Routes, Route } from 'react-router-dom'
import Home from './pages/Home'
import AuditDetail from './pages/AuditDetail'
import AuditVault from './pages/AuditVault'
import ChecksCatalog from './pages/ChecksCatalog'
import LocalSEO from './pages/LocalSEO'
import SearchIntegrations from './pages/SearchIntegrations'
import Settings from './pages/Settings'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/audit/:id" element={<AuditDetail />} />
      <Route path="/vault" element={<AuditVault />} />
      <Route path="/checks" element={<ChecksCatalog />} />
      <Route path="/local-seo" element={<LocalSEO />} />
      <Route path="/search-integrations" element={<SearchIntegrations />} />
      <Route path="/settings" element={<Settings />} />
    </Routes>
  )
}
