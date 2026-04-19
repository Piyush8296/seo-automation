import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Play, AlertCircle, Table as TableIcon, FileText } from 'lucide-react'
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
  const [reportTab, setReportTab] = useState('table') // 'table' | 'html'

  // Load initial record
  useEffect(() => {
    api.getAudit(id)
      .then(setAudit)
      .catch(() => setLoadErr('Audit not found'))
  }, [id])

  // SSE — only connect while running
  const sseURL = audit?.status === 'running' ? api.eventsURL(id) : null
  const { lastEvent, done } = useSSE(sseURL)

  // Merge SSE progress into local state
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
      // Reload the full record once crawl is done
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
      <div className="min-h-screen bg-gray-950 flex items-center justify-center">
        <div className="card p-8 text-center max-w-sm">
          <AlertCircle className="mx-auto text-red-400 mb-3" size={32} />
          <p className="text-gray-300">{loadErr}</p>
          <button onClick={() => navigate('/')} className="btn-ghost mt-4 mx-auto">
            <ArrowLeft size={14} /> Go back
          </button>
        </div>
      </div>
    )
  }

  if (!audit) {
    return (
      <div className="min-h-screen bg-gray-950 flex items-center justify-center">
        <span className="w-8 h-8 border-2 border-gray-700 border-t-emerald-500 rounded-full animate-spin" />
      </div>
    )
  }

  const isRunning = audit.status === 'running'
  const isComplete = audit.status === 'complete'
  const isFailed = audit.status === 'failed' || audit.status === 'cancelled'

  // Use SSE data if available, fall back to persisted record
  const score = finalEvent?.health_score ?? audit.health_score
  const grade = finalEvent?.grade ?? audit.grade
  const errors = finalEvent?.error_count ?? audit.error_count
  const warnings = finalEvent?.warn_count ?? audit.warn_count
  const notices = finalEvent?.notice_count ?? audit.notice_count
  const pages = finalEvent?.page_count ?? audit.page_count

  return (
    <div className="min-h-screen bg-gray-950">
      {/* Header */}
      <header className="border-b border-gray-800 bg-gray-900/50 backdrop-blur sticky top-0 z-40">
        <div className="max-w-5xl mx-auto px-6 h-14 flex items-center gap-4">
          <button onClick={() => navigate('/')} className="btn-ghost">
            <ArrowLeft size={15} /> Back
          </button>
          <div className="h-4 w-px bg-gray-700" />
          <StatusBadge status={audit.status} />
          <span className="text-sm text-gray-400 truncate flex-1 min-w-0" title={audit.url}>
            {audit.url}
          </span>
          <button onClick={handleRerun} className="btn-ghost text-xs hidden sm:flex">
            <Play size={12} /> Re-run
          </button>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-6 py-8 flex flex-col gap-6">

        {/* Running — show live progress */}
        {isRunning && (
          <CrawlProgress
            crawled={crawled}
            currentURL={currentURL}
            maxPages={audit.config?.max_pages ?? 0}
            onCancel={handleCancel}
          />
        )}

        {/* Failed / Cancelled */}
        {isFailed && (
          <div className="card p-6 border-red-500/20">
            <div className="flex items-center gap-3 mb-2">
              <AlertCircle className="text-red-400 shrink-0" size={20} />
              <span className="font-semibold text-red-300 capitalize">{audit.status}</span>
            </div>
            {audit.error && <p className="text-sm text-gray-500 ml-8">{audit.error}</p>}
            <button onClick={handleRerun} className="btn-primary mt-4 w-fit">
              <Play size={14} /> Retry
            </button>
          </div>
        )}

        {/* Complete — show scores + report */}
        {(isComplete || finalEvent?.type === 'complete') && (
          <>
            {/* Score cards */}
            <div className="grid grid-cols-3 gap-4">
              <ScoreCard label="Overall"  score={score}              grade={grade} />
              <ScoreCard label="Desktop"  score={audit.desktop_score} grade={audit.desktop_score >= 90 ? 'A' : audit.desktop_score >= 80 ? 'B' : audit.desktop_score >= 70 ? 'C' : audit.desktop_score >= 50 ? 'D' : 'F'} />
              <ScoreCard label="Mobile"   score={audit.mobile_score}  grade={audit.mobile_score  >= 90 ? 'A' : audit.mobile_score  >= 80 ? 'B' : audit.mobile_score  >= 70 ? 'C' : audit.mobile_score  >= 50 ? 'D' : 'F'} />
            </div>

            {/* Issue summary */}
            <IssueSummary
              errors={errors}
              warnings={warnings}
              notices={notices}
              pages={pages}
            />

            {/* Report — tabbed view */}
            <div className="flex flex-col gap-3">
              <div className="flex items-center gap-1 border-b border-gray-800">
                <TabButton
                  active={reportTab === 'table'}
                  onClick={() => setReportTab('table')}
                  icon={TableIcon}
                  label="Issue Table"
                />
                <TabButton
                  active={reportTab === 'html'}
                  onClick={() => setReportTab('html')}
                  icon={FileText}
                  label="HTML Report"
                />
              </div>

              {reportTab === 'table' ? (
                <IssueTable auditId={id} />
              ) : (
                <ReportViewer auditId={id} />
              )}
            </div>
          </>
        )}

        {/* Audit metadata */}
        <div className="card px-5 py-4">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
            {[
              { label: 'Audit ID',    value: audit.id },
              { label: 'Started',     value: new Date(audit.created_at).toLocaleString() },
              { label: 'Concurrency', value: audit.config?.concurrency ?? '—' },
              { label: 'Max depth',   value: audit.config?.max_depth === -1 ? 'unlimited' : audit.config?.max_depth },
            ].map(({ label, value }) => (
              <div key={label}>
                <div className="text-xs text-gray-500 uppercase tracking-wider">{label}</div>
                <div className="text-gray-300 font-mono text-xs mt-0.5 truncate" title={String(value)}>{value}</div>
              </div>
            ))}
          </div>
        </div>
      </main>
    </div>
  )
}

function TabButton({ active, onClick, icon: Icon, label }) {
  return (
    <button
      onClick={onClick}
      className={`inline-flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors ${
        active
          ? 'text-emerald-400 border-emerald-500'
          : 'text-gray-400 border-transparent hover:text-gray-200 hover:border-gray-700'
      }`}
    >
      <Icon size={14} />
      {label}
    </button>
  )
}
