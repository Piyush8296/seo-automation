import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { RefreshCw } from 'lucide-react'
import AuditForm from '../components/AuditForm'
import AuditHistory from '../components/AuditHistory'
import { api } from '../lib/api'

export default function Home() {
  const [audits, setAudits] = useState([])
  const [historyLoading, setHistoryLoading] = useState(true)
  const [starting, setStarting] = useState(false)
  const [checkCount, setCheckCount] = useState(null)
  const navigate = useNavigate()

  const fetchAudits = useCallback(async () => {
    try {
      const data = await api.listAudits()
      setAudits(data ?? [])
    } catch (e) {
      console.error('Failed to load audits:', e)
    } finally {
      setHistoryLoading(false)
    }
  }, [])

  useEffect(() => { fetchAudits() }, [fetchAudits])

  useEffect(() => {
    api.getCheckCatalog()
      .then((c) => setCheckCount(c?.total ?? null))
      .catch((e) => console.error('Failed to load check catalog:', e))
  }, [])

  // Auto-refresh while any audit is still running
  useEffect(() => {
    const hasRunning = audits.some((a) => a.status === 'running')
    if (!hasRunning) return
    const t = setInterval(fetchAudits, 3000)
    return () => clearInterval(t)
  }, [audits, fetchAudits])

  const handleStart = async (req) => {
    setStarting(true)
    try {
      const record = await api.startAudit(req)
      navigate(`/audit/${record.id}`)
    } finally {
      setStarting(false)
    }
  }

  const handleDelete = async (id) => {
    await api.deleteAudit(id)
    setAudits((prev) => prev.filter((a) => a.id !== id))
  }

  const handleRerun = async (audit) => {
    const record = await api.startAudit(audit.config)
    navigate(`/audit/${record.id}`)
  }

  return (
    <div className="min-h-screen bg-gray-950">
      {/* Header */}
      <header className="border-b border-gray-800 bg-gray-900/50 backdrop-blur sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-6 h-14 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-xl">🔍</span>
            <span className="font-bold text-gray-100 tracking-tight">SEO Audit</span>
            <span className="hidden sm:inline text-xs bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded-full">
              {checkCount !== null ? `${checkCount} checks` : '… checks'}
            </span>
          </div>
          <button
            onClick={fetchAudits}
            className="btn-ghost"
            title="Refresh history"
          >
            <RefreshCw size={14} />
            Refresh
          </button>
        </div>
      </header>

      {/* Body */}
      <main className="max-w-7xl mx-auto px-6 py-8 grid grid-cols-1 lg:grid-cols-[420px_1fr] gap-8 items-start">
        {/* Left — new audit form */}
        <div className="lg:sticky lg:top-20">
          <AuditForm onSubmit={handleStart} loading={starting} />
        </div>

        {/* Right — history */}
        <div>
          <AuditHistory
            audits={audits}
            loading={historyLoading}
            onDelete={handleDelete}
            onRerun={handleRerun}
          />
        </div>
      </main>
    </div>
  )
}
