import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Activity, ArrowLeft, Play, AlertCircle, Table as TableIcon, FileText, Globe2, Settings, CheckSquare, Database, MapPinned, Search, Code2 } from 'lucide-react'
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

  const handleRerunUncapped = async () => {
    if (!audit) return
    try {
      const record = await api.startAudit({ ...audit.config, max_pages: 0 })
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
  const maxPages = audit.config?.max_pages ?? 0
  const capReached = maxPages > 0 && pages >= maxPages

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
            <button
              onClick={() => navigate('/local-seo')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <MapPinned size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Local SEO</span>
            </button>
            <button
              onClick={() => navigate('/search-integrations')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <Search size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>GSC + Bing</span>
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
            { label: 'Max pages',   value: maxPages === 0 ? 'unlimited' : maxPages.toLocaleString() },
            { label: 'Scope',       value: audit.config?.scope ?? 'host' },
            { label: 'Sitemap',     value: audit.config?.sitemap_mode ?? 'discover' },
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

            {capReached && (
              <div
                className="rounded-xl px-4 py-3 flex items-center gap-3"
                style={{ background: 'rgba(255,183,174,0.08)', border: '1px solid rgba(255,183,174,0.22)' }}
              >
                <AlertCircle size={18} style={{ color: '#ffb7ae' }} className="shrink-0" />
                <div className="min-w-0">
                  <div className="text-on-surface text-sm font-medium">
                    Crawl stopped at the configured {maxPages.toLocaleString()} page cap.
                  </div>
                  <div className="text-on-surface-variant mt-0.5" style={{ fontSize: '12px' }}>
                    Increase Max Pages or set it to 0 for unlimited to crawl beyond this sample.
                  </div>
                </div>
                <button onClick={handleRerunUncapped} className="btn-ghost ml-auto shrink-0">
                  <Play size={13} /> Re-run uncapped
                </button>
              </div>
            )}

            <div className="flex flex-col gap-3">
              <div
                className="flex items-center gap-1"
                style={{ borderBottom: '1px solid rgba(60,74,60,0.4)' }}
              >
                <TabButton active={reportTab === 'table'} onClick={() => setReportTab('table')} icon={TableIcon} label="Issue Table" />
                <TabButton active={reportTab === 'evidence'} onClick={() => setReportTab('evidence')} icon={Activity} label="Crawler Evidence" />
                <TabButton active={reportTab === 'rendered'} onClick={() => setReportTab('rendered')} icon={Code2} label="Rendered SEO" />
                <TabButton active={reportTab === 'html'}  onClick={() => setReportTab('html')}  icon={FileText}  label="HTML Report" />
              </div>
              {reportTab === 'table' && <IssueTable auditId={id} />}
              {reportTab === 'evidence' && <CrawlerEvidencePanel auditId={id} />}
              {reportTab === 'rendered' && <RenderedSEOPanel auditId={id} />}
              {reportTab === 'html' && <ReportViewer auditId={id} />}
            </div>
          </>
        )}

      </main>
      </div>
    </div>
  )
}

const EVIDENCE_META = {
  pass: { label: 'Pass', color: '#3fe56c', bg: 'rgba(63,229,108,0.08)' },
  warning: { label: 'Warning', color: '#ffb7ae', bg: 'rgba(255,183,174,0.08)' },
  fail: { label: 'Fail', color: '#ffb4ab', bg: 'rgba(255,180,171,0.1)' },
  needs_input: { label: 'Needs input', color: '#9cc7ff', bg: 'rgba(156,199,255,0.08)' },
  info: { label: 'Info', color: '#dde2f1', bg: 'rgba(221,226,241,0.06)' },
}

function CrawlerEvidencePanel({ auditId }) {
  return (
    <EvidencePanel
      auditId={auditId}
      field="crawler_evidence"
      emptyMessage="Crawler evidence was not enabled for this audit."
    />
  )
}

function RenderedSEOPanel({ auditId }) {
  return (
    <EvidencePanel
      auditId={auditId}
      field="rendered_seo"
      emptyMessage="Rendered JavaScript SEO checks were not enabled for this audit."
    />
  )
}

function EvidencePanel({ auditId, field, emptyMessage }) {
  const [items, setItems] = useState(null)
  const [err, setErr] = useState('')

  useEffect(() => {
    api.getReportJSON(auditId)
      .then((report) => {
        setItems(report?.[field] ?? [])
        setErr('')
      })
      .catch((e) => setErr(e.message || 'Failed to load evidence'))
  }, [auditId, field])

  if (err) {
    return (
      <div className="card p-6 text-sm" style={{ color: '#ffb4ab' }}>
        {err}
      </div>
    )
  }

  if (!items) {
    return (
      <div className="card p-12 flex items-center justify-center">
        <span className="w-8 h-8 rounded-full animate-spin" style={{ border: '2px solid #2f3540', borderTopColor: '#3fe56c' }} />
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <div className="card p-6 text-on-surface-variant text-sm">
        {emptyMessage}
      </div>
    )
  }

  return (
    <div className="card overflow-hidden">
      <div
        className="grid gap-3 px-4 py-2 text-on-surface-variant uppercase tracking-widest"
        style={{ gridTemplateColumns: '100px minmax(220px, 1.2fr) 120px minmax(240px, 1fr)', fontSize: '9px', borderBottom: '1px solid rgba(60,74,60,0.24)' }}
      >
        <span>ID</span>
        <span>Check</span>
        <span>Status</span>
        <span>Evidence</span>
      </div>
      {items.map((item) => {
        const meta = EVIDENCE_META[item.status] ?? EVIDENCE_META.info
        return (
          <div
            key={item.id}
            className="grid gap-3 px-4 py-3 items-start"
            style={{ gridTemplateColumns: '100px minmax(220px, 1.2fr) 120px minmax(240px, 1fr)', borderTop: '1px solid rgba(60,74,60,0.16)' }}
          >
            <code className="font-mono text-primary" style={{ fontSize: '11px' }}>{item.id}</code>
            <div>
              <div className="text-on-surface text-sm leading-snug">{item.name}</div>
              {item.details && <div className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '11px' }}>{item.details}</div>}
            </div>
            <span className="w-fit rounded-full px-2 py-0.5 font-mono" style={{ background: meta.bg, color: meta.color, fontSize: '10px' }}>
              {meta.label}
            </span>
            <div>
              <div className="text-on-surface" style={{ fontSize: '12px' }}>{item.message}</div>
              {item.evidence?.length > 0 && (
                <div className="mt-2 flex flex-col gap-1">
                  {item.evidence.slice(0, 8).map((line, idx) => (
                    <code key={`${item.id}-${idx}`} className="text-on-surface-variant break-all" style={{ fontSize: '10px' }}>{line}</code>
                  ))}
                </div>
              )}
            </div>
          </div>
        )
      })}
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
