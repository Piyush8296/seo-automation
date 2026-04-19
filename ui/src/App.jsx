import { Routes, Route } from 'react-router-dom'
import Home from './pages/Home'
import AuditDetail from './pages/AuditDetail'
import AuditVault from './pages/AuditVault'
import ChecksCatalog from './pages/ChecksCatalog'
import Settings from './pages/Settings'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/audit/:id" element={<AuditDetail />} />
      <Route path="/vault" element={<AuditVault />} />
      <Route path="/checks" element={<ChecksCatalog />} />
      <Route path="/settings" element={<Settings />} />
    </Routes>
  )
}
