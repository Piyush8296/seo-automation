import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Globe2, Settings as SettingsIcon, CheckSquare, Database, X, Plus, RotateCcw, KeyRound, Flag, MapPinned, Search } from 'lucide-react'
import { api } from '../lib/api'

const DEFAULT_HOSTS = [
  'linkedin.com', 'www.linkedin.com',
  'twitter.com', 'www.twitter.com',
  'x.com', 'www.x.com',
  'instagram.com', 'www.instagram.com',
  'facebook.com', 'www.facebook.com',
  'tiktok.com', 'www.tiktok.com',
]

export default function Settings() {
  const navigate = useNavigate()
  const [settings, setSettings] = useState(null)
  const [hosts, setHosts] = useState([])
  const [externalCatalog, setExternalCatalog] = useState(null)
  const [input, setInput] = useState('')
  const [inputError, setInputError] = useState('')
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const inputRef = useRef(null)

  useEffect(() => {
    api.getSettings()
      .then((s) => {
        setSettings(s)
        setHosts(s?.skip_link_hosts ?? DEFAULT_HOSTS)
      })
      .catch(() => setHosts(DEFAULT_HOSTS))
    api.getExternalCheckCatalog()
      .then(setExternalCatalog)
      .catch(() => setExternalCatalog(null))
  }, [])

  const addHost = () => {
    const h = input.trim().toLowerCase().replace(/^https?:\/\//, '').replace(/\/.*$/, '')
    if (!h) return
    if (hosts.includes(h)) {
      setInputError('Already in the list')
      return
    }
    setInputError('')
    setHosts((prev) => [...prev, h])
    setInput('')
  }

  const removeHost = (h) => setHosts((prev) => prev.filter((x) => x !== h))

  const handleKeyDown = (e) => {
    if (e.key === 'Enter') { e.preventDefault(); addHost() }
  }

  const handleSave = async () => {
    setSaving(true)
    setSaved(false)
    try {
      const nextSettings = { ...(settings ?? {}), skip_link_hosts: hosts }
      const savedSettings = await api.updateSettings(nextSettings)
      setSettings(savedSettings)
      setSaved(true)
      setTimeout(() => setSaved(false), 2500)
    } finally {
      setSaving(false)
    }
  }

  const handleReset = () => {
    setHosts(DEFAULT_HOSTS)
    setInput('')
    setInputError('')
  }

  return (
    <div className="h-screen bg-surface flex overflow-hidden">
      {/* Sidebar */}
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
            <button
              onClick={() => navigate('/')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <Globe2 size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Observatory</span>
            </button>
            <button
              onClick={() => navigate('/vault')}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant hover:bg-surface-bright hover:text-on-surface transition-colors"
            >
              <Database size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Audit Vault</span>
            </button>
            <div
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg"
              style={{ background: '#1a202a', borderRight: '2px solid #00c853', color: '#3fe56c' }}
            >
              <SettingsIcon size={15} />
              <span className="uppercase tracking-widest font-medium" style={{ fontSize: '9px' }}>Settings</span>
            </div>
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
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden">
        {/* Header */}
        <header
          className="h-14 flex items-center gap-3 px-8 shrink-0"
          style={{
            background: 'rgba(14,19,30,0.72)',
            backdropFilter: 'blur(16px)',
            WebkitBackdropFilter: 'blur(16px)',
            borderBottom: '1px solid rgba(221,226,241,0.06)',
          }}
        >
          <SettingsIcon size={15} style={{ color: '#3fe56c' }} />
          <span className="font-display font-semibold text-on-surface text-sm">Settings</span>
        </header>

        <main className="flex-1 overflow-y-auto p-8">
          <div className="max-w-2xl mx-auto flex flex-col gap-6">

            {/* Skip Links section */}
            <div className="rounded-xl p-6 flex flex-col gap-5" style={{ background: '#161c26' }}>
              <div>
                <h2 className="text-on-surface font-semibold text-sm mb-1">Skip Links</h2>
                <p className="text-on-surface-variant" style={{ fontSize: '12px' }}>
                  Domains listed here are skipped during external link validation.
                  Use this to suppress false positives from platforms that block automated requests (e.g. LinkedIn, Twitter).
                </p>
              </div>

              {/* Tag list */}
              <div className="flex flex-wrap gap-2 min-h-[40px]">
                {hosts.map((h) => (
                  <span
                    key={h}
                    className="flex items-center gap-1.5 px-3 py-1 rounded-full font-mono"
                    style={{ background: '#1a202a', border: '1px solid rgba(60,74,60,0.4)', fontSize: '11px', color: '#dde2f1' }}
                  >
                    {h}
                    <button
                      onClick={() => removeHost(h)}
                      className="text-on-surface-variant hover:text-on-surface transition-colors"
                      title="Remove"
                    >
                      <X size={11} />
                    </button>
                  </span>
                ))}
                {hosts.length === 0 && (
                  <span className="text-on-surface-variant italic" style={{ fontSize: '12px' }}>No domains — all external links will be validated.</span>
                )}
              </div>

              {/* Add input */}
              <div className="flex gap-2">
                <div className="flex-1">
                  <input
                    ref={inputRef}
                    type="text"
                    className="input text-sm w-full"
                    placeholder="e.g. example.com or www.example.com"
                    value={input}
                    onChange={(e) => { setInput(e.target.value); setInputError('') }}
                    onKeyDown={handleKeyDown}
                  />
                  {inputError && (
                    <p className="mt-1 text-xs" style={{ color: '#ffb4ab' }}>{inputError}</p>
                  )}
                </div>
                <button
                  onClick={addHost}
                  className="btn-ghost shrink-0"
                  disabled={!input.trim()}
                >
                  <Plus size={14} />
                  Add
                </button>
              </div>

              {/* Actions */}
              <div className="flex items-center gap-3 pt-1" style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}>
                <button
                  onClick={handleSave}
                  disabled={saving}
                  className="btn-primary py-2 text-sm"
                >
                  {saving ? 'Saving…' : saved ? '✓ Saved' : 'Save Changes'}
                </button>
                <button
                  onClick={handleReset}
                  className="btn-ghost text-sm"
                  title="Restore default skip list"
                >
                  <RotateCcw size={13} />
                  Reset to defaults
                </button>
              </div>
            </div>

            {/* External API readiness section */}
            <div className="rounded-xl p-6 flex flex-col gap-5" style={{ background: '#161c26' }}>
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <KeyRound size={14} style={{ color: '#3fe56c' }} />
                  <h2 className="text-on-surface font-semibold text-sm">External API Check Packs</h2>
                </div>
                <p className="text-on-surface-variant" style={{ fontSize: '12px' }}>
                  These non-Screaming-Frog checks run only after the relevant provider flag,
                  property IDs, OAuth/API credentials, and UI inputs are configured. Secrets should live in a
                  secret store or environment variables, not in this settings payload.
                </p>
              </div>

              {externalCatalog ? (
                <>
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                    <div className="rounded-lg p-4" style={{ background: '#1a202a', border: '1px solid rgba(60,74,60,0.35)' }}>
                      <p className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>Mapped Checks</p>
                      <p className="text-on-surface font-display font-semibold mt-1" style={{ fontSize: '24px' }}>
                        {externalCatalog.total_checks ?? 0}
                      </p>
                    </div>
                    <div className="rounded-lg p-4" style={{ background: '#1a202a', border: '1px solid rgba(60,74,60,0.35)' }}>
                      <p className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>Providers</p>
                      <p className="text-on-surface font-display font-semibold mt-1" style={{ fontSize: '24px' }}>
                        {externalCatalog.providers?.length ?? 0}
                      </p>
                    </div>
                    <div className="rounded-lg p-4" style={{ background: '#1a202a', border: '1px solid rgba(60,74,60,0.35)' }}>
                      <p className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>Feature Flags</p>
                      <p className="text-on-surface font-display font-semibold mt-1" style={{ fontSize: '24px' }}>
                        {externalCatalog.feature_flags?.length ?? 0}
                      </p>
                    </div>
                  </div>

                  <div className="flex flex-col gap-3">
                    {externalCatalog.providers?.map((provider) => (
                      <div
                        key={provider.id}
                        className="rounded-lg p-4"
                        style={{ background: '#1a202a', border: '1px solid rgba(60,74,60,0.35)' }}
                      >
                        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-2">
                          <div>
                            <h3 className="text-on-surface font-semibold text-sm">{provider.name}</h3>
                            <p className="text-on-surface-variant mt-1" style={{ fontSize: '11px' }}>
                              {provider.cost_model}
                            </p>
                          </div>
                          <span
                            className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 font-mono"
                            style={{ background: 'rgba(63,229,108,0.08)', color: '#3fe56c', fontSize: '10px' }}
                          >
                            <Flag size={10} />
                            {provider.feature_flag}
                          </span>
                        </div>
                        <div className="mt-3 flex flex-wrap gap-2">
                          <span className="text-on-surface-variant" style={{ fontSize: '11px' }}>
                            {provider.check_ids?.length ?? 0} checks
                          </span>
                          <span className="text-on-surface-variant" style={{ fontSize: '11px' }}>•</span>
                          <span className="text-on-surface-variant" style={{ fontSize: '11px' }}>
                            {provider.inputs?.filter((inputItem) => inputItem.source === 'ui').length ?? 0} UI inputs
                          </span>
                          <span className="text-on-surface-variant" style={{ fontSize: '11px' }}>•</span>
                          <span className="text-on-surface-variant" style={{ fontSize: '11px' }}>
                            {provider.inputs?.filter((inputItem) => inputItem.secret).length ?? 0} secret inputs
                          </span>
                        </div>
                        <p className="text-on-surface-variant mt-2" style={{ fontSize: '11px' }}>
                          Auth: {provider.auth_model}
                        </p>
                      </div>
                    ))}
                  </div>
                </>
              ) : (
                <p className="text-on-surface-variant italic" style={{ fontSize: '12px' }}>
                  External check catalog could not be loaded.
                </p>
              )}
            </div>

          </div>
        </main>
      </div>
    </div>
  )
}
