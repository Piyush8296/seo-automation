import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft, Building2, CheckSquare, ClipboardList, Database, FileText,
  Globe2, MapPinned, RefreshCw, Search, Send, Settings, Upload,
} from 'lucide-react'
import { api } from '../lib/api'

const PRIORITY_COLORS = {
  Critical: '#ffb4ab',
  High: '#ffb7ae',
  Medium: '#8ed793',
}

const CHANNEL_COLORS = {
  GBP: '#3fe56c',
  Website: '#8ed793',
  Listings: '#ffb7ae',
  Monitoring: '#9cc7ff',
  Manual: '#bbcbb8',
  'Off-page': '#dde2f1',
}

function StatusCard({ label, value, tone = '#dde2f1' }) {
  return (
    <div className="rounded-lg p-4" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
      <div className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>{label}</div>
      <div className="font-display font-semibold mt-1" style={{ color: tone, fontSize: '24px' }}>{value}</div>
    </div>
  )
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

function CheckRow({ check }) {
  const priorityColor = PRIORITY_COLORS[check.priority] ?? '#bbcbb8'
  const channelColor = CHANNEL_COLORS[check.channel] ?? '#bbcbb8'
  return (
    <div
      className="grid gap-3 px-4 py-3 items-start"
      style={{
        gridTemplateColumns: '86px minmax(220px, 1.5fr) 108px minmax(190px, 1fr) minmax(180px, 1fr)',
        borderTop: '1px solid rgba(60,74,60,0.18)',
      }}
    >
      <code className="font-mono text-primary" style={{ fontSize: '11px' }}>{check.id}</code>
      <div>
        <div className="text-on-surface text-sm leading-snug">{check.name}</div>
        {check.notes && (
          <div className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '11px' }}>
            {check.notes}
          </div>
        )}
      </div>
      <span
        className="w-fit rounded-full px-2 py-0.5 font-mono"
        style={{ background: 'rgba(255,255,255,0.05)', color: priorityColor, fontSize: '10px' }}
      >
        {check.priority}
      </span>
      <div>
        <span className="font-medium" style={{ color: channelColor, fontSize: '12px' }}>{check.channel}</span>
        <div className="text-on-surface-variant mt-1" style={{ fontSize: '11px' }}>{check.primary_source}</div>
      </div>
      <div>
        <div className="text-on-surface" style={{ fontSize: '12px' }}>{check.automation}</div>
        {check.operation_types?.length > 0 && (
          <div className="flex gap-1.5 mt-1">
            {check.operation_types.map((op) => (
              <span
                key={op}
                className="uppercase tracking-widest rounded px-1.5 py-0.5"
                style={{ background: 'rgba(63,229,108,0.08)', color: '#3fe56c', fontSize: '8px' }}
              >
                {op}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default function LocalSEO() {
  const navigate = useNavigate()
  const [workspace, setWorkspace] = useState(null)
  const [loading, setLoading] = useState(true)
  const [operation, setOperation] = useState('update')
  const [locationID, setLocationID] = useState('')
  const [title, setTitle] = useState('')
  const [summary, setSummary] = useState('')
  const [payloadText, setPayloadText] = useState('{\n  "primary_phone": "+91-00000-00000",\n  "primary_category": "Used car dealer"\n}')
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState('')
  const [result, setResult] = useState(null)

  const fetchWorkspace = async () => {
    setLoading(true)
    try {
      const data = await api.getLocalSEO()
      setWorkspace(data)
      if (!locationID && data?.status?.configured_locations?.[0]) {
        setLocationID(data.status.configured_locations[0])
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchWorkspace().catch(() => setLoading(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const visibleChecks = useMemo(() => workspace?.checks ?? [], [workspace])

  const setOp = (next) => {
    setOperation(next)
    setResult(null)
    setFormError('')
    if (next === 'post') {
      setPayloadText('{\n  "topic_type": "STANDARD",\n  "cta_url": "https://www.cars24.com/"\n}')
    } else {
      setPayloadText('{\n  "primary_phone": "+91-00000-00000",\n  "primary_category": "Used car dealer"\n}')
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setFormError('')
    setResult(null)

    let payload = {}
    try {
      payload = payloadText.trim() ? JSON.parse(payloadText) : {}
    } catch {
      setFormError('Payload must be valid JSON.')
      return
    }
    if (Array.isArray(payload) || payload === null || typeof payload !== 'object') {
      setFormError('Payload must be a JSON object.')
      return
    }

    setSubmitting(true)
    try {
      const res = await api.submitGBPAction({
        type: operation,
        location_id: locationID,
        title,
        summary,
        payload,
      })
      setResult(res)
    } catch (err) {
      setFormError(err.message)
    } finally {
      setSubmitting(false)
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
            <NavButton icon={Search} label="GSC + Bing" onClick={() => navigate('/search-integrations')} />
            <div
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg"
              style={{ background: '#1a202a', borderRight: '2px solid #00c853', color: '#3fe56c' }}
            >
              <MapPinned size={15} />
              <span className="uppercase tracking-widest font-medium" style={{ fontSize: '9px' }}>Local SEO</span>
            </div>
          </nav>
        </div>

        <div className="px-4 py-4 flex flex-col gap-3" style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}>
          <span className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>GBP API</span>
          <div className="rounded-lg p-3" style={{ background: '#1a202a', border: '1px solid rgba(60,74,60,0.35)' }}>
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full" style={{ background: workspace?.status?.configured ? '#3fe56c' : '#ffb7ae' }} />
              <span className="uppercase tracking-widest text-on-surface" style={{ fontSize: '9px' }}>
                {workspace?.status?.mode ?? 'placeholder'}
              </span>
            </div>
            <div className="text-on-surface-variant mt-2 leading-snug" style={{ fontSize: '11px' }}>
              {workspace?.status?.configured ? 'Settings present' : 'Awaiting OAuth and approved quota'}
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
          <MapPinned size={16} style={{ color: '#3fe56c' }} />
          <span className="font-display font-semibold text-on-surface text-sm">Local SEO / GBP</span>
          <span
            className="ml-1 px-2.5 py-0.5 rounded-full text-on-surface-variant"
            style={{ background: '#1a202a', fontSize: '10px' }}
          >
            {workspace?.summary?.total_checks ?? 0} sheet checks
          </span>
          <div className="flex-1" />
          <button onClick={fetchWorkspace} className="btn-ghost" title="Refresh">
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
            <div className="grid gap-6" style={{ gridTemplateColumns: 'minmax(0, 1fr) 380px' }}>
              <section className="flex flex-col gap-5 min-w-0">
                <div className="grid grid-cols-4 gap-3">
                  <StatusCard label="Total Checks" value={workspace?.summary?.total_checks ?? 0} tone="#3fe56c" />
                  <StatusCard label="GBP Direct" value={workspace?.summary?.gbp_direct_checks ?? 0} tone="#3fe56c" />
                  <StatusCard label="Website" value={workspace?.summary?.website_checks ?? 0} tone="#8ed793" />
                  <StatusCard label="Vendor / Manual" value={(workspace?.summary?.vendor_workflow_checks ?? 0) + (workspace?.summary?.manual_only_checks ?? 0)} tone="#ffb7ae" />
                </div>

                <div className="rounded-xl overflow-hidden" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-center gap-3 px-4 py-3">
                    <ClipboardList size={15} style={{ color: '#3fe56c' }} />
                    <h2 className="text-on-surface font-semibold text-sm">Sheet Checks</h2>
                  </div>
                  <div
                    className="grid gap-3 px-4 py-2 text-on-surface-variant uppercase tracking-widest"
                    style={{
                      gridTemplateColumns: '86px minmax(220px, 1.5fr) 108px minmax(190px, 1fr) minmax(180px, 1fr)',
                      fontSize: '9px',
                      borderTop: '1px solid rgba(60,74,60,0.24)',
                    }}
                  >
                    <span>ID</span>
                    <span>Check</span>
                    <span>Priority</span>
                    <span>Source</span>
                    <span>Automation</span>
                  </div>
                  {visibleChecks.map((check) => <CheckRow key={check.id} check={check} />)}
                </div>
              </section>

              <aside className="flex flex-col gap-5 min-w-0">
                <section className="rounded-xl p-5 flex flex-col gap-4" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-start gap-3">
                    <Building2 size={16} style={{ color: '#3fe56c' }} />
                    <div>
                      <h2 className="text-on-surface font-semibold text-sm">Connection</h2>
                      <p className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '12px' }}>
                        {workspace?.status?.message}
                      </p>
                    </div>
                  </div>
                  <div className="rounded-lg p-3 font-mono text-on-surface-variant break-all" style={{ background: '#1a202a', fontSize: '11px' }}>
                    {workspace?.status?.oauth_scope}
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <StatusCard label="Accounts" value={workspace?.status?.configured_accounts?.length ?? 0} />
                    <StatusCard label="Locations" value={workspace?.status?.configured_locations?.length ?? 0} />
                  </div>
                </section>

                <section className="rounded-xl p-5 flex flex-col gap-4" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-start gap-3">
                    {operation === 'post'
                      ? <Send size={16} style={{ color: '#3fe56c' }} />
                      : <Upload size={16} style={{ color: '#3fe56c' }} />}
                    <div>
                      <h2 className="text-on-surface font-semibold text-sm">GBP Operation</h2>
                      <p className="text-on-surface-variant mt-1" style={{ fontSize: '12px' }}>
                        Placeholder endpoint: no live GBP write is performed.
                      </p>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    <button
                      type="button"
                      onClick={() => setOp('update')}
                      className="rounded-lg px-3 py-2 text-sm font-medium transition-colors"
                      style={operation === 'update' ? { background: '#3fe56c', color: '#003912' } : { background: '#1a202a', color: '#bbcbb8' }}
                    >
                      UPDATE
                    </button>
                    <button
                      type="button"
                      onClick={() => setOp('post')}
                      className="rounded-lg px-3 py-2 text-sm font-medium transition-colors"
                      style={operation === 'post' ? { background: '#3fe56c', color: '#003912' } : { background: '#1a202a', color: '#bbcbb8' }}
                    >
                      POST
                    </button>
                  </div>

                  <form onSubmit={handleSubmit} className="flex flex-col gap-3">
                    <div>
                      <label className="label">GBP Location ID</label>
                      <input
                        className="input text-sm"
                        value={locationID}
                        onChange={(e) => setLocationID(e.target.value)}
                        placeholder="locations/123456789"
                        required
                      />
                    </div>

                    {operation === 'post' && (
                      <>
                        <div>
                          <label className="label">Post Title</label>
                          <input
                            className="input text-sm"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            placeholder="Weekend used-car inspection camp"
                          />
                        </div>
                        <div>
                          <label className="label">Post Summary</label>
                          <textarea
                            className="input text-sm min-h-[96px]"
                            value={summary}
                            onChange={(e) => setSummary(e.target.value)}
                            placeholder="Branch-specific GBP post content"
                            required
                          />
                        </div>
                      </>
                    )}

                    <div>
                      <label className="label">Payload JSON</label>
                      <textarea
                        className="input font-mono min-h-[148px]"
                        style={{ fontSize: '11px' }}
                        value={payloadText}
                        onChange={(e) => setPayloadText(e.target.value)}
                        spellCheck={false}
                      />
                    </div>

                    {formError && (
                      <div className="text-sm rounded-lg px-3 py-2" style={{ color: '#ffb4ab', background: 'rgba(147,0,10,0.2)' }}>
                        {formError}
                      </div>
                    )}
                    {result && (
                      <div className="rounded-lg px-3 py-2" style={{ background: 'rgba(63,229,108,0.08)', border: '1px solid rgba(63,229,108,0.18)' }}>
                        <div className="text-primary font-medium text-sm">{result.status}</div>
                        <div className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '12px' }}>{result.message}</div>
                      </div>
                    )}

                    <button type="submit" className="btn-primary justify-center" disabled={submitting}>
                      {submitting
                        ? <span className="w-4 h-4 rounded-full animate-spin" style={{ border: '2px solid rgba(0,57,18,0.4)', borderTopColor: '#003912' }} />
                        : operation === 'post' ? <Send size={14} /> : <Upload size={14} />}
                      {submitting ? 'Submitting...' : `Submit ${operation.toUpperCase()}`}
                    </button>
                  </form>
                </section>

                <section className="rounded-xl p-5 flex flex-col gap-3" style={{ background: '#161c26', border: '1px solid rgba(60,74,60,0.24)' }}>
                  <div className="flex items-center gap-2">
                    <FileText size={14} style={{ color: '#3fe56c' }} />
                    <h2 className="text-on-surface font-semibold text-sm">Action Scope</h2>
                  </div>
                  {(workspace?.actions ?? []).map((action) => (
                    <div key={action.type} className="rounded-lg p-3" style={{ background: '#1a202a' }}>
                      <div className="uppercase tracking-widest text-primary" style={{ fontSize: '9px' }}>{action.type}</div>
                      <div className="text-on-surface text-sm mt-1">{action.label}</div>
                      <div className="text-on-surface-variant mt-1 leading-snug" style={{ fontSize: '11px' }}>{action.description}</div>
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
