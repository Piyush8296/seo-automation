import { Routes, Route } from 'react-router-dom'
import Home from './pages/Home'
import AuditDetail from './pages/AuditDetail'
import ChecksCatalog from './pages/ChecksCatalog'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/audit/:id" element={<AuditDetail />} />
      <Route path="/checks" element={<ChecksCatalog />} />
    </Routes>
  )
}
