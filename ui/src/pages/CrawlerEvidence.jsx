import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Activity, ArrowLeft, CheckCircle2, CheckSquare, Database, FileSearch,
  Globe2, Image, MapPinned, RefreshCw, Search, Settings, ShieldCheck,
} from 'lucide-react'
import { api } from '../lib/api'

const STATUS_STYLE = {
  pass: { label: 'Pass', color: '#3fe56c', bg: 'rgba(63,229,108,0.08)' },
  warning: { label: 'Warning', color: '#ffb7ae', bg: 'rgba(255,183,174,0.08)' },
  fail: { label: 'Fail', color: '#ffb4ab', bg: 'rgba(255,180,171,0.1)' },
  needs_input: { label: 'Needs input', color: '#9cc7ff', bg: 'rgba(156,199,255,0.08)' },
  info: { label: 'Info', color: '#dde2f1', bg: 'rgba(221,226,241,0.06)' },
}

function NavButton({ icon: Icon, label, onClick }) {
  return (
    <button
      onClick={onClick}
      className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
    >
      <Icon size={15} />
      <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>{label}</span>
    </button>
  )
}

function StatusCard({ label, value, tone = '#dde2f1' }) {
  return (
    <div className="rounded-lg p-4" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
      <div className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>{label}</div>
      <div className="font-display font-semibold mt-1" style={{ color: tone, fontSize: '24px' }}>{value}</div>
    </div>
  )
}

function splitLines(value) {
  return value
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
}

function EvidenceRow({ item }) {
  const style = STATUS_STYLE[item.status] ?? STATUS_STYLE.info
  return (
    <div
      className="grid gap-3 px-4 py-3 items-start"
      style={{
        gridTemplateColumns: '96px minmax(220px, 1.2fr) 118px minmax(240px, 1fr)',
        borderTop: '1px solid rgba(60,74,60,0.18)',
      }}
    >
      <code className="font-mono text-primary" style={{ fontSize: '11px' }}>{item.id}</code>
      <div>
        <div className="text-on-surface text-sm leading-snug">{item.name}</div>
        {item.details && (
          <div className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '11px' }}>
            {item.details}
          </div>
        )}
      </div>
      <span
        className="w-fit rounded-full px-2 py-0.5 font-mono"
        style={{ background: style.bg, color: style.color, fontSize: '10px' }}
      >
        {style.label}
      </span>
      <div>
        <div className="text-on-surface" style={{ fontSize: '12px' }}>{item.message}</div>
        {item.evidence?.length > 0 && (
          <div className="mt-2 flex flex-col gap-1">
            {item.evidence.slice(0, 8).map((line, idx) => (
              <code key={`${item.id}-${idx}`} className="text-on-surface-variant break-all" style={{ fontSize: '10px' }}>
                {line}
              </code>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default function CrawlerEvidence() {
  const navigate = useNavigate()
  const [workspace, setWorkspace] = useState(null)
  const [loading, setLoading] = useState(true)
  const [running, setRunning] = useState(false)
  const [error, setError] = useState('')
  const [report, setReport] = useState(null)
  const [form, setForm] = useState({
    url: '',
    sitemapUrl: '',
    maxPages: 40,
    maxDepth: 2,
    concurrency: 5,
    timeout: '20s',
    sitemapMode: 'discover',
    respectRobots: true,
    expectedInventory: '',
    importantPages: '',
    expectedParams: 'sort\nfilter\ncity\nprice',
    cdnHosts: '',
    requiredLiveText: '',
  })

  const refresh = async () => {
    setLoading(true)
    setError('')
    try {
      const data = await api.getCrawlerEvidence()
      setWorkspace(data)
      setForm((prev) => ({
        ...prev,
        url: prev.url || data?.default_config?.url || '',
        maxPages: prev.maxPages || data?.default_config?.max_pages || 40,
        maxDepth: prev.maxDepth || data?.default_config?.max_depth || 2,
        concurrency: prev.concurrency || data?.default_config?.concurrency || 5,
        timeout: prev.timeout || data?.default_config?.timeout || '20s',
        sitemapMode: prev.sitemapMode || data?.default_config?.sitemap_mode || 'discover',
      }))
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
  }, [])

  const checks = useMemo(() => workspace?.checks ?? [], [workspace])

  const updateForm = (key, value) => {
    setError('')
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const runEvidence = async (e) => {
    e.preventDefault()
    setError('')
    setReport(null)
    setRunning(true)
    try {
      const res = await api.runCrawlerEvidence({
        url: form.url.trim(),
        sitemap_url: form.sitemapUrl.trim(),
        max_pages: Number(form.maxPages),
        max_depth: Number(form.maxDepth),
        concurrency: Number(form.concurrency),
        timeout: form.timeout.trim(),
        sitemap_mode: form.sitemapMode,
        respect_robots: form.respectRobots,
        expected_inventory_urls: splitLines(form.expectedInventory),
        important_page_urls: splitLines(form.importantPages),
        expected_parameter_names: splitLines(form.expectedParams),
        allowed_image_cdn_hosts: splitLines(form.cdnHosts),
        required_live_text: splitLines(form.requiredLiveText),
      })
      setReport(res)
    } catch (err) {
      setError(err.message)
    } finally {
      setRunning(false)
    }
  }

  return (
    <div className="h-screen bg-surface flex overflow-hidden">
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
            <NavButton icon={ArrowLeft} label="Observatory" onClick={() => navigate('/')} />
            <NavButton icon={Database} label="Audit Vault" onClick={() => navigate('/vault')} />
            <NavButton icon={Settings} label="Settings" onClick={() => navigate('/settings')} />
            <NavButton icon={CheckSquare} label="Checks Catalog" onClick={() => navigate('/checks')} />
            <NavButton icon={MapPinned} label="Local SEO" onClick={() => navigate('/local-seo')} />
            <NavButton icon={Search} label="GSC + Bing" onClick={() => navigate('/search-integrations')} />
            <div
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg"
              style={{ background: '#1a202a', borderRight: '2px solid #00c853', color: '#3fe56c' }}
            >
              <Activity size={15} />
              <span className="uppercase tracking-widest font-medium" style={{ fontSize: '9px' }}>Crawler Evidence</span>
            </div>
          </nav>
        </div>

        <div className="px-4 py-4 flex flex-col gap-3" style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}>
          <span className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>Run State</span>
          <div className="rounded-lg p-3" style={{ background: '#1a202a', border: '1px solid rgba(60,74,60,0.35)' }}>
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full" style={{ background: running ? '#ffb7ae' : report ? '#3fe56c' : '#9cc7ff' }} />
              <span className="uppercase tracking-widest text-on-surface" style={{ fontSize: '9px' }}>
                {running ? 'Running' : report ? 'Complete' : 'Ready'}
              </span>
            </div>
            <div className="text-on-surface-variant mt-2 leading-snug" style={{ fontSize: '11px' }}>
              {report ? `${report.summary?.pages_crawled ?? 0} pages crawled` : `${checks.length} crawler checks loaded`}
            </div>
          </div>
        </div>
      </aside>

      <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden">
        <header
          className="h-14 flex items-center gap-4 px-6 shrink-0"
          style={{
            background: 'rgba(14,19,30,0.72)',
            backdropFilter: 'blur(16px)',
            WebkitBackdropFilter: 'blur(16px)',
            borderBottom: '1px solid rgba(221,226,241,0.06)',
          }}
        >
          <Activity size={16} style={{ color: '#3fe56c' }} />
          <span className="font-display font-semibold text-on-surface text-sm">Crawler Evidence</span>
          <span className="ml-1 px-2.5 py-0.5 rounded-full text-on-surface-variant" style={{ background: '#1a202a', fontSize: '10px' }}>
            {checks.length} checks
          </span>
          <div className="flex-1" />
          <button onClick={refresh} className="btn-ghost" title="Refresh" disabled={running}>
            <RefreshCw size={13} />
            Refresh
          </button>
        </header>

        <main className="flex-1 overflow-y-auto p-6">
          {loading ? (
            <div className="flex justify-center py-20">
              <span className="w-8 h-8 rounded-full animate-spin" style={{ border: '2px solid #2f3540', borderTopColor: '#3fe56c' }} />
            </div>
          ) : (
            <div className="grid gap-6" style={{ gridTemplateColumns: 'minmax(0, 1fr) 400px' }}>
              <section className="flex flex-col gap-5 min-w-0">
                <div className="grid grid-cols-5 gap-3">
                  <StatusCard label="Pages" value={report?.summary?.pages_crawled ?? 0} tone="#3fe56c" />
                  <StatusCard label="Pass" value={report?.summary?.pass ?? 0} tone="#3fe56c" />
                  <StatusCard label="Warnings" value={report?.summary?.warning ?? 0} tone="#ffb7ae" />
                  <StatusCard label="Fails" value={report?.summary?.fail ?? 0} tone="#ffb4ab" />
                  <StatusCard label="Needs Input" value={report?.summary?.needs_input ?? 0} tone="#9cc7ff" />
                </div>

                {error && (
                  <div className="text-sm rounded-lg px-4 py-3" style={{ color: '#ffb4ab', background: 'rgba(147,0,10,0.2)', border: '1px solid rgba(255,180,171,0.18)' }}>
                    {error}
                  </div>
                )}

                <section className="rounded-xl overflow-hidden" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-center gap-3 px-4 py-3">
                    <ShieldCheck size={15} style={{ color: '#3fe56c' }} />
                    <h2 className="text-on-surface font-semibold text-sm">Evidence Report</h2>
                  </div>
                  {report ? (
                    <>
                      <div
                        className="grid gap-3 px-4 py-2 text-on-surface-variant uppercase tracking-widest"
                        style={{
                          gridTemplateColumns: '96px minmax(220px, 1.2fr) 118px minmax(240px, 1fr)',
                          fontSize: '9px',
                          borderTop: '1px solid rgba(60,74,60,0.24)',
                        }}
                      >
                        <span>ID</span>
                        <span>Check</span>
                        <span>Status</span>
                        <span>Evidence</span>
                      </div>
                      {report.report?.map((item) => <EvidenceRow key={item.id} item={item} />)}
                    </>
                  ) : (
                    <div className="px-5 pb-5">
                      <div className="rounded-lg p-4 flex items-start gap-3" style={{ background: '#1a202a' }}>
                        <CheckCircle2 size={16} style={{ color: '#9cc7ff' }} />
                        <div>
                          <p className="text-on-surface text-sm">No crawler evidence run yet.</p>
                          <p className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '12px' }}>
                            Configure a bounded crawl and run the pack.
                          </p>
                        </div>
                      </div>
                    </div>
                  )}
                </section>

                <section className="rounded-xl overflow-hidden" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-center gap-3 px-4 py-3">
                    <FileSearch size={15} style={{ color: '#3fe56c' }} />
                    <h2 className="text-on-surface font-semibold text-sm">Page Snapshot</h2>
                  </div>
                  {report?.pages?.length ? (
                    <div className="flex flex-col">
                      {report.pages.slice(0, 20).map((page) => (
                        <div key={`${page.url}-${page.content_hash}`} className="grid gap-3 px-4 py-3 items-center" style={{ gridTemplateColumns: '72px minmax(220px, 1fr) 130px 80px', borderTop: '1px solid rgba(60,74,60,0.18)' }}>
                          <span className="font-mono" style={{ color: page.status_code >= 400 ? '#ffb4ab' : '#3fe56c', fontSize: '11px' }}>{page.status_code || '—'}</span>
                          <div className="min-w-0">
                            <div className="text-on-surface text-sm truncate">{page.title || 'Untitled page'}</div>
                            <code className="block text-on-surface-variant truncate mt-1" style={{ fontSize: '10px' }}>{page.url}</code>
                          </div>
                          <code className="text-on-surface-variant" style={{ fontSize: '10px' }}>{page.content_hash}</code>
                          <span className="text-on-surface-variant" style={{ fontSize: '11px' }}>{page.image_count} images</span>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="px-5 pb-5 text-on-surface-variant" style={{ fontSize: '12px' }}>
                      Page hashes appear after a run.
                    </div>
                  )}
                </section>
              </section>

              <aside className="flex flex-col gap-5 min-w-0">
                <form onSubmit={runEvidence} className="rounded-xl p-5 flex flex-col gap-4" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-start gap-3">
                    <Activity size={16} style={{ color: '#3fe56c' }} />
                    <div>
                      <h2 className="text-on-surface font-semibold text-sm">Run Crawler Pack</h2>
                      <p className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '12px' }}>
                        Bounded crawl with robots, sitemap, link, image, and live-content evidence.
                      </p>
                    </div>
                  </div>

                  <div>
                    <label className="label">Start URL</label>
                    <input
                      className="input text-sm"
                      value={form.url}
                      onChange={(e) => updateForm('url', e.target.value)}
                      placeholder="https://www.cars24.com/"
                      required
                    />
                  </div>
                  <div>
                    <label className="label">Sitemap URL</label>
                    <input
                      className="input text-sm"
                      value={form.sitemapUrl}
                      onChange={(e) => updateForm('sitemapUrl', e.target.value)}
                      placeholder="Auto-discover"
                    />
                  </div>
                  <div className="grid grid-cols-3 gap-3">
                    <div>
                      <label className="label">Pages</label>
                      <input className="input text-sm" type="number" min="1" max="250" value={form.maxPages} onChange={(e) => updateForm('maxPages', e.target.value)} />
                    </div>
                    <div>
                      <label className="label">Depth</label>
                      <input className="input text-sm" type="number" min="-1" max="8" value={form.maxDepth} onChange={(e) => updateForm('maxDepth', e.target.value)} />
                    </div>
                    <div>
                      <label className="label">Threads</label>
                      <input className="input text-sm" type="number" min="1" max="12" value={form.concurrency} onChange={(e) => updateForm('concurrency', e.target.value)} />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="label">Timeout</label>
                      <input className="input text-sm" value={form.timeout} onChange={(e) => updateForm('timeout', e.target.value)} />
                    </div>
                    <div>
                      <label className="label">Sitemap Mode</label>
                      <select className="input text-sm" value={form.sitemapMode} onChange={(e) => updateForm('sitemapMode', e.target.value)}>
                        <option value="discover">discover</option>
                        <option value="seed">seed</option>
                        <option value="off">off</option>
                      </select>
                    </div>
                  </div>
                  <label className="flex items-center gap-2 text-on-surface-variant" style={{ fontSize: '12px' }}>
                    <input
                      type="checkbox"
                      checked={form.respectRobots}
                      onChange={(e) => updateForm('respectRobots', e.target.checked)}
                    />
                    Respect robots.txt
                  </label>

                  <div>
                    <label className="label">Expected Inventory URLs</label>
                    <textarea className="input text-sm min-h-[86px]" value={form.expectedInventory} onChange={(e) => updateForm('expectedInventory', e.target.value)} placeholder="One URL per line" />
                  </div>
                  <div>
                    <label className="label">Important Homepage-Linked URLs</label>
                    <textarea className="input text-sm min-h-[76px]" value={form.importantPages} onChange={(e) => updateForm('importantPages', e.target.value)} placeholder="One priority URL per line" />
                  </div>
                  <div>
                    <label className="label">Expected Parameters</label>
                    <textarea className="input text-sm min-h-[76px]" value={form.expectedParams} onChange={(e) => updateForm('expectedParams', e.target.value)} />
                  </div>
                  <div>
                    <label className="label">Allowed Image CDN Hosts</label>
                    <textarea className="input text-sm min-h-[76px]" value={form.cdnHosts} onChange={(e) => updateForm('cdnHosts', e.target.value)} placeholder="images.examplecdn.com" />
                  </div>
                  <div>
                    <label className="label">Required Live Text</label>
                    <textarea className="input text-sm min-h-[76px]" value={form.requiredLiveText} onChange={(e) => updateForm('requiredLiveText', e.target.value)} placeholder="One required snippet per line" />
                  </div>

                  <button type="submit" className="btn-primary justify-center" disabled={running}>
                    {running
                      ? <span className="w-4 h-4 rounded-full animate-spin" style={{ border: '2px solid rgba(0,57,18,0.4)', borderTopColor: '#003912' }} />
                      : <Activity size={14} />}
                    {running ? 'Running...' : 'Run Evidence'}
                  </button>
                </form>

                <section className="rounded-xl p-5 flex flex-col gap-3" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-center gap-2">
                    <Image size={14} style={{ color: '#3fe56c' }} />
                    <h2 className="text-on-surface font-semibold text-sm">Crawler Checks</h2>
                  </div>
                  {checks.map((check) => (
                    <div key={check.id} className="rounded-lg p-3" style={{ background: '#1a202a' }}>
                      <div className="flex items-center justify-between gap-3">
                        <code className="font-mono text-primary" style={{ fontSize: '10px' }}>{check.id}</code>
                        <span className="text-on-surface-variant" style={{ fontSize: '10px' }}>{check.priority}</span>
                      </div>
                      <div className="text-on-surface text-sm mt-1">{check.name}</div>
                      <div className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '11px' }}>{check.crawler_role}</div>
                    </div>
                  ))}
                </section>
              </aside>
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
