import { useEffect, useState } from 'react'
import { ChevronDown, ChevronUp, Play } from 'lucide-react'
import { api } from '../lib/api'
import { auditDefaultsToForm, auditFormToRequest, optionsOrCurrent } from '../lib/auditDefaults'

export default function AuditForm({ onSubmit, loading }) {
  const [form, setForm] = useState(() => auditDefaultsToForm())
  const [auditControls, setAuditControls] = useState(null)
  const [advanced, setAdvanced] = useState(false)
  const [error, setError] = useState('')

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

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    const url = form.url.trim()
    if (!url) { setError('URL is required'); return }
    if (!/^https?:\/\//i.test(url)) { setError('URL must start with http:// or https://'); return }
    try {
      await onSubmit({ ...auditFormToRequest(form), url })
    } catch (err) {
      setError(err.message)
    }
  }

  const concurrencyControl = auditControls?.concurrency ?? { min: 1, max: 20, step: 1 }
  const maxPagesControl = auditControls?.max_pages ?? { min: 0, step: 50 }
  const maxDepthOptions = optionsOrCurrent(auditControls?.max_depth_options, form.max_depth, form.max_depth === -1 ? 'Unlimited' : `${form.max_depth} levels`)
  const timeoutOptions = optionsOrCurrent(auditControls?.timeout_options, form.timeout, form.timeout)
  const platformOptions = optionsOrCurrent(auditControls?.platform_options, form.platform, 'Both (bifurcated)')
  const sitemapModeOptions = optionsOrCurrent(auditControls?.sitemap_mode_options, form.sitemap_mode, form.sitemap_mode)

  return (
    <form onSubmit={handleSubmit} className="card p-6 flex flex-col gap-5">
      <div>
        <h2 className="font-display text-lg font-bold text-on-surface mb-1">New Audit</h2>
        <p className="text-sm text-on-surface-variant">Enter a URL to start a full technical SEO crawl</p>
      </div>

      {/* URL */}
      <div>
        <label className="label">Website URL</label>
        <input
          type="url"
          placeholder="https://example.com"
          value={form.url}
          onChange={(e) => set('url', e.target.value)}
          className="input text-base"
          required
          disabled={loading}
        />
      </div>

      {/* Advanced toggle */}
      <button
        type="button"
        onClick={() => setAdvanced((v) => !v)}
        className="flex items-center gap-2 text-sm text-on-surface-variant hover:text-on-surface transition-colors w-fit"
      >
        {advanced ? <ChevronUp size={15} /> : <ChevronDown size={15} />}
        Advanced configuration
      </button>

      {advanced && (
        <div className="grid grid-cols-2 gap-4 pt-1">
          <div>
            <label className="label">Max depth</label>
            <select
              value={form.max_depth}
              onChange={(e) => set('max_depth', e.target.value)}
              className="input"
              disabled={loading}
            >
              {maxDepthOptions.map((option) => (
                <option key={option.value} value={option.value}>{option.label}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="label">Max pages</label>
            <input
              type="number"
              min={maxPagesControl.min}
              step={maxPagesControl.step}
              value={form.max_pages}
              onChange={(e) => set('max_pages', e.target.value)}
              className="input"
              disabled={loading}
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
                  disabled={loading}
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
            <label className="label">Concurrency — {form.concurrency} workers</label>
            <input
              type="range"
              min={concurrencyControl.min} max={concurrencyControl.max} step={concurrencyControl.step}
              value={form.concurrency}
              onChange={(e) => set('concurrency', e.target.value)}
              className="w-full mt-1"
              style={{ accentColor: '#3fe56c' }}
              disabled={loading}
            />
            <div className="flex justify-between text-xs text-on-surface-variant/60 mt-1">
              <span>{concurrencyControl.min} (gentle)</span>
              <span>{concurrencyControl.max} (fast)</span>
            </div>
          </div>

          <div>
            <label className="label">Request timeout</label>
            <select
              value={form.timeout}
              onChange={(e) => set('timeout', e.target.value)}
              className="input"
              disabled={loading}
            >
              {timeoutOptions.map((option) => (
                <option key={option.value} value={option.value}>{option.label}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="label">Platform</label>
            <select
              value={form.platform}
              onChange={(e) => set('platform', e.target.value)}
              className="input"
              disabled={loading}
            >
              {platformOptions.map((option) => (
                <option key={option.value || 'all'} value={option.value}>{option.label}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="label">Crawl scope</label>
            <select
              value={form.scope}
              onChange={(e) => set('scope', e.target.value)}
              className="input"
              disabled={loading}
            >
              <option value="host">Same host</option>
              <option value="subfolder">Seed subfolder only</option>
            </select>
          </div>

          <div>
            <label className="label">Scope prefix override</label>
            <input
              type="text"
              placeholder="/buy-used-cars"
              value={form.scope_prefix}
              onChange={(e) => set('scope_prefix', e.target.value)}
              className="input"
              disabled={loading}
            />
          </div>

          <div>
            <label className="label">Sitemap mode</label>
            <select
              value={form.sitemap_mode}
              onChange={(e) => set('sitemap_mode', e.target.value)}
              className="input"
              disabled={loading}
            >
              {sitemapModeOptions.map((option) => (
                <option key={option.value} value={option.value}>{option.label}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="label">Custom output dir</label>
            <input
              type="text"
              placeholder="~/.seo-reports (default)"
              value={form.output_dir}
              onChange={(e) => set('output_dir', e.target.value)}
              className="input"
              disabled={loading}
            />
          </div>

          <div className="col-span-2">
            <label className="label">Sitemap URL override</label>
            <input
              type="url"
              placeholder="https://example.com/sitemap.xml"
              value={form.sitemap_url}
              onChange={(e) => set('sitemap_url', e.target.value)}
              className="input"
              disabled={loading}
            />
          </div>

          <div className="col-span-2">
            <label className="label">Desktop user-agent override</label>
            <input
              type="text"
              placeholder="Leave blank for default SEOAuditBot UA"
              value={form.user_agent}
              onChange={(e) => set('user_agent', e.target.value)}
              className="input"
              disabled={loading}
            />
          </div>

          <div className="col-span-2">
            <label className="label">Mobile user-agent override</label>
            <input
              type="text"
              placeholder="Leave blank for default mobile Chrome UA"
              value={form.mobile_user_agent}
              onChange={(e) => set('mobile_user_agent', e.target.value)}
              className="input"
              disabled={loading}
            />
          </div>

          <div>
            <label className="label">Max redirects</label>
            <input
              type="number"
              min={1}
              value={form.max_redirects}
              onChange={(e) => set('max_redirects', e.target.value)}
              className="input"
              disabled={loading}
            />
          </div>

          <div>
            <label className="label">Max page size (KB)</label>
            <input
              type="number"
              min={1}
              value={form.max_page_size_kb}
              onChange={(e) => set('max_page_size_kb', e.target.value)}
              className="input"
              disabled={loading}
            />
          </div>

          <div>
            <label className="label">Max URL length</label>
            <input
              type="number"
              min={0}
              value={form.max_url_length}
              onChange={(e) => set('max_url_length', e.target.value)}
              className="input"
              disabled={loading}
            />
          </div>

          <div>
            <label className="label">Max query params</label>
            <input
              type="number"
              min={0}
              value={form.max_query_params}
              onChange={(e) => set('max_query_params', e.target.value)}
              className="input"
              disabled={loading}
            />
          </div>

          <div>
            <label className="label">Max links per page</label>
            <input
              type="number"
              min={0}
              value={form.max_links_per_page}
              onChange={(e) => set('max_links_per_page', e.target.value)}
              className="input"
              disabled={loading}
            />
          </div>

          <div
            className="col-span-2 pt-4 mt-1 flex flex-col gap-3"
            style={{ borderTop: '1px solid rgba(60,74,60,0.35)' }}
          >
            <label className="flex items-start gap-3 cursor-pointer group">
              <input
                type="checkbox"
                checked={form.respect_robots}
                onChange={(e) => set('respect_robots', e.target.checked)}
                className="mt-0.5 h-4 w-4 rounded border-gray-700 bg-gray-800 accent-emerald-500"
                disabled={loading}
              />
              <div>
                <div className="text-sm text-gray-200 group-hover:text-white">
                  Respect robots.txt
                </div>
                <div className="text-xs text-gray-500">
                  Keep crawl behavior aligned with robots.txt allow and disallow rules.
                </div>
              </div>
            </label>

            <label className="flex items-start gap-3 cursor-pointer group">
              <input
                type="checkbox"
                checked={form.follow_nofollow_links}
                onChange={(e) => set('follow_nofollow_links', e.target.checked)}
                className="mt-0.5 h-4 w-4 rounded border-gray-700 bg-gray-800 accent-emerald-500"
                disabled={loading}
              />
              <div>
                <div className="text-sm text-gray-200 group-hover:text-white">
                  Follow internal nofollow links
                </div>
                <div className="text-xs text-gray-500">
                  Allow rel=nofollow links to expand the crawl frontier.
                </div>
              </div>
            </label>

            <label className="flex items-start gap-3 cursor-pointer group">
              <input
                type="checkbox"
                checked={form.expand_noindex_pages}
                onChange={(e) => set('expand_noindex_pages', e.target.checked)}
                className="mt-0.5 h-4 w-4 rounded border-gray-700 bg-gray-800 accent-emerald-500"
                disabled={loading}
              />
              <div>
                <div className="text-sm text-gray-200 group-hover:text-white">
                  Expand links on noindex pages
                </div>
                <div className="text-xs text-gray-500">
                  Continue discovering URLs from pages that declare noindex.
                </div>
              </div>
            </label>

            <label className="flex items-start gap-3 cursor-pointer group">
              <input
                type="checkbox"
                checked={form.expand_canonicalized_pages}
                onChange={(e) => set('expand_canonicalized_pages', e.target.checked)}
                className="mt-0.5 h-4 w-4 rounded border-gray-700 bg-gray-800 accent-emerald-500"
                disabled={loading}
              />
              <div>
                <div className="text-sm text-gray-200 group-hover:text-white">
                  Expand links on canonicalized pages
                </div>
                <div className="text-xs text-gray-500">
                  Continue discovering URLs from pages that canonically point elsewhere.
                </div>
              </div>
            </label>

            <label className="flex items-start gap-3 cursor-pointer group">
              <input
                type="checkbox"
                checked={form.validate_external_links}
                onChange={(e) => set('validate_external_links', e.target.checked)}
                className="mt-0.5 h-4 w-4 rounded"
                style={{ accentColor: '#3fe56c' }}
                disabled={loading}
              />
              <div>
                <div className="text-sm text-on-surface">Validate external links</div>
                <div className="text-xs text-on-surface-variant">
                  HEAD-check every outbound link for 4xx/5xx/timeouts. Adds significant crawl time.
                </div>
              </div>
            </label>

            <label className="flex items-start gap-3 cursor-pointer group">
              <input
                type="checkbox"
                checked={form.discover_resources}
                onChange={(e) => set('discover_resources', e.target.checked)}
                className="mt-0.5 h-4 w-4 rounded"
                style={{ accentColor: '#3fe56c' }}
                disabled={loading}
              />
              <div>
                <div className="text-sm text-on-surface">Discover sub-resources (CSS, JS, fonts)</div>
                <div className="text-xs text-on-surface-variant">
                  Fetch and validate every stylesheet, script and font. Detects broken assets. Slow.
                </div>
              </div>
            </label>
          </div>
        </div>
      )}

      {error && (
        <div
          className="text-sm rounded-xl px-4 py-2.5"
          style={{ color: '#ffb4ab', background: 'rgba(147,0,10,0.2)' }}
        >
          {error}
        </div>
      )}

      <button type="submit" className="btn-primary justify-center py-3" disabled={loading}>
        {loading ? (
          <>
            <span
              className="w-4 h-4 rounded-full animate-spin"
              style={{ border: '2px solid rgba(0,57,18,0.4)', borderTopColor: '#003912' }}
            />
            Starting…
          </>
        ) : (
          <>
            <Play size={16} />
            Launch Crawl
          </>
        )}
      </button>
    </form>
  )
}
