import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft, BarChart3, CheckCircle2, CheckSquare, Database, ExternalLink,
  Globe2, KeyRound, MapPinned, RefreshCw, Search, Send, Settings,
} from 'lucide-react'
import { api } from '../lib/api'

const STATUS_STYLES = {
  ready: { label: 'Ready', color: '#3fe56c', bg: 'rgba(63,229,108,0.08)' },
  needs_evidence: { label: 'Evidence', color: '#ffb7ae', bg: 'rgba(255,183,174,0.08)' },
  verify_first: { label: 'Verify first', color: '#9cc7ff', bg: 'rgba(156,199,255,0.08)' },
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

function ProviderCard({ provider, onConnect, connecting }) {
  const connected = provider?.connected
  const configured = provider?.configured
  return (
    <div className="rounded-xl p-5 flex flex-col gap-4" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full shrink-0" style={{ background: connected ? '#3fe56c' : configured ? '#ffb7ae' : '#9cc7ff' }} />
            <h2 className="text-on-surface font-semibold text-sm">{provider.name}</h2>
          </div>
          <p className="text-on-surface-variant mt-2 leading-snug" style={{ fontSize: '12px' }}>
            {provider.connection_message}
          </p>
        </div>
        <a href={provider.docs_url} target="_blank" rel="noreferrer" className="btn-ghost shrink-0" title="Open docs">
          <ExternalLink size={13} />
        </a>
      </div>

      <div className="rounded-lg p-3 font-mono text-on-surface-variant break-all" style={{ background: '#1a202a', fontSize: '11px' }}>
        {provider.oauth_scope}
      </div>

      <button
        type="button"
        onClick={() => onConnect(provider.id)}
        disabled={!configured || connecting}
        className="btn-primary justify-center text-sm"
        title={configured ? `Connect ${provider.name}` : provider.verify_first_message}
      >
        <KeyRound size={14} />
        {connecting ? 'Connecting...' : connected ? 'Reconnect OAuth' : 'Connect OAuth'}
      </button>

      {!configured && (
        <p className="text-on-surface-variant leading-snug" style={{ fontSize: '11px' }}>
          {provider.verify_first_message}
        </p>
      )}
    </div>
  )
}

function CheckRow({ item }) {
  const style = STATUS_STYLES[item.status] ?? STATUS_STYLES.verify_first
  return (
    <div
      className="grid gap-3 px-4 py-3 items-start"
      style={{
        gridTemplateColumns: '92px minmax(220px, 1.4fr) 92px minmax(170px, 1fr) 112px',
        borderTop: '1px solid rgba(60,74,60,0.18)',
      }}
    >
      <code className="font-mono text-primary" style={{ fontSize: '11px' }}>{item.id}</code>
      <div>
        <div className="text-on-surface text-sm leading-snug">{item.name}</div>
        <div className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '11px' }}>{item.notes}</div>
      </div>
      <span className="uppercase tracking-widest" style={{ color: item.provider === 'gsc' ? '#3fe56c' : '#9cc7ff', fontSize: '9px' }}>
        {item.provider}
      </span>
      <div>
        <div className="text-on-surface" style={{ fontSize: '12px' }}>{item.primary_api}</div>
        <div className="text-on-surface-variant mt-1" style={{ fontSize: '11px' }}>{item.automation}</div>
      </div>
      <span
        className="w-fit rounded-full px-2 py-0.5 font-mono"
        style={{ background: style.bg, color: style.color, fontSize: '10px' }}
      >
        {style.label}
      </span>
    </div>
  )
}

function splitSitemaps(value) {
  return value
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
}

export default function SearchIntegrations() {
  const navigate = useNavigate()
  const [workspace, setWorkspace] = useState(null)
  const [settings, setSettings] = useState(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [connecting, setConnecting] = useState('')
  const [notice, setNotice] = useState('')
  const [error, setError] = useState('')
  const [form, setForm] = useState({
    gscProperty: '',
    gscCountry: '',
    gscDevice: 'mobile',
    gscDateRange: 'last_28_days',
    bingSite: '',
    bingSitemaps: '',
  })
  const [actionType, setActionType] = useState('inspect_url')
  const [actionProvider, setActionProvider] = useState('gsc')
  const [actionURL, setActionURL] = useState('')
  const [sitemapURL, setSitemapURL] = useState('')
  const [payloadText, setPayloadText] = useState('{}')
  const [actionResult, setActionResult] = useState(null)
  const [actionError, setActionError] = useState('')
  const [submittingAction, setSubmittingAction] = useState(false)

  const hydrateForm = (cfg) => {
    setForm({
      gscProperty: cfg?.integrations?.gsc?.property_url ?? '',
      gscCountry: cfg?.integrations?.gsc?.country ?? '',
      gscDevice: cfg?.integrations?.gsc?.device ?? 'mobile',
      gscDateRange: cfg?.integrations?.gsc?.date_range ?? 'last_28_days',
      bingSite: cfg?.integrations?.bing_webmaster?.site_url ?? '',
      bingSitemaps: (cfg?.integrations?.bing_webmaster?.sitemap_urls ?? []).join('\n'),
    })
  }

  const refresh = async () => {
    setLoading(true)
    setError('')
    try {
      const [workspaceData, settingsData] = await Promise.all([
        api.getSearchIntegrations(),
        api.getSettings(),
      ])
      setWorkspace(workspaceData)
      setSettings(settingsData)
      hydrateForm(settingsData)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
  }, [])

  const providerByID = useMemo(() => {
    const out = {}
    for (const provider of workspace?.providers ?? []) {
      out[provider.id] = provider
    }
    return out
  }, [workspace])

  const connectedProviderIDs = useMemo(() => {
    return new Set((workspace?.providers ?? []).filter((provider) => provider.connected).map((provider) => provider.id))
  }, [workspace])

  const canSubmitAction = connectedProviderIDs.has(actionProvider)

  const updateForm = (key, value) => {
    setNotice('')
    setError('')
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const saveConfig = async () => {
    setSaving(true)
    setNotice('')
    setError('')
    try {
      const next = {
        ...(settings ?? {}),
        feature_flags: {
          ...(settings?.feature_flags ?? {}),
          gsc: Boolean(form.gscProperty.trim()),
          bing_webmaster: Boolean(form.bingSite.trim()),
        },
        integrations: {
          ...(settings?.integrations ?? {}),
          gsc: {
            ...(settings?.integrations?.gsc ?? {}),
            property_url: form.gscProperty.trim(),
            country: form.gscCountry.trim(),
            device: form.gscDevice,
            date_range: form.gscDateRange.trim() || 'last_28_days',
          },
          bing_webmaster: {
            ...(settings?.integrations?.bing_webmaster ?? {}),
            site_url: form.bingSite.trim(),
            sitemap_urls: splitSitemaps(form.bingSitemaps),
          },
        },
      }
      const saved = await api.updateSettings(next)
      setSettings(saved)
      hydrateForm(saved)
      const workspaceData = await api.getSearchIntegrations()
      setWorkspace(workspaceData)
      setNotice('Provider settings saved.')
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  const connectOAuth = async (provider) => {
    setConnecting(provider)
    setNotice('')
    setError('')
    try {
      const res = await api.connectSearchOAuth({ provider })
      setNotice(res.message)
      await refresh()
    } catch (err) {
      setError(err.message)
    } finally {
      setConnecting('')
    }
  }

  const chooseAction = (nextType) => {
    setActionType(nextType)
    setActionResult(null)
    setActionError('')
    if (nextType === 'inspect_url') {
      setActionProvider('gsc')
    }
  }

  const submitAction = async (e) => {
    e.preventDefault()
    setActionError('')
    setActionResult(null)

    let payload = {}
    try {
      payload = payloadText.trim() ? JSON.parse(payloadText) : {}
    } catch {
      setActionError('Payload must be valid JSON.')
      return
    }
    if (Array.isArray(payload) || payload === null || typeof payload !== 'object') {
      setActionError('Payload must be a JSON object.')
      return
    }

    setSubmittingAction(true)
    try {
      const res = await api.submitSearchIntegrationAction({
        type: actionType,
        provider: actionProvider,
        url: actionURL,
        sitemap_url: sitemapURL,
        payload,
      })
      setActionResult(res)
    } catch (err) {
      setActionError(err.message)
    } finally {
      setSubmittingAction(false)
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
            <div
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg"
              style={{ background: '#1a202a', borderRight: '2px solid #00c853', color: '#3fe56c' }}
            >
              <Search size={15} />
              <span className="uppercase tracking-widest font-medium" style={{ fontSize: '9px' }}>GSC + Bing</span>
            </div>
          </nav>
        </div>

        <div className="px-4 py-4 flex flex-col gap-3" style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}>
          <span className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>Report Gate</span>
          <div className="rounded-lg p-3" style={{ background: '#1a202a', border: '1px solid rgba(60,74,60,0.35)' }}>
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full" style={{ background: workspace?.summary?.report_ready ? '#3fe56c' : '#ffb7ae' }} />
              <span className="uppercase tracking-widest text-on-surface" style={{ fontSize: '9px' }}>
                {workspace?.summary?.report_ready ? 'Connected' : 'Verify first'}
              </span>
            </div>
            <div className="text-on-surface-variant mt-2 leading-snug" style={{ fontSize: '11px' }}>
              {workspace?.summary?.report_ready ? 'At least one provider is connected.' : 'Connect OAuth after verification.'}
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
          <Search size={16} style={{ color: '#3fe56c' }} />
          <span className="font-display font-semibold text-on-surface text-sm">GSC + Bing Pack</span>
          <span className="ml-1 px-2.5 py-0.5 rounded-full text-on-surface-variant" style={{ background: '#1a202a', fontSize: '10px' }}>
            {workspace?.summary?.total_checks ?? 0} checks
          </span>
          <div className="flex-1" />
          <button onClick={refresh} className="btn-ghost" title="Refresh">
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
                <div className="grid grid-cols-4 gap-3">
                  <StatusCard label="Total Checks" value={workspace?.summary?.total_checks ?? 0} tone="#3fe56c" />
                  <StatusCard label="GSC" value={workspace?.summary?.gsc_checks ?? 0} tone="#3fe56c" />
                  <StatusCard label="Bing" value={workspace?.summary?.bing_checks ?? 0} tone="#9cc7ff" />
                  <StatusCard label="POST Ready" value={workspace?.summary?.write_capable_checks ?? 0} tone="#ffb7ae" />
                </div>

                {(notice || error) && (
                  <div
                    className="rounded-lg px-4 py-3 text-sm"
                    style={{
                      background: error ? 'rgba(147,0,10,0.2)' : 'rgba(63,229,108,0.08)',
                      border: error ? '1px solid rgba(255,180,171,0.18)' : '1px solid rgba(63,229,108,0.18)',
                      color: error ? '#ffb4ab' : '#3fe56c',
                    }}
                  >
                    {error || notice}
                  </div>
                )}

                <div className="grid grid-cols-2 gap-4">
                  {(workspace?.providers ?? []).map((provider) => (
                    <ProviderCard
                      key={provider.id}
                      provider={provider}
                      onConnect={connectOAuth}
                      connecting={connecting === provider.id}
                    />
                  ))}
                </div>

                <section className="rounded-xl overflow-hidden" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-center gap-3 px-4 py-3">
                    <BarChart3 size={15} style={{ color: '#3fe56c' }} />
                    <h2 className="text-on-surface font-semibold text-sm">
                      {workspace?.summary?.report_ready ? 'Report' : 'Verification Required'}
                    </h2>
                  </div>

                  {workspace?.report ? (
                    <>
                      <div className="px-4 pb-3 text-on-surface-variant leading-snug" style={{ fontSize: '12px' }}>
                        {workspace.report.message}
                      </div>
                      <div
                        className="grid gap-3 px-4 py-2 text-on-surface-variant uppercase tracking-widest"
                        style={{
                          gridTemplateColumns: '92px minmax(220px, 1.4fr) 92px minmax(170px, 1fr) 112px',
                          fontSize: '9px',
                          borderTop: '1px solid rgba(60,74,60,0.24)',
                        }}
                      >
                        <span>ID</span>
                        <span>Check</span>
                        <span>Provider</span>
                        <span>API</span>
                        <span>Status</span>
                      </div>
                      {workspace.report.items.map((item) => <CheckRow key={item.id} item={item} />)}
                    </>
                  ) : (
                    <div className="px-5 pb-5">
                      <div className="rounded-lg p-4 flex items-start gap-3" style={{ background: '#1a202a' }}>
                        <CheckCircle2 size={16} style={{ color: '#9cc7ff' }} />
                        <div>
                          <p className="text-on-surface text-sm">Verify the provider first.</p>
                          <p className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '12px' }}>
                            {workspace?.verification_message}
                          </p>
                        </div>
                      </div>
                    </div>
                  )}
                </section>
              </section>

              <aside className="flex flex-col gap-5 min-w-0">
                <section className="rounded-xl p-5 flex flex-col gap-4" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-start gap-3">
                    <KeyRound size={16} style={{ color: '#3fe56c' }} />
                    <div>
                      <h2 className="text-on-surface font-semibold text-sm">Provider Setup</h2>
                      <p className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '12px' }}>
                        Store only property references here. OAuth tokens stay outside this settings payload.
                      </p>
                    </div>
                  </div>

                  <div>
                    <label className="label">GSC Property URL</label>
                    <input
                      className="input text-sm"
                      value={form.gscProperty}
                      onChange={(e) => updateForm('gscProperty', e.target.value)}
                      placeholder="sc-domain:cars24.com"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="label">Country</label>
                      <input
                        className="input text-sm"
                        value={form.gscCountry}
                        onChange={(e) => updateForm('gscCountry', e.target.value)}
                        placeholder="ind"
                      />
                    </div>
                    <div>
                      <label className="label">Device</label>
                      <select className="input text-sm" value={form.gscDevice} onChange={(e) => updateForm('gscDevice', e.target.value)}>
                        <option value="mobile">mobile</option>
                        <option value="desktop">desktop</option>
                        <option value="tablet">tablet</option>
                      </select>
                    </div>
                  </div>
                  <div>
                    <label className="label">GSC Date Range</label>
                    <input
                      className="input text-sm"
                      value={form.gscDateRange}
                      onChange={(e) => updateForm('gscDateRange', e.target.value)}
                      placeholder="last_28_days"
                    />
                  </div>
                  <div>
                    <label className="label">Bing Site URL</label>
                    <input
                      className="input text-sm"
                      value={form.bingSite}
                      onChange={(e) => updateForm('bingSite', e.target.value)}
                      placeholder="https://www.cars24.com/"
                    />
                  </div>
                  <div>
                    <label className="label">Bing Sitemap URLs</label>
                    <textarea
                      className="input text-sm min-h-[92px]"
                      value={form.bingSitemaps}
                      onChange={(e) => updateForm('bingSitemaps', e.target.value)}
                      placeholder="https://www.cars24.com/sitemap.xml"
                    />
                  </div>
                  <button type="button" className="btn-primary justify-center text-sm" onClick={saveConfig} disabled={saving}>
                    {saving ? 'Saving...' : 'Save Setup'}
                  </button>
                </section>

                <section className="rounded-xl p-5 flex flex-col gap-4" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-start gap-3">
                    <Send size={16} style={{ color: '#3fe56c' }} />
                    <div>
                      <h2 className="text-on-surface font-semibold text-sm">POST Operation</h2>
                      <p className="text-on-surface-variant mt-1" style={{ fontSize: '12px' }}>
                        Placeholder endpoint for URL inspection and sitemap submission.
                      </p>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    <button
                      type="button"
                      onClick={() => chooseAction('inspect_url')}
                      className="rounded-lg px-3 py-2 text-sm font-medium transition-colors"
                      style={actionType === 'inspect_url' ? { background: '#3fe56c', color: '#003912' } : { background: '#1a202a', color: '#bbcbb8' }}
                    >
                      Inspect URL
                    </button>
                    <button
                      type="button"
                      onClick={() => chooseAction('submit_sitemap')}
                      className="rounded-lg px-3 py-2 text-sm font-medium transition-colors"
                      style={actionType === 'submit_sitemap' ? { background: '#3fe56c', color: '#003912' } : { background: '#1a202a', color: '#bbcbb8' }}
                    >
                      Submit Sitemap
                    </button>
                  </div>

                  <form onSubmit={submitAction} className="flex flex-col gap-3">
                    <div>
                      <label className="label">Provider</label>
                      <select
                        className="input text-sm"
                        value={actionProvider}
                        onChange={(e) => setActionProvider(e.target.value)}
                        disabled={actionType === 'inspect_url'}
                      >
                        <option value="gsc">GSC</option>
                        <option value="bing">Bing</option>
                      </select>
                    </div>

                    {actionType === 'inspect_url' ? (
                      <div>
                        <label className="label">URL</label>
                        <input
                          className="input text-sm"
                          value={actionURL}
                          onChange={(e) => setActionURL(e.target.value)}
                          placeholder="https://www.cars24.com/"
                          required
                        />
                      </div>
                    ) : (
                      <div>
                        <label className="label">Sitemap URL</label>
                        <input
                          className="input text-sm"
                          value={sitemapURL}
                          onChange={(e) => setSitemapURL(e.target.value)}
                          placeholder={providerByID.bing?.configured_sitemaps?.[0] ?? 'https://www.cars24.com/sitemap.xml'}
                          required
                        />
                      </div>
                    )}

                    <div>
                      <label className="label">Payload JSON</label>
                      <textarea
                        className="input font-mono min-h-[104px]"
                        style={{ fontSize: '11px' }}
                        value={payloadText}
                        onChange={(e) => setPayloadText(e.target.value)}
                        spellCheck={false}
                      />
                    </div>

                    {!canSubmitAction && (
                      <div className="rounded-lg px-3 py-2 text-on-surface-variant" style={{ background: '#1a202a', fontSize: '12px' }}>
                        Connect {actionProvider === 'gsc' ? 'GSC' : 'Bing'} OAuth before running this operation.
                      </div>
                    )}
                    {actionError && (
                      <div className="text-sm rounded-lg px-3 py-2" style={{ color: '#ffb4ab', background: 'rgba(147,0,10,0.2)' }}>
                        {actionError}
                      </div>
                    )}
                    {actionResult && (
                      <div className="rounded-lg px-3 py-2" style={{ background: 'rgba(63,229,108,0.08)', border: '1px solid rgba(63,229,108,0.18)' }}>
                        <div className="text-primary font-medium text-sm">{actionResult.status}</div>
                        <div className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '12px' }}>{actionResult.message}</div>
                      </div>
                    )}

                    <button type="submit" className="btn-primary justify-center" disabled={submittingAction || !canSubmitAction}>
                      {submittingAction
                        ? <span className="w-4 h-4 rounded-full animate-spin" style={{ border: '2px solid rgba(0,57,18,0.4)', borderTopColor: '#003912' }} />
                        : <Send size={14} />}
                      {submittingAction ? 'Submitting...' : 'Submit'}
                    </button>
                  </form>
                </section>
              </aside>
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
