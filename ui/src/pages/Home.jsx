import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Play, Settings, Globe2, ChevronDown, ChevronUp, CheckSquare, Database, Info, MapPinned, Search } from 'lucide-react'
import { api } from '../lib/api'
import { auditDefaultsToForm, auditFormToRequest, optionsOrCurrent } from '../lib/auditDefaults'

const TIPS = {
  url:                    'The root URL to start crawling from. All pages within the same domain will be discovered and audited.',
  max_depth:              'How many link-hops away from the root URL to crawl. Unlimited follows every internal link regardless of depth.',
  max_pages:              'Hard cap on the total number of pages crawled. Unlimited means the entire site will be processed.',
  timeout:                'Maximum time allowed per individual HTTP request. Increase for slow servers.',
  concurrency:            'Number of pages fetched in parallel. Higher values finish faster but put more load on the target server.',
  platform:               'Choose "Both" to run a full desktop + mobile bifurcated audit. "Desktop only" skips the mobile fetch. "Mobile focus" surfaces only mobile and mobile-vs-desktop issues.',
  sitemap_mode:           'Off starts crawling immediately from the URL. Discover samples sitemap URLs for coverage checks. Seed also adds sitemap URLs to the crawl queue and can be slow on large sites.',
  output_dir:             'Folder where HTML, JSON and Markdown reports are saved. Defaults to ~/.seo-reports if left blank.',
  validate_external_links:'Sends a HEAD request to every outbound link to check for broken URLs. Significantly increases total crawl time.',
  discover_resources:     'Validates every CSS, JavaScript and font file referenced by crawled pages. Very slow on large sites.',
  enable_rendered_seo:    'Runs a small Playwright/Chrome pass against sampled pages and compares raw HTML with the rendered DOM for JavaScript SEO risks.',
  rendered_sample_limit:  'Number of crawled pages to render in the browser pass. Higher values improve coverage but add time.',
  rendered_timeout:       'Maximum browser render time per sampled page.',
  important_page_urls:    'Priority category, hub, city, or model URLs that should be linked directly from the homepage.',
}

function Tooltip({ text }) {
  const [visible, setVisible] = useState(false)
  const [pos, setPos] = useState('top')
  const ref = useRef(null)

  const show = () => {
    if (ref.current) {
      const rect = ref.current.getBoundingClientRect()
      setPos(rect.top < 80 ? 'bottom' : 'top')
    }
    setVisible(true)
  }

  return (
    <span
      ref={ref}
      className="relative inline-flex items-center"
      onMouseEnter={show}
      onMouseLeave={() => setVisible(false)}
    >
      <Info size={11} className="text-on-surface-variant/50 hover:text-on-surface-variant cursor-help transition-colors" />
      {visible && (
        <span
          className="absolute z-50 w-64 rounded-lg px-3 py-2 text-xs leading-relaxed pointer-events-none"
          style={{
            background: '#1a202a',
            border: '1px solid rgba(60,74,60,0.5)',
            color: '#bbcbb8',
            boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
            left: '50%',
            transform: 'translateX(-50%)',
            ...(pos === 'top'
              ? { bottom: 'calc(100% + 8px)' }
              : { top: 'calc(100% + 8px)' }),
          }}
        >
          {text}
          <span
            className="absolute left-1/2 -translate-x-1/2 w-2 h-2 rotate-45"
            style={{
              background: '#1a202a',
              border: '1px solid rgba(60,74,60,0.5)',
              ...(pos === 'top'
                ? { bottom: '-5px', borderTop: 'none', borderLeft: 'none' }
                : { top: '-5px', borderBottom: 'none', borderRight: 'none' }),
            }}
          />
        </span>
      )}
    </span>
  )
}

function Label({ children, tip }) {
  return (
    <label className="label mb-2 flex items-center gap-1.5">
      {children}
      {tip && <Tooltip text={tip} />}
    </label>
  )
}

export default function Home() {
  const [starting, setStarting] = useState(false)
  const [checkCount, setCheckCount] = useState(null)
  const [form, setForm] = useState(() => auditDefaultsToForm())
  const [auditControls, setAuditControls] = useState(null)
  const [advanced, setAdvanced] = useState(false)
  const [formError, setFormError] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    api.getCheckCatalog()
      .then((c) => setCheckCount(c?.total ?? null))
      .catch(() => {})
  }, [])

  useEffect(() => {
    api.getAuditDefaults()
      .then((data) => {
        setAuditControls(data?.controls ?? null)
        setForm((current) => ({
          ...auditDefaultsToForm(data?.default_config),
          url: current.url || data?.default_config?.url || '',
        }))
      })
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
      const record = await api.startAudit({ ...auditFormToRequest(form), url })
      navigate(`/audit/${record.id}`)
    } catch (err) {
      setFormError(err.message)
      setStarting(false)
    }
  }

  const concurrencyControl = auditControls?.concurrency ?? { min: 1, max: 20, step: 1 }
  const maxPagesControl = auditControls?.max_pages ?? { min: 0, step: 50 }
  const maxDepthOptions = optionsOrCurrent(auditControls?.max_depth_options, form.max_depth, form.max_depth === -1 ? 'Unlimited' : `${form.max_depth} levels`)
  const timeoutOptions = optionsOrCurrent(auditControls?.timeout_options, form.timeout, form.timeout)
  const platformOptions = optionsOrCurrent(auditControls?.platform_options, form.platform, 'Both (bifurcated)')
  const sitemapModeOptions = optionsOrCurrent(auditControls?.sitemap_mode_options, form.sitemap_mode, form.sitemap_mode)
  const renderedSampleLimitOptions = optionsOrCurrent(auditControls?.rendered_sample_limit_options, form.rendered_sample_limit, `${form.rendered_sample_limit} pages`)
  const renderedTimeoutOptions = optionsOrCurrent(auditControls?.rendered_timeout_options, form.rendered_timeout, form.rendered_timeout)

  return (
    <div className="h-screen bg-surface flex overflow-hidden">

      {/* ── Left Sidebar: nav ── */}
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
              {checkCount != null && (
                <span className="ml-auto text-on-surface-variant font-mono" style={{ fontSize: '9px', background: '#2f3540', padding: '1px 5px', borderRadius: '8px' }}>
                  {checkCount}
                </span>
              )}
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

            {/* URL input */}
            <div>
              <Label tip={TIPS.url}>Target Domain URL</Label>
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
                <Label tip={TIPS.max_depth}>Crawl Depth</Label>
                <select value={form.max_depth} onChange={(e) => set('max_depth', e.target.value)} className="input" disabled={starting}>
                  {maxDepthOptions.map((option) => (
                    <option key={option.value} value={option.value}>{option.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <Label tip={TIPS.max_pages}>Max Pages</Label>
                <input
                  type="number"
                  min={maxPagesControl.min}
                  step={maxPagesControl.step}
                  value={form.max_pages}
                  onChange={(e) => set('max_pages', e.target.value)}
                  className="input"
                  disabled={starting}
                />
                <div className="mt-2 flex flex-wrap gap-1.5">
                  {(auditControls?.max_page_presets ?? []).map((preset) => (
                    <button
                      key={preset.value}
                      type="button"
                      onClick={() => set('max_pages', preset.value)}
                      className="rounded-md px-2 py-1 text-on-surface-variant hover:text-on-surface transition-colors"
                      style={{
                        fontSize: '10px',
                        background: Number(form.max_pages) === preset.value ? 'rgba(63,229,108,0.12)' : '#1a202a',
                        border: Number(form.max_pages) === preset.value ? '1px solid rgba(63,229,108,0.35)' : '1px solid rgba(60,74,60,0.28)',
                        color: Number(form.max_pages) === preset.value ? '#3fe56c' : undefined,
                      }}
                      disabled={starting}
                    >
                      {preset.label}
                    </button>
                  ))}
                </div>
                <div className="mt-1 text-on-surface-variant" style={{ fontSize: '10px' }}>
                  Use 0 for unlimited.
                </div>
              </div>
              <div>
                <Label tip={TIPS.timeout}>Timeout</Label>
                <select value={form.timeout} onChange={(e) => set('timeout', e.target.value)} className="input" disabled={starting}>
                  {timeoutOptions.map((option) => (
                    <option key={option.value} value={option.value}>{option.label}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label tip={TIPS.concurrency}>Concurrent Workers — {form.concurrency}</Label>
                <input
                  type="range" min={concurrencyControl.min} max={concurrencyControl.max} step={concurrencyControl.step}
                  value={form.concurrency}
                  onChange={(e) => set('concurrency', e.target.value)}
                  className="w-full mt-1"
                  style={{ accentColor: '#3fe56c' }}
                  disabled={starting}
                />
                <div className="flex justify-between text-on-surface-variant/50 mt-1" style={{ fontSize: '9px' }}>
                  <span>{concurrencyControl.min} (gentle)</span><span>{concurrencyControl.max} (fast)</span>
                </div>
              </div>
              <div>
                <Label tip={TIPS.platform}>Platform</Label>
                <select value={form.platform} onChange={(e) => set('platform', e.target.value)} className="input" disabled={starting}>
                  {platformOptions.map((option) => (
                    <option key={option.value || 'all'} value={option.value}>{option.label}</option>
                  ))}
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
                  <Label tip={TIPS.output_dir}>Custom output directory</Label>
                  <input
                    type="text"
                    placeholder="~/.seo-reports (default)"
                    value={form.output_dir}
                    onChange={(e) => set('output_dir', e.target.value)}
                    className="input"
                    disabled={starting}
                  />
                </div>
                <div>
                  <Label tip={TIPS.sitemap_mode}>Sitemap mode</Label>
                  <select
                    value={form.sitemap_mode}
                    onChange={(e) => set('sitemap_mode', e.target.value)}
                    className="input"
                    disabled={starting}
                  >
                    {sitemapModeOptions.map((option) => (
                      <option key={option.value} value={option.value}>{option.label}</option>
                    ))}
                  </select>
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
                      <div className="text-sm text-on-surface flex items-center gap-1.5">
                        Validate external links
                        <Tooltip text={TIPS.validate_external_links} />
                      </div>
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
                      <div className="text-sm text-on-surface flex items-center gap-1.5">
                        Discover sub-resources (CSS, JS, fonts)
                        <Tooltip text={TIPS.discover_resources} />
                      </div>
                      <div className="text-xs text-on-surface-variant">Validates stylesheets, scripts and fonts. Slow.</div>
                    </div>
                  </label>
                  <label className="flex items-start gap-3 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={form.enable_crawler_evidence}
                      onChange={(e) => set('enable_crawler_evidence', e.target.checked)}
                      className="mt-0.5 h-4 w-4 rounded"
                      style={{ accentColor: '#3fe56c' }}
                      disabled={starting}
                    />
                    <div>
                      <div className="text-sm text-on-surface">Run crawler evidence checks</div>
                      <div className="text-xs text-on-surface-variant">Adds robots, sitemap, image CDN, and live-content evidence to this audit.</div>
                    </div>
                  </label>
                  <label className="flex items-start gap-3 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={form.enable_rendered_seo}
                      onChange={(e) => set('enable_rendered_seo', e.target.checked)}
                      className="mt-0.5 h-4 w-4 rounded"
                      style={{ accentColor: '#3fe56c' }}
                      disabled={starting}
                    />
                    <div>
                      <div className="text-sm text-on-surface flex items-center gap-1.5">
                        Run rendered JavaScript SEO checks
                        <Tooltip text={TIPS.enable_rendered_seo} />
                      </div>
                      <div className="text-xs text-on-surface-variant">Browser-renders sampled pages and compares raw HTML vs rendered DOM.</div>
                    </div>
                  </label>
                </div>
                {form.enable_rendered_seo && (
                  <div className="grid grid-cols-2 gap-4 pt-4" style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}>
                    <div>
                      <Label tip={TIPS.rendered_sample_limit}>Rendered sample size</Label>
                      <select
                        value={form.rendered_sample_limit}
                        onChange={(e) => set('rendered_sample_limit', e.target.value)}
                        className="input"
                        disabled={starting}
                      >
                        {renderedSampleLimitOptions.map((option) => (
                          <option key={option.value} value={option.value}>{option.label}</option>
                        ))}
                      </select>
                    </div>
                    <div>
                      <Label tip={TIPS.rendered_timeout}>Rendered page timeout</Label>
                      <select
                        value={form.rendered_timeout}
                        onChange={(e) => set('rendered_timeout', e.target.value)}
                        className="input"
                        disabled={starting}
                      >
                        {renderedTimeoutOptions.map((option) => (
                          <option key={option.value} value={option.value}>{option.label}</option>
                        ))}
                      </select>
                    </div>
                  </div>
                )}
                {form.enable_crawler_evidence && (
                  <div className="grid grid-cols-2 gap-4 pt-4" style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}>
                    <div>
                      <label className="label">Expected inventory URLs</label>
                      <textarea
                        className="input text-sm min-h-[88px]"
                        value={form.expected_inventory_urls}
                        onChange={(e) => set('expected_inventory_urls', e.target.value)}
                        placeholder="One URL per line"
                        disabled={starting}
                      />
                    </div>
                    <div>
                      <Label tip={TIPS.important_page_urls}>Important homepage-linked URLs</Label>
                      <textarea
                        className="input text-sm min-h-[88px]"
                        value={form.important_page_urls}
                        onChange={(e) => set('important_page_urls', e.target.value)}
                        placeholder="One priority URL per line"
                        disabled={starting}
                      />
                    </div>
                    <div>
                      <label className="label">Expected URL parameters</label>
                      <textarea
                        className="input text-sm min-h-[88px]"
                        value={form.expected_parameter_names}
                        onChange={(e) => set('expected_parameter_names', e.target.value)}
                        placeholder={'sort\nfilter\ncity'}
                        disabled={starting}
                      />
                    </div>
                    <div>
                      <label className="label">Allowed image CDN hosts</label>
                      <textarea
                        className="input text-sm min-h-[88px]"
                        value={form.allowed_image_cdn_hosts}
                        onChange={(e) => set('allowed_image_cdn_hosts', e.target.value)}
                        placeholder="images.examplecdn.com"
                        disabled={starting}
                      />
                    </div>
                    <div>
                      <label className="label">Required live text</label>
                      <textarea
                        className="input text-sm min-h-[88px]"
                        value={form.required_live_text}
                        onChange={(e) => set('required_live_text', e.target.value)}
                        placeholder="One expected snippet per line"
                        disabled={starting}
                      />
                    </div>
                  </div>
                )}
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
