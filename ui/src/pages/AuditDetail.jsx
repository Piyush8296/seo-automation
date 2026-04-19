import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Play, AlertCircle, Table as TableIcon, FileText, Globe2, Settings, CheckSquare, Database } from 'lucide-react'
import { useSSE } from '../hooks/useSSE'
import { api } from '../lib/api'
import CrawlProgress from '../components/CrawlProgress'
import ScoreCard from '../components/ScoreCard'
import IssueSummary from '../components/IssueSummary'
import ReportViewer from '../components/ReportViewer'
import IssueTable from '../components/IssueTable'
import StatusBadge from '../components/StatusBadge'

export default function AuditDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [audit, setAudit] = useState(null)
  const [loadErr, setLoadErr] = useState('')
  const [reportTab, setReportTab] = useState('table')

  useEffect(() => {
    api.getAudit(id)
      .then(setAudit)
      .catch(() => setLoadErr('Audit not found'))
  }, [id])

  const sseURL = audit?.status === 'running' ? api.eventsURL(id) : null
  const { lastEvent } = useSSE(sseURL)

  const [crawled, setCrawled] = useState(0)
  const [currentURL, setCurrentURL] = useState('')
  const [finalEvent, setFinalEvent] = useState(null)

  useEffect(() => {
    if (!lastEvent) return
    if (lastEvent.type === 'progress') {
      setCrawled(lastEvent.pages_crawled ?? 0)
      setCurrentURL(lastEvent.current_url ?? '')
    } else if (lastEvent.type === 'complete') {
      setFinalEvent(lastEvent)
      api.getAudit(id).then(setAudit).catch(() => {})
    } else if (lastEvent.type === 'error' || lastEvent.type === 'cancelled') {
      setFinalEvent(lastEvent)
      api.getAudit(id).then(setAudit).catch(() => {})
    }
  }, [lastEvent, id])

  const handleCancel = async () => {
    try { await api.cancelAudit(id) } catch {}
  }

  const handleRerun = async () => {
    if (!audit) return
    try {
      const record = await api.startAudit(audit.config)
      navigate(`/audit/${record.id}`)
    } catch {}
  }

  if (loadErr) {
    return (
      <div className="min-h-screen bg-surface flex items-center justify-center">
        <div className="card p-8 text-center max-w-sm">
          <AlertCircle className="mx-auto mb-3" size={32} style={{ color: '#ffb4ab' }} />
          <p className="text-on-surface-variant">{loadErr}</p>
          <button onClick={() => navigate('/')} className="btn-ghost mt-4 mx-auto">
            <ArrowLeft size={14} /> Go back
          </button>
        </div>
      </div>
    )
  }

  if (!audit) {
    return (
      <div className="min-h-screen bg-surface flex items-center justify-center">
        <span
          className="w-8 h-8 rounded-full animate-spin"
          style={{ border: '2px solid #2f3540', borderTopColor: '#3fe56c' }}
        />
      </div>
    )
  }

  const isRunning = audit.status === 'running'
  const isComplete = audit.status === 'complete'
  const isFailed = audit.status === 'failed' || audit.status === 'cancelled'

  const score = finalEvent?.health_score ?? audit.health_score
  const grade = finalEvent?.grade ?? audit.grade
  const errors = finalEvent?.error_count ?? audit.error_count
  const warnings = finalEvent?.warn_count ?? audit.warn_count
  const notices = finalEvent?.notice_count ?? audit.notice_count
  const pages = finalEvent?.page_count ?? audit.page_count

  return (
    <div className="h-screen bg-surface flex overflow-hidden">
      {/* Sidebar */}
      <aside
        className="w-64 flex flex-col shrink-0 h-full"
        style={{ background: '#161c26' }}
      >
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
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Back to Observatory</span>
            </button>
            <button
              onClick={() => navigate('/vault')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <Database size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Audit Vault</span>
            </button>
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

        {/* Audit metadata in sidebar */}
        <div
          className="px-4 py-4 flex flex-col gap-3 flex-1"
          style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}
        >
          <span className="uppercase tracking-widest text-on-surface-variant mb-1" style={{ fontSize: '9px' }}>
            Audit Info
          </span>
          <StatusBadge status={audit.status} />
          <div className="text-xs text-on-surface/70 font-mono break-all">{audit.url}</div>
          {[
            { label: 'ID',          value: audit.id?.slice(0, 12) + '…' },
            { label: 'Started',     value: new Date(audit.created_at).toLocaleString() },
            { label: 'Concurrency', value: audit.config?.concurrency ?? '—' },
            { label: 'Max depth',   value: audit.config?.max_depth === -1 ? 'unlimited' : audit.config?.max_depth },
          ].map(({ label, value }) => (
            <div key={label}>
              <div className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '8px' }}>{label}</div>
              <div className="text-xs text-on-surface/70 font-mono mt-0.5">{value}</div>
            </div>
          ))}
          <button onClick={handleRerun} className="btn-ghost text-xs mt-auto justify-start">
            <Play size={12} /> Re-run audit
          </button>
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
          <StatusBadge status={audit.status} />
          <span className="text-sm text-on-surface-variant truncate flex-1 min-w-0" title={audit.url}>
            {audit.url}
          </span>
        </header>

      <main className="flex-1 overflow-y-auto px-6 py-8 flex flex-col gap-6">

        {isRunning && (
          <CrawlProgress
            crawled={crawled}
            currentURL={currentURL}
            maxPages={audit.config?.max_pages ?? 0}
            onCancel={handleCancel}
          />
        )}

        {isFailed && (
          <div className="card p-6" style={{ borderLeft: '4px solid #ffb4ab' }}>
            <div className="flex items-center gap-3 mb-2">
              <AlertCircle size={20} style={{ color: '#ffb4ab' }} className="shrink-0" />
              <span className="font-semibold capitalize" style={{ color: '#ffb4ab' }}>{audit.status}</span>
            </div>
            {audit.error && <p className="text-sm text-on-surface-variant ml-8">{audit.error}</p>}
            <button onClick={handleRerun} className="btn-primary mt-4 w-fit">
              <Play size={14} /> Retry
            </button>
          </div>
        )}

        {(isComplete || finalEvent?.type === 'complete') && (
          <>
            <div className="grid grid-cols-3 gap-4">
              <ScoreCard label="Overall"  score={score}              grade={grade} />
              <ScoreCard label="Desktop"  score={audit.desktop_score} grade={audit.desktop_score >= 90 ? 'A' : audit.desktop_score >= 80 ? 'B' : audit.desktop_score >= 70 ? 'C' : audit.desktop_score >= 50 ? 'D' : 'F'} />
              <ScoreCard label="Mobile"   score={audit.mobile_score}  grade={audit.mobile_score  >= 90 ? 'A' : audit.mobile_score  >= 80 ? 'B' : audit.mobile_score  >= 70 ? 'C' : audit.mobile_score  >= 50 ? 'D' : 'F'} />
            </div>

            <IssueSummary
              errors={errors}
              warnings={warnings}
              notices={notices}
              pages={pages}
            />

            <div className="flex flex-col gap-3">
              <div
                className="flex items-center gap-1"
                style={{ borderBottom: '1px solid rgba(60,74,60,0.4)' }}
              >
                <TabButton active={reportTab === 'table'} onClick={() => setReportTab('table')} icon={TableIcon} label="Issue Table" />
                <TabButton active={reportTab === 'html'}  onClick={() => setReportTab('html')}  icon={FileText}  label="HTML Report" />
              </div>
              {reportTab === 'table' ? <IssueTable auditId={id} /> : <ReportViewer auditId={id} />}
            </div>
          </>
        )}

      </main>
      </div>
    </div>
  )
}

function TabButton({ active, onClick, icon: Icon, label }) {
  return (
    <button
      onClick={onClick}
      className={`inline-flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors ${
        active
          ? 'text-primary border-primary'
          : 'text-on-surface-variant border-transparent hover:text-on-surface'
      }`}
      style={active ? { borderBottomColor: '#3fe56c', color: '#3fe56c' } : {}}
    >
      <Icon size={14} />
      {label}
    </button>
  )
}
