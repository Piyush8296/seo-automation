import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { Play, Settings, RefreshCw, Globe2, ChevronDown, ChevronUp, Trash2, ExternalLink } from 'lucide-react'
import { api } from '../lib/api'

const DEFAULTS = {
  url: '',
  max_depth: -1,
  max_pages: 0,
  concurrency: 10,
  timeout: '30s',
  platform: '',
  output_dir: '',
  validate_external_links: true,
  discover_resources: true,
}

function accentColor(status) {
  switch (status) {
    case 'complete':  return '#3fe56c'
    case 'running':   return '#8ed793'
    case 'failed':    return '#ffb4ab'
    default:          return '#3c4a3c'
  }
}

function statusColor(status) {
  switch (status) {
    case 'complete':  return '#3fe56c'
    case 'running':   return '#8ed793'
    case 'failed':    return '#ffb4ab'
    default:          return '#bbcbb8'
  }
}

function fmt(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleString(undefined, {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

function SidebarAuditItem({ audit, onDelete, onRerun, navigate }) {
  const [deleting, setDeleting] = useState(false)
  const handleDelete = async (e) => {
    e.stopPropagation()
    setDeleting(true)
    try { await onDelete(audit.id) } finally { setDeleting(false) }
  }
  return (
    <div
      className="relative group px-4 py-3 hover:bg-surface-bright cursor-pointer transition-colors"
      onClick={() => navigate(`/audit/${audit.id}`)}
    >
      <div
        className="absolute left-0 top-2 bottom-2 w-[3px] rounded-r-full"
        style={{ background: accentColor(audit.status) }}
      />
      <div className="flex items-center justify-between gap-2">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5 mb-0.5">
            <span className="w-1.5 h-1.5 rounded-full shrink-0" style={{ background: statusColor(audit.status) }} />
            <span className="text-on-surface text-xs font-medium truncate">{
              audit.url.replace(/^https?:\/\//, '').replace(/\/$/, '')
            }</span>
          </div>
          <div className="flex items-center gap-2" style={{ fontSize: '9px' }}>
            <span className="text-on-surface-variant">{fmt(audit.created_at)}</span>
            {audit.grade && (
              <span className="font-bold font-display" style={{ color: statusColor(audit.status) }}>
                {audit.grade}
              </span>
            )}
            {audit.error_count > 0 && (
              <span style={{ color: '#ffb4ab' }}>{audit.error_count} err</span>
            )}
          </div>
        </div>
        <div
          className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={(e) => { e.stopPropagation(); onRerun(audit) }}
            className="p-1 text-on-surface-variant hover:text-primary rounded transition-colors"
            title="Re-run"
          >
            <Play size={11} />
          </button>
          <a
            href={api.reportURL(audit.id)}
            target="_blank"
            rel="noopener noreferrer"
            className="p-1 text-on-surface-variant hover:text-on-surface rounded transition-colors"
            title="Open report"
          >
            <ExternalLink size={11} />
          </a>
          <button
            onClick={handleDelete}
            disabled={deleting}
            className="p-1 rounded transition-colors"
            style={{ color: '#ffb4ab' }}
            title="Delete"
          >
            <Trash2 size={11} />
          </button>
        </div>
      </div>
    </div>
  )
}

export default function Home() {
  const [audits, setAudits] = useState([])
  const [historyLoading, setHistoryLoading] = useState(true)
  const [starting, setStarting] = useState(false)
  const [checkCount, setCheckCount] = useState(null)
  const [form, setForm] = useState(DEFAULTS)
  const [advanced, setAdvanced] = useState(false)
  const [formError, setFormError] = useState('')
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
      .catch(() => {})
  }, [])

  useEffect(() => {
    const hasRunning = audits.some((a) => a.status === 'running')
    if (!hasRunning) return
    const t = setInterval(fetchAudits, 3000)
    return () => clearInterval(t)
  }, [audits, fetchAudits])

  const set = (k, v) => setForm((f) => ({ ...f, [k]: v }))

  const handleStart = async (e) => {
    e.preventDefault()
    setFormError('')
    const url = form.url.trim()
    if (!url) { setFormError('URL is required'); return }
    if (!/^https?:\/\//i.test(url)) { setFormError('URL must start with http:// or https://'); return }
    setStarting(true)
    try {
      const record = await api.startAudit({
        ...form, url,
        max_depth: Number(form.max_depth),
        max_pages: Number(form.max_pages),
        concurrency: Number(form.concurrency),
      })
      navigate(`/audit/${record.id}`)
    } catch (err) {
      setFormError(err.message)
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
    <div className="h-screen bg-surface flex overflow-hidden">

      {/* ── Left Sidebar: nav + audit history ── */}
      <aside
        className="w-64 flex flex-col shrink-0 h-full"
        style={{ background: '#161c26' }}
      >
        {/* Brand */}
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

          {/* Nav links */}
          <nav className="space-y-0.5">
            <div
              className="flex items-center gap-3 px-3 py-2.5 rounded-lg"
              style={{ background: '#1a202a', borderRight: '2px solid #00c853', color: '#3fe56c' }}
            >
              <Globe2 size={15} />
              <span className="uppercase tracking-widest font-medium" style={{ fontSize: '9px' }}>Observatory</span>
            </div>
            <button
              onClick={() => navigate('/settings')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <Settings size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Settings</span>
            </button>
          </nav>
        </div>

        {/* Audit history */}
        <div
          className="shrink-0 flex items-center justify-between px-4 py-2"
          style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}
        >
          <span className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>
            Audit Vault
          </span>
          <div className="flex items-center gap-1.5">
            <span className="text-on-surface-variant" style={{ fontSize: '9px' }}>{audits.length}</span>
            <button
              onClick={fetchAudits}
              className="p-1 text-on-surface-variant hover:text-primary transition-colors rounded"
              title="Refresh"
            >
              <RefreshCw size={10} />
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto">
          {historyLoading ? (
            <div className="flex justify-center py-8">
              <span
                className="w-5 h-5 rounded-full animate-spin"
                style={{ border: '2px solid #2f3540', borderTopColor: '#3fe56c' }}
              />
            </div>
          ) : audits.length === 0 ? (
            <p className="text-center text-on-surface-variant/50 text-xs py-8 px-4">
              No audits yet
            </p>
          ) : (
            <div className="flex flex-col">
              {audits.map((a) => (
                <SidebarAuditItem
                  key={a.id}
                  audit={a}
                  onDelete={handleDelete}
                  onRerun={handleRerun}
                  navigate={navigate}
                />
              ))}
            </div>
          )}
        </div>

        {/* Check count badge */}
        {checkCount !== null && (
          <div
            className="px-4 py-3 shrink-0"
            style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}
          >
            <span
              className="text-xs px-2 py-1 rounded-full"
              style={{ background: 'rgba(63,229,108,0.1)', color: '#3fe56c' }}
            >
              {checkCount} checks active
            </span>
          </div>
        )}
      </aside>

      {/* ── Main Content ── */}
      <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden">
        {/* Header */}
        <header
          className="h-14 flex items-center justify-between px-8 shrink-0"
          style={{
            background: 'rgba(14,19,30,0.72)',
            backdropFilter: 'blur(16px)',
            WebkitBackdropFilter: 'blur(16px)',
            borderBottom: '1px solid rgba(221,226,241,0.06)',
          }}
        >
          <div className="flex items-center gap-3">
            <span className="font-bold text-primary text-sm tracking-widest font-display">NEW SCAN</span>
            <span className="w-2 h-2 bg-primary rounded-full" style={{ boxShadow: '0 0 8px #00C853' }} />
          </div>
          <div className="flex items-center gap-4">
            <div className="h-4 w-px" style={{ background: 'rgba(60,74,60,0.5)' }} />
            <span className="text-on-surface-variant uppercase tracking-widest" style={{ fontSize: '10px' }}>
              Configure crawl parameters
            </span>
          </div>
        </header>

        {/* Form content */}
        <main className="flex-1 overflow-y-auto p-8">
          <form onSubmit={handleStart} className="max-w-2xl mx-auto flex flex-col gap-6">

            {/* URL input — hero input */}
            <div>
              <label className="label mb-2">Target Domain URL</label>
              <div className="relative">
                <input
                  type="url"
                  placeholder="https://example.com"
                  value={form.url}
                  onChange={(e) => set('url', e.target.value)}
                  className="input text-base py-4 pl-5 pr-16"
                  required
                  disabled={starting}
                />
                <Globe2
                  size={16}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-on-surface-variant"
                />
              </div>
            </div>

            {/* Crawl parameter grid */}
            <div className="grid grid-cols-3 gap-4">
              <div>
                <label className="label">Crawl Depth</label>
                <select value={form.max_depth} onChange={(e) => set('max_depth', e.target.value)} className="input" disabled={starting}>
                  <option value={-1}>Unlimited</option>
                  <option value={1}>1 level</option>
                  <option value={2}>2 levels</option>
                  <option value={3}>3 levels</option>
                  <option value={5}>5 levels</option>
                </select>
              </div>
              <div>
                <label className="label">Max Pages</label>
                <select value={form.max_pages} onChange={(e) => set('max_pages', e.target.value)} className="input" disabled={starting}>
                  <option value={0}>Unlimited</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                  <option value={500}>500</option>
                  <option value={1000}>1,000</option>
                </select>
              </div>
              <div>
                <label className="label">Timeout</label>
                <select value={form.timeout} onChange={(e) => set('timeout', e.target.value)} className="input" disabled={starting}>
                  <option value="10s">10 seconds</option>
                  <option value="30s">30 seconds</option>
                  <option value="1m">1 minute</option>
                  <option value="2m">2 minutes</option>
                </select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="label">Concurrent Workers — {form.concurrency}</label>
                <input
                  type="range" min={1} max={20} step={1}
                  value={form.concurrency}
                  onChange={(e) => set('concurrency', e.target.value)}
                  className="w-full mt-1"
                  style={{ accentColor: '#3fe56c' }}
                  disabled={starting}
                />
                <div className="flex justify-between text-on-surface-variant/50 mt-1" style={{ fontSize: '9px' }}>
                  <span>1 (gentle)</span><span>20 (fast)</span>
                </div>
              </div>
              <div>
                <label className="label">Platform</label>
                <select value={form.platform} onChange={(e) => set('platform', e.target.value)} className="input" disabled={starting}>
                  <option value="">Both (bifurcated)</option>
                  <option value="desktop">Desktop only</option>
                  <option value="mobile">Mobile focus</option>
                </select>
              </div>
            </div>

            {/* Advanced toggle */}
            <button
              type="button"
              onClick={() => setAdvanced((v) => !v)}
              className="flex items-center gap-2 text-sm text-on-surface-variant hover:text-on-surface transition-colors w-fit"
            >
              {advanced ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
              Advanced options
            </button>

            {advanced && (
              <div className="flex flex-col gap-4 rounded-xl p-5" style={{ background: '#161c26' }}>
                <div>
                  <label className="label">Custom output directory</label>
                  <input
                    type="text"
                    placeholder="~/.seo-reports (default)"
                    value={form.output_dir}
                    onChange={(e) => set('output_dir', e.target.value)}
                    className="input"
                    disabled={starting}
                  />
                </div>
                <div className="flex flex-col gap-3" style={{ borderTop: '1px solid rgba(60,74,60,0.3)', paddingTop: '1rem' }}>
                  <label className="flex items-start gap-3 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={form.validate_external_links}
                      onChange={(e) => set('validate_external_links', e.target.checked)}
                      className="mt-0.5 h-4 w-4 rounded"
                      style={{ accentColor: '#3fe56c' }}
                      disabled={starting}
                    />
                    <div>
                      <div className="text-sm text-on-surface">Validate external links</div>
                      <div className="text-xs text-on-surface-variant">HEAD-check every outbound link. Adds crawl time.</div>
                    </div>
                  </label>
                  <label className="flex items-start gap-3 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={form.discover_resources}
                      onChange={(e) => set('discover_resources', e.target.checked)}
                      className="mt-0.5 h-4 w-4 rounded"
                      style={{ accentColor: '#3fe56c' }}
                      disabled={starting}
                    />
                    <div>
                      <div className="text-sm text-on-surface">Discover sub-resources (CSS, JS, fonts)</div>
                      <div className="text-xs text-on-surface-variant">Validates stylesheets, scripts and fonts. Slow.</div>
                    </div>
                  </label>
                </div>
              </div>
            )}

            {formError && (
              <div className="text-sm rounded-xl px-4 py-2.5" style={{ color: '#ffb4ab', background: 'rgba(147,0,10,0.2)' }}>
                {formError}
              </div>
            )}

            {/* CTA */}
            <button type="submit" className="btn-primary justify-center py-4 text-base" disabled={starting}>
              {starting ? (
                <>
                  <span
                    className="w-4 h-4 rounded-full animate-spin"
                    style={{ border: '2px solid rgba(0,57,18,0.4)', borderTopColor: '#003912' }}
                  />
                  Initiating…
                </>
              ) : (
                <>
                  <Play size={16} />
                  Initiate Crawl
                </>
              )}
            </button>
          </form>
        </main>
      </div>
    </div>
  )
}
