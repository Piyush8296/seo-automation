import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Globe2, Settings, ArrowLeft, CheckSquare, Search, ChevronDown, ChevronRight, Database, MapPinned } from 'lucide-react'
import { api } from '../lib/api'

const CAT_ICONS = {
  'AMP': '⚡',
  'Canonical': '🎯',
  'Content': '📄',
  'Core Web Vitals': '🏎',
  'Crawl Budget': '💰',
  'Crawlability': '🕷',
  'E-E-A-T': '⭐',
  'Headings': '🏷',
  'HTTPS & Security': '🔒',
  'Images': '🖼',
  'Internal Linking': '🔗',
  'International': '🌍',
  'Meta Descriptions': '📋',
  'Mobile': '📱',
  'Mobile vs Desktop': '💻',
  'Pagination': '📑',
  'Performance': '⚡',
  'Resources': '📦',
  'Sitemap': '🗺',
  'Social': '📣',
  'Structured Data': '🧩',
  'Titles': '📝',
  'URL Structure': '🌐',
}

function groupByCategory(checks) {
  const map = {}
  for (const c of checks) {
    if (!map[c.category]) map[c.category] = []
    map[c.category].push(c)
  }
  return Object.entries(map).sort(([a], [b]) => a.localeCompare(b))
}

export default function ChecksCatalog() {
  const navigate = useNavigate()
  const [catalog, setCatalog] = useState(null)
  const [search, setSearch] = useState('')
  const [expanded, setExpanded] = useState({})

  useEffect(() => {
    api.getCheckCatalog()
      .then(setCatalog)
      .catch(() => {})
  }, [])

  const toggle = (cat) => setExpanded((prev) => ({ ...prev, [cat]: !prev[cat] }))

  const query = search.trim().toLowerCase()
  const groups = catalog
    ? groupByCategory(
        query
          ? catalog.checks.filter((c) =>
              c.id.includes(query) ||
              c.category.toLowerCase().includes(query) ||
              c.description?.toLowerCase().includes(query)
            )
          : catalog.checks
      )
    : []

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
              <ArrowLeft size={15} />
              <span className="uppercase tracking-widest" style={{ fontSize: '9px' }}>Observatory</span>
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
            <div
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg"
              style={{ background: '#1a202a', borderRight: '2px solid #00c853', color: '#3fe56c' }}
            >
              <CheckSquare size={15} />
              <span className="uppercase tracking-widest font-medium" style={{ fontSize: '9px' }}>Checks Catalog</span>
            </div>
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

        {/* Sidebar stats */}
        <div className="px-4 py-4 flex flex-col gap-3" style={{ borderTop: '1px solid rgba(60,74,60,0.3)' }}>
          <span className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '9px' }}>Catalog Stats</span>
          {catalog ? (
            <>
              {[
                { label: 'Total Check IDs', value: catalog.check_ids },
                { label: 'Page Runners', value: catalog.page_checks },
                { label: 'Site Runners', value: catalog.site_checks },
                { label: 'Categories', value: groups.length || groupByCategory(catalog.checks).length },
              ].map(({ label, value }) => (
                <div key={label}>
                  <div className="uppercase tracking-widest text-on-surface-variant" style={{ fontSize: '8px' }}>{label}</div>
                  <div className="text-sm font-display font-bold text-primary mt-0.5">{value}</div>
                </div>
              ))}
            </>
          ) : (
            <div className="flex justify-center py-4">
              <span className="w-4 h-4 rounded-full animate-spin" style={{ border: '2px solid #2f3540', borderTopColor: '#3fe56c' }} />
            </div>
          )}
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
          <CheckSquare size={16} style={{ color: '#3fe56c' }} />
          <span className="font-display font-semibold text-on-surface text-sm">Checks Catalog</span>
          {catalog && (
            <span
              className="ml-2 px-2.5 py-0.5 rounded-full text-on-surface-variant"
              style={{ background: '#1a202a', fontSize: '10px' }}
            >
              {catalog.check_ids} checks · {groupByCategory(catalog.checks).length} categories
            </span>
          )}
        </header>

        <main className="flex-1 overflow-y-auto px-6 py-6">
          {/* Search bar */}
          <div className="mb-5 relative">
            <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2" style={{ color: '#bbcbb8' }} />
            <input
              className="input pl-9 text-sm"
              placeholder="Filter by check ID or category…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>

          {!catalog ? (
            <div className="flex justify-center py-16">
              <span className="w-8 h-8 rounded-full animate-spin" style={{ border: '2px solid #2f3540', borderTopColor: '#3fe56c' }} />
            </div>
          ) : groups.length === 0 ? (
            <div className="text-center py-16 text-on-surface-variant text-sm">No checks match "{search}"</div>
          ) : (
            <div className="flex flex-col gap-2">
              {groups.map(([cat, checks]) => {
                const isOpen = expanded[cat] !== false
                return (
                  <div
                    key={cat}
                    className="rounded-xl overflow-hidden"
                    style={{ background: '#161c26' }}
                  >
                    {/* Category header */}
                    <button
                      className="w-full flex items-center gap-3 px-4 py-3 hover:bg-surface-bright transition-colors"
                      onClick={() => toggle(cat)}
                    >
                      <span style={{ fontSize: '16px', lineHeight: 1 }}>{CAT_ICONS[cat] ?? '📌'}</span>
                      <span className="font-display font-semibold text-on-surface text-sm flex-1 text-left">{cat}</span>
                      <span
                        className="text-on-surface-variant font-mono px-2 py-0.5 rounded-full"
                        style={{ fontSize: '10px', background: '#2f3540' }}
                      >
                        {checks.length}
                      </span>
                      {isOpen
                        ? <ChevronDown size={13} style={{ color: '#bbcbb8' }} />
                        : <ChevronRight size={13} style={{ color: '#bbcbb8' }} />
                      }
                    </button>

                    {/* Check list */}
                    {isOpen && (
                      <div
                        className="divide-y"
                        style={{ borderTop: '1px solid rgba(60,74,60,0.25)', borderColor: 'rgba(60,74,60,0.18)' }}
                      >
                        {checks.map((c) => (
                          <div
                            key={c.id}
                            className="flex items-start gap-3 px-4 py-2.5 hover:bg-surface-bright transition-colors"
                          >
                            <code
                              className="font-mono shrink-0 rounded px-2 py-0.5 text-primary"
                              style={{ fontSize: '10px', background: 'rgba(63,229,108,0.08)', border: '1px solid rgba(63,229,108,0.15)' }}
                            >
                              {c.id}
                            </code>
                            <span className="text-on-surface-variant leading-snug" style={{ fontSize: '12px' }}>
                              {c.description}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
