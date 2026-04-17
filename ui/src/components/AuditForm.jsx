import { useState } from 'react'
import { ChevronDown, ChevronUp, Play } from 'lucide-react'

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

export default function AuditForm({ onSubmit, loading }) {
  const [form, setForm] = useState(DEFAULTS)
  const [advanced, setAdvanced] = useState(false)
  const [error, setError] = useState('')

  const set = (k, v) => setForm((f) => ({ ...f, [k]: v }))

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    const url = form.url.trim()
    if (!url) { setError('URL is required'); return }
    if (!/^https?:\/\//i.test(url)) { setError('URL must start with http:// or https://'); return }
    try {
      await onSubmit({
        ...form,
        url,
        max_depth:   Number(form.max_depth),
        max_pages:   Number(form.max_pages),
        concurrency: Number(form.concurrency),
      })
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="card p-6 flex flex-col gap-5">
      <div>
        <h2 className="text-lg font-bold text-gray-100 mb-1">New Audit</h2>
        <p className="text-sm text-gray-500">Enter a URL to start a full technical SEO crawl</p>
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
        className="flex items-center gap-2 text-sm text-gray-400 hover:text-gray-200 transition-colors w-fit"
      >
        {advanced ? <ChevronUp size={15} /> : <ChevronDown size={15} />}
        Advanced configuration
      </button>

      {advanced && (
        <div className="grid grid-cols-2 gap-4 pt-1">
          {/* Max Depth */}
          <div>
            <label className="label">Max depth</label>
            <select
              value={form.max_depth}
              onChange={(e) => set('max_depth', e.target.value)}
              className="input"
              disabled={loading}
            >
              <option value={-1}>Unlimited</option>
              <option value={1}>1 (homepage only)</option>
              <option value={2}>2 levels</option>
              <option value={3}>3 levels</option>
              <option value={5}>5 levels</option>
            </select>
          </div>

          {/* Max Pages */}
          <div>
            <label className="label">Max pages</label>
            <select
              value={form.max_pages}
              onChange={(e) => set('max_pages', e.target.value)}
              className="input"
              disabled={loading}
            >
              <option value={0}>Unlimited</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
              <option value={500}>500</option>
              <option value={1000}>1,000</option>
            </select>
          </div>

          {/* Concurrency */}
          <div>
            <label className="label">Concurrency — {form.concurrency} workers</label>
            <input
              type="range"
              min={1} max={20} step={1}
              value={form.concurrency}
              onChange={(e) => set('concurrency', e.target.value)}
              className="w-full accent-emerald-500 mt-1"
              disabled={loading}
            />
            <div className="flex justify-between text-xs text-gray-600 mt-1">
              <span>1 (gentle)</span>
              <span>20 (fast)</span>
            </div>
          </div>

          {/* Timeout */}
          <div>
            <label className="label">Request timeout</label>
            <select
              value={form.timeout}
              onChange={(e) => set('timeout', e.target.value)}
              className="input"
              disabled={loading}
            >
              <option value="10s">10 seconds</option>
              <option value="30s">30 seconds</option>
              <option value="1m">1 minute</option>
              <option value="2m">2 minutes</option>
            </select>
          </div>

          {/* Platform */}
          <div>
            <label className="label">Platform</label>
            <select
              value={form.platform}
              onChange={(e) => set('platform', e.target.value)}
              className="input"
              disabled={loading}
            >
              <option value="">Both (bifurcated)</option>
              <option value="desktop">Desktop only</option>
              <option value="mobile">Mobile focus</option>
            </select>
          </div>

          {/* Output dir */}
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

          {/* Opt-in slow checks */}
          <div className="col-span-2 border-t border-gray-800 pt-4 mt-1 flex flex-col gap-3">
            <label className="flex items-start gap-3 cursor-pointer group">
              <input
                type="checkbox"
                checked={form.validate_external_links}
                onChange={(e) => set('validate_external_links', e.target.checked)}
                className="mt-0.5 h-4 w-4 rounded border-gray-700 bg-gray-800 accent-emerald-500"
                disabled={loading}
              />
              <div>
                <div className="text-sm text-gray-200 group-hover:text-white">
                  Validate external links
                </div>
                <div className="text-xs text-gray-500">
                  HEAD-check every outbound link for 4xx/5xx/timeouts. Adds significant crawl time.
                </div>
              </div>
            </label>

            <label className="flex items-start gap-3 cursor-pointer group">
              <input
                type="checkbox"
                checked={form.discover_resources}
                onChange={(e) => set('discover_resources', e.target.checked)}
                className="mt-0.5 h-4 w-4 rounded border-gray-700 bg-gray-800 accent-emerald-500"
                disabled={loading}
              />
              <div>
                <div className="text-sm text-gray-200 group-hover:text-white">
                  Discover sub-resources (CSS, JS, fonts)
                </div>
                <div className="text-xs text-gray-500">
                  Fetch and validate every stylesheet, script and font. Detects broken assets and `font-display` issues. Slow.
                </div>
              </div>
            </label>
          </div>
        </div>
      )}

      {error && (
        <div className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg px-4 py-2.5">
          {error}
        </div>
      )}

      <button type="submit" className="btn-primary justify-center py-3" disabled={loading}>
        {loading ? (
          <>
            <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            Starting…
          </>
        ) : (
          <>
            <Play size={16} />
            Start Crawl
          </>
        )}
      </button>
    </form>
  )
}
