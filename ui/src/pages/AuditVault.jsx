import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Globe2, Settings, CheckSquare, ArrowLeft, Database,
  Play, ExternalLink, Trash2, RefreshCw, Search, AlertCircle,
} from 'lucide-react'
import { api } from '../lib/api'
import StatusBadge from '../components/StatusBadge'

function fmt(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleString(undefined, {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

function gradeColor(grade) {
  switch (grade) {
    case 'A': return '#3fe56c'
    case 'B': return '#8ed793'
    case 'C': return '#ffb7ae'
    default:  return '#ffb4ab'
  }
}

function ScorePill({ score, grade }) {
  if (!grade) return null
  return (
    <div className="flex flex-col items-center justify-center w-12 h-12 rounded-xl shrink-0" style={{ background: '#1a202a' }}>
      <span className="font-display font-bold text-base leading-none" style={{ color: gradeColor(grade) }}>{grade}</span>
      {score != null && (
        <span className="text-on-surface-variant leading-none mt-0.5" style={{ fontSize: '9px' }}>
          {Math.round(score)}
        </span>
      )}
    </div>
  )
}

function AuditRow({ audit, onDelete, onRerun }) {
  const navigate = useNavigate()
  const [deleting, setDeleting] = useState(false)

  const handleDelete = async (e) => {
    e.stopPropagation()
    setDeleting(true)
    try { await onDelete(audit.id) } finally { setDeleting(false) }
  }

  const domain = audit.url?.replace(/^https?:\/\//, '').replace(/\/$/, '') ?? '—'

  return (
    <div
      className="group flex items-center gap-4 px-5 py-4 hover:bg-surface-bright transition-colors cursor-pointer rounded-xl"
      style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.2)' }}
      onClick={() => navigate(`/audit/${audit.id}`)}
    >
      <ScorePill score={audit.health_score} grade={audit.grade} />

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <StatusBadge status={audit.status} />
          <span className="text-on-surface font-medium text-sm truncate">{domain}</span>
        </div>
        <div className="flex items-center gap-3 flex-wrap" style={{ fontSize: '11px' }}>
          <span className="text-on-surface-variant">{fmt(audit.created_at)}</span>
          {audit.page_count > 0 && (
            <span className="text-on-surface-variant">{audit.page_count} pages</span>
          )}
          {audit.error_count > 0 && (
            <span style={{ color: '#ffb4ab' }}>{audit.error_count} errors</span>
          )}
          {audit.warn_count > 0 && (
            <span style={{ color: '#ffb7ae' }}>{audit.warn_count} warnings</span>
          )}
          {audit.notice_count > 0 && (
            <span style={{ color: '#8ed793' }}>{audit.notice_count} notices</span>
          )}
        </div>
      </div>

      {/* Actions — visible on hover */}
      <div
        className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
        onClick={(e) => e.stopPropagation()}
      >
        <button
          onClick={(e) => { e.stopPropagation(); onRerun(audit) }}
          className="p-2 text-on-surface-variant hover:text-primary rounded-lg hover:bg-surface-container-high transition-colors"
          title="Re-run"
        >
          <Play size={13} />
        </button>
        <a
          href={api.reportURL(audit.id)}
          target="_blank"
          rel="noopener noreferrer"
          className="p-2 text-on-surface-variant hover:text-on-surface rounded-lg hover:bg-surface-container-high transition-colors"
          title="Open HTML report"
        >
          <ExternalLink size={13} />
        </a>
        <button
          onClick={handleDelete}
          disabled={deleting}
          className="p-2 rounded-lg hover:bg-surface-container-high transition-colors disabled:opacity-30"
          style={{ color: '#ffb4ab' }}
          title="Delete"
        >
          <Trash2 size={13} />
        </button>
      </div>
    </div>
  )
}

const STATUS_FILTERS = [
  { value: '',         label: 'All' },
  { value: 'complete', label: 'Complete' },
  { value: 'running',  label: 'Running' },
  { value: 'failed',   label: 'Failed' },
]

export default function AuditVault() {
  const navigate = useNavigate()
  const [audits, setAudits] = useState([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')

  const fetchAudits = useCallback(async () => {
    try {
      const data = await api.listAudits()
      setAudits(data ?? [])
    } catch (e) {
      console.error('Failed to load audits:', e)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { fetchAudits() }, [fetchAudits])

  // Poll while any audit is running
  useEffect(() => {
    const hasRunning = audits.some((a) => a.status === 'running')
    if (!hasRunning) return
    const t = setInterval(fetchAudits, 3000)
    return () => clearInterval(t)
  }, [audits, fetchAudits])

  const handleDelete = async (id) => {
    await api.deleteAudit(id)
    setAudits((prev) => prev.filter((a) => a.id !== id))
  }

  const handleRerun = async (audit) => {
    const record = await api.startAudit(audit.config)
    navigate(`/audit/${record.id}`)
  }

  const q = search.trim().toLowerCase()
  const visible = audits
    .filter((a) => !statusFilter || a.status === statusFilter)
    .filter((a) => !q || (a.url ?? '').toLowerCase().includes(q))

  const counts = {
    complete: audits.filter((a) => a.status === 'complete').length,
    running:  audits.filter((a) => a.status === 'running').length,
    failed:   audits.filter((a) => a.status === 'failed').length,
  }

  return (
    <div className="h-screen bg-surface flex overflow-hidden">
      {/* Sidebar */}
      <aside className="w-64 flex flex-col shrink-0 h-full" style={{ background: '#161c26' }}>
        <div className="px-6 pt-6 pb-4 shrink-0">
          <div className="flex items-center gap-3 mb-5">
            <div
              className="w-8 h-8 rounded-lg flex items-center justify-center shrink-0"
              style={{ background: 'linear-gradient(135deg, #3fe56c, #00c853)' }}
            >
              <Globe2 size={14} style={{ color: '#003912' }} />
            </div>
            <span className="font-display font-bold text-on-surface tracking-tight text-sm">SEO Observatory</span>
          </div>
          <nav className="space-y-0.5">
            <button
              onClick={() => navigate('/')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <ArrowLeft size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Observatory</span>
            </button>
            <div className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg"
              style={{ background: '#1a202a', borderRight: '2px solid #00c853', color: '#3fe56c' }}
            >
              <Database size={15} />
              <span className="uppercase tracking-widest font-medium" style={{ fontSize: '9px' }}>Audit Vault</span>
            </div>
            <button
              onClick={() => navigate('/settings')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <Settings size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Settings</span>
            </button>
            <button
              onClick={() => navigate('/checks')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <CheckSquare size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Checks Catalog</span>
            </button>
          </nav>
        </div>

        {/* Sidebar stats */}
        <div className="px-4 py-4 flex flex-col gap-3" style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}>
          <span className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>Summary</span>
          {[
            { label: 'Total Audits',    value: audits.length,    color: '#dde2f1' },
            { label: 'Complete',        value: counts.complete,  color: '#3fe56c' },
            { label: 'Running',         value: counts.running,   color: '#8ed793' },
            { label: 'Failed',          value: counts.failed,    color: '#ffb4ab' },
          ].map(({ label, value, color }) => (
            <div key={label} className="flex items-center justify-between">
              <span className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '8px' }}>{label}</span>
              <span className="text-sm font-display font-bold" style={{ color }}>{value}</span>
            </div>
          ))}
        </div>
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden">
        {/* Header */}
        <header
          className="h-14 flex items-center gap-4 px-6 shrink-0"
          style={{
            background: 'rgba(14,19,30,0.72)',
            backdropFilter: 'blur(16px)',
            WebkitBackdropFilter: 'blur(16px)',
            borderBottom: '1px solid rgba(221,226,241,0.06)',
          }}
        >
          <Database size={16} style={{ color: '#3fe56c' }} />
          <span className="font-display font-semibold text-on-surface text-sm">Audit Vault</span>
          <span
            className="ml-1 px-2.5 py-0.5 rounded-full text-on-surface-variant"
            style={{ background: '#1a202a', fontSize: '10px' }}
          >
            {audits.length} {audits.length === 1 ? 'audit' : 'audits'}
          </span>
          <div className="flex-1" />
          <button
            onClick={fetchAudits}
            className="btn-ghost"
            title="Refresh"
          >
            <RefreshCw size={13} />
            Refresh
          </button>
          <button
            onClick={() => navigate('/')}
            className="btn-primary text-sm py-2"
          >
            <Play size={13} />
            New Audit
          </button>
        </header>

        <main className="flex-1 overflow-y-auto px-6 py-6">
          {/* Filter bar */}
          <div className="flex items-center gap-3 mb-5">
            <div className="relative flex-1 max-w-sm">
              <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant" />
              <input
                className="input pl-9 text-sm"
                placeholder="Filter by URL…"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <div className="flex items-center gap-1 rounded-lg p-1" style={{ background: '#161c26' }}>
              {STATUS_FILTERS.map((f) => (
                <button
                  key={f.value}
                  onClick={() => setStatusFilter(f.value)}
                  className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors"
                  style={
                    statusFilter === f.value
                      ? { background: '#242a35', color: '#3fe56c' }
                      : { color: '#bbcbb8' }
                  }
                >
                  {f.label}
                  {f.value && (
                    <span className="ml-1.5 opacity-60">{counts[f.value] ?? 0}</span>
                  )}
                </button>
              ))}
            </div>
          </div>

          {/* Audit list */}
          {loading ? (
            <div className="flex justify-center py-20">
              <span className="w-8 h-8 rounded-full animate-spin" style={{ border: '2px solid #2f3540', borderTopColor: '#3fe56c' }} />
            </div>
          ) : visible.length === 0 ? (
            <div className="text-center py-20">
              <AlertCircle className="mx-auto mb-3" size={28} style={{ color: '#3c4a3c' }} />
              <p className="text-on-surface-variant text-sm">
                {audits.length === 0 ? 'No audits yet — start one from the Observatory.' : 'No audits match your filters.'}
              </p>
            </div>
          ) : (
            <div className="flex flex-col gap-2">
              {visible.map((audit) => (
                <AuditRow
                  key={audit.id}
                  audit={audit}
                  onDelete={handleDelete}
                  onRerun={handleRerun}
                />
              ))}
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
