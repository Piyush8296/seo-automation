import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Play, Settings, Globe2, ChevronDown, ChevronUp, CheckSquare, Database } from 'lucide-react'
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


export default function Home() {
  const [starting, setStarting] = useState(false)
  const [checkCount, setCheckCount] = useState(null)
  const [form, setForm] = useState(DEFAULTS)
  const [advanced, setAdvanced] = useState(false)
  const [formError, setFormError] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    api.getCheckCatalog()
      .then((c) => setCheckCount(c?.total ?? null))
      .catch(() => {})
  }, [])

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
            <button
              onClick={() => navigate('/vault')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <Database size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Audit Vault</span>
            </button>
            <button
              onClick={() => navigate('/checks')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <CheckSquare size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Checks Catalog</span>
              {checkCount != null && (
                <span className="ml-auto text-on-surface-variant font-mono" style={{ fontSize: '9px', background: '#2f3540', padding: '1px 5px', borderRadius: '8px' }}>
                  {checkCount}
                </span>
              )}
            </button>
          </nav>
        </div>
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
